package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	goaway "github.com/TwiN/go-away"
	"github.com/microcosm-cc/bluemonday"

	"github.com/ninja-software/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

var profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(Profanities, []string{}, []string{})

// AssetController holds handlers for as
type AssetController struct {
	Log *zerolog.Logger
	API *API
}

// NewAssetController creates the asset hub
func NewAssetController(log *zerolog.Logger, api *API) *AssetController {
	assetHub := &AssetController{
		Log: log_helpers.NamedLogger(log, "asset_hub"),
		API: api,
	}

	// assets list
	api.SecureCommand(HubKeyAssetList, assetHub.AssetListHandler)
	// asset subscribe
	api.Command(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)

	// asset set name
	api.SecureCommand(HubKeyAssetUpdateName, assetHub.AssetUpdateNameHandler)
	// api.SecureUserSubscribeCommand(HubKeyAssetRepairStatUpdate, assetHub.AssetRepairStatUpdateSubscriber)
	// api.SecureUserSubscribeCommand(HubKeyAssetQueueCostUpdate, assetHub.AssetQueueCostUpdateSubscriber)

	return assetHub
}

// AssetJoinQueueRequest contain the asset token id that user want to join/leave the battle queue
type AssetJoinQueueRequest struct {
	Payload struct {
		AssetHash   string `json:"asset_hash"`
		NeedInsured bool   `json:"need_insured"`
	} `json:"payload"`
}

// AssetsUpdatedSubscribeRequest requests holds the filter for user list
type AssetsUpdatedSubscribeRequest struct {
	Payload struct {
		UserID              types.UserID               `json:"user_id"`
		SortDir             db.SortByDir               `json:"sortDir"`
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

// AssetListResponse is the response from get asset list
type AssetListResponse struct {
	Total       int      `json:"total"`
	AssetHashes []string `json:"asset_hashes"`
}

const HubKeyAssetList = "ASSET:LIST"

func (ac *AssetController) AssetListHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetsUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, items, err := db.PurchaseItemsList(
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
		return terror.Error(err, "Could not get list of assets, try again or contact support.")
	}

	itemHashes := make([]string, 0)
	for _, s := range items {
		itemHashes = append(itemHashes, s.Hash)
	}

	resp := &AssetListResponse{
		total,
		itemHashes,
	}

	reply(resp)
	return nil
}

// AssetUpdatedSubscribeRequest requests an update for an xsyn_metadata
type AssetUpdatedSubscribeRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

type AssetUpdatedSubscribeResponse struct {
	CollectionSlug string                `json:"collection_slug"`
	PurchasedItem  *boiler.PurchasedItem `json:"purchased_item"`
	OwnerUsername  string                `json:"owner_username"`
	HostURL        string                `json:"host_url"`
}

const HubKeyAssetSubscribe = "ASSET:SUBSCRIBE"

func (ac *AssetController) AssetUpdatedSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue subscribing to asset updates, try again or contact support."
	req := &AssetUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	asset, err := db.PurchasedItemByHash(req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if asset == nil {
		return terror.Error(fmt.Errorf("asset doesn't exist"), "Asset doesn't exist.")
	}

	owner, err := users.ID(asset.OwnerID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	collection, err := db.Collection(uuid.Must(uuid.FromString(asset.CollectionID)))
	if err != nil {
		return terror.Error(err, errMsg)
	}
	reply(&AssetUpdatedSubscribeResponse{
		PurchasedItem:  asset,
		OwnerUsername:  owner.Username,
		CollectionSlug: collection.Slug,
		HostURL:        ac.API.GameserverHostUrl,
	})
	return nil
}

// AssetRepairStatUpdateRequest request the repair stat of the asset
type AssetRepairStatUpdateRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

// AssetSetNameRequest requests an update for an xsyn_metadata
type AssetSetNameRequest struct {
	Payload struct {
		AssetHash string        `json:"asset_hash"`
		UserID    *types.UserID `json:"user_id"`
		Name      string        `json:"name"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetUpdateName, AssetController.AssetUpdateNameHandler)
const HubKeyAssetUpdateName = "ASSET:UPDATE:NAME"

func (ac *AssetController) AssetUpdateNameHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating asset name, try again or contact support."
	bm := bluemonday.StrictPolicy()

	req := &AssetSetNameRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if user.RenameBanned.Valid && user.RenameBanned.Bool {
		return terror.Warn(fmt.Errorf("user rename banned"), "You have been banned from renaming.")
	}

	if profanityDetector.IsProfane(req.Payload.Name) {
		return terror.Error(err, "Profanity is not allowed.")
	}

	// get item
	item, err := db.PurchasedItemByHash(req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	if item == nil {
		return terror.Error(fmt.Errorf("item doesn't exist"), "This item does not exist.")
	}

	// check if user owns asset
	if item.OwnerID != user.ID {
		return terror.Error(err, "User must own item to update it's name, try again or contact support.")
	}

	name := html.UnescapeString(bm.Sanitize(req.Payload.Name))

	if len(name) > 25 {
		return terror.Error(err, "Name must be less than 25 characters.")
	}

	// update asset name
	item, err = db.PurchasedItemSetName(uuid.Must(uuid.FromString(item.ID)), name)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	collection, err := db.Collection(uuid.Must(uuid.FromString(item.CollectionID)))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	data, _ := json.Marshal(&AssetUpdatedSubscribeResponse{
		PurchasedItem:  item,
		OwnerUsername:  user.Username,
		CollectionSlug: collection.Slug,
	})

	ws.Publish(fmt.Sprintf("/ws/public/%s", user.Username), data)

	//go ac.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, req.Payload.AssetHash)))

	reply(item)
	return nil
}
