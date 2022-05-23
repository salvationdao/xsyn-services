package api

import (
	"context"
	"encoding/json"
	"fmt"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)


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
	api.Command(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)

	return assetHub
}

type AssetListRequest struct {
	Payload struct {
		UserID              types.UserID               `json:"user_id"`
		Sort              *db.ListSortRequest                     `json:"sort,omitempty"`
		Filter              *db.ListFilterRequest      `json:"filter,omitempty"`
		AttributeFilter     *db.AttributeFilterRequest `json:"attribute_filter,omitempty"`
		AssetType           string                     `json:"asset_type"`
		Search              string                     `json:"search"`
		PageSize            int                        `json:"page_size"`
		Page                int                        `json:"page"`
	} `json:"payload"`
}

// AssetListResponse is the response from get asset list
type AssetListResponse struct {
	Total       int64                `json:"total"`
	Assets []*types.UserAsset `json:"assets"` // TODO: create api type for user assets
}

const HubKeyAssetList = "ASSET:LIST"

func (ac *AssetController) AssetListHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	total, assets, err := db.AssetList(&db.AssetListOpts{
		UserID:          req.Payload.UserID,
		Sort:            req.Payload.Sort,
		Filter:          req.Payload.Filter,
		AttributeFilter: req.Payload.AttributeFilter,
		AssetType:       req.Payload.AssetType,
		Search:          req.Payload.Search,
		PageSize:        req.Payload.PageSize,
		Page:            req.Payload.Page,
	})
	if err != nil {
		return terror.Error(err, "Unable to retrieve assets at this time, please try again or contact support.")
	}

	reply(&AssetListResponse{
		Total: total,
		Assets: assets,
	})
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
	PurchasedItem  *boiler.PurchasedItemsOld `json:"purchased_item"`
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
		return terror.Error(fmt.Errorf("asset doesn't exist"), "UserAsset doesn't exist.")
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
