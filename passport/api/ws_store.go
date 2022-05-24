package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/items"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/log_helpers"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)


// StoreControllerWS holds handlers for serverClienting serverClient status
type StoreControllerWS struct {
	Log *zerolog.Logger
	API *API
}

// NewStoreController creates the serverClient hub
func NewStoreController(log *zerolog.Logger, api *API) *StoreControllerWS {
	storeHub := &StoreControllerWS{
		Log: log_helpers.NamedLogger(log, "store_hub"),
		API: api,
	}

	api.SecureCommand(HubKeyStoreList, storeHub.StoreListHandler)
	api.SecureCommand(HubKeyLootbox, storeHub.PurchaseLootboxHandler)
	api.SecureCommand(HubKeyLootboxAmount, storeHub.LootboxAmountHandler)

	api.SecureCommand(HubKeyPurchaseItem, storeHub.PurchaseItemHandler)

	api.SecureCommand(types.HubKeyStoreItemSubscribe, storeHub.StoreItemHandler)

	return storeHub
}

const HubKeyPurchaseItem = "STORE:PURCHASE"

type PurchaseRequest struct {
	Payload struct {
		StoreItemID types.StoreItemID `json:"store_item_id"`
	} `json:"payload"`
}

func (sc *StoreControllerWS) PurchaseItemHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PurchaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	//  sc.API.MessageBus, messagebus.BusKey(HubKeyStoreItemSubscribe),
	err = items.Purchase(sc.Log, decimal.New(12, -2), sc.API.userCacheMap.Transact, user, req.Payload.StoreItemID)
	if err != nil {
		return err
	}

	reply(true)

	// broadcast available mech amount
	go func() {
		fsa, err := db.StoreItemsAvailable()
		if err != nil {
			sc.API.Log.Err(err)
			return
		}
		ws.PublishMessage("/store/availability", types.HubKeyAvailableItemAmount, fsa)
		//sc.API.MessageBus.Send(messagebus.BusKey(HubKeyAvailableItemAmountSubscribe), fsa)
	}()

	return nil
}

type PurchaseLootboxRequest struct {
	Payload struct {
		FactionID types.FactionID `json:"faction_id"`
	} `json:"payload"`
}

const HubKeyLootbox = "STORE:LOOTBOX"

func (sc *StoreControllerWS) PurchaseLootboxHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	return terror.Warn(fmt.Errorf("store closed"), "The XSYN Store is currently closed.")

	//req := &PurchaseLootboxRequest{}
	//err := json.Unmarshal(payload, req)
	//if err != nil {
	//	return terror.Error(err, "Invalid request received.")
	//}
	//
	//item, err := items.PurchaseLootbox(sc.Log, sc.API.userCacheMap.Transact, user.User, req.Payload.FactionID)
	//if err != nil {
	//	return err
	//}
	//
	//err = item.L.LoadCollection(passdb.StdConn, true, item, nil)
	//
	//reply(&AssetUpdatedSubscribeResponse{
	//	PurchasedItem:  item,
	//	OwnerUsername:  user.Username,
	//	CollectionSlug: item.R.Collection.Slug,
	//	HostURL:        sc.API.GameserverHostUrl,
	//})
	//
	//// broadcast available mech amount
	//go func() {
	//	fsa, err := db.StoreItemsAvailable()
	//	if err != nil {
	//		sc.API.Log.Err(err)
	//		return
	//	}
	//	ws.PublishMessage("/store/availability", types.HubKeyAvailableItemAmount, fsa)
	//}()

	return nil
}

const HubKeyLootboxAmount = "STORE:LOOTBOX:AMOUNT"

func (sc *StoreControllerWS) LootboxAmountHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &PurchaseLootboxRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}



	reply(-1)
	return nil
}

const HubKeyStoreList = "STORE:LIST"

type StoreListRequest struct {
	Payload struct {
		UserID              types.UserID               `json:"user_id"`
		SortDir             db.SortByDir               `json:"sort_dir"`
		SortBy              string                     `json:"sortBy"`
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
	Total        int                 `json:"total"`
	StoreItemIDs []types.StoreItemID `json:"store_item_ids"`
}

func (sc *StoreControllerWS) StoreListHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue getting list of store items, try again or contact support."
	req := &StoreListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	
	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	// TODO: remove megas filters later (?)
	total, items, err := db.StoreItemsList(
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.Filter,
		req.Payload.AttributeFilter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	storeItemIDs := make([]types.StoreItemID, 0)
	for _, storeItem := range items {
		storeItemIDs = append(storeItemIDs, storeItem.ID)
	}

	reply(&StoreListResponse{
		total,
		storeItemIDs,
	})
	return nil
}

type StoreItemSubscribeRequest struct {
	Payload struct {
		StoreItemID types.StoreItemID `json:"store_item_id"`
	} `json:"payload"`
}

type StoreItemSubscribeResponse struct {
	PriceInSUPS string            `json:"price_in_sups"`
	Item        *boiler.StoreItem `json:"item"`
	HostURL     string            `json:"host_url"`
}

func (sc *StoreControllerWS) StoreItemHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &StoreItemSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	item, err := db.StoreItem(uuid.UUID(req.Payload.StoreItemID))
	if err != nil {
		return terror.Error(err, "Could not get store item, try again or contact support.")
	}

	if !user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user has no faction"), "Please select a syndicate to view this item.")
	}

	//if user.FactionID.String != item.FactionID {
	//	return terror.Warn(fmt.Errorf("user has wrong faction, need %s, got %s", item.FactionID, user.FactionID), "You do not belong to the correct faction.")
	//}

	supsAsCents, err := db.SupInCents()
	if err != nil {
		return terror.Error(err, "Could not get SUP price, try again or contact support.")
	}

	priceAsCents := decimal.New(int64(item.UsdCentCost), 0)
	priceAsSups := priceAsCents.Div(supsAsCents).Mul(decimal.New(1, 18)).BigInt().String()

	result := &StoreItemSubscribeResponse{
		PriceInSUPS: priceAsSups,
		Item:        item,
		HostURL:     sc.API.GameserverHostUrl,
	}

	reply(result)
	return nil
}
