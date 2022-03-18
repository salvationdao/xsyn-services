package items

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/rpcclient"
	"time"

	"github.com/ninja-syndicate/hub/ext/messagebus"

	"github.com/rs/zerolog"

	"github.com/jackc/pgx/v4"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gofrs/uuid"

	"github.com/shopspring/decimal"

	"github.com/ninja-software/terror/v2"
)

// Purchase attempts to make a purchase for a given user ID and a given
func Purchase(
	ctx context.Context,
	conn *pgxpool.Pool,
	log *zerolog.Logger,
	bus *messagebus.MessageBus,
	busKey messagebus.BusKey,
	supPrice decimal.Decimal,
	ucmProcess func(*passport.NewTransaction) (*decimal.Decimal, *decimal.Decimal, string, error),
	user passport.User,
	storeItemID passport.StoreItemID,
	externalUrl string,
) error {
	storeItem, err := db.StoreItem(uuid.UUID(storeItemID))
	if err != nil {
		return terror.Error(err)
	}

	if storeItem.AmountSold >= storeItem.AmountAvailable {
		return terror.Error(fmt.Errorf("all sold out"), "This item has sold out.")
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to purchase faction specific items.")
	}

	if user.FactionID.String() != storeItem.FactionID {
		return terror.Error(fmt.Errorf("user is wrong faction"), "You cannot buy items for another faction")
	}

	if storeItem.RestrictionGroup == "WHITELIST" || storeItem.RestrictionGroup == "LOOTBOX" {
		return terror.Error(fmt.Errorf("cannot purchase whitelist item or lootbox"), "Item currently not available.")
	}

	if storeItem.Tier == db.TierMega {
		count, err := db.PurchasedItemsbyOwnerIDAndTier(uuid.UUID(user.ID), db.TierMega)
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
	txRef := fmt.Sprintf("PURCHASE OF %s | %d", template.BlueprintChassis.Label, time.Now().UnixNano())

	supsAsCents := supPrice.Mul(decimal.New(100, 0))
	priceAsCents := decimal.New(int64(storeItem.UsdCentCost), 0)
	priceAsSupsDecimal := priceAsCents.Div(supsAsCents)
	priceAsSupsBigInt := priceAsSupsDecimal.Mul(decimal.New(1, 18)).BigInt()

	// resultChan := make(chan *passport.TransactionResult, 1)
	trans := &passport.NewTransaction{
		To:                   passport.XsynTreasuryUserID,
		From:                 user.ID,
		Amount:               decimal.NewFromBigInt(priceAsSupsBigInt, 0),
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Purchase on Supremacy storefront.",
		Group:                passport.TransactionGroupStore,
		SubGroup:             "Purchase",
	}

	nfb, ntb, _, err := ucmProcess(trans)
	if err != nil {
		return terror.Error(err)
	}

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())

	// refund callback
	refund := func(reason string) {
		trans := &passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               decimal.NewFromBigInt(priceAsSupsBigInt, 0),
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
			Group:                passport.TransactionGroupStore,
			SubGroup:             "Refund",
		}

		nfb, ntb, _, err := ucmProcess(trans)
		if err != nil {
			log.Err(err).Msg("failed to process refund")
			return
		}

		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())

	}

	// let's assign the item.
	tx, err := conn.Begin(ctx)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	_, err = db.PurchasedItemRegister(uuid.Must(uuid.FromString(storeItem.ID)), uuid.UUID(user.ID))
	if err != nil {
		refund(err.Error())
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	resp := struct {
		PriceInSUPS string            `json:"price_in_sups"`
		Item        *boiler.StoreItem `json:"item"`
	}{
		PriceInSUPS: priceAsSupsBigInt.String(),
		Item:        storeItem,
	}

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", busKey, storeItem.ID)), resp)
	return nil
}

// PurchaseLootbox attempts to make a purchase for a given user ID and a given
func PurchaseLootbox(ctx context.Context, conn *pgxpool.Pool, log *zerolog.Logger, bus *messagebus.MessageBus, busKey messagebus.BusKey,
	ucmProcess func(*passport.NewTransaction) (*decimal.Decimal, *decimal.Decimal, string, error), user passport.User, factionID passport.FactionID, externalURL string) (string, error) {

	// get all faction items marked as loot box
	items, err := db.StoreItemsByFactionIDAndRestrictionGroup(uuid.UUID(factionID), db.RestrictionGroupLootbox)
	if err != nil {
		return "", terror.Error(err, "failed to get loot box items")
	}
	itemIDs := []passport.StoreItemID{}
	for _, m := range items {
		for i := 0; i < m.AmountAvailable-m.AmountSold; i++ {
			itemIDs = append(itemIDs, passport.StoreItemID(uuid.Must(uuid.FromString(m.ID))))
		}
	}

	if len(itemIDs) == 0 {
		return "", terror.Error(fmt.Errorf("all sold out"), "This item has sold out.")
	}

	chosenIdx := rand.Intn(len(itemIDs))
	var storeItem *boiler.StoreItem

	for _, m := range items {
		if m.ID == itemIDs[chosenIdx].String() {
			storeItem = m
			continue
		}
	}

	if storeItem == nil {
		return "", terror.Error(fmt.Errorf("store item nil"), "Internal error, contact support or try again.")
	}

	data := &rpcclient.TemplateContainer{}
	err = storeItem.Data.Unmarshal(data)
	if err != nil {
		return "", terror.Error(err, "failed to get store item info")
	}
	txRef := fmt.Sprintf("Lootbox %s | %d", data.Template.Label, time.Now().UnixNano())

	// resultChan := make(chan *passport.TransactionResult, 1)

	price := decimal.New(2500, 18)

	trans := &passport.NewTransaction{
		To:                   passport.XsynTreasuryUserID,
		From:                 user.ID,
		Amount:               price,
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Mystery crate purchase.",
		Group:                passport.TransactionGroupStore,
		SubGroup:             "Purchase",
	}

	// process user cache map
	nfb, ntb, _, txerr := ucmProcess(trans)
	if txerr != nil {
		return "", terror.Error(txerr)
	}

	if !trans.From.IsSystemUser() {
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
	}

	if !trans.To.IsSystemUser() {
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())
	}

	// refund callback
	refund := func(reason string) {
		if txerr != nil {
			return
		}
		trans := &passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               price,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
			Group:                passport.TransactionGroupStore,
			SubGroup:             "Refund",
		}

		nfb, ntb, _, err := ucmProcess(trans)
		if err != nil {
			log.Err(err).Msg("failed to process refund")
			return
		}

		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())
	}

	// let's assign the item.
	tx, err := conn.Begin(ctx)
	if err != nil {
		refund(err.Error())
		return "", terror.Error(err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	newItem, err := db.PurchasedItemRegister(uuid.Must(uuid.FromString(storeItem.ID)), uuid.UUID(user.ID))
	if err != nil {
		refund(err.Error())
		return "", terror.Error(err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		refund(err.Error())
		return "", terror.Error(err)
	}

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", busKey, storeItem.ID)), storeItem)
	return newItem.Hash, nil
}

func LootboxAmountPerFaction(
	ctx context.Context,
	conn *pgxpool.Pool,
	log *zerolog.Logger,
	bus *messagebus.MessageBus,
	busKey messagebus.BusKey,
	factionID passport.FactionID,
) (int, error) {
	collection, err := db.GenesisCollection()
	if err != nil {
		return 0, err
	}
	// get all faction items marked as loot box
	remainingLootboxesForFaction, err := db.StoreItemsRemainingByFactionIDAndRestrictionGroup(uuid.Must(uuid.FromString(collection.ID)), uuid.UUID(factionID), db.RestrictionGroupLootbox)
	if err != nil {
		return -1, terror.Error(err, "failed to get loot box items")
	}

	return remainingLootboxesForFaction, nil
}
