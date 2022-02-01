package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

// CollectionController holds handlers for Collections
type CollectionController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewCollectionController creates the collection hub
func NewCollectionController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *CollectionController {
	collectionHub := &CollectionController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "collection_hub"),
		API:  api,
	}

	// collection list
	api.SecureUserSubscribeCommand(HubKeyCollectionsSubscribe, collectionHub.CollectionsUpdatedSubscribeHandler)

	// collection get
	api.SecureUserSubscribeCommand(HubKeyCollectionSubscribe, collectionHub.CollectionUpdatedSubscribeHandler)

	return collectionHub
}

// rootHub.SecureCommand(HubKeyAssetList, AssetController.GetHandler)

// AssetListHandlerRequest requests holds the filter for user list
type CollectionsUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID           passport.UserID       `json:"user_id"`
		SortDir          db.SortByDir          `json:"sortDir"`
		SortBy           db.AssetColumn        `json:"sortBy"`
		IncludedTokenIDs []int                 `json:"includedTokenIDs"`
		Filter           *db.ListFilterRequest `json:"filter"`
		Archived         bool                  `json:"archived"`
		Search           string                `json:"search"`
		PageSize         int                   `json:"pageSize"`
		Page             int                   `json:"page"`
	} `json:"payload"`
}

// CollectionListResponse is the response from get asset list
type CollectionListResponse struct {
	Records []*passport.XsynNftMetadata `json:"records"`
	Total   int                         `json:"total"`
}

const HubKeyCollectionsSubscribe hub.HubCommandKey = "COLLECTION_LIST:SUBSCRIBE"

func (ctrlr *CollectionController) CollectionsUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &AssetsUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	assets := []*passport.XsynNftMetadata{}
	total, err := db.AssetList(
		ctx, ctrlr.Conn, &assets,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.IncludedTokenIDs,
		req.Payload.Filter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	resp := &AssetListResponse{
		Total:   total,
		Records: assets,
	}

	reply(resp)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetsSubscribe, req.Payload.UserID.String())), nil

}

// CollectionUpdatedSubscribeRequest requests an update for a collection
type CollectionUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		id uint64 `json:"id"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyCollectionSubscribe, CollectionController.CollectionSubscribe)
const HubKeyCollectionSubscribe hub.HubCommandKey = "ASSET:SUBSCRIBE"

func (ctrlr *CollectionController) CollectionUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &AssetUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	asset, err := db.AssetGet(ctx, ctrlr.Conn, req.Payload.TokenID)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	reply(asset)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.TokenID)), nil
}
