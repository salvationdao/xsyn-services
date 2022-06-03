package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/kevinms/leakybucket-go"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/supremacy_rpcclient"
	xsynTypes "xsyn-services/types"

	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ninja-software/log_helpers"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// AssetController holds handlers for as
type AssetController struct {
	Log *zerolog.Logger
	API *API
}

// NewAssetController creates the asset hub
func NewAssetController(log *zerolog.Logger, api *API) *AssetController {
	assetHub := &AssetController{
		Log: log_helpers.NamedLogger(log, "asset_hub"),
		API: api,
	}

	//const HubKeyAssetTransferFromSupremacy = "ASSET:TRANSFER:FROM:SUPREMACY"
	//AssetTransferFromSupremacyHandler

	// assets list
	api.SecureCommand(HubKeyAssetList, assetHub.AssetList721Handler)
	api.SecureCommand(HubKey1155AssetList, assetHub.AssetList1155Handler)
	api.SecureCommand(HubKeyAssetTransferToSupremacy, assetHub.AssetTransferToSupremacyHandler)
	api.SecureCommand(HubKeyAssetTransferFromSupremacy, assetHub.AssetTransferFromSupremacyHandler)
	api.SecureCommand(HubKeyAsset1155TransferToSupremacy, assetHub.Asset1155TransferToSupremacyHandler)
	api.SecureCommand(HubKeyAsset1155TransferFromSupremacy, assetHub.Asset1155TransferFromSupremacyHandler)
	api.SecureCommand(HubKeyDeposit1155Asset, assetHub.DepositAsset1155Handler)
	api.SecureCommand(HubKeyDepositAsset1155List, assetHub.DepositAsset1155ListHandler)
	api.Command(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)
	api.Command(HubKeyAsset1155Subscribe, assetHub.Asset1155UpdatedSubscribeHandler)

	return assetHub
}

type AssetListRequest struct {
	Payload struct {
		UserID          xsynTypes.UserID           `json:"user_id"`
		Sort            *db.ListSortRequest        `json:"sort,omitempty"`
		Filter          *db.ListFilterRequest      `json:"filter,omitempty"`
		AttributeFilter *db.AttributeFilterRequest `json:"attribute_filter,omitempty"`
		AssetType       string                     `json:"asset_type"`
		Search          string                     `json:"search"`
		PageSize        int                        `json:"page_size"`
		Page            int                        `json:"page"`
	} `json:"payload"`
}

// AssetListResponse is the response from get asset list
type AssetListResponse struct {
	Total  int64                  `json:"total"`
	Assets []*xsynTypes.UserAsset `json:"assets"` // TODO: create api type for user assets
}

const HubKeyAssetList = "ASSET:LIST:721"

func (ac *AssetController) AssetList721Handler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	total, assets, err := db.AssetList721(&db.AssetListOpts{
		UserID:          req.Payload.UserID,
		Sort:            req.Payload.Sort,
		Filter:          req.Payload.Filter,
		AttributeFilter: req.Payload.AttributeFilter,
		AssetType:       "mech", // for now this is hardcoded to hide all the other assets
		Search:          req.Payload.Search,
		PageSize:        req.Payload.PageSize,
		Page:            req.Payload.Page,
	})
	if err != nil {
		return terror.Error(err, "Unable to retrieve assets at this time, please try again or contact support.")
	}

	reply(&AssetListResponse{
		Total:  total,
		Assets: assets,
	})
	return nil
}

// Asset1155ListResponse is the response from get asset list
type Asset1155ListResponse struct {
	Total  int64                      `json:"total"`
	Assets []*xsynTypes.User1155Asset `json:"assets"` // TODO: create api type for user assets
}

const HubKey1155AssetList = "ASSET:LIST:1155"

func (ac *AssetController) AssetList1155Handler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	total, assets, err := db.AssetList1155(&db.AssetListOpts{
		UserID:          req.Payload.UserID,
		Sort:            req.Payload.Sort,
		Filter:          req.Payload.Filter,
		AttributeFilter: req.Payload.AttributeFilter,
		AssetType:       req.Payload.AssetType,
		Search:          req.Payload.Search,
		PageSize:        req.Payload.PageSize,
		Page:            req.Payload.Page,
	})
	if err != nil {
		return terror.Error(err, "Unable to retrieve assets at this time, please try again or contact support.")
	}

	reply(&Asset1155ListResponse{
		Total:  total,
		Assets: assets,
	})
	return nil
}

// AssetUpdatedSubscribeRequest requests an update for an xsyn_metadata
type AssetUpdatedSubscribeRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

type AssetUpdatedSubscribeResponse struct {
	CollectionSlug string                    `json:"collection_slug"`
	PurchasedItem  *boiler.PurchasedItemsOld `json:"purchased_item"`
	OwnerUsername  string                    `json:"owner_username"`
	HostURL        string                    `json:"host_url"`
}

type UserAsset struct {
	ID                  string      `json:"id"`
	CollectionID        string      `json:"collection_id"`
	TokenID             int64       `json:"token_id"`
	Tier                string      `json:"tier"`
	Hash                string      `json:"hash"`
	OwnerID             string      `json:"owner_id"`
	Data                types.JSON  `json:"data"`
	Attributes          types.JSON  `json:"attributes"`
	Name                string      `json:"name"`
	AssetType           null.String `json:"asset_type,omitempty"`
	ImageURL            null.String `json:"image_url,omitempty"`
	ExternalURL         null.String `json:"external_url,omitempty"`
	CardAnimationURL    null.String `json:"card_animation_url,omitempty"`
	AvatarURL           null.String `json:"avatar_url,omitempty"`
	LargeImageURL       null.String `json:"large_image_url,omitempty"`
	Description         null.String `json:"description,omitempty"`
	BackgroundColor     null.String `json:"background_color,omitempty"`
	AnimationURL        null.String `json:"animation_url,omitempty"`
	YoutubeURL          null.String `json:"youtube_url,omitempty"`
	UnlockedAt          time.Time   `json:"unlocked_at"`
	MintedAt            null.Time   `json:"minted_at,omitempty"`
	OnChainStatus       string      `json:"on_chain_status"`
	XsynLocked          null.Bool   `json:"xsyn_locked,omitempty"`
	DeletedAt           null.Time   `json:"deleted_at,omitempty"`
	DataRefreshedAt     time.Time   `json:"data_refreshed_at"`
	UpdatedAt           time.Time   `json:"updated_at"`
	CreatedAt           time.Time   `json:"created_at"`
	LockedToServiceName null.String `json:"locked_to_service_name,omitempty"`
}

type Collection struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	LogoBlobID    null.String `json:"logo_blob_id,omitempty"`
	Keywords      null.String `json:"keywords,omitempty"`
	DeletedAt     null.Time   `json:"deleted_at,omitempty"`
	UpdatedAt     time.Time   `json:"updated_at"`
	CreatedAt     time.Time   `json:"created_at"`
	Slug          string      `json:"slug"`
	MintContract  null.String `json:"mint_contract,omitempty"`
	StakeContract null.String `json:"stake_contract,omitempty"`
	IsVisible     null.Bool   `json:"is_visible,omitempty"`
	ContractType  null.String `json:"contract_type,omitempty"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type AssetResponse struct {
	UserAsset  *UserAsset  `json:"user_asset"`
	Collection *Collection `json:"collection"`
	Owner      *User       `json:"owner"`
}

const HubKeyAssetSubscribe = "ASSET:GET:721"

func (ac *AssetController) AssetUpdatedSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	// errMsg := "Issue subscribing to asset updates, try again or contact support."
	req := &AssetUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userAsset, err := boiler.UserAssets(
		boiler.UserAssetWhere.Hash.EQ(req.Payload.AssetHash),
		qm.Load(boiler.UserAssetRels.Collection),
		qm.Load(
			boiler.UserAssetRels.Owner,
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		),
		qm.Load(boiler.UserAssetRels.LockedToServiceUser),
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user asset from db")
	}

	serviceName := null.NewString("", false)
	if userAsset.R.LockedToServiceUser != nil {
		serviceName = null.StringFrom(userAsset.R.LockedToServiceUser.Username)
	}

	reply(&AssetResponse{
		UserAsset: &UserAsset{
			ID:                  userAsset.ID,
			CollectionID:        userAsset.CollectionID,
			TokenID:             userAsset.TokenID,
			Tier:                userAsset.Tier,
			Hash:                userAsset.Hash,
			OwnerID:             userAsset.OwnerID,
			Data:                userAsset.Data,
			Attributes:          userAsset.Attributes,
			Name:                userAsset.Name,
			ImageURL:            userAsset.ImageURL,
			AssetType:           userAsset.AssetType,
			ExternalURL:         userAsset.ExternalURL,
			CardAnimationURL:    userAsset.CardAnimationURL,
			AvatarURL:           userAsset.AvatarURL,
			LargeImageURL:       userAsset.LargeImageURL,
			Description:         userAsset.Description,
			BackgroundColor:     userAsset.BackgroundColor,
			AnimationURL:        userAsset.AnimationURL,
			YoutubeURL:          userAsset.YoutubeURL,
			UnlockedAt:          userAsset.UnlockedAt,
			MintedAt:            userAsset.MintedAt,
			OnChainStatus:       userAsset.OnChainStatus,
			DeletedAt:           userAsset.DeletedAt,
			DataRefreshedAt:     userAsset.DataRefreshedAt,
			UpdatedAt:           userAsset.UpdatedAt,
			CreatedAt:           userAsset.CreatedAt,
			LockedToServiceName: serviceName,
		},
		Collection: &Collection{
			ID:            userAsset.R.Collection.ID,
			Name:          userAsset.R.Collection.Name,
			LogoBlobID:    userAsset.R.Collection.LogoBlobID,
			Keywords:      userAsset.R.Collection.Keywords,
			DeletedAt:     userAsset.R.Collection.DeletedAt,
			UpdatedAt:     userAsset.R.Collection.UpdatedAt,
			CreatedAt:     userAsset.R.Collection.CreatedAt,
			Slug:          userAsset.R.Collection.Slug,
			MintContract:  userAsset.R.Collection.MintContract,
			StakeContract: userAsset.R.Collection.StakeContract,
			IsVisible:     userAsset.R.Collection.IsVisible,
			ContractType:  userAsset.R.Collection.ContractType,
		},
		Owner: &User{
			ID:       userAsset.R.Owner.ID,
			Username: userAsset.R.Owner.Username,
		},
	})
	return nil
}

// Asset1155UpdatedSubscribeRequest requests an update for an xsyn_metadata
type Asset1155UpdatedSubscribeRequest struct {
	Payload struct {
		CollectionSlug string `json:"collection_slug"`
		Locked         bool   `json:"locked"`
		TokenID        int    `json:"token_id"`
		OwnerID        string `json:"owner_id"`
	} `json:"payload"`
}

type User1155AssetView struct {
	ID                  string      `json:"id"`
	OwnerID             string      `json:"owner_id"`
	ExternalTokenID     int         `json:"external_token_id"`
	Count               int         `json:"count"`
	Label               string      `json:"label"`
	Description         string      `json:"description"`
	ImageURL            string      `json:"image_url"`
	AnimationURL        null.String `json:"animation_url"`
	Attributes          types.JSON  `json:"attributes"`
	ServiceNameLockedIn null.String `json:"service_name_locked_in"`
}

type Asset1155Response struct {
	UserAsset  *User1155AssetView `json:"user_asset"`
	Collection *Collection        `json:"collection"`
	Owner      *User              `json:"owner"`
}

const HubKeyAsset1155Subscribe = "ASSET:GET:1155"

func (ac *AssetController) Asset1155UpdatedSubscribeHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	// errMsg := "Issue subscribing to asset updates, try again or contact support."
	req := &Asset1155UpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	queries := []qm.QueryMod{
		boiler.UserAssets1155Where.OwnerID.EQ(req.Payload.OwnerID),
		boiler.UserAssets1155Where.ExternalTokenID.EQ(req.Payload.TokenID),
		qm.Load(boiler.UserAssets1155Rels.Collection),
		qm.Load(boiler.UserAssets1155Rels.Owner),
	}

	if req.Payload.Locked {
		queries = append(queries, boiler.UserAssets1155Where.ServiceID.IsNotNull())
	} else {
		queries = append(queries, boiler.UserAssets1155Where.ServiceID.IsNull())
	}

	userAsset, err := boiler.UserAssets1155S(
		queries...,
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user asset from db")
	}

	serviceName := null.NewString("", false)
	if userAsset.ServiceID.Valid {
		service, err := boiler.Users(
			boiler.UserWhere.ID.EQ(userAsset.ServiceID.String),
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		).One(passdb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to get service name")
		}

		serviceName = null.StringFrom(service.Username)
	}

	reply(&Asset1155Response{
		UserAsset: &User1155AssetView{
			ID:                  userAsset.ID,
			OwnerID:             userAsset.OwnerID,
			ExternalTokenID:     userAsset.ExternalTokenID,
			Count:               userAsset.Count,
			Label:               userAsset.Label,
			Description:         userAsset.Description,
			ImageURL:            userAsset.ImageURL,
			AnimationURL:        userAsset.AnimationURL,
			Attributes:          userAsset.Attributes,
			ServiceNameLockedIn: serviceName,
		},
		Collection: &Collection{
			ID:            userAsset.R.Collection.ID,
			Name:          userAsset.R.Collection.Name,
			LogoBlobID:    userAsset.R.Collection.LogoBlobID,
			Keywords:      userAsset.R.Collection.Keywords,
			DeletedAt:     userAsset.R.Collection.DeletedAt,
			UpdatedAt:     userAsset.R.Collection.UpdatedAt,
			CreatedAt:     userAsset.R.Collection.CreatedAt,
			Slug:          userAsset.R.Collection.Slug,
			MintContract:  userAsset.R.Collection.MintContract,
			StakeContract: userAsset.R.Collection.StakeContract,
			IsVisible:     userAsset.R.Collection.IsVisible,
			ContractType:  userAsset.R.Collection.ContractType,
		},
		Owner: &User{
			ID:       userAsset.R.Owner.ID,
			Username: userAsset.R.Owner.Username,
		},
	})
	return nil
}

type AssetTransferToSupremacyRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

const HubKeyAssetTransferToSupremacy = "ASSET:TRANSFER:TO:SUPREMACY"

func (ac *AssetController) AssetTransferToSupremacyHandler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetTransferToSupremacyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userAsset, err := boiler.UserAssets(
		boiler.UserAssetWhere.Hash.EQ(req.Payload.AssetHash),
		qm.Load(boiler.UserAssetRels.Collection),
		qm.Load(
			boiler.UserAssetRels.Owner,
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		),
		qm.Load(boiler.UserAssetRels.LockedToServiceUser),
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user asset from db")
	}

	if userAsset.OnChainStatus == "STAKABLE" {
		return terror.Error(fmt.Errorf("trying to transfer unstaked asset to supremacy"), "Asset needs to be On-World before being able to transfer to Supremacy.")
	}

	if userAsset.OwnerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "You don't own this asset.")
	}

	// pay 5 sups
	userUUID := uuid.Must(uuid.FromString(user.ID))

	tx := &xsynTypes.NewTransaction{
		From:                 xsynTypes.UserID(userUUID),
		To:                   xsynTypes.XsynTreasuryUserID,
		TransactionReference: xsynTypes.TransactionReference(fmt.Sprintf("asset_transfer_fee|%s|%s|%d", "SUPREMACY", req.Payload.AssetHash, time.Now().UnixNano())),
		Description:          fmt.Sprintf("Transfer of asset: %s to Supremacy", req.Payload.AssetHash),
		Amount:               decimal.New(5, 18), // 5 sups
		Group:                xsynTypes.TransactionGroupAssetManagement,
		SubGroup:             "Transfer",
		NotSafe:              true,
	}

	_, _, txID, err := ac.API.userCacheMap.Transact(tx)
	if err != nil {
		return terror.Error(err, "failed to process sups")
	}

	transferLog := &boiler.AssetServiceTransferEvent{
		UserAssetID:   userAsset.ID,
		UserID:        userAsset.OwnerID,
		InitiatedFrom: "XSYN",
		ToService:     null.StringFrom(xsynTypes.SupremacyGameUserID.String()),
		TransferTXID:  txID,
	}
	err = transferLog.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to transfer asset to supremacy")
	}

	marketLocked := true
	if userAsset.OnChainStatus != "UNSTAKABLE_OLD" {
		marketLocked = false
	}

	err = supremacy_rpcclient.AssetLockToSupremacy(xsynTypes.UserAssetFromBoiler(userAsset), transferLog.ID, marketLocked)
	if err != nil {
		_, _ = reverseAssetServiceTransaction(
			ac.API.userCacheMap,
			tx,
			transferLog,
			"Failed to transfer asset to supremacy")
		return terror.Error(err, "Failed to transfer asset to supremacy")
	}

	userAsset.LockedToService = null.StringFrom(xsynTypes.SupremacyGameUserID.String())
	_, err = userAsset.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		reverseTransaction, reverseTransfer := reverseAssetServiceTransaction(
			ac.API.userCacheMap,
			tx,
			transferLog,
			"Failed to transfer asset to supremacy")
		err = supremacy_rpcclient.AssetUnlockFromSupremacy(xsynTypes.UserAssetFromBoiler(userAsset), reverseTransfer.ID)
		if err != nil {
			_, _ = reverseAssetServiceTransaction(
				ac.API.userCacheMap,
				reverseTransaction,
				reverseTransfer,
				"Failed to transfer asset to supremacy")
			passlog.L.Error().Err(err).Msg("failed to unlock asset from supremacy after we failed to update asset lock")
			return terror.Error(err, "Failed to transfer asset to supremacy")
		}
		return terror.Error(err, "Failed to transfer asset to supremacy")
	}

	reply(&AssetResponse{
		UserAsset: &UserAsset{
			ID:                  userAsset.ID,
			CollectionID:        userAsset.CollectionID,
			TokenID:             userAsset.TokenID,
			Tier:                userAsset.Tier,
			Hash:                userAsset.Hash,
			OwnerID:             userAsset.OwnerID,
			Data:                userAsset.Data,
			Attributes:          userAsset.Attributes,
			Name:                userAsset.Name,
			ImageURL:            userAsset.ImageURL,
			ExternalURL:         userAsset.ExternalURL,
			AssetType:           userAsset.AssetType,
			CardAnimationURL:    userAsset.CardAnimationURL,
			AvatarURL:           userAsset.AvatarURL,
			LargeImageURL:       userAsset.LargeImageURL,
			Description:         userAsset.Description,
			BackgroundColor:     userAsset.BackgroundColor,
			AnimationURL:        userAsset.AnimationURL,
			YoutubeURL:          userAsset.YoutubeURL,
			UnlockedAt:          userAsset.UnlockedAt,
			MintedAt:            userAsset.MintedAt,
			OnChainStatus:       userAsset.OnChainStatus,
			DeletedAt:           userAsset.DeletedAt,
			DataRefreshedAt:     userAsset.DataRefreshedAt,
			UpdatedAt:           userAsset.UpdatedAt,
			CreatedAt:           userAsset.CreatedAt,
			LockedToServiceName: null.StringFrom("Supremacy"),
		},
		Collection: &Collection{
			ID:            userAsset.R.Collection.ID,
			Name:          userAsset.R.Collection.Name,
			LogoBlobID:    userAsset.R.Collection.LogoBlobID,
			Keywords:      userAsset.R.Collection.Keywords,
			DeletedAt:     userAsset.R.Collection.DeletedAt,
			UpdatedAt:     userAsset.R.Collection.UpdatedAt,
			CreatedAt:     userAsset.R.Collection.CreatedAt,
			Slug:          userAsset.R.Collection.Slug,
			MintContract:  userAsset.R.Collection.MintContract,
			StakeContract: userAsset.R.Collection.StakeContract,
			IsVisible:     userAsset.R.Collection.IsVisible,
			ContractType:  userAsset.R.Collection.ContractType,
		},
		Owner: &User{
			ID:       userAsset.R.Owner.ID,
			Username: userAsset.R.Owner.Username,
		},
	})
	return nil
}

type AssetTransferFromSupremacyRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

const HubKeyAssetTransferFromSupremacy = "ASSET:TRANSFER:FROM:SUPREMACY"

func (ac *AssetController) AssetTransferFromSupremacyHandler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetTransferFromSupremacyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userAsset, err := boiler.UserAssets(
		boiler.UserAssetWhere.Hash.EQ(req.Payload.AssetHash),
		qm.Load(boiler.UserAssetRels.Collection),
		qm.Load(
			boiler.UserAssetRels.Owner,
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		),
		qm.Load(boiler.UserAssetRels.LockedToServiceUser),
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user asset from db")
	}

	if userAsset.OwnerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "You don't own this asset.")
	}

	// pay 5 sups
	userUUID := uuid.Must(uuid.FromString(user.ID))

	tx := &xsynTypes.NewTransaction{
		From:                 xsynTypes.UserID(userUUID),
		To:                   xsynTypes.XsynTreasuryUserID,
		TransactionReference: xsynTypes.TransactionReference(fmt.Sprintf("asset_transfer_fee|%s|%s|%d", "XSYN", req.Payload.AssetHash, time.Now().UnixNano())),
		Description:          fmt.Sprintf("Transfer of asset: %s to XSYN", req.Payload.AssetHash),
		Amount:               decimal.New(5, 18), // 5 sups
		Group:                xsynTypes.TransactionGroupAssetManagement,
		SubGroup:             "Transfer",
		NotSafe:              true,
	}

	_, _, txID, err := ac.API.userCacheMap.Transact(tx)
	if err != nil {
		return err
	}

	transferLog := &boiler.AssetServiceTransferEvent{
		UserAssetID:   userAsset.ID,
		UserID:        userAsset.OwnerID,
		InitiatedFrom: "XSYN",
		FromService:   null.StringFrom(xsynTypes.SupremacyGameUserID.String()),
		TransferTXID:  txID,
	}
	err = transferLog.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to transfer asset to XSYN")
	}

	err = supremacy_rpcclient.AssetUnlockFromSupremacy(xsynTypes.UserAssetFromBoiler(userAsset), transferLog.ID)
	if err != nil {
		// TODO: in the future we should forcibly be able to pull assets from services back to xsyn
		reverseAssetServiceTransaction(
			ac.API.userCacheMap,
			tx,
			transferLog,
			"Failed to transfer asset from supremacy",
		)
		return terror.Error(err, "Failed to transfer asset from supremacy")
	}

	userAsset.LockedToService = null.NewString("", false)
	_, err = userAsset.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		reverseTransaction, reverseTransferLog := reverseAssetServiceTransaction(
			ac.API.userCacheMap,
			tx,
			transferLog,
			"Failed to transfer asset from supremacy",
		)
		marketLocked := true
		if userAsset.OnChainStatus != "UNSTAKABLE_OLD" {
			marketLocked = false
		}

		err = supremacy_rpcclient.AssetLockToSupremacy(xsynTypes.UserAssetFromBoiler(userAsset), reverseTransferLog.ID, marketLocked)
		if err != nil {
			_, _ = reverseAssetServiceTransaction(
				ac.API.userCacheMap,
				reverseTransaction,
				reverseTransferLog,
				"Failed to transfer asset from supremacy",
			)
			passlog.L.Error().Err(err).Msg("failed to lock asset from supremacy after we failed to update asset lock")
			return terror.Error(err, "Failed to transfer asset from supremacy")
		}
		return terror.Error(err, "Failed to transfer asset from supremacy")
	}

	reply(&AssetResponse{
		UserAsset: &UserAsset{
			ID:                  userAsset.ID,
			CollectionID:        userAsset.CollectionID,
			TokenID:             userAsset.TokenID,
			Tier:                userAsset.Tier,
			Hash:                userAsset.Hash,
			OwnerID:             userAsset.OwnerID,
			Data:                userAsset.Data,
			Attributes:          userAsset.Attributes,
			Name:                userAsset.Name,
			ImageURL:            userAsset.ImageURL,
			ExternalURL:         userAsset.ExternalURL,
			AssetType:           userAsset.AssetType,
			CardAnimationURL:    userAsset.CardAnimationURL,
			AvatarURL:           userAsset.AvatarURL,
			LargeImageURL:       userAsset.LargeImageURL,
			Description:         userAsset.Description,
			BackgroundColor:     userAsset.BackgroundColor,
			AnimationURL:        userAsset.AnimationURL,
			YoutubeURL:          userAsset.YoutubeURL,
			UnlockedAt:          userAsset.UnlockedAt,
			MintedAt:            userAsset.MintedAt,
			OnChainStatus:       userAsset.OnChainStatus,
			DeletedAt:           userAsset.DeletedAt,
			DataRefreshedAt:     userAsset.DataRefreshedAt,
			UpdatedAt:           userAsset.UpdatedAt,
			CreatedAt:           userAsset.CreatedAt,
			LockedToServiceName: null.NewString("", false),
		},
		Collection: &Collection{
			ID:            userAsset.R.Collection.ID,
			Name:          userAsset.R.Collection.Name,
			LogoBlobID:    userAsset.R.Collection.LogoBlobID,
			Keywords:      userAsset.R.Collection.Keywords,
			DeletedAt:     userAsset.R.Collection.DeletedAt,
			UpdatedAt:     userAsset.R.Collection.UpdatedAt,
			CreatedAt:     userAsset.R.Collection.CreatedAt,
			Slug:          userAsset.R.Collection.Slug,
			MintContract:  userAsset.R.Collection.MintContract,
			StakeContract: userAsset.R.Collection.StakeContract,
			IsVisible:     userAsset.R.Collection.IsVisible,
			ContractType:  userAsset.R.Collection.ContractType,
		},
		Owner: &User{
			ID:       userAsset.R.Owner.ID,
			Username: userAsset.R.Owner.Username,
		},
	})
	return nil
}

type Asset1155TransferToSupremacyRequest struct {
	Payload struct {
		TokenID        int    `json:"token_id"`
		Amount         int    `json:"amount"`
		CollectionSlug string `json:"collection_slug"`
	} `json:"payload"`
}

const HubKeyAsset1155TransferToSupremacy = "ASSET:1155:TRANSFER:TO:SUPREMACY"

var TransferBucket = leakybucket.NewCollector(0.5, 1, true)

func (ac *AssetController) Asset1155TransferToSupremacyHandler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &Asset1155TransferToSupremacyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	b := TransferBucket.Add(fmt.Sprintf("%s_%s_%d", user.ID, req.Payload.CollectionSlug, req.Payload.TokenID), 1)
	if b == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many request made for transfer")
	}

	asset, err := boiler.UserAssets1155S(
		boiler.UserAssets1155Where.ExternalTokenID.EQ(req.Payload.TokenID),
		boiler.UserAssets1155Where.OwnerID.EQ(user.ID),
		boiler.UserAssets1155Where.ServiceID.IsNull(),
		qm.Load(boiler.UserAssets1155Rels.Collection),
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user 1155 asset")
	}
	if asset.OwnerID != user.ID {
		return terror.Error(terror.ErrUnauthorised, "You don't own this asset.")
	}
	if asset.Count-req.Payload.Amount < 0 {
		return terror.Error(fmt.Errorf("asset count below 0 after transfer"), "Cannot process transfer. Amount after transfer is below 0")
	}

	userUUID := uuid.Must(uuid.FromString(user.ID))

	tx := &xsynTypes.NewTransaction{
		From:                 xsynTypes.UserID(userUUID),
		To:                   xsynTypes.XsynTreasuryUserID,
		TransactionReference: xsynTypes.TransactionReference(fmt.Sprintf("asset_transfer_fee|%s|%s|%d", "XSYN", req.Payload.TokenID, time.Now().UnixNano())),
		Description:          fmt.Sprintf("Transfer of asset: %s to Supremacy", asset.Label),
		Amount:               decimal.New(5, 18), // 5 sups
		Group:                xsynTypes.TransactionGroupAssetManagement,
		SubGroup:             "Transfer",
		NotSafe:              true,
	}

	_, _, txID, err := ac.API.userCacheMap.Transact(tx)
	if err != nil {
		return terror.Error(err, "Failed to process asset transfer transaction")
	}

	transferLog := &boiler.Asset1155ServiceTransferEvent{
		User1155AssetID: asset.ID,
		UserID:          asset.OwnerID,
		Amount:          req.Payload.Amount,
		ToService:       null.StringFrom(xsynTypes.SupremacyGameUserID.String()),
		TransferTXID:    txID,
	}

	err = transferLog.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to log transfer")
	}

	err = supremacy_rpcclient.KeycardsTransferToSupremacy(xsynTypes.UserAsset1155FromBoiler(asset), transferLog.ID, req.Payload.Amount)
	if err != nil {
		reverseAsset1155ServiceTransaction(
			ac.API.userCacheMap,
			tx,
			transferLog,
			"Failed to transfer asset to supremacy",
		)

		return terror.Error(err, "Failed to transfer asset to supremacy")
	}

	asset.Count -= req.Payload.Amount

	_, err = asset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.Count))
	if err != nil {
		passlog.L.Error().Err(err).Str("asset.id", asset.ID).Str("asset.owenr_id", asset.OwnerID).Msg("Failed to update asset count during transfer")
		return terror.Error(err, "Failed to update asset count during transfer")
	}

	offXsynAsset, err := boiler.UserAssets1155S(
		boiler.UserAssets1155Where.ExternalTokenID.EQ(req.Payload.TokenID),
		boiler.UserAssets1155Where.OwnerID.EQ(user.ID),
		boiler.UserAssets1155Where.ServiceID.EQ(null.StringFrom(xsynTypes.SupremacyGameUserID.String())),
	).One(passdb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		offXsynAsset = &boiler.UserAssets1155{
			OwnerID:         asset.OwnerID,
			CollectionID:    asset.CollectionID,
			ExternalTokenID: asset.ExternalTokenID,
			Label:           asset.Label,
			Description:     asset.Description,
			ImageURL:        asset.ImageURL,
			AnimationURL:    asset.AnimationURL,
			KeycardGroup:    asset.KeycardGroup,
			Attributes:      asset.Attributes,
			ServiceID:       null.StringFrom(xsynTypes.SupremacyGameUserID.String()),
		}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get off xsyn asset")
	}

	offXsynAsset.Count += req.Payload.Amount

	_, err = offXsynAsset.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to update off xsyn asset")
	}

	// TODO reply new asset1155 count and status

	serviceName := null.NewString("", false)
	if asset.ServiceID.Valid {
		service, err := boiler.Users(
			boiler.UserWhere.ID.EQ(asset.ServiceID.String),
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		).One(passdb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to get service name")
		}

		serviceName = null.StringFrom(service.Username)
	}

	owner, err := asset.Owner().One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get owner data")
	}

	reply(&Asset1155Response{
		UserAsset: &User1155AssetView{
			ID:                  asset.ID,
			OwnerID:             asset.OwnerID,
			ExternalTokenID:     asset.ExternalTokenID,
			Count:               asset.Count,
			Label:               asset.Label,
			Description:         asset.Description,
			ImageURL:            asset.ImageURL,
			AnimationURL:        asset.AnimationURL,
			Attributes:          asset.Attributes,
			ServiceNameLockedIn: serviceName,
		},
		Collection: &Collection{
			ID:            asset.R.Collection.ID,
			Name:          asset.R.Collection.Name,
			LogoBlobID:    asset.R.Collection.LogoBlobID,
			Keywords:      asset.R.Collection.Keywords,
			DeletedAt:     asset.R.Collection.DeletedAt,
			UpdatedAt:     asset.R.Collection.UpdatedAt,
			CreatedAt:     asset.R.Collection.CreatedAt,
			Slug:          asset.R.Collection.Slug,
			MintContract:  asset.R.Collection.MintContract,
			StakeContract: asset.R.Collection.StakeContract,
			IsVisible:     asset.R.Collection.IsVisible,
			ContractType:  asset.R.Collection.ContractType,
		},
		Owner: &User{
			ID:       owner.ID,
			Username: owner.Username,
		},
	})

	return nil
}

const HubKeyAsset1155TransferFromSupremacy = "ASSET:1155:TRANSFER:FROM:SUPREMACY"

func (ac *AssetController) Asset1155TransferFromSupremacyHandler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &Asset1155TransferToSupremacyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	b := TransferBucket.Add(fmt.Sprintf("%s_%s_%d", user.ID, req.Payload.CollectionSlug, req.Payload.TokenID), 1)
	if b == 0 {
		return terror.Error(fmt.Errorf("too many requests"), "Too many request made for transfer")
	}

	asset, err := boiler.UserAssets1155S(
		boiler.UserAssets1155Where.ExternalTokenID.EQ(req.Payload.TokenID),
		boiler.UserAssets1155Where.OwnerID.EQ(user.ID),
		boiler.UserAssets1155Where.ServiceID.EQ(null.StringFrom(xsynTypes.SupremacyGameUserID.String())),
		qm.Load(boiler.UserAssets1155Rels.Collection),
		qm.Load(boiler.UserAssets1155Rels.Owner),
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user 1155 asset")
	}

	userUUID := uuid.Must(uuid.FromString(user.ID))

	tx := &xsynTypes.NewTransaction{
		From:                 xsynTypes.UserID(userUUID),
		To:                   xsynTypes.XsynTreasuryUserID,
		TransactionReference: xsynTypes.TransactionReference(fmt.Sprintf("asset_transfer_fee|%s|%s|%d", "XSYN", req.Payload.TokenID, time.Now().UnixNano())),
		Description:          fmt.Sprintf("Transfer of asset with token id of %d from collection %s to Xsyn", req.Payload.TokenID, req.Payload.CollectionSlug),
		Amount:               decimal.New(5, 18), // 5 sups
		Group:                xsynTypes.TransactionGroupAssetManagement,
		SubGroup:             "Transfer",
		NotSafe:              true,
	}

	_, _, txID, err := ac.API.userCacheMap.Transact(tx)
	if err != nil {
		return terror.Error(err, "Failed to process asset transfer transaction")
	}

	transferLog := &boiler.Asset1155ServiceTransferEvent{
		User1155AssetID: asset.ID,
		UserID:          asset.OwnerID,
		Amount:          req.Payload.Amount,
		TransferTXID:    txID,
	}

	err = transferLog.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to log transfer")
	}

	_, err = supremacy_rpcclient.KeycardsTransferFromSupremacy(xsynTypes.UserAsset1155FromBoiler(asset), transferLog.ID, req.Payload.Amount)
	if err != nil {
		reverseAsset1155ServiceTransaction(
			ac.API.userCacheMap,
			tx,
			transferLog,
			"Failed to transfer asset to XSYN",
		)

		return terror.Error(err, "Failed to transfer asset to XSYN")
	}

	asset.Count -= req.Payload.Amount

	_, err = asset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.Count))
	if err != nil {
		passlog.L.Error().Err(err).Str("asset.id", asset.ID).Str("asset.owner_id", asset.OwnerID).Msg("Failed to update asset count during transfer")
		return terror.Error(err, "Failed to update asset count during transfer")
	}

	onXsynAsset, err := boiler.UserAssets1155S(
		boiler.UserAssets1155Where.ExternalTokenID.EQ(req.Payload.TokenID),
		boiler.UserAssets1155Where.OwnerID.EQ(user.ID),
		boiler.UserAssets1155Where.ServiceID.IsNull(),
	).One(passdb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {
		onXsynAsset = &boiler.UserAssets1155{
			OwnerID:         asset.OwnerID,
			CollectionID:    asset.CollectionID,
			ExternalTokenID: asset.ExternalTokenID,
			Label:           asset.Label,
			Description:     asset.Description,
			ImageURL:        asset.ImageURL,
			AnimationURL:    asset.AnimationURL,
			KeycardGroup:    asset.KeycardGroup,
			Attributes:      asset.Attributes,
		}

		if err := onXsynAsset.Insert(passdb.StdConn, boil.Infer()); err != nil {
			return terror.Error(err, "Failed to insert new xsyn asset")
		}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, "Failed to get on XSYN asset")
	}

	onXsynAsset.Count += req.Payload.Amount

	_, err = onXsynAsset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.Count))
	if err != nil {
		return terror.Error(err, "Failed to update on XSYN asset")
	}

	// TODO reply new asset1155 count and status

	serviceName := null.NewString("", false)
	if asset.ServiceID.Valid {
		service, err := boiler.Users(
			boiler.UserWhere.ID.EQ(asset.ServiceID.String),
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		).One(passdb.StdConn)
		if err != nil {
			return terror.Error(err, "Failed to get service name")
		}

		serviceName = null.StringFrom(service.Username)
	}

	owner, err := asset.Owner().One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get owner info")
	}

	reply(&Asset1155Response{
		UserAsset: &User1155AssetView{
			ID:                  asset.ID,
			OwnerID:             asset.OwnerID,
			ExternalTokenID:     asset.ExternalTokenID,
			Count:               asset.Count,
			Label:               asset.Label,
			Description:         asset.Description,
			ImageURL:            asset.ImageURL,
			AnimationURL:        asset.AnimationURL,
			Attributes:          asset.Attributes,
			ServiceNameLockedIn: serviceName,
		},
		Collection: &Collection{
			ID:            asset.R.Collection.ID,
			Name:          asset.R.Collection.Name,
			LogoBlobID:    asset.R.Collection.LogoBlobID,
			Keywords:      asset.R.Collection.Keywords,
			DeletedAt:     asset.R.Collection.DeletedAt,
			UpdatedAt:     asset.R.Collection.UpdatedAt,
			CreatedAt:     asset.R.Collection.CreatedAt,
			Slug:          asset.R.Collection.Slug,
			MintContract:  asset.R.Collection.MintContract,
			StakeContract: asset.R.Collection.StakeContract,
			IsVisible:     asset.R.Collection.IsVisible,
			ContractType:  asset.R.Collection.ContractType,
		},
		Owner: &User{
			ID:       owner.ID,
			Username: owner.Username,
		},
	})

	return nil
}

type Asset115DepositRequest struct {
	Payload struct {
		TransactionHash string `json:"transaction_hash"`
		Amount          int    `json:"amount"`
		CollectionSlug  string `json:"collection_slug"`
		TokenID         int    `json:"token_id"`
	} `json:"payload"`
}

const HubKeyDeposit1155Asset = "ASSET:1155:DEPOSIT"

func (ac *AssetController) DepositAsset1155Handler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing 11555 asset transaction, try again or contact support."

	req := &Asset115DepositRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.TransactionHash == "" {
		passlog.L.Error().Str("func", "DepositAsset1155Handler").Msg("deposit transaction hash was not provided")
		return terror.Error(fmt.Errorf("transaction hash was not provided"), errMsg)
	}

	if req.Payload.Amount <= 0 {
		passlog.L.Error().Str("func", "DepositAsset1155Handler").Msg("deposit transaction amount is lower than the minimum required amount")
		return terror.Error(fmt.Errorf("deposit transaction amount is lower than the minimum required amount"), "Deposit transaction amount is lower than the minimum required amount.")
	}

	dtx := boiler.DepositAsset1155Transaction{
		UserID:         user.ID,
		TXHash:         req.Payload.TransactionHash,
		Amount:         req.Payload.Amount,
		CollectionSlug: req.Payload.CollectionSlug,
		TokenID:        req.Payload.TokenID,
	}
	err = dtx.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Str("func", "DepositSupHandler").Msg("failed to create deposit transaction in db")
		return terror.Error(err, errMsg)
	}

	reply(true)
	return nil
}

type DepositAsset1155ListResponse struct {
	Total        int                       `json:"total"`
	Transactions []*depositTransactionItem `json:"transactions"`
}

type depositTransactionItem struct {
	Username       string    `json:"username"`
	TxHash         string    `json:"tx_hash"`
	Amount         int       `json:"amount"`
	TokenID        int       `json:"token_id"`
	CollectionName string    `json:"collection_name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      null.Time `json:"updated_at"`
}

const HubKeyDepositAsset1155List = "ASSET:1155:DEPOSIT:LIST"

func (ac *AssetController) DepositAsset1155ListHandler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue getting deposit transaction list, try again or contact support."

	dtxs, err := boiler.DepositAsset1155Transactions(boiler.DepositAsset1155TransactionWhere.UserID.EQ(user.ID), qm.Load(boiler.DepositAsset1155TransactionRels.User), qm.Limit(10), qm.OrderBy("created_at DESC")).All(passdb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	transactions := make([]*depositTransactionItem, 0)

	for _, depositTx := range dtxs {
		collection, err := db.CollectionBySlug(depositTx.CollectionSlug)
		if err != nil {
			continue
		}

		deposit := &depositTransactionItem{
			Username:       depositTx.R.User.Username,
			TxHash:         depositTx.TXHash,
			Amount:         depositTx.Amount,
			TokenID:        depositTx.TokenID,
			CollectionName: collection.Name,
			Status:         depositTx.Status,
			CreatedAt:      depositTx.CreatedAt,
			UpdatedAt:      depositTx.UpdatedAt,
		}

		transactions = append(transactions, deposit)
	}

	resp := &DepositAsset1155ListResponse{
		Total:        len(dtxs),
		Transactions: transactions,
	}

	reply(resp)
	return nil
}

func reverseAssetServiceTransaction(
	ucm *Transactor,
	transactionToReverse *xsynTypes.NewTransaction,
	transferToReverse *boiler.AssetServiceTransferEvent,
	reason string,
) (transaction *xsynTypes.NewTransaction, returnTransferLog *boiler.AssetServiceTransferEvent) {
	transaction = &xsynTypes.NewTransaction{
		From:                 transactionToReverse.To,
		To:                   transactionToReverse.From,
		TransactionReference: xsynTypes.TransactionReference(fmt.Sprintf("REFUND - %s", transactionToReverse.TransactionReference)),
		Description:          fmt.Sprintf("Reverse transaction - %s. Reason: %s", transactionToReverse.Description, reason),
		Amount:               transactionToReverse.Amount,
		Group:                xsynTypes.TransactionGroupAssetManagement,
		SubGroup:             "Transfer",
		RelatedTransactionID: null.StringFrom(transactionToReverse.ID),
	}

	_, _, reverseID, err := ucm.Transact(transaction)
	if err != nil {
		passlog.L.Error().
			Err(err).
			Interface("transaction", transaction).
			Msg("reverse failed")
		return
	}

	returnTransferLog = &boiler.AssetServiceTransferEvent{
		UserAssetID:   transferToReverse.UserAssetID,
		UserID:        transferToReverse.UserID,
		InitiatedFrom: transferToReverse.InitiatedFrom,
		FromService:   transferToReverse.ToService,
		ToService:     transferToReverse.FromService,
		TransferTXID:  reverseID,
	}
	err = returnTransferLog.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().
			Err(err).
			Interface("returnTransferLog", returnTransferLog).
			Msg("failed to update transfer log in refund")
		return
	}
	return
}

func reverseAsset1155ServiceTransaction(
	ucm *Transactor,
	transactionToReverse *xsynTypes.NewTransaction,
	transferToReverse *boiler.Asset1155ServiceTransferEvent,
	reason string,
) (transaction *xsynTypes.NewTransaction, returnTransferLog *boiler.Asset1155ServiceTransferEvent) {
	transaction = &xsynTypes.NewTransaction{
		From:                 transactionToReverse.To,
		To:                   transactionToReverse.From,
		TransactionReference: xsynTypes.TransactionReference(fmt.Sprintf("REFUND - %s", transactionToReverse.TransactionReference)),
		Description:          fmt.Sprintf("Reverse transaction - %s. Reason: %s", transactionToReverse.Description, reason),
		Amount:               transactionToReverse.Amount,
		Group:                xsynTypes.TransactionGroupAssetManagement,
		SubGroup:             "Transfer",
		RelatedTransactionID: null.StringFrom(transactionToReverse.ID),
	}

	_, _, reverseID, err := ucm.Transact(transaction)
	if err != nil {
		passlog.L.Error().
			Err(err).
			Interface("transaction", transaction).
			Msg("reverse failed")
		return
	}

	returnTransferLog = &boiler.Asset1155ServiceTransferEvent{
		User1155AssetID: transferToReverse.User1155AssetID,
		UserID:          transferToReverse.UserID,
		InitiatedFrom:   transferToReverse.InitiatedFrom,
		Amount:          transferToReverse.Amount,
		FromService:     transferToReverse.FromService,
		ToService:       transferToReverse.ToService,
		TransferTXID:    reverseID,
	}
	err = returnTransferLog.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().
			Err(err).
			Interface("returnTransferLog", returnTransferLog).
			Msg("failed to update transfer log in refund")
		return
	}
	return
}
