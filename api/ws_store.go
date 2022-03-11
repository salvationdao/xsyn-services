package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/items"

	"github.com/ninja-software/log_helpers"

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
	api.Command(HubKeyLootbox, storeHub.PurchaseLootboxHandler)
	api.Command(HubKeyLootboxAmount, storeHub.LootboxAmountHandler)

	api.SecureCommand(HubKeyPurchaseItem, storeHub.PurchaseItemHandler)

	api.SubscribeCommand(HubKeyStoreItemSubscribe, storeHub.StoreItemSubscribeHandler)
	api.SubscribeCommand(HubKeyAvailableItemAmountSubscribe, storeHub.AvailableItemAmountSubscribeHandler)

	return storeHub
}

const HubKeyPurchaseItem = hub.HubCommandKey("STORE:PURCHASE")

type PurchaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		StoreItemID passport.StoreItemID `json:"store_item_id"`
	} `json:"payload"`
}

func (sc *StoreControllerWS) PurchaseItemHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
	user, err := db.UserGet(ctx, sc.Conn, passport.UserID(uid))
	if err != nil {
		return terror.Error(err)
	}

	err = items.Purchase(ctx, sc.Conn, sc.Log, sc.API.MessageBus, messagebus.BusKey(HubKeyStoreItemSubscribe), decimal.New(12, -2), sc.API.userCacheMap.Process, *user, req.Payload.StoreItemID, sc.API.storeItemExternalUrl)
	if err != nil {
		return terror.Error(err)
	}

	reply(true)

	// broadcast available mech amount
	go func() {
		fsa, err := db.StoreItemsAvailable()
		if err != nil {
			sc.API.Log.Err(err)
			return
		}
		sc.API.MessageBus.Send(ctx, messagebus.BusKey(HubKeyAvailableItemAmountSubscribe), fsa)
	}()

	return nil
}

type PurchaseLootboxRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"faction_id"`
	} `json:"payload"`
}

const HubKeyLootbox = hub.HubCommandKey("STORE:LOOTBOX")

func (sc *StoreControllerWS) PurchaseLootboxHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PurchaseLootboxRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err)
	}
	user, err := db.UserGet(ctx, sc.Conn, passport.UserID(uid))
	if err != nil {
		return terror.Error(err)
	}

	tokenID, err := items.PurchaseLootbox(ctx, sc.Conn, sc.Log, sc.API.MessageBus, messagebus.BusKey(HubKeyStoreItemSubscribe), sc.API.userCacheMap.Process, *user, req.Payload.FactionID, sc.API.storeItemExternalUrl)
	if err != nil {
		return terror.Error(err)
	}

	reply(tokenID)

	// broadcast available mech amount
	go func() {
		fsa, err := db.StoreItemsAvailable()
		if err != nil {
			sc.API.Log.Err(err)
			return
		}
		sc.API.MessageBus.Send(ctx, messagebus.BusKey(HubKeyAvailableItemAmountSubscribe), fsa)
	}()

	return nil
}

const HubKeyLootboxAmount = hub.HubCommandKey("STORE:LOOTBOX:AMOUNT")

func (sc *StoreControllerWS) LootboxAmountHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &PurchaseLootboxRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	amount, err := items.LootboxAmountPerFaction(ctx, sc.Conn, sc.Log, sc.API.MessageBus, messagebus.BusKey(HubKeyLootboxAmount), req.Payload.FactionID)
	if err != nil {
		return terror.Error(err, "Could not get mystery crate amount")
	}

	reply(amount)

	return nil
}

const HubKeyStoreList = hub.HubCommandKey("STORE:LIST")

type StoreListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID              passport.UserID            `json:"user_id"`
		SortDir             db.SortByDir               `json:"sort_dir"`
		SortBy              string                     `json:"sortBy"`
		IncludedAssetHashes []string                   `json:"included_asset_hashes"`
		Filter              *db.ListFilterRequest      `json:"filter,omitempty"`
		AttributeFilter     *db.AttributeFilterRequest `json:"attribute_filter,omitempty"`
		AssetType           string                     `json:"asset_type"`
		Archived            bool                       `json:"archived"`
		Search              string                     `json:"search"`
		PageSize            int                        `json:"page_size"`
		Page                int                        `json:"page"`
	} `json:"payload"`
}

type StoreListResponse struct {
	Total        int                    `json:"total"`
	StoreItemIDs []passport.StoreItemID `json:"store_item_ids"`
}

func (sc *StoreControllerWS) StoreListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &StoreListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get user to check faction
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not find user")
	}
	user, err := db.UserGet(ctx, sc.Conn, passport.UserID(uid))
	if err != nil {
		return terror.Error(err, "Could not find user")
	}

	if user.FactionID == nil {
		return terror.Error(fmt.Errorf("user not enlisted: %s", user.ID), "User is not enlisted")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, items, err := db.StoreItemsList(
		ctx, sc.Conn,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.IncludedAssetHashes,
		req.Payload.Filter,
		req.Payload.AttributeFilter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return terror.Error(err)
	}

	storeItemIDs := make([]passport.StoreItemID, 0)
	// filter by megas for now
	for _, storeItem := range items {
		if storeItem.RestrictionGroup == db.RestrictionGroupPrize {
			continue
		}
		if storeItem.Tier != db.TierMega {
			continue
		}
		if storeItem.IsDefault {
			continue
		}
		storeItemIDs = append(storeItemIDs, storeItem.ID)
	}

	reply(&StoreListResponse{
		total,
		storeItemIDs,
	})
	return nil
}

const HubKeyStoreItemSubscribe hub.HubCommandKey = "STORE:ITEM"

type StoreItemSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		StoreItemID passport.StoreItemID `json:"store_item_id"`
	} `json:"payload"`
}

type StoreItemSubscribeResponse struct {
	PriceInSUPS string            `json:"price_in_sups"`
	Item        *boiler.StoreItem `json:"item"`
}

func (sc *StoreControllerWS) StoreItemSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &StoreItemSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	item, err := db.StoreItem(uuid.UUID(req.Payload.StoreItemID))
	if err != nil {
		return "", "", terror.Error(err)
	}

	if client.Identifier() == "" || client.Level < 1 {
		return "", "", terror.Error(fmt.Errorf("user not logged in"), "You must be logged in and enlisted in the faction to view this item.")
	}

	// get user to check faction
	uid, err := uuid.FromString(client.Identifier())
	if err != nil {
		return "", "", terror.Error(err)
	}
	user, err := db.UserGet(ctx, sc.Conn, passport.UserID(uid))
	if err != nil {
		return "", "", terror.Error(err)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return "", "", terror.Error(fmt.Errorf("user has no faction"), "Please select a syndicate to view this item.")
	}

	if user.FactionID.String() != item.FactionID {
		return "", "", terror.Warn(fmt.Errorf("user has wrong faction, need %s, got %s", item.FactionID, user.FactionID), "You do not belong to the correct faction.")
	}

	supsAsCents, err := db.SupInCents()
	if err != nil {
		return "", "", terror.Error(err, "Could not get SUP price")
	}

	priceAsCents := decimal.New(int64(item.UsdCentCost), 0)
	priceAsSups := priceAsCents.Div(supsAsCents).Mul(decimal.New(1, 18)).BigInt().String()

	result := &StoreItemSubscribeResponse{
		PriceInSUPS: priceAsSups,
		Item:        item,
	}

	reply(result)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyStoreItemSubscribe, item.ID)), nil
}

const HubKeyAvailableItemAmountSubscribe hub.HubCommandKey = "AVAILABLE:ITEM:AMOUNT"

func (sc *StoreControllerWS) AvailableItemAmountSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	fsa, err := db.StoreItemsAvailable()
	if err != nil {
		return "", "", terror.Error(err)
	}

	reply(fsa)

	return req.TransactionID, messagebus.BusKey(HubKeyAvailableItemAmountSubscribe), nil
}
