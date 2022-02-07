package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/log_helpers"

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
	api.SecureUserSubscribeCommand(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)

	// asset set name
	api.SecureCommand(HubKeyAssetUpdateName, assetHub.AssetUpdateNameHandler)

	api.SecureCommand(HubKeyAssetQueueJoin, assetHub.JoinQueueHandler)
	api.SecureCommand(HubKeyAssetQueueLeave, assetHub.LeaveQueueHandler)

	return assetHub
}

// AssetQueueRequest contain the asset token id that user want to join/leave the battle queue
type AssetQueueRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetTokenID uint64 `json:"assetTokenID"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetQueueJoin, AssetController.RegisterHandler)
const HubKeyAssetQueueLeave hub.HubCommandKey = "ASSET:QUEUE:LEAVE"

func (ac *AssetController) LeaveQueueHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AssetQueueRequest{}
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
	user, err := db.UserGet(ctx, ac.Conn, userID, ac.API.HostUrl)
	if err != nil {
		return terror.Error(err)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User need to join a faction")
	}

	err = db.XsynAssetUnfreezeableCheck(ctx, ac.Conn, req.Payload.AssetTokenID, userID)
	if err != nil {
		return terror.Error(terror.ErrInvalidInput, "Current asset is unable to leave the battle queue")
	}

	warMachineMetadata := &passport.WarMachineMetadata{
		TokenID:   req.Payload.AssetTokenID,
		OwnedByID: userID,
		FactionID: *user.FactionID,
	}

	// release the asset from the queue
	ac.API.SendToAllServerClient(&ServerClientMessage{
		Key: AssetQueueLeave,
		Payload: struct {
			// TODO: change this to metadata
			WarMachineNFT *passport.WarMachineMetadata `json:"warMachineNFT"`
		}{
			// TODO: change this to metadata
			WarMachineNFT: warMachineMetadata,
		},
	})

	return nil
}

// 	rootHub.SecureCommand(HubKeyAssetQueueJoin, AssetController.RegisterHandler)
const HubKeyAssetQueueJoin hub.HubCommandKey = "ASSET:QUEUE:JOIN"

// JoinQueueHandler join user's asset to queue
func (ac *AssetController) JoinQueueHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AssetQueueRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// parse user id
	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	// TODO: In the future, check user has enough sups to join their war machine into battle queue

	// get user
	user, err := db.UserGet(ctx, ac.Conn, userID, ac.API.HostUrl)
	if err != nil {
		return terror.Error(err)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User needs to join a faction to Deploy War Machine")
	}

	// check user own this asset and it has not joined the queue yet
	metadata, err := db.XsynMetadataAvailableGet(ctx, ac.Conn, userID, req.Payload.AssetTokenID)
	if err != nil {
		return terror.Error(err)
	}

	warMachineMetadata := &passport.WarMachineMetadata{
		OwnedByID: userID,
	}

	// parse metadata
	for _, att := range metadata.Attributes {
		if att.TraitType != "Asset Type" {
			continue
		}

		switch att.Value {
		case string(passport.WarMachine):
			passport.ParseWarMachineMetadata(metadata, warMachineMetadata)
		case string(passport.Weapon):
		case string(passport.Utility):
		}
	}

	// assign faction id
	warMachineMetadata.FactionID = *user.FactionID

	// join the asset to the queue
	ac.API.SendToAllServerClient(&ServerClientMessage{
		Key: AssetQueueJoin,
		Payload: struct {
			// TODO: change this to metadata
			WarMachineNFT *passport.WarMachineMetadata `json:"warMachineNFT"`
		}{
			// TODO: change this to metadata
			WarMachineNFT: warMachineMetadata,
		},
	})

	reply(true)
	return nil
}

// AssetsUpdatedSubscribeRequest requests holds the filter for user list
type AssetsUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID           passport.UserID       `json:"user_id"`
		SortDir          db.SortByDir          `json:"sortDir"`
		SortBy           db.AssetColumn        `json:"sortBy"`
		IncludedTokenIDs []int                 `json:"includedTokenIDs"`
		Filter           *db.ListFilterRequest `json:"filter,omitempty"`
		AssetType        string                `json:"assetType"`
		Archived         bool                  `json:"archived"`
		Search           string                `json:"search"`
		PageSize         int                   `json:"pageSize"`
		Page             int                   `json:"page"`
	} `json:"payload"`
}

// AssetListResponse is the response from get asset list
type AssetListResponse struct {
	Records []*passport.XsynMetadata `json:"records"`
	Total   int                      `json:"total"`
}

const HubKeyAssetList hub.HubCommandKey = "ASSET:LIST"

func (ctrlr *AssetController) AssetListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	req := &AssetsUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	assets := []*passport.XsynMetadata{}
	total, err := db.AssetList(
		ctx, ctrlr.Conn, &assets,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.IncludedTokenIDs,
		req.Payload.Filter,
		req.Payload.AssetType,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return terror.Error(err)
	}

	resp := &AssetListResponse{
		Total:   total,
		Records: assets,
	}

	reply(resp)
	return nil
}

// AssetUpdatedSubscribeRequest requests an update for an xsyn_metadata
type AssetUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TokenID uint64 `json:"tokenID"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetSubscribe, AssetController.AssetSubscribe)
const HubKeyAssetSubscribe hub.HubCommandKey = "ASSET:SUBSCRIBE"

func (ctrlr *AssetController) AssetUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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

// AssetSetNameRequest requests an update for an xsyn_metadata
type AssetSetNameRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TokenID uint64           `json:"tokenID"`
		UserID  *passport.UserID `json:"userID"`
		Name    string           `json:"name"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetUpdateName, AssetController.AssetUpdateNameHandler)
const HubKeyAssetUpdateName hub.HubCommandKey = "ASSET:UPDATE:NAME"

// AssetSetName update's name of an asset
func (ac *AssetController) AssetUpdateNameHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AssetSetNameRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get asset
	asset, err := db.AssetGet(ctx, ac.Conn, req.Payload.TokenID)
	if err != nil {
		return terror.Error(err)
	}

	// check if user owns asset
	if *asset.UserID != *req.Payload.UserID {
		return terror.Error(err, "Must own Asset to update it's name")
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

	// update asset name
	err = db.AssetUpdate(ctx, ac.Conn, asset.TokenID, req.Payload.Name)
	if err != nil {
		return terror.Error(err, "Failed to update Asset name")
	}

	// get asset
	asset, err = db.AssetGet(ctx, ac.Conn, req.Payload.TokenID)
	if err != nil {
		return terror.Error(err)
	}

	ac.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, req.Payload.TokenID)), asset)

	ac.API.SendToAllServerClient(&ServerClientMessage{
		Key: AssetUpdated,
		Payload: struct {
			Asset *passport.XsynMetadata `json:"asset"`
		}{
			Asset: asset,
		},
	})

	reply(asset)
	return nil
}
