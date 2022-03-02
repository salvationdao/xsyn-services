package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"passport/db"
	"time"

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
		AssetHash   string `json:"assetHash"`
		NeedInsured bool   `json:"needInsured"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetQueueJoin, AssetController.RegisterHandler)
const HubKeyAssetQueueJoin hub.HubCommandKey = "ASSET:QUEUE:JOIN"

// JoinQueueHandler join user's asset to queue
func (ac *AssetController) JoinQueueHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AssetJoinQueueRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// parse user id
	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	// get user
	user, err := db.UserGet(ctx, ac.Conn, userID)
	if err != nil {
		return terror.Error(err)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User needs to join a faction to deploy war machine.")
	}

	// check user own this asset, and it has not joined the queue yet
	asset, err := db.AssetGet(ctx, ac.Conn, req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err)
	}
	if asset == nil {
		return terror.Error(fmt.Errorf("asset doesn't exist"))
	}

	if asset.UserID == nil || *asset.UserID != userID {
		return terror.Error(terror.ErrForbidden)
	}

	warMachineMetadata := &passport.WarMachineMetadata{
		OwnedByID: userID,
	}

	// parse metadata
	for _, att := range asset.Attributes {
		if att.TraitType != "Asset Type" {
			continue
		}

		switch att.Value {
		case string(passport.WarMachine):
			passport.ParseWarMachineMetadata(asset, warMachineMetadata)
		case string(passport.Weapon):
		case string(passport.Utility):
		}
	}

	// assign faction id
	warMachineMetadata.FactionID = *user.FactionID
	warMachineMetadata.Hash = req.Payload.AssetHash

	var resp struct {
		Position       *int    `json:"position"`
		ContractReward *string `json:"contractReward"`
	}
	err = ac.API.GameserverRequest(http.MethodPost, "/war_machine_join", struct {
		WarMachineMetadata *passport.WarMachineMetadata `json:"warMachineMetadata"`
		NeedInsured        bool                         `json:"needInsured"`
	}{
		WarMachineMetadata: warMachineMetadata,
		NeedInsured:        req.Payload.NeedInsured,
	}, &resp)
	if err != nil {
		return terror.Error(err)
	}

	ac.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyWarMachineQueueStatSubscribe, warMachineMetadata.Hash)), resp)

	reply(true)
	return nil
}

// AssetsUpdatedSubscribeRequest requests holds the filter for user list
type AssetsUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID              passport.UserID            `json:"user_id"`
		SortDir             db.SortByDir               `json:"sortDir"`
		SortBy              db.AssetColumn             `json:"sortBy"`
		IncludedAssetHashes []string                   `json:"includedAssetHashes"`
		Filter              *db.ListFilterRequest      `json:"filter,omitempty"`
		AttributeFilter     *db.AttributeFilterRequest `json:"attributeFilter,omitempty"`
		AssetType           string                     `json:"assetType"`
		Archived            bool                       `json:"archived"`
		Search              string                     `json:"search"`
		PageSize            int                        `json:"pageSize"`
		Page                int                        `json:"page"`
	} `json:"payload"`
}

// AssetListResponse is the response from get asset list
type AssetListResponse struct {
	Total       int      `json:"total"`
	AssetHashes []string `json:"assetHashes"`
}

const HubKeyAssetList hub.HubCommandKey = "ASSET:LIST"

func (ac *AssetController) AssetListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	req := &AssetsUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, assets, err := db.AssetList(
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
		return terror.Error(err)
	}

	assetHashes := make([]string, 0)
	for _, s := range assets {
		assetHashes = append(assetHashes, s.Hash)
	}

	resp := &AssetListResponse{
		total,
		assetHashes,
	}

	reply(resp)
	return nil
}

// AssetUpdatedSubscribeRequest requests an update for an xsyn_metadata
type AssetUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"assetHash"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetSubscribe, AssetController.AssetSubscribe)
const HubKeyAssetSubscribe hub.HubCommandKey = "ASSET:SUBSCRIBE"

func (ac *AssetController) AssetUpdatedSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &AssetUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	asset, err := db.AssetGet(ctx, ac.Conn, req.Payload.AssetHash)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}
	if asset == nil {
		return req.TransactionID, "", terror.Error(fmt.Errorf("asset doesn't exist"))
	}

	reply(asset)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.Hash)), nil
}

// AssetRepairStatUpdateRequest request the repair stat of the asset
type AssetRepairStatUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"assetHash"`
	} `json:"payload"`
}

const HubKeyAssetRepairStatUpdate hub.HubCommandKey = "ASSET:DURABILITY:SUBSCRIBE"

func (ac *AssetController) AssetRepairStatUpdateSubscriber(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	ac.Log.Debug().RawJSON("req", payload).Str("fn", "AssetRepairStatUpdateSubscriber").Msg("ws handler")

	req := &AssetRepairStatUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return req.TransactionID, "", terror.Error(fmt.Errorf("no auth"))
	}

	if req.Payload.AssetHash == "" {
		return req.TransactionID, "", terror.Error(fmt.Errorf("empty asset hash"), "Issue subscripting to asset repair status.")
	}

	// check ownership
	// get asset
	asset, err := db.AssetGet(ctx, ac.Conn, req.Payload.AssetHash)
	if err != nil {
		return "", "", terror.Error(err)
	}
	if asset == nil {
		return "", "", terror.Error(fmt.Errorf("asset doesn't exist"), "This asset does not exist.")
	}

	// check if user owns asset
	if *asset.UserID != userID {
		return "", "", terror.Error(err, "Must own Asset to repair it.")
	}

	// get repair stat from gameserver
	var resp struct {
		AssetRepairRecord *passport.AssetRepairRecord `json:"assetRepairRecord"`
	}

	err = ac.API.GameserverRequest(http.MethodPost, "/asset_repair_stat", struct {
		Hash string `json:"hash"`
	}{
		Hash: req.Payload.AssetHash,
	}, &resp)
	if err != nil {
		return "", "", terror.Error(err)
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
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	faction, err := db.FactionGetByUserID(context.Background(), ac.Conn, userID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	var resp struct {
		Length int `json:"length"`
	}

	err = ac.API.GameserverRequest(http.MethodPost, "/faction_queue_cost", struct {
		FactionID passport.FactionID `json:"factionID"`
	}{
		FactionID: faction.ID,
	}, &resp)
	if err != nil {
		return "", "", terror.Error(err)
	}

	cost := big.NewInt(1000000000000000000)
	cost.Mul(cost, big.NewInt(int64(resp.Length)+1))

	reply(cost.String())

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetRepairStatUpdate, faction.ID)), nil
}

func ReverseAssetRepairStartTime(record *passport.AssetRepairRecord) time.Time {
	secondPerPoint := 18
	if record.RepairMode == passport.RepairModeStandard {
		secondPerPoint = 864
	}

	return record.ExpectCompletedAt.Add(time.Duration(-100*secondPerPoint) * time.Second)
}

// AssetSetNameRequest requests an update for an xsyn_metadata
type AssetSetNameRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string           `json:"assetHash"`
		UserID    *passport.UserID `json:"userID"`
		Name      string           `json:"name"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetUpdateName, AssetController.AssetUpdateNameHandler)
const HubKeyAssetUpdateName hub.HubCommandKey = "ASSET:UPDATE:NAME"

func (ac *AssetController) AssetUpdateNameHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	bm := bluemonday.StrictPolicy()

	req := &AssetSetNameRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get asset
	asset, err := db.AssetGet(ctx, ac.Conn, req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err)
	}
	if asset == nil {
		return terror.Error(fmt.Errorf("asset doesn't exist"), "This asset does not exist.")
	}

	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err)
	}

	userID := passport.UserID(uid)

	// check if user owns asset
	if *asset.UserID != userID {
		return terror.Error(err, "Must own Asset to update it's name.")
	}

	// check if war machine
	isWarMachine := false
	for _, att := range asset.Attributes {
		if att.TraitType != "Asset Type" {
			continue
		}
		switch att.Value {
		case string(passport.WarMachine):
			isWarMachine = true
		}
	}
	if !isWarMachine {
		return terror.Error(err, "Asset must be a War Machine")
	}

	name := bm.Sanitize(req.Payload.Name)

	if len(name) > 10 {
		return terror.Error(err, "Name must be less than 10 characters")
	}

	// update asset name
	err = db.AssetUpdate(ctx, ac.Conn, asset.Hash, name)
	if err != nil {
		return terror.Error(err)
	}

	// get asset
	asset, err = db.AssetGet(ctx, ac.Conn, req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err)
	}
	if asset == nil {
		return terror.Error(fmt.Errorf("asset doesn't exist"), "This asset does not exist.")
	}

	go ac.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, req.Payload.AssetHash)), asset)

	reply(asset)
	return nil
}
