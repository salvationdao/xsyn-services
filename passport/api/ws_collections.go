package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ninja-software/log_helpers"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
	"github.com/rs/zerolog"
)

// CollectionController holds handlers for Collections
type CollectionController struct {
	Log           *zerolog.Logger
	API           *API
	isTestnetwork bool
}

// NewCollectionController creates the collection hub
func NewCollectionController(log *zerolog.Logger, api *API, isTestnetwork bool) *CollectionController {
	collectionHub := &CollectionController{
		Log:           log_helpers.NamedLogger(log, "collection_hub"),
		API:           api,
		isTestnetwork: isTestnetwork,
	}

	// collection list
	api.SecureCommand(HubKeyCollectionList, collectionHub.CollectionsList)
	api.Command(HubKeyWalletCollectionList, collectionHub.WalletCollectionsList)

	// collection subscribe
	api.Command(HubKeyCollectionSubscribe, collectionHub.Collection)

	//api.SubscribeCommand(HubKeyCollectionSubscribe, collectionHub.CollectionUpdatedSubscribeHandler)

	return collectionHub
}

// CollectionListRequest requests holds the filter for collections list
type CollectionListRequest struct {
	Payload struct {
		UserID types.UserID `json:"user_id"`
	} `json:"payload"`
}

// CollectionListResponse is the response from get collection list
type CollectionListResponse struct {
	Records []*boiler.Collection `json:"records"`
	Total   int                  `json:"total"`
}

const HubKeyCollectionList = "COLLECTION:LIST"

func (ctrlr *CollectionController) CollectionsList(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Could not get list of collections, try again or contact support."
	req := &CollectionListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	collections, err := db.CollectionsVisibleList()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(&CollectionListResponse{
		Records: collections,
		Total:   len(collections),
	})
	return nil

}

type WalletCollectionsListRequest struct {
	Payload struct {
		Username        string                     `json:"username"`
		SortDir         db.SortByDir               `json:"sort_dir"`
		SortBy          string                     `json:"sort_by"`
		AttributeFilter *db.AttributeFilterRequest `json:"attribute_filter,omitempty"`
		AssetType       string                     `json:"asset_type"`
		Archived        bool                       `json:"archived"`
		Search          string                     `json:"search"`
		PageSize        int                        `json:"page_size"`
		Page            int                        `json:"page"`
	} `json:"payload"`
}

// WalletCollectionListResponse is the response from get WalletCollection list
type WalletCollectionListResponse struct {
	Total       int      `json:"total"`
	AssetHashes []string `json:"asset_hashes"`
}

const HubKeyWalletCollectionList = "COLLECTION:WALLET:LIST"

func (ctrlr *CollectionController) WalletCollectionsList(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to get user's NFT assets, try again or contact support."
	req := &WalletCollectionsListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	un := ctx.Value("username")
	username, ok := un.(string)
	if !ok {
		return terror.Error(fmt.Errorf("username not found"), errMsg)
	}

	user, err := boiler.Users(boiler.UserWhere.Username.EQ(strings.ToLower(username))).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "user not found")
	}

	o := bridge.NewOracle(ctrlr.API.BridgeParams.MoralisKey)

	network := bridge.NetworkGoerli
	if !ctrlr.isTestnetwork {
		network = bridge.NetworkEth
	}

	// get all collections
	collections, err := db.CollectionsList()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// for each collection get all nfts
	items := []*boiler.PurchasedItemsOld{}
	for _, c := range collections {
		walletCollections, err := o.NFTOwners(common.HexToAddress(c.MintContract.String), network)
		if err != nil {
			return terror.Error(err, errMsg)
		}

		// for all nfts
		for _, nft := range walletCollections.Result {
			// get metadata
			// if asset is owned by user anbd matches filter, add to result
			if nft.OwnerOf == user.PublicAddress.String {
				tokenID, err := strconv.ParseInt(nft.TokenID, 10, 64)
				if err != nil {
					return terror.Error(err, errMsg)
				}

				item, err := db.PurchasedItemByMintContractAndTokenID(common.HexToAddress(nft.TokenAddress), int(tokenID))
				if err != nil {
					return terror.Error(err, errMsg)
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
	assets []*types.XsynMetadata,
	search string,
	attributeFilter *db.AttributeFilterRequest,
	offset int,
	pageSize int,
	sortBy string,
	sortDir db.SortByDir,
) (int, []*types.XsynMetadata, error) {
	filtered := make([]*types.XsynMetadata, 0)
	filtered = append(filtered, assets...)
	for _, a := range assets {
		if attributeFilter != nil {
			for _, f := range attributeFilter.Items {
				column := db.TraitType(f.Trait)
				err := column.IsValid()
				if err != nil {
					return 0, nil, terror.Error(err, "Error filtering by attribute, try again or contact support.")
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
	Payload struct {
		Slug string `json:"slug"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyCollectionSubscribe, CollectionController.CollectionSubscribe)
const HubKeyCollectionSubscribe = "COLLECTION:SUBSCRIBE"

func (ctrlr *CollectionController) Collection(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to subscribe to collection updates, try again or contact support."
	req := &CollectionUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	collection, err := db.CollectionBySlug(req.Payload.Slug)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(collection)
	return nil
}
