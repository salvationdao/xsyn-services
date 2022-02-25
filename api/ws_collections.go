package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"

	"github.com/ninja-software/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
	"github.com/rs/zerolog"
)

// CollectionController holds handlers for Collections
type CollectionController struct {
	Conn          *pgxpool.Pool
	Log           *zerolog.Logger
	API           *API
	isTestnetwork bool
}

// NewCollectionController creates the collection hub
func NewCollectionController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, isTestnetwork bool) *CollectionController {
	collectionHub := &CollectionController{
		Conn:          conn,
		Log:           log_helpers.NamedLogger(log, "collection_hub"),
		API:           api,
		isTestnetwork: isTestnetwork,
	}

	// collection list
	api.Command(HubKeyCollectionList, collectionHub.CollectionsList)
	api.Command(HubKeyWalletCollectionList, collectionHub.WalletCollectionsList)

	// collection subscribe
	api.SubscribeCommand(HubKeyCollectionSubscribe, collectionHub.CollectionUpdatedSubscribeHandler)

	return collectionHub
}

// CollectionsUpdatedSubscribeRequest requests holds the filter for collections list
type CollectionsUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID           passport.UserID       `json:"userID"`
		SortDir          db.SortByDir          `json:"sortDir"`
		SortBy           db.CollectionColumn   `json:"sortBy"`
		IncludedTokenIDs []int                 `json:"includedTokenIDs"`
		Filter           *db.ListFilterRequest `json:"filter"`
		Archived         bool                  `json:"archived"`
		Search           string                `json:"search"`
		PageSize         int                   `json:"pageSize"`
		Page             int                   `json:"page"`
	} `json:"payload"`
}

// CollectionListResponse is the response from get collection list
type CollectionListResponse struct {
	Records []*passport.Collection `json:"records"`
	Total   int                    `json:"total"`
}

type WalletCollectionListResponse struct {
	NFTOwners []*bridge.NFTOwner
}

const HubKeyCollectionList hub.HubCommandKey = "COLLECTION:LIST"

func (ctrlr *CollectionController) CollectionsList(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &CollectionsUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	collections := []*passport.Collection{}
	total, err := db.CollectionsList(
		ctx, ctrlr.Conn, &collections,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.Filter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return terror.Error(err)
	}

	resp := &CollectionListResponse{
		Total:   total,
		Records: collections,
	}

	reply(resp)
	return nil

}

const HubKeyWalletCollectionList hub.HubCommandKey = "COLLECTION:WALLET:LIST"

func (ctrlr *CollectionController) WalletCollectionsList(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	//o := bridge.NewOracle(ctrlr.API.BridgeParams.MoralisKey)
	//
	//network := bridge.NetworkGoerli
	//if !ctrlr.isTestnetwork {
	//	network = bridge.NetworkEth
	//}
	//
	////walletCollections, err := o.NFTOwners(ctrlr.API.BridgeParams.EthNftAddr, network)
	//
	//if err != nil {
	//	return terror.Error(err)
	//}
	//
	//resp := &WalletCollectionListResponse{
	//	NFTOwners: walletCollections.Result,
	//}
	//reply(resp)
	return nil

}

// CollectionUpdatedSubscribeRequest requests an update for a collection
type CollectionUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Slug string `json:"slug"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyCollectionSubscribe, CollectionController.CollectionSubscribe)
const HubKeyCollectionSubscribe hub.HubCommandKey = "COLLECTION:SUBSCRIBE"

func (ctrlr *CollectionController) CollectionUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &CollectionUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	collection, err := db.CollectionGet(ctx, ctrlr.Conn, req.Payload.Slug)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	reply(collection)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyCollectionSubscribe, collection.ID)), nil
}
