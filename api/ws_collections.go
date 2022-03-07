package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
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

// CollectionListRequest requests holds the filter for collections list
type CollectionListRequest struct {
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

const HubKeyCollectionList hub.HubCommandKey = "COLLECTION:LIST"

func (ctrlr *CollectionController) CollectionsList(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &CollectionListRequest{}
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

type WalletCollectionsListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Username        string                     `json:"username"`
		SortDir         db.SortByDir               `json:"sortDir"`
		SortBy          string                     `json:"sortBy"`
		AttributeFilter *db.AttributeFilterRequest `json:"attributeFilter,omitempty"`
		AssetType       string                     `json:"assetType"`
		Archived        bool                       `json:"archived"`
		Search          string                     `json:"search"`
		PageSize        int                        `json:"pageSize"`
		Page            int                        `json:"page"`
	} `json:"payload"`
}

// WalletCollectionListResponse is the response from get WalletCollection list
type WalletCollectionListResponse struct {
	Total       int      `json:"total"`
	AssetHashes []string `json:"assetHashes"`
}

const HubKeyWalletCollectionList hub.HubCommandKey = "COLLECTION:WALLET:LIST"

func (ctrlr *CollectionController) WalletCollectionsList(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &WalletCollectionsListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	o := bridge.NewOracle(ctrlr.API.BridgeParams.MoralisKey)

	network := bridge.NetworkGoerli
	if !ctrlr.isTestnetwork {
		network = bridge.NetworkEth
	}

	// get user
	user, err := db.UserGetByUsername(ctx, ctrlr.Conn, req.Payload.Username)
	if err != nil {
		return terror.Error(err)
	}

	// get all collections
	collections := []*passport.Collection{}
	_, err = db.CollectionsList(ctx, ctrlr.Conn, &collections, "", false, nil, 0, 100, "", db.SortByDirAsc)
	if err != nil {
		return terror.Error(err)
	}

	// for each collection get all nfts
	items := []*boiler.PurchasedItem{}
	for _, c := range collections {
		walletCollections, err := o.NFTOwners(common.HexToAddress(c.MintContract), network)
		if err != nil {
			return terror.Error(err)
		}

		// for all nfts
		for _, nft := range walletCollections.Result {
			// get metadata
			// if asset is owned by user anbd matches filter, add to result
			if nft.OwnerOf == user.PublicAddress.String {
				tokenID, err := strconv.ParseInt(nft.TokenID, 10, 64)
				if err != nil {
					return terror.Error(err)
				}

				item, err := db.PurchasedItemByMintContractAndTokenID(common.HexToAddress(nft.TokenAddress), int(tokenID))
				if err != nil {
					return terror.Error(err)
				}
				items = append(items, item)
			}
		}
	}

	itemHashes := make([]string, 0)
	for _, item := range items {
		itemHashes = append(itemHashes, item.Hash)
	}

	resp := &WalletCollectionListResponse{
		len(itemHashes),
		itemHashes,
	}
	reply(resp)
	return nil
}

func FilterAssetList(
	assets []*passport.XsynMetadata,
	search string,
	attributeFilter *db.AttributeFilterRequest,
	offset int,
	pageSize int,
	sortBy string,
	sortDir db.SortByDir,
) (int, []*passport.XsynMetadata, error) {
	filtered := make([]*passport.XsynMetadata, 0)
	filtered = append(filtered, assets...)
	for _, a := range assets {
		if attributeFilter != nil {
			for _, f := range attributeFilter.Items {
				column := db.TraitType(f.Trait)
				err := column.IsValid()
				if err != nil {
					return 0, nil, terror.Error(err)
				}
				for _, att := range a.Attributes {
					if !(att.TraitType == f.Trait && att.Value == f.Value) {
						filtered = append(filtered, a)
					}
				}
			}
		}
	}

	return len(filtered), filtered, nil
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
