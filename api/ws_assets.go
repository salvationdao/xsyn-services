package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"passport"
	"passport/db"
	"passport/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
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
	api.Command(HubKeyAssetList, assetHub.AssetList)

	// asset subscribe
	api.SecureUserSubscribeCommand(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)

	api.SecureCommand(HubKeyAssetRegister, assetHub.RegisterHandler)
	api.SecureCommand(HubKeyAssetQueueJoin, assetHub.JoinQueueHandler)
	api.SecureCommand(HubKeyAssetQueueLeave, assetHub.LeaveQueueHandler)

	return assetHub
}

// AssetRegisterRequest contain the nft that user want to plug into server
type AssetRegisterRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		XsynNftMetadata *passport.XsynNftMetadata `json:"xsynNftMetadata"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyAssetRegister, AssetController.RegisterHandler)
const HubKeyAssetRegister hub.HubCommandKey = "ASSET:REGISTER"

// RegisterHandler allow user to register their nft to their passport
func (ac *AssetController) RegisterHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AssetRegisterRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// parse user id
	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	tx, err := ac.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ac.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	// insert asset to db
	err = db.XsynNftMetadataInsert(ctx, tx, req.Payload.XsynNftMetadata, req.Payload.XsynNftMetadata.Collection.ID)
	if err != nil {
		return terror.Error(err)
	}

	// assign asset to user
	err = db.XsynNftMetadataAssignUser(ctx, tx, req.Payload.XsynNftMetadata.TokenID, userID)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	return nil
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

	warMachineNFT := &passport.WarMachineNFT{
		TokenID:   req.Payload.AssetTokenID,
		OwnedByID: userID,
		FactionID: *user.FactionID,
	}

	// release the asset from the queue
	ac.API.SendToAllServerClient(&ServerClientMessage{
		Key: AssetQueueLeave,
		Payload: struct {
			WarMachineNFT *passport.WarMachineNFT `json:"warMachineNFT"`
		}{
			WarMachineNFT: warMachineNFT,
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
	nft, err := db.XsynNftMetadataAvailableGet(ctx, ac.Conn, userID, req.Payload.AssetTokenID)
	if err != nil {
		return terror.Error(err)
	}

	warMachineNFT := &passport.WarMachineNFT{
		OwnedByID: userID,
	}

	// parse nft
	for _, att := range nft.Attributes {
		if att.TraitType != "Asset Type" {
			continue
		}

		switch att.Value {
		case string(passport.WarMachine):
			passport.ParseWarMachineNFT(nft, warMachineNFT)
		case string(passport.Weapon):
		case string(passport.Utility):
		}
	}

	// assign faction id
	warMachineNFT.FactionID = *user.FactionID

	// join the asset to the queue
	ac.API.SendToAllServerClient(&ServerClientMessage{
		Key: AssetQueueJoin,
		Payload: struct {
			WarMachineNFT *passport.WarMachineNFT `json:"warMachineNFT"`
		}{
			WarMachineNFT: warMachineNFT,
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
		Filter           *db.ListFilterRequest `json:"filter"`
		AssetType        string                `json:"assetType"`
		Archived         bool                  `json:"archived"`
		Search           string                `json:"search"`
		PageSize         int                   `json:"pageSize"`
		Page             int                   `json:"page"`
	} `json:"payload"`
}

// AssetListResponse is the response from get asset list
type AssetListResponse struct {
	Records []*passport.XsynNftMetadata `json:"records"`
	Total   int                         `json:"total"`
}

const HubKeyAssetList hub.HubCommandKey = "ASSET:LIST"

func (ctrlr *AssetController) AssetList(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	req := &AssetsUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		terror.Error(err)
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
		req.Payload.AssetType,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		terror.Error(err)
	}

	resp := &AssetListResponse{
		Total:   total,
		Records: assets,
	}

	reply(resp)

	return nil

}

// AssetUpdatedSubscribeRequest requests an update for an xsyn_nft_metadata
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
