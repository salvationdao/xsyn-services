package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/items"
	"passport/log_helpers"

	"github.com/shopspring/decimal"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub/ext/messagebus"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
)

// StoreControllerWS holds handlers for serverClienting serverClient status
type StoreControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewStoreController creates the serverClient hub
func NewStoreController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *StoreControllerWS {
	storeHub := &StoreControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "store_hub"),
		API:  api,
	}

	api.Command(HubKeyStoreList, storeHub.StoreListHandler)
	api.SecureCommand(HubKeyPurchaseItem, storeHub.PurchaseItemHandler)

	api.SecureUserSubscribeCommand(HubKeyStoreItemSubscribe, storeHub.StoreItemSubscribeHandler)

	return storeHub
}

const HubKeyPurchaseItem = hub.HubCommandKey("STORE:PURCHASE")

type PurchaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		StoreItemID passport.StoreItemID `json:"storeItemID"`
	} `json:"payload"`
}

func (ctrlr *StoreControllerWS) PurchaseItemHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err)
	}
	user, err := db.UserGet(ctx, ctrlr.Conn, passport.UserID(uid))
	if err != nil {
		return terror.Error(err)
	}
	err = items.Purchase(ctx, ctrlr.Conn, ctrlr.Log, ctrlr.API.MessageBus, messagebus.BusKey(HubKeyStoreItemSubscribe), decimal.New(12, -2), ctrlr.API.transaction, *user, req.Payload.StoreItemID)
	if err != nil {
		return terror.Error(err)
	}

	reply(true)
	return nil
}

const HubKeyStoreList = hub.HubCommandKey("STORE:LIST")

type StoreListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		CollectionID passport.CollectionID `json:"collectionID"`
		FactionID    passport.FactionID    `json:"factionID"`
		Page         int                   `json:"page"`
		PageSize     int                   `json:"pageSize"`
	} `json:"payload"`
}

type StoreListResponse struct {
	StoreItemIDs []*passport.StoreItemID `json:"storeItemIDs"`
}

func (ctrlr *StoreControllerWS) StoreListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &StoreListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	storeItemIDs, err := db.StoreItemListByCollectionAndFaction(ctx, ctrlr.Conn, req.Payload.CollectionID, req.Payload.FactionID, req.Payload.Page, req.Payload.PageSize)
	if err != nil {
		return terror.Error(err)
	}

	reply(storeItemIDs)
	return nil
}

const HubKeyStoreItemSubscribe hub.HubCommandKey = "STORE:ITEM"

type StoreItemSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		StoreItemID passport.StoreItemID `json:"storeItemID"`
	} `json:"payload"`
}

func (ctrlr *StoreControllerWS) StoreItemSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &StoreItemSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	item, err := db.StoreItemByID(ctx, ctrlr.Conn, req.Payload.StoreItemID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	// if item isn't faction specific, return item
	if item.FactionID.IsNil() {
		reply(item)
		return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyStoreItemSubscribe, item.ID)), nil
	}

	// get user to check faction
	uid, err := uuid.FromString(client.Identifier())
	if err != nil {
		return "", "", terror.Error(err)
	}
	user, err := db.UserGet(ctx, ctrlr.Conn, passport.UserID(uid))
	if err != nil {
		return "", "", terror.Error(err)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return "", "", terror.Error(fmt.Errorf("user has no faction"), "Please select a syndicate to view this item.")
	}

	if *user.FactionID != item.FactionID {
		return "", "", terror.Warn(fmt.Errorf("user has wrong faction, need %s, got %s", item.FactionID, user.FactionID), "You do not belong to the correct faction.")
	}

	priceAsDecimal := decimal.New(int64(item.UsdCentCost), 0).Div(ctrlr.API.SupUSD).Ceil()
	priceAsSups := decimal.New(priceAsDecimal.IntPart(), 18).BigInt()
	item.SupCost = priceAsSups.String()

	reply(item)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyStoreItemSubscribe, item.ID)), nil
}
