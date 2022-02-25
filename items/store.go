package items

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"passport"
	"passport/db"
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
func Purchase(ctx context.Context, conn *pgxpool.Pool, log *zerolog.Logger, bus *messagebus.MessageBus, busKey messagebus.BusKey,
	supPrice decimal.Decimal, ucmProcess func(*passport.NewTransaction) (*big.Int, *big.Int, string, error), user passport.User, storeItemID passport.StoreItemID, externalUrl string) error {
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
	txRef := fmt.Sprintf("PURCHASE OF %s %s %d", storeItem.Name, txID, time.Now().UnixNano())

	supsAsCents := supPrice.Mul(decimal.New(100, 0))
	priceAsCents := decimal.New(int64(storeItem.UsdCentCost), 0)
	priceAsSupsDecimal := priceAsCents.Div(supsAsCents)
	priceAsSupsBigInt := priceAsSupsDecimal.Mul(decimal.New(1, 18)).BigInt()

	// resultChan := make(chan *passport.TransactionResult, 1)
	trans := &passport.NewTransaction{
		To:                   passport.XsynTreasuryUserID,
		From:                 user.ID,
		Amount:               *priceAsSupsBigInt,
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Purchase on Supremacy storefront.",
	}

	nfb, ntb, _, err := ucmProcess(trans)
	if err != nil {
		return terror.Error(err, "failed to process user sups")
	}

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())

	// refund callback
	refund := func(reason string) {
		trans := &passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               *priceAsSupsBigInt,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
		}

		nfb, ntb, _, err := ucmProcess(trans)
		if err != nil {
			log.Err(errors.New("failed to process user sups"))
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

	// create metadata object
	newItem := &passport.XsynMetadata{
		CollectionID: storeItem.CollectionID,
		Description:  &storeItem.Description,
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
	err = db.XsynMetadataAssignUser(ctx, conn, newItem.Hash, user.ID, newItem.CollectionID, newItem.ExternalTokenID)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	// update item amounts
	err = db.StoreItemPurchased(ctx, conn, storeItem)
	if err != nil {
		fmt.Println("here33")
		refund(err.Error())
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		refund(err.Error())
		return terror.Error(err)
	}

	storeItem.SupCost = priceAsSupsBigInt.String()

	go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", busKey, storeItem.ID)), storeItem)
	return nil
}

// PurchaseLootbox attempts to make a purchase for a given user ID and a given
func PurchaseLootbox(ctx context.Context, conn *pgxpool.Pool, log *zerolog.Logger, bus *messagebus.MessageBus, busKey messagebus.BusKey,
	ucmProcess func(*passport.NewTransaction) (*big.Int, *big.Int, string, error), user passport.User, factionID passport.FactionID, externalURL string) (string, error) {

	// get all faction items marked as loot box
	mechs, err := db.StoreItemListByFactionLootbox(ctx, conn, factionID)
	if err != nil {
		return "", terror.Error(err, "failed to get loot box items")
	}

	mechIDPool := []*passport.StoreItemID{}
	for _, m := range mechs {
		for i := 0; i < m.AmountAvailable-m.AmountSold; i++ {
			mechIDPool = append(mechIDPool, &m.ID)
		}
	}

	if len(mechs) == 0 {
		return "", terror.Error(fmt.Errorf("all sold out"), "This item has sold out.")
	}

	chosenIdx := rand.Intn(len(mechIDPool))
	var storeItem *passport.StoreItem

	for _, m := range mechs {
		if m.ID == *mechIDPool[chosenIdx] {
			storeItem = m
			continue
		}
	}

	if storeItem == nil {
		return "", terror.Error(fmt.Errorf("store item nil"), "Internal error, contact support or try again.")
	}

	txID := uuid.Must(uuid.NewV4())
	txRef := fmt.Sprintf("Lootbox %s %s %d", storeItem.Name, txID, time.Now().UnixNano())

	// resultChan := make(chan *passport.TransactionResult, 1)

	price := *decimal.New(2500, 18).BigInt()

	trans := &passport.NewTransaction{
		To:                   passport.XsynTreasuryUserID,
		From:                 user.ID,
		Amount:               price,
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Mystery Create purchase.",
	}

	// process user cache map
	nfb, ntb, _, err := ucmProcess(trans)
	if err != nil {
		return "", terror.Error(err)
	}

	if !trans.From.IsSystemUser() {
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.From)), nfb.String())
	}

	if !trans.To.IsSystemUser() {
		go bus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", "USER:SUPS:SUBSCRIBE", trans.To)), ntb.String())
	}

	// refund callback
	refund := func(reason string) {
		trans := &passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               price,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
		}

		nfb, ntb, _, err := ucmProcess(trans)
		if err != nil {
			log.Err(errors.New("failed to process user sups"))
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

	// create metadata object
	newItem := &passport.XsynMetadata{
		CollectionID: storeItem.CollectionID,
		Description:  &storeItem.Description,
		Image:        storeItem.Image,
		Attributes:   storeItem.Attributes,
		AnimationURL: storeItem.AnimationURL,
	}

	// create item on metadata table
	err = db.XsynMetadataInsert(ctx, conn, newItem, externalURL)
	if err != nil {
		refund(err.Error())
		return "", terror.Error(err)
	}

	// assign new item to user
	err = db.XsynMetadataAssignUser(ctx, conn, newItem.Hash, user.ID, newItem.CollectionID, newItem.ExternalTokenID)
	if err != nil {
		refund(err.Error())
		return "", terror.Error(err)
	}

	// update item amounts
	err = db.StoreItemPurchased(ctx, conn, storeItem)
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
