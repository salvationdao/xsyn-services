package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
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
		StoreItemID passport.StoreItemID `json:"storeItemID"`
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
		fsa, err := db.AssetSaleAvailable(ctx, sc.Conn)
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
		FactionID passport.FactionID `json:"factionID"`
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
		fsa, err := db.AssetSaleAvailable(ctx, sc.Conn)
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
		UserID               passport.UserID            `json:"user_id"`
		SortDir              db.SortByDir               `json:"sortDir"`
		SortBy               db.StoreColumn             `json:"sortBy"`
		IncludedStoreItemIDs []passport.StoreItemID     `json:"includedTokenIDs"`
		Filter               *db.ListFilterRequest      `json:"filter,omitempty"`
		AttributeFilter      *db.AttributeFilterRequest `json:"attributeFilter,omitempty"`
		AssetType            string                     `json:"assetType"`
		Archived             bool                       `json:"archived"`
		Search               string                     `json:"search"`
		PageSize             int                        `json:"pageSize"`
		Page                 int                        `json:"page"`
	} `json:"payload"`
}

type StoreListResponse struct {
	Total        int                     `json:"total"`
	StoreItemIDs []*passport.StoreItemID `json:"storeItemIDs"`
}

func (sc *StoreControllerWS) StoreListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &StoreListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	// genesisCollection, err := db.CollectionGet(ctx, sc.Conn, "supremacy-genesis")
	// if err != nil {
	// 	return terror.Error(err, "Error getting collection details, please contact support.")
	// }

	// collectionFilter := &db.ListFilterRequest{
	// 	LinkOperator: db.LinkOperatorTypeAnd,
	// 	Items: []*db.ListFilterRequestItem{{
	// 		ColumnField:   string(db.StoreColumnCollectionID),
	// 		OperatorValue: db.OperatorValueTypeEquals,
	// 		Value:         genesisCollection.ID.String(),
	// 	}},
	// }

	// megaFilter := &db.AttributeFilterRequest{
	// 	LinkOperator: db.LinkOperatorTypeAnd,
	// 	Items: []*db.AttributeFilterRequestItem{
	// 		{
	// 			Trait:         "Rarity",
	// 			Value:         "Mega",
	// 			OperatorValue: db.OperatorValueTypeEquals,
	// 		},
	// 	},
	// }
	// notMegaFilter := &db.AttributeFilterRequest{
	// 	LinkOperator: db.LinkOperatorTypeAnd,
	// 	Items: []*db.AttributeFilterRequestItem{
	// 		{
	// 			Trait:         "Rarity",
	// 			Value:         "Mega",
	// 			OperatorValue: db.OperatorValueTypeIsNot,
	// 		},
	// 	},
	// }
	// genesisMegaCount, _, err := db.AssetList(ctx, sc.Conn,
	// 	"", false, nil, collectionFilter, megaFilter, 0, 5000, "", "")
	// if err != nil {
	// 	return terror.Error(err)
	// }
	// collectionFilter = &db.ListFilterRequest{
	// 	LinkOperator: db.LinkOperatorTypeAnd,
	// 	Items: []*db.ListFilterRequestItem{{
	// 		ColumnField:   string(db.StoreColumnCollectionID),
	// 		OperatorValue: db.OperatorValueTypeEquals,
	// 		Value:         genesisCollection.ID.String(),
	// 	}},
	// }
	// genesisNonMegaCount, _, err := db.AssetList(ctx, sc.Conn,
	// 	"", false, nil, collectionFilter, notMegaFilter, 0, 5000, "", "")
	// if err != nil {
	// 	return terror.Error(err)
	// }

	total, storeItems, err := db.StoreList(
		ctx, sc.Conn,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.IncludedStoreItemIDs,
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
	storeItemIDs := make([]*passport.StoreItemID, 0)
	for _, s := range storeItems {
		// if s.UsdCentCost == 100 && genesisMegaCount >= 2 {
		// 	continue
		// }
		// if s.Restriction == "LOOTBOX" && genesisNonMegaCount >= 10 {
		// 	continue
		// }
		storeItemIDs = append(storeItemIDs, &s.ID)
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
		StoreItemID passport.StoreItemID `json:"storeItemID"`
	} `json:"payload"`
}

func (sc *StoreControllerWS) StoreItemSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &StoreItemSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	item, err := db.StoreItemGet(ctx, sc.Conn, req.Payload.StoreItemID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	// if item isn't faction specific, return item
	if item.FactionID.IsNil() {
		reply(item)
		return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyStoreItemSubscribe, item.ID)), nil
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

	if *user.FactionID != item.FactionID {
		return "", "", terror.Warn(fmt.Errorf("user has wrong faction, need %s, got %s", item.FactionID, user.FactionID), "You do not belong to the correct faction.")
	}

	supsAsCents := sc.API.SupUSD.Mul(decimal.New(100, 0))
	priceAsCents := decimal.New(int64(item.UsdCentCost), 0)
	priceAsSups := priceAsCents.Div(supsAsCents)
	item.SupCost = priceAsSups.Mul(decimal.New(1, 18)).BigInt().String()

	reply(item)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyStoreItemSubscribe, item.ID)), nil
}

const HubKeyAvailableItemAmountSubscribe hub.HubCommandKey = "AVAILABLE:ITEM:AMOUNT"

func (sc *StoreControllerWS) AvailableItemAmountSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	fsa, err := db.AssetSaleAvailable(ctx, sc.Conn)
	if err != nil {
		return "", "", terror.Error(err)
	}

	reply(fsa)

	return req.TransactionID, messagebus.BusKey(HubKeyAvailableItemAmountSubscribe), nil
}
