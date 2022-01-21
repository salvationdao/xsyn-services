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
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// AssetController holds handlers for roles
type AssetController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewAssetController creates the role hub
func NewAssetController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *AssetController {
	assetHub := &AssetController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "role_hub"),
		API:  api,
	}

	// collections list
	// api.SecureCommand(HubKeyCollectionList, assetHub.CollectionListHandler)

	// assets list
	api.SecureCommand(HubKeyAssetList, assetHub.ListHandler)

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
	err = db.XsynNftMetadataInsert(ctx, tx, req.Payload.XsynNftMetadata)
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
	user, err := db.UserGet(ctx, ac.Conn, userID)
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
	user, err := db.UserGet(ctx, ac.Conn, userID)
	if err != nil {
		return terror.Error(err)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User need to join a faction")
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
		case passport.WarMachine:
			parseWarMachineNFT(nft, warMachineNFT)
		case passport.Weapon:
		case passport.Utility:
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

	return nil
}

type WarMachineAttField string

const (
	WarMachineAttFieldMaxHitPoint      WarMachineAttField = "Max Structure Hit Points"
	WarMachineAttFieldSpeed            WarMachineAttField = "Speed"
	WarMachineAttFieldPowerGrid        WarMachineAttField = "Power Grid"
	WarMachineAttFieldCPU              WarMachineAttField = "CPU"
	WarMachineAttFieldWeaponHardpoints WarMachineAttField = "Weapon Hardpoints"
	WarMachineAttFieldTurretHardpoints WarMachineAttField = "Turret Hardpoints"
	WarMachineAttFieldUtilitySlots     WarMachineAttField = "Utility Slots"
)

// parseWarMachineNFT
func parseWarMachineNFT(nft *passport.XsynNftMetadata, warMachineNFT *passport.WarMachineNFT) {
	warMachineNFT.TokenID = nft.TokenID
	warMachineNFT.Name = nft.Name
	warMachineNFT.Description = nft.Description
	warMachineNFT.ExternalUrl = nft.ExternalUrl
	warMachineNFT.Image = nft.Image
	warMachineNFT.Durability = nft.Durability

	for _, att := range nft.Attributes {
		switch att.TraitType {
		case string(WarMachineAttFieldMaxHitPoint):
			warMachineNFT.MaxHitPoint = int(att.Value.(float64))
			warMachineNFT.RemainHitPoint = int(att.Value.(float64))
		case string(WarMachineAttFieldSpeed):
			warMachineNFT.Speed = int(att.Value.(float64))
		case string(WarMachineAttFieldPowerGrid):
			warMachineNFT.PowerGrid = int(att.Value.(float64))
		case string(WarMachineAttFieldCPU):
			warMachineNFT.CPU = int(att.Value.(float64))
		case string(WarMachineAttFieldWeaponHardpoints):
			warMachineNFT.WeaponHardpoint = int(att.Value.(float64))
		case string(WarMachineAttFieldTurretHardpoints):
			warMachineNFT.TurretHardpoint = int(att.Value.(float64))
		case string(WarMachineAttFieldUtilitySlots):
			warMachineNFT.UtilitySlots = int(att.Value.(float64))
		}
	}
}

// rootHub.SecureCommand(HubKeyAssetList, AssetController.GetHandler)
const HubKeyAssetList hub.HubCommandKey = "ASSET:LIST"

// AssetListHandlerRequest requests holds the filter for user list
type AssetListHandlerRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir          `json:"sortDir"`
		SortBy   db.AssetColumn        `json:"sortBy"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Archived bool                  `json:"archived"`
		Search   string                `json:"search"`
		PageSize int                   `json:"pageSize"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// AssetListResponse is the response from get asset list
type AssetListResponse struct {
	Records []*passport.Asset `json:"records"`
	Total   int               `json:"total"`
}

// ListHandler list assets with pagination
func (ctrlr *AssetController) ListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Something went wrong, please try again."
	req := &AssetListHandlerRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	fmt.Println("this is req", req.Payload.Filter)
	assets := []*passport.Asset{}
	total, err := db.AssetList(
		ctx, ctrlr.Conn, &assets,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.Filter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := &AssetListResponse{
		Total:   total,
		Records: assets,
	}

	reply(resp)

	return nil
}

// CollectionListHandler list collections with pagination
func (ctrlr *AssetController) CollectionListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	return nil
}
