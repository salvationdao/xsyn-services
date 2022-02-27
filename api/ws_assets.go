package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"math/big"
	"passport"
	"passport/db"

	"github.com/ninja-software/log_helpers"

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
	api.Command(HubKeyAssetList, assetHub.AssetListHandler)

	// asset subscribe
	api.SubscribeCommand(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyAssetDurabilitySubscribe, assetHub.AssetDurabilityUpdatedSubscribeHandler)

	// asset set name
	api.SecureCommand(HubKeyAssetUpdateName, assetHub.AssetUpdateNameHandler)

	api.SecureCommand(HubKeyAssetQueueJoin, assetHub.JoinQueueHandler)
	api.SecureCommand(HubKeyAssetQueueLeave, assetHub.LeaveQueueHandler)
	api.SecureCommand(HubKeyAssetInsurancePay, assetHub.PayAssetInsuranceHandler)
	api.SecureUserSubscribeCommand(HubKeyAssetQueueContractReward, assetHub.AssetQueueContractRewardSubscriber)

	return assetHub
}

// AssetQueueRequest contain the asset token id that user want to join/leave the battle queue
type AssetQueueRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"assetHash"`
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

	metadata, err := db.XsynMetadataOwnerGet(ctx, ac.Conn, userID, req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err)
	}

	if metadata.LockedByID != nil && !metadata.LockedByID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "Current asset is locked")
	}

	warMachineMetadata := &passport.WarMachineMetadata{
		Hash:      req.Payload.AssetHash,
		OwnedByID: userID,
		FactionID: *user.FactionID,
	}

	// release the asset from the queue
	ac.API.SendToAllServerClient(ctx, &ServerClientMessage{
		Key: AssetQueueLeave,
		Payload: struct {
			WarMachineMetadata *passport.WarMachineMetadata `json:"warMachineMetadata"`
		}{
			WarMachineMetadata: warMachineMetadata,
		},
	})

	reply(true)
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

	if !asset.IsUsable() {
		return terror.Error(fmt.Errorf("asset is locked"), "Asset is locked.")
	}

	if asset.Durability < 100 {
		return terror.Warn(fmt.Errorf("current assets durability is low"), "Current asset's durability is low.")
	}

	warMachineMetadata := &passport.WarMachineMetadata{
		OwnedByID:      userID,
		ContractReward: *big.NewInt(0),
	}

	// get current faction contract reward
	contractRewardChan := make(chan big.Int)
	if _, ok := ac.API.factionWarMachineContractMap[*user.FactionID]; ok {
		ac.API.factionWarMachineContractMap[*user.FactionID] <- func(wmc *WarMachineContract) {
			contractRewardChan <- wmc.CurrentReward
		}
		warMachineMetadata.ContractReward = <-contractRewardChan
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
	// TODO: commented out by vinnie 25/02/2022, no mech specific abilities exist yet
	//if len(warMachineMetadata.Abilities) > 0 {
	// get abilities asset
	//for _, abilityMetadata := range warMachineMetadata.Abilities {
	//	err := db.AbilityAssetGet(ctx, ac.Conn, abilityMetadata)
	//	if err != nil {
	//		return terror.Error(err)
	//	}
	//	if asset == nil {
	//		return terror.Error(fmt.Errorf("asset doesn't exist"))
	//	}
	//
	//	supsCost, err := db.WarMachineAbilityCostGet(ctx, ac.Conn, warMachineMetadata.Hash, abilityMetadata.TokenID)
	//	if err != nil {
	//		return terror.Error(err)
	//	}
	//
	//	abilityMetadata.SupsCost = supsCost
	//}
	//}
	// get faction from id
	f, err := db.FactionGet(ctx, ac.Conn, *user.FactionID)
	if err != nil {
		return terror.Error(err)
	}

	// assign faction id
	warMachineMetadata.FactionID = *user.FactionID
	warMachineMetadata.Faction = f

	// join the asset to the queue
	ac.API.SendToAllServerClient(ctx, &ServerClientMessage{
		Key: AssetQueueJoin,
		Payload: struct {
			WarMachineMetadata *passport.WarMachineMetadata `json:"warMachineMetadata"`
		}{
			WarMachineMetadata: warMachineMetadata,
		},
	})

	reply(true)
	return nil
}

type AssetsInsurancePayRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetHash string `json:"assetHash"`
	} `json:"payload"`
}

const HubKeyAssetInsurancePay hub.HubCommandKey = "ASSET:INSURANCE:PAY"

func (ac *AssetController) PayAssetInsuranceHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AssetsInsurancePayRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
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
		return terror.Error(terror.ErrInvalidInput, "User needs to join a faction to Deploy War Machine")
	}

	// check user own this asset and it has not joined the queue yet
	metadata, err := db.XsynMetadataOwnerGet(ctx, ac.Conn, userID, req.Payload.AssetHash)
	if err != nil {
		return terror.Error(err)
	}

	if metadata.FrozenAt == nil {
		return terror.Error(terror.ErrForbidden, "Error - current asset has not joined the queue")
	}

	if metadata.LockedByID != nil {
		return terror.Error(terror.ErrForbidden, "Error - current asset has already joined the battle ")
	}

	// fire request to server client
	ac.API.SendToAllServerClient(ctx, &ServerClientMessage{
		Key: AssetInsurancePay,
		Payload: struct {
			FactionID passport.FactionID `json:"factionID"`
			AssetHash string             `json:"assetHash"`
		}{
			FactionID: *user.FactionID,
			AssetHash: metadata.Hash,
		},
	})

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
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.ExternalTokenID)), nil
}

const HubKeyAssetDurabilitySubscribe hub.HubCommandKey = "ASSET:DURABILITY:SUBSCRIBE"

type DurabilityResponse struct {
	Durability int64      `json:"durability"`
	RepairType RepairType `json:"repairType"`
}

func (ac *AssetController) AssetDurabilityUpdatedSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &AssetUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return req.TransactionID, "", terror.Error(err)
	}

	// get asset durability from db
	durability, err := db.AssetDurabilityGet(ctx, ac.Conn, userID, req.Payload.AssetHash)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	// if durability less that 100
	if durability < 100 {
		// find the war machine in repair center
		check := &struct {
			isFound bool
		}{
			isFound: false,
		}
		ac.API.fastAssetRepairCenter <- func(rq RepairQueue) {
			if _, ok := rq[req.Payload.AssetHash]; ok {
				reply(&DurabilityResponse{
					Durability: durability,
					RepairType: RepairTypeFast,
				})
				check.isFound = true
				return
			}
			check.isFound = false
		}

		if !check.isFound {
			ac.API.standardAssetRepairCenter <- func(rq RepairQueue) {
				if _, ok := rq[req.Payload.AssetHash]; ok {
					reply(&DurabilityResponse{
						Durability: durability,
						RepairType: RepairTypeStandard,
					})
					return
				}
			}
		}
	} else {
		reply(&DurabilityResponse{
			Durability: durability,
			RepairType: "",
		})
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetDurabilitySubscribe, req.Payload.AssetHash)), nil
}

const HubKeyAssetQueueContractReward hub.HubCommandKey = "ASSET:QUEUE:CONTRACT:REWARD"

func (ac *AssetController) AssetQueueContractRewardSubscriber(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	// get user faction
	faction, err := db.FactionGetByUserID(ctx, ac.Conn, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", terror.Error(err)
	}

	if _, ok := ac.API.factionWarMachineContractMap[faction.ID]; ok {
		ac.API.factionWarMachineContractMap[faction.ID] <- func(wmc *WarMachineContract) {
			reply(wmc.CurrentReward.String())
		}
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetQueueContractReward, faction.ID)), nil
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

	ac.API.SendToAllServerClient(ctx, &ServerClientMessage{
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
