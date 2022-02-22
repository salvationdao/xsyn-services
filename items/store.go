package items

import (
	"context"
	"errors"
	"fmt"
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
	supPrice decimal.Decimal, txChan chan<- *passport.NewTransaction, user passport.User, storeItemID passport.StoreItemID) error {
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

	txID := uuid.Must(uuid.NewV4())
	txRef := fmt.Sprintf("PURCHASE OF %s %s", storeItem.Name, txID)

	// convert price to sups
	asDecimal := decimal.New(int64(storeItem.UsdCentCost), 0).Div(supPrice).Ceil()
	asSups := decimal.New(asDecimal.IntPart(), 18).BigInt()

	resultChan := make(chan *passport.TransactionResult, 1)

	select {
	case txChan <- &passport.NewTransaction{
		To:                   passport.XsynTreasuryUserID,
		From:                 user.ID,
		Amount:               *asSups,
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Purchase on Supremacy storefront.",
		ResultChan:           resultChan,
	}:

	case <-time.After(10 * time.Second):
		log.Err(errors.New("timeout on channel send exceeded"))
		panic("Purchase on Supremacy storefront.")
	}

	result := <-resultChan

	if result.Error != nil {
		return terror.Error(result.Error)
	}

	if result.Transaction.Status != passport.TransactionSuccess {
		return terror.Error(fmt.Errorf("purchase failed: %s", result.Transaction.Reason), fmt.Sprintf("Purchase failed: %s.", result.Transaction.Reason))
	}

	// refund callback
	refund := func(reason string) {
		select {
		case txChan <- &passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               *asSups,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of purchase on Supremacy storefront.",
		}:

		case <-time.After(10 * time.Second):
			log.Err(errors.New("timeout on channel send exceeded"))
			panic("Refund of purchase on Supremacy storefront.")
		}
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
		ExternalUrl:  "TODO",
		Image:        storeItem.Image,
		Attributes:   storeItem.Attributes,
	}

	// create item on metadata table
	err = db.XsynMetadataInsert(ctx, conn, newItem)
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
