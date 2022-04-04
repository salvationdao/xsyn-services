package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"math/big"
	"net/http"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/microcosm-cc/bluemonday"

	"github.com/ninja-software/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

// AssetController holds handlers for as
type AssetController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewAssetController creates the asset hub
func NewAssetController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *AssetController {
	assetHub := &AssetController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "asset_hub"),
		API:  api,
	}

	// assets list
	api.Command(HubKeyAssetList, assetHub.AssetListHandler)

	// asset subscribe
	api.SubscribeCommand(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)

	// asset set name
	api.SecureCommand(HubKeyAssetUpdateName, assetHub.AssetUpdateNameHandler)

	api.SecureCommand(HubKeyAssetQueueJoin, assetHub.JoinQueueHandler)
	api.SecureUserSubscribeCommand(HubKeyAssetRepairStatUpdate, assetHub.AssetRepairStatUpdateSubscriber)
	api.SecureUserSubscribeCommand(HubKeyAssetQueueCostUpdate, assetHub.AssetQueueCostUpdateSubscriber)

	return assetHub
}

// AssetJoinQueueRequest contain the asset token id that user want to join/leave the battle queue
type AssetJoinQueueRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash   string `json:"asset_hash"`
		NeedInsured bool   `json:"need_insured"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetQueueJoin, AssetController.RegisterHandler)
const HubKeyAssetQueueJoin hub.HubCommandKey = "ASSET:QUEUE:JOIN"

// JoinQueueHandler join user's asset to queue
func (ac *AssetController) JoinQueueHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Issue joining user's asset to queue, try again or contact support."
	req := &AssetJoinQueueRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// parse user id
	userID := types.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User is not logged in, access forbidden.")
	}

	// get user
	user, err := db.UserGet(ctx, ac.Conn, userID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User needs to join a faction to deploy war machine.")
	}

	// check user own this asset, and it has not joined the queue yet
	asset, err := db.PurchasedItemByHash(req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	if asset == nil {
		return terror.Error(fmt.Errorf("asset doesn't exist"), "Asset doesn't exist.")
	}

	if asset.OwnerID != userID.String() {
		return terror.Error(terror.ErrForbidden, "Asset is not owned by this user.")
	}

	warMachineMetadata := &types.WarMachineMetadata{
		OwnedByID: userID,
	}

	// assign faction id
	warMachineMetadata.FactionID = *user.FactionID
	warMachineMetadata.Hash = req.Payload.AssetHash

	var resp struct {
		Position       *int    `json:"position"`
		ContractReward *string `json:"contract_reward"`
	}
	err = ac.API.GameserverRequest(http.MethodPost, "/war_machine_join", struct {
		WarMachineMetadata *types.WarMachineMetadata `json:"war_machine_metadata"`
		NeedInsured        bool                      `json:"need_insured"`
	}{
		WarMachineMetadata: warMachineMetadata,
		NeedInsured:        req.Payload.NeedInsured,
	}, &resp)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	ac.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyWarMachineQueueStatSubscribe, warMachineMetadata.Hash)), resp)
	reply(true)
	return nil
}

// AssetsUpdatedSubscribeRequest requests holds the filter for user list
type AssetsUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
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

const HubKeyAssetList hub.HubCommandKey = "ASSET:LIST"

func (ac *AssetController) AssetListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	req := &AssetsUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	userID := types.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(fmt.Errorf("no auth: user ID %s", userID), "User is not logged in, access forbidden.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, items, err := db.PurchaseItemsList(
		ctx, ac.Conn,
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
	*hub.HubCommandRequest
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

const HubKeyAssetSubscribe hub.HubCommandKey = "ASSET:SUBSCRIBE"

func (ac *AssetController) AssetUpdatedSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	errMsg := "Issue subscribing to asset updates, try again or contact support."
	req := &AssetUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	asset, err := db.PurchasedItemByHash(req.Payload.AssetHash)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, errMsg)
	}
	if asset == nil {
		return req.TransactionID, "", terror.Error(fmt.Errorf("asset doesn't exist"), "Asset doesn't exist.")
	}

	owner, err := db.UserGet(context.Background(), ac.Conn, types.UserID(uuid.Must(uuid.FromString(asset.OwnerID))))
	if err != nil {
		return req.TransactionID, "", terror.Error(err, errMsg)
	}

	collection, err := db.Collection(uuid.Must(uuid.FromString(asset.CollectionID)))
	if err != nil {
		return req.TransactionID, "", terror.Error(err, errMsg)
	}
	reply(&AssetUpdatedSubscribeResponse{
		PurchasedItem:  asset,
		OwnerUsername:  owner.Username,
		CollectionSlug: collection.Slug,
		HostURL:        ac.API.GameserverHostUrl,
	})
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.Hash)), nil
}

// AssetRepairStatUpdateRequest request the repair stat of the asset
type AssetRepairStatUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

const HubKeyAssetRepairStatUpdate hub.HubCommandKey = "ASSET:DURABILITY:SUBSCRIBE"

func (ac *AssetController) AssetRepairStatUpdateSubscriber(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	ac.Log.Debug().RawJSON("req", payload).Str("fn", "AssetRepairStatUpdateSubscriber").Msg("ws handler")

	errMsg := "Issue subscribing to asset repair status updates, try again or contact support."
	req := &AssetRepairStatUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	userID := types.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return req.TransactionID, "", terror.Error(fmt.Errorf("no auth"), "User is not logged in, access forbidden.")
	}

	if req.Payload.AssetHash == "" {
		return req.TransactionID, "", terror.Error(fmt.Errorf("empty asset hash"), errMsg)
	}

	// check ownership
	// get asset
	asset, err := db.PurchasedItemByHash(req.Payload.AssetHash)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}
	if asset == nil {
		return "", "", terror.Error(fmt.Errorf("asset doesn't exist"), "This asset does not exist, try again or contact support.")
	}

	// check if user owns asset
	if asset.OwnerID != userID.String() {
		return "", "", terror.Error(err, "Must own asset to repair it, try again or contact support.")
	}

	// get repair stat from gameserver
	var resp struct {
		AssetRepairRecord *types.AssetRepairRecord `json:"asset_repair_record"`
	}

	err = ac.API.GameserverRequest(http.MethodPost, "/asset_repair_stat", struct {
		Hash string `json:"hash"`
	}{
		Hash: req.Payload.AssetHash,
	}, &resp)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	if resp.AssetRepairRecord.Hash == req.Payload.AssetHash {
		resp.AssetRepairRecord.StartedAt = ReverseAssetRepairStartTime(resp.AssetRepairRecord)
		reply(resp)
	} else {
		reply(nil)
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetRepairStatUpdate, req.Payload.AssetHash)), nil
}

const HubKeyAssetQueueCostUpdate hub.HubCommandKey = "ASSET:QUEUE:COST:UPDATE"

func (ac *AssetController) AssetQueueCostUpdateSubscriber(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	errMsg := "Issue subscribing to asset queue cost updates, try again or contact support."
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	userID := types.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden, "User is not logged in, access forbidden.")
	}

	faction, err := db.FactionGetByUserID(context.Background(), ac.Conn, userID)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	var resp struct {
		Length int `json:"length"`
	}

	err = ac.API.GameserverRequest(http.MethodPost, "/faction_queue_cost", struct {
		FactionID types.FactionID `json:"faction_id"`
	}{
		FactionID: faction.ID,
	}, &resp)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	cost := big.NewInt(1000000000000000000)
	cost.Mul(cost, big.NewInt(int64(resp.Length)+1))

	reply(cost.String())

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetRepairStatUpdate, faction.ID)), nil
}

func ReverseAssetRepairStartTime(record *types.AssetRepairRecord) time.Time {
	secondPerPoint := 18
	if record.RepairMode == types.RepairModeStandard {
		secondPerPoint = 864
	}

	return record.ExpectCompletedAt.Add(time.Duration(-100*secondPerPoint) * time.Second)
}

// AssetSetNameRequest requests an update for an xsyn_metadata
type AssetSetNameRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string        `json:"asset_hash"`
		UserID    *types.UserID `json:"user_id"`
		Name      string        `json:"name"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetUpdateName, AssetController.AssetUpdateNameHandler)
const HubKeyAssetUpdateName hub.HubCommandKey = "ASSET:UPDATE:NAME"

func (ac *AssetController) AssetUpdateNameHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Issue updating asset name, try again or contact support."
	bm := bluemonday.StrictPolicy()

	req := &AssetSetNameRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// get user
	user, err := boiler.FindUser(passdb.StdConn, hubc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if user.RenameBanned.Valid && user.RenameBanned.Bool {
		return terror.Warn(fmt.Errorf("user rename banned"), "You have been banned from renaming, asshole.")
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
	go ac.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, req.Payload.AssetHash)), &AssetUpdatedSubscribeResponse{
		PurchasedItem:  item,
		OwnerUsername:  user.Username,
		CollectionSlug: collection.Slug,
	})

	reply(item)
	return nil
}