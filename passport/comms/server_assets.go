package comms

import (
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
)

func (s *S) AssetOnChainStatusHandler(req AssetOnChainStatusReq, resp *AssetOnChainStatusResp) error {
	item, err := boiler.PurchasedItemsOlds(boiler.PurchasedItemsOldWhere.ID.EQ(req.AssetID)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetID", req.AssetID).Err(err).Msg("failed to get asset")
		return err
	}

	resp.OnChainStatus = item.OnChainStatus
	return nil
}

func (s *S) AssetsOnChainStatusHandler(req AssetsOnChainStatusReq, resp *AssetsOnChainStatusResp) error {
	items, err := boiler.PurchasedItemsOlds(boiler.PurchasedItemsOldWhere.ID.IN(req.AssetIDs)).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetIDs", strings.Join(req.AssetIDs, ", ")).Err(err).Msg("failed to get assets")
		return err
	}

	assetMap := make(map[string]string)
	for _, asset := range items {
		assetMap[asset.ID] = asset.OnChainStatus
	}

	resp.OnChainStatuses = assetMap
	return nil
}

type UpdateAssetIDReq struct {
	AssetID    string `json:"asset_ID"`
	OldAssetID string `json:"old_asset_ID"`
}

type UpdateAssetIDResp struct {
	AssetID    string `json:"asset_ID"`
}

// UpdateAssetIDHandler updates the asset id if it happened to change for some reason
func (s *S) UpdateAssetIDHandler(req UpdateAssetIDReq, resp *UpdateAssetIDResp) error {
	err := db.ChangePurchasedItemID(req.OldAssetID, req.AssetID)
	if err != nil {
		passlog.L.Error().Str("req.AssetID", req.AssetID).Str("req.OldAssetID",req.OldAssetID).Err(err).Msg("failed to update asset")
		return err
	}

	resp.AssetID = req.AssetID
	return nil
}

type UpdateAssetsIDReq struct {
	AssetsToUpdate []*UpdateAssetIDReq `json:"assets_to_update"`
}

type UpdateAssetsIDResp struct {
	Success bool `json:"success"`
}

// UpdateAssetsIDHandler updates the assets' id if it happened to change for some reason
func (s *S) UpdateAssetsIDHandler(req UpdateAssetsIDReq, resp *UpdateAssetsIDResp) error {
	for _, ass := range req.AssetsToUpdate {
		err := db.ChangePurchasedItemID(ass.OldAssetID, ass.AssetID)
		if err != nil {
			passlog.L.Error().Str("req.AssetID", ass.AssetID).Str("req.OldAssetID",ass.OldAssetID).Err(err).Msg("failed to update asset")
			return err
		}
	}

	resp.Success = true
	return nil
}

type UpdateStoreItemIDsReq struct {
	StoreItemsToUpdate []*TemplatesToUpdate `json:"store_items_to_update"`
}

type TemplatesToUpdate struct {
	OldTemplateID string `json:"old_template_id"`
	NewTemplateID string `json:"new_template_id"`
}

type UpdateStoreItemIDsResp struct {
	Success bool `json:"success"`
}

// UpdateStoreItemIDsHandler updates the store item's template ID
func (s *S) UpdateStoreItemIDsHandler(req UpdateStoreItemIDsReq, resp *UpdateStoreItemIDsResp) error {
	for _, ass := range req.StoreItemsToUpdate {
		err := db.ChangeStoreItemsTemplateID(ass.OldTemplateID, ass.NewTemplateID)
		if err != nil {
			passlog.L.Error().Str("req.NewTemplateID", ass.NewTemplateID).Str("req.OldTemplateID",ass.OldTemplateID).Err(err).Msg("failed to update store item id")
			return err
		}
	}

	resp.Success = true
	return nil
}


type RegisterAssetReq struct {
	ID              string      `json:"id"`
	CollectionSlug    string      `json:"collection_slug"`
	TokenID         int         `json:"token_id"`
	Tier            string      `json:"tier"`
	Hash            string      `json:"hash"`
	OwnerID         string      `json:"owner_id"`
	Data            types.JSON  `json:"data"`
	Attributes      types.JSON  `json:"attributes"`
	Name            string      `json:"name"`
	ImageURL        null.String `json:"image_url,omitempty"`
	ExternalURL     null.String `json:"external_url,omitempty"`
	Description     null.String `json:"description,omitempty"`
	BackgroundColor null.String `json:"background_color,omitempty"`
	AnimationURL    null.String `json:"animation_url,omitempty"`
	YoutubeURL      null.String `json:"youtube_url,omitempty"`
	UnlockedAt      time.Time   `json:"unlocked_at"`
	MintedAt        null.Time   `json:"minted_at,omitempty"`
	OnChainStatus   string      `json:"on_chain_status"`
	XsynLocked      null.Bool   `json:"xsyn_locked,omitempty"`
}

type RegisterAssetResp struct {
	Success bool `json:"success"`
}


// RegisterAsset registers a new asset
func (s *S) RegisterAsset(req RegisterAssetReq, resp *RegisterAssetResp) error {
	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(req.CollectionSlug)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't get collection")
		return err
	}

	boilerAsset := boiler.UserAsset{
		ID: req.ID,
		CollectionID: collection.ID,
		TokenID: req.TokenID,
		Tier: req.Tier,
		Hash: req.Hash,
		OwnerID: req.OwnerID,
		Data: req.Data,
		Attributes: req.Attributes,
		Name: req.Name,
		ImageURL: req.ImageURL,
		ExternalURL: req.ExternalURL,
		Description: req.Description,
		BackgroundColor: req.BackgroundColor,
		AnimationURL: req.AnimationURL,
		YoutubeURL: req.YoutubeURL,
		UnlockedAt: req.UnlockedAt,
		MintedAt: req.MintedAt,
		OnChainStatus: req.OnChainStatus,
		XsynLocked: req.XsynLocked,
	}

	err = boilerAsset.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't insert asset")
		return err
	}

	resp.Success = true
	return nil
}

