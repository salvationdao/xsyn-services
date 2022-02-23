package items

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"passport"
	"passport/db"

	"github.com/ninja-syndicate/hub/ext/messagebus"

	"github.com/rs/zerolog"

	"github.com/jackc/pgx/v4"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gofrs/uuid"

	"github.com/shopspring/decimal"

	"github.com/ninja-software/terror/v2"
)

// Purchase attempts to make a purchase for a given user ID and a given
func Purchase(ctx context.Context, conn *pgxpool.Pool, log *zerolog.Logger, bus *messagebus.MessageBus, busKey messagebus.BusKey,
	supPrice decimal.Decimal, ucmProcess func(fromID, toID string, amount big.Int) (*big.Int, *big.Int, error), txProcess func(t passport.NewTransaction) string, user passport.User, storeItemID passport.StoreItemID, externalUrl string) error {
	storeItem, err := db.StoreItemGet(ctx, conn, storeItemID)
	if err != nil {
		return terror.Error(err)
	}

	if storeItem.AmountSold >= storeItem.AmountAvailable {
		return terror.Error(fmt.Errorf("all sold out"), "This item has sold out.")
	}

	if !storeItem.FactionID.IsNil() && (user.FactionID == nil || user.FactionID.IsNil()) {
		return terror.Error(fmt.Errorf("user has no faction"), "You need a faction to purchase faction specific items.")
	}

	if !storeItem.FactionID.IsNil() && *user.FactionID != storeItem.FactionID {
		return terror.Error(fmt.Errorf("user is wrong faction"), "You cannot buy items for another faction")
	}

	if storeItem.Restriction == "WHITELIST" || storeItem.Restriction == "LOOTBOX" {
		return terror.Error(fmt.Errorf("cannot purchase whitelist item or lootbox"), "Item currently not available.")
	}

	txID := uuid.Must(uuid.NewV4())
	txRef := fmt.Sprintf("PURCHASE OF %s %s", storeItem.Name, txID)

	// convert price to sups
	asDecimal := decimal.New(int64(storeItem.UsdCentCost), 0).Div(supPrice).Ceil()
	asSups := decimal.New(asDecimal.IntPart(), 18).BigInt()

	// resultChan := make(chan *passport.TransactionResult, 1)
	trans := passport.NewTransaction{
		To:                   passport.XsynTreasuryUserID,
		From:                 user.ID,
		Amount:               *asSups,
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Purchase on Supremacy storefront.",
	}

	nfb, ntb, err := ucmProcess(trans.From.String(), trans.To.String(), trans.Amount)
	if err != nil {
		return terror.Error(err, "failed to process user sups")
	}

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())

	txProcess(trans)
	// select {
	// case txChan <- &passport.NewTransaction{
	// 	To:                   passport.XsynTreasuryUserID,
	// 	From:                 user.ID,
	// 	Amount:               *asSups,
	// 	TransactionReference: passport.TransactionReference(txRef),
	// 	Description:          "Purchase on Supremacy storefront.",
	// 	// ResultChan:           resultChan,
	// }:

	// case <-time.After(10 * time.Second):
	// 	log.Err(errors.New("timeout on channel send exceeded"))
	// 	panic("Purchase on Supremacy storefront.")
	// }

	// result := <-resultChan

	// if result.Error != nil {
	// 	return terror.Error(result.Error)
	// }

	// if result.Transaction.Status != passport.TransactionSuccess {
	// 	return terror.Error(fmt.Errorf("purchase failed: %s", result.Transaction.Reason), fmt.Sprintf("Purchase failed: %s.", result.Transaction.Reason))
	// }

	// refund callback
	refund := func(reason string) {
		trans := passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               *asSups,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
		}

		nfb, ntb, err := ucmProcess(trans.From.String(), trans.To.String(), trans.Amount)
		if err != nil {
			log.Err(errors.New("failed to process user sups"))
			return
		}

		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())

		txProcess(trans)
		// select {
		// case txChan <- &passport.NewTransaction{
		// 	To:                   user.ID,
		// 	From:                 passport.XsynTreasuryUserID,
		// 	Amount:               *asSups,
		// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
		// 	Description:          "Refund of purchase on Supremacy storefront.",
		// }:

		// case <-time.After(10 * time.Second):
		// 	log.Err(errors.New("timeout on channel send exceeded"))
		// 	panic("Refund of purchase on Supremacy storefront.")
		// }
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

	// create metadata object
	newItem := &passport.XsynMetadata{
		CollectionID: storeItem.CollectionID,
		Description:  storeItem.Description,
		Image:        storeItem.Image,
		Attributes:   storeItem.Attributes,
		AnimationURL: storeItem.AnimationURL,
	}

	// create item on metadata table
	err = db.XsynMetadataInsert(ctx, conn, newItem, externalUrl)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	// assign new item to user
	err = db.XsynMetadataAssignUser(ctx, conn, newItem.TokenID, user.ID)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	// update item amounts
	err = db.StoreItemPurchased(ctx, conn, storeItem)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	priceAsDecimal := decimal.New(int64(storeItem.UsdCentCost), 0).Div(supPrice).Ceil()
	priceAsSups := decimal.New(priceAsDecimal.IntPart(), 18).BigInt()
	storeItem.SupCost = priceAsSups.String()

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", busKey, storeItem.ID)), storeItem)
	return nil
}

// Purchase attempts to make a purchase for a given user ID and a given
func PurchaseLootbox(ctx context.Context, conn *pgxpool.Pool, log *zerolog.Logger, bus *messagebus.MessageBus, busKey messagebus.BusKey,
	supPrice decimal.Decimal, ucmProcess func(fromID, toID string, amount big.Int) (*big.Int, *big.Int, error), txProcess func(t passport.NewTransaction) string, user passport.User, factionID passport.FactionID, externalURL string) (*uint64, error) {

	// get all faction items marked as loot box
	mechs, err := db.StoreItemListByFactionLootbox(ctx, conn, passport.FactionID(factionID))
	if err != nil {
		return nil, terror.Error(err, "failed to get loot box items")
	}

	if len(mechs) == 0 {
		return nil, terror.Error(fmt.Errorf("all sold out"), "This item has sold out.")
	}

	chosenIdx := rand.Intn(len(mechs))
	storeItem := mechs[chosenIdx]

	// check available
	if storeItem.AmountAvailable < 1 || storeItem.AmountSold >= storeItem.AmountAvailable {
		return nil, terror.Error(fmt.Errorf("all sold out"), "This item has sold out.")
	}

	txID := uuid.Must(uuid.NewV4())
	txRef := fmt.Sprintf("Lootbox %s %s ", storeItem.Name, txID)

	// convert price to sups
	asDecimal := decimal.New(int64(storeItem.UsdCentCost), 0).Div(supPrice).Ceil()
	asSups := decimal.New(asDecimal.IntPart(), 18).BigInt()

	// resultChan := make(chan *passport.TransactionResult, 1)

	trans := passport.NewTransaction{
		To:                   passport.XsynTreasuryUserID,
		From:                 user.ID,
		Amount:               *asSups,
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Lootbox prize.",
		// ResultChan:           resultChan,
	}
	// process user cache map
	nfb, ntb, err := ucmProcess(trans.From.String(), trans.To.String(), trans.Amount)
	if err != nil {
		return nil, terror.Error(err, "Failed to transfer the fund")
	}
	// resultChan := make(chan *passport.TransactionResult)

	if !trans.From.IsSystemUser() {
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
	}

	if !trans.To.IsSystemUser() {
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())
	}

	txProcess(trans)

	// txChan <- &passport.NewTransaction{
	// 	To:                   passport.XsynTreasuryUserID,
	// 	From:                 user.ID,
	// 	Amount:               *asSups,
	// 	TransactionReference: passport.TransactionReference(txRef),
	// 	Description:          "Lootbox prize.",
	// 	ResultChan:           resultChan,
	// }

	// result := <-resultChan

	// if result.Error != nil {
	// 	return nil, terror.Error(result.Error)
	// }

	// if result.Transaction.Status != passport.TransactionSuccess {
	// 	return nil, terror.Error(fmt.Errorf("lootbox failed: %s", result.Transaction.Reason), fmt.Sprintf("lootbox failed: %s.", result.Transaction.Reason))
	// }

	// refund callback
	refund := func(reason string) {
		// txChan <- &passport.NewTransaction{
		// 	To:                   user.ID,
		// 	From:                 passport.XsynTreasuryUserID,
		// 	Amount:               *asSups,
		// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
		// 	Description:          "Refund of lootbox",
		// }
		trans := passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               *asSups,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
		}

		nfb, ntb, err := ucmProcess(trans.From.String(), trans.To.String(), trans.Amount)
		if err != nil {
			log.Err(errors.New("failed to process user sups"))
			return
		}

		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())

		txProcess(trans)
	}

	// let's assign the item.
	tx, err := conn.Begin(ctx)
	if err != nil {
		refund(err.Error())
		return nil, terror.Error(err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	// create metadata object
	newItem := &passport.XsynMetadata{
		CollectionID: storeItem.CollectionID,
		Description:  storeItem.Description,
		Image:        storeItem.Image,
		Attributes:   storeItem.Attributes,
		AnimationURL: storeItem.AnimationURL,
	}

	// create item on metadata table
	err = db.XsynMetadataInsert(ctx, conn, newItem, "test")
	if err != nil {
		refund(err.Error())
		return nil, terror.Error(err)
	}

	// assign new item to user
	err = db.XsynMetadataAssignUser(ctx, conn, newItem.TokenID, user.ID)
	if err != nil {
		refund(err.Error())
		return nil, terror.Error(err)
	}

	// update item amounts
	err = db.StoreItemPurchased(ctx, conn, storeItem)
	if err != nil {
		refund(err.Error())
		return nil, terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		refund(err.Error())
		return nil, terror.Error(err)
	}

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", busKey, storeItem.ID)), storeItem)
	return &newItem.TokenID, nil
}
