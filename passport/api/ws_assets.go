package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
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
	api.SecureCommand(HubKeyAssetList, assetHub.AssetListHandler)
	api.SecureCommand(HubKeyAssetTransferToSupremacy, assetHub.AssetTransferToSupremacyHandler)
	api.SecureCommand(HubKeyAssetTransferFromSupremacy, assetHub.AssetTransferFromSupremacyHandler)
	api.Command(HubKeyAssetSubscribe, assetHub.AssetUpdatedSubscribeHandler)

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

const HubKeyAssetList = "ASSET:LIST"

func (ac *AssetController) AssetListHandler(ctx context.Context, user *xsynTypes.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AssetListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	total, assets, err := db.AssetList(&db.AssetListOpts{
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

const HubKeyAssetSubscribe = "ASSET:GET"

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

		err = supremacy_rpcclient.AssetLockToSupremacy(xsynTypes.UserAssetFromBoiler(userAsset),  reverseTransferLog.ID, marketLocked)
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
