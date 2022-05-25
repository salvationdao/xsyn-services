package items

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/rpcclient"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/rs/zerolog"

	"github.com/shopspring/decimal"
)


// Purchase attempts to make a purchase for a given user ID and a given
func Purchase(
	log *zerolog.Logger,
	supPrice decimal.Decimal,
	ucmProcess func(*types.NewTransaction) (decimal.Decimal, decimal.Decimal, string, error),
	user *types.User,
	storeItemID types.StoreItemID,
) error {
	storeItem, err := db.StoreItem(uuid.UUID(storeItemID))
	if err != nil {
		return terror.Error(err)
	}

	isLocked := user.CheckUserIsLocked("account")
	if isLocked {
		return terror.Error(fmt.Errorf("user: %s, attempting to purchase item while account is locked", user.ID), "Account is locked, contact support to unlock.")
	}

	if storeItem.AmountSold >= storeItem.AmountAvailable {
		return terror.Error(fmt.Errorf("all sold out"), "This item has sold out.")
	}

	if storeItem.RestrictionGroup == "WHITELIST" || storeItem.RestrictionGroup == "LOOTBOX" {
		return terror.Error(fmt.Errorf("cannot purchase whitelist item or lootbox"), "Item currently not available.")
	}

	if storeItem.Tier == db.TierMega {
		count, err := db.PurchasedItemsByOwnerIDAndTierDEPRECATE(user.ID, db.TierMega)
		if err != nil {
			return terror.Error(err)
		}
		if count >= 2 {
			return terror.Warn(fmt.Errorf("user bought 2 starter mechs"), "You have reached your 2 Mega War Machine limit.")
		}
	}

	template := &rpcclient.TemplateContainer{}
	err = storeItem.Data.Unmarshal(template)
	if err != nil {
		return terror.Error(err)
	}
	txRef := fmt.Sprintf("PURCHASE OF %s | %d", template.Label, time.Now().UnixNano())

	supsAsCents := supPrice.Mul(decimal.New(100, 0))
	priceAsCents := decimal.New(int64(storeItem.UsdCentCost), 0)
	priceAsSupsDecimal := priceAsCents.Div(supsAsCents)
	priceAsSupsBigInt := priceAsSupsDecimal.Mul(decimal.New(1, 18)).BigInt()

	// resultChan := make(chan *passport.TransactionResult, 1)
	trans := &types.NewTransaction{
		To:                   types.XsynTreasuryUserID,
		From:                 types.UserID(uuid.Must(uuid.FromString(user.ID))),
		Amount:               decimal.NewFromBigInt(priceAsSupsBigInt, 0),
		TransactionReference: types.TransactionReference(txRef),
		Description:          "Purchase on Supremacy storefront.",
		Group:                types.TransactionGroupStore,
		SubGroup:             "Purchase",
	}

	nfb, ntb, txID, err := ucmProcess(trans)
	if err != nil {
		return err
	}

	ws.PublishMessage("/user/"+trans.From.String()+"/sups", "USER:SUPS", nfb.String())
	ws.PublishMessage("/user/"+trans.To.String()+"/sups", "USER:SUPS", ntb.String())

	//go bus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
	//go bus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())

	// refund callback
	refund := func(reason string) {
		trans := &types.NewTransaction{
			To:                   types.UserID(uuid.Must(uuid.FromString(user.ID))),
			RelatedTransactionID: null.StringFrom(txID),
			From:                 types.XsynTreasuryUserID,
			Amount:               decimal.NewFromBigInt(priceAsSupsBigInt, 0),
			TransactionReference: types.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
			Group:                types.TransactionGroupStore,
			SubGroup:             "Refund",
		}

		nfb, ntb, _, err := ucmProcess(trans)
		if err != nil {
			log.Err(err).Msg("failed to process refund")
			return
		}

		ws.PublishMessage("/user/"+trans.From.String()+"/sups", "USER:SUPS", nfb.String())
		ws.PublishMessage("/user/"+trans.To.String()+"/sups", "USER:SUPS", ntb.String())
	}

	// let's assign the item.
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		refund(err.Error())
		return err
	}

	defer tx.Rollback()

	_, err = db.PurchasedItemRegister(uuid.Must(uuid.FromString(storeItem.ID)), uuid.Must(uuid.FromString(user.ID)))
	if err != nil {
		refund(err.Error())
		return err
	}

	err = tx.Commit()
	if err != nil {
		refund(err.Error())
		return err
	}

	resp := struct {
		PriceInSUPS string            `json:"price_in_sups"`
		Item        *boiler.StoreItem `json:"item"`
	}{
		PriceInSUPS: priceAsSupsBigInt.String(),
		Item:        storeItem,
	}

	//go bus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", busKey, storeItem.ID)), resp)
	ws.PublishMessage("/user/"+user.ID+"/purchase", types.HubKeyStoreItemSubscribe, resp)
	return nil
}

