package comms

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
	"xsyn-services/passport/supremacy_rpcclient"
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
			passlog.L.Error().Str("req.NewTemplateID", ass.NewTemplateID).Str("req.OldTemplateID", ass.OldTemplateID).Err(err).Msg("failed to update store item id")
			return err
		}
	}

	resp.Success = true
	return nil
}

type RegisterAssetReq struct {
	ApiKey string                        `json:"api_key"`
	Asset *supremacy_rpcclient.XsynAsset `json:"asset"`
}

type RegisterAssetResp struct {
	Success bool `json:"success"`
}

// AssetRegisterHandler registers a new asset
func (s *S) AssetRegisterHandler(req RegisterAssetReq, resp *RegisterAssetResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to register new asset")
		return err
	}

	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(req.Asset.CollectionSlug)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't get collection")
		return err
	}

	var attributeJson types.JSON
	if req.Asset.Attributes != nil {
		err = attributeJson.Marshal(req.Asset.Attributes)
		if err != nil {
			passlog.L.Error().Interface("req", req).Interface("req.Asset.Attributes", req.Asset.Attributes).Err(err).Msg("failed to register new asset - can't marshall attributes")
			return err
		}
	} else {
		err = attributeJson.Marshal("{}")
		if err != nil {
			passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't marshall '{}' attributes")
			return err
		}
	}

	boilerAsset := &boiler.UserAsset{
		ID:              req.Asset.ID,
		CollectionID:    collection.ID,
		TokenID:         req.Asset.TokenID,
		Tier:            req.Asset.Tier,
		AssetType: req.Asset.AssetType,
		Hash:            req.Asset.Hash,
		OwnerID:         req.Asset.OwnerID,
		Data:            req.Asset.Data,
		Attributes:      attributeJson,
		Name:            req.Asset.Name,
		ImageURL:        req.Asset.ImageURL,
		ExternalURL:     req.Asset.ExternalURL,
		Description:     req.Asset.Description,
		BackgroundColor: req.Asset.BackgroundColor,
		AnimationURL:    req.Asset.AnimationURL,
		YoutubeURL:      req.Asset.YoutubeURL,
		UnlockedAt:      req.Asset.UnlockedAt,
		MintedAt:        req.Asset.MintedAt,
		OnChainStatus:   req.Asset.OnChainStatus,
		LockedToService: null.StringFrom(serviceID),
	}

	err = boilerAsset.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't insert asset")
		return err
	}

	resp.Success = true
	return nil
}

type RegisterAssetsReq struct {
	ApiKey string                           `json:"api_key"`
	Assets []*supremacy_rpcclient.XsynAsset `json:"assets"`
}

type RegisterAssetsResp struct {
	Success bool `json:"success"`
}

// AssetsRegisterHandler registers new assets
func (s *S) AssetsRegisterHandler(req RegisterAssetsReq, resp *RegisterAssetsResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to register new asset")
		return err
	}

	for _, asset := range req.Assets {
		collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(asset.CollectionSlug)).One(passdb.StdConn)
		if err != nil {
			passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't get collection")
			return err
		}

		var attributeJson types.JSON
		if asset.Attributes != nil {
			err = attributeJson.Marshal(asset.Attributes)
			if err != nil {
				passlog.L.Error().Interface("req", req).Interface("asset.Attributes", asset.Attributes).Err(err).Msg("failed to register new asset - can't marshall attributes")
				return err
			}
		} else {
			err = attributeJson.Marshal("{}")
			if err != nil {
				passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't marshall '{}' attributes")
				return err
			}
		}

		boilerAsset := &boiler.UserAsset{
			ID:              asset.ID,
			CollectionID:    collection.ID,
			TokenID:         int64(int(asset.TokenID)),
			Tier:            asset.Tier,
			Hash:            asset.Hash,
			AssetType: asset.AssetType,
			OwnerID:         asset.OwnerID,
			Data:            asset.Data,
			Attributes:      attributeJson,
			Name:            asset.Name,
			ImageURL:        asset.ImageURL,
			ExternalURL:     asset.ExternalURL,
			Description:     asset.Description,
			BackgroundColor: asset.BackgroundColor,
			AnimationURL:    asset.AnimationURL,
			YoutubeURL:      asset.YoutubeURL,
			UnlockedAt:      asset.UnlockedAt,
			MintedAt:        asset.MintedAt,
			OnChainStatus:   asset.OnChainStatus,
			LockedToService: null.StringFrom(serviceID),
		}

		err = boilerAsset.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			passlog.L.Error().Interface("req", req).Err(err).Msg(" failed to register new asset - can't insert asset")
			return err
		}
	}
	resp.Success = true
	return nil
}

type UpdateUser1155AssetReq struct {
	PublicAddress string               `json:"public_address"`
	AssetData     []Supremacy1155Asset `json:"asset_data"`
}

type Supremacy1155Asset struct {
	Label          string                      `json:"label"`
	Description    string                      `json:"description"`
	CollectionSlug string                      `json:"collection_slug"`
	TokenID        int                         `json:"token_id"`
	Count          int                         `json:"count"`
	ImageURL       string                      `json:"image_url"`
	AnimationURL   string                      `json:"animation_url"`
	KeycardGroup   string                      `json:"keycard_group"`
	Attributes     []SupremacyKeycardAttribute `json:"attributes"`
}

type SupremacyKeycardAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value,omitempty"`
}

type UpdateUser1155AssetResp struct {
	UserID        string      `json:"user_id"`
	Username      string      `json:"username"`
	FactionID     null.String `json:"faction_id"`
	PublicAddress null.String `json:"public_address"`
}

//InsertUser1155Asset inserts keycards
func (s *S) InsertUser1155Asset(req UpdateUser1155AssetReq, resp *UpdateUser1155AssetResp) error {
	user, err := payments.CreateOrGetUser(common.HexToAddress(req.PublicAddress))
	if err != nil {
		passlog.L.Error().Str("req.PublicAddress", req.PublicAddress).Err(err).Msg("Failed to get or create user while updating 1155 asset")
		return terror.Error(err, "Failed to create or get user")
	}

	for _, asset := range req.AssetData {
		var assetJson types.JSON

		if asset.Attributes != nil {
			err = assetJson.Marshal(asset.Attributes)
			if err != nil {

				return terror.Error(err, "Failed to get asset attributes")
			}
		} else {
			err = assetJson.Marshal("{}")
			if err != nil {

				return terror.Error(err, "Failed to get asset attributes")
			}
		}

		collection, err := db.CollectionBySlug(asset.CollectionSlug)
		if err != nil {
			return terror.Error(err, "Failed to get collection from DB")
		}

		newAsset := &boiler.UserAssets1155{
			OwnerID:         user.ID,
			ExternalTokenID: asset.TokenID,
			Count:           asset.Count,
			Label:           asset.Label,
			Description:     asset.Description,
			ImageURL:        asset.ImageURL,
			AnimationURL:    null.StringFrom(asset.AnimationURL),
			KeycardGroup:    asset.KeycardGroup,
			Attributes:      assetJson,
			CollectionID:    collection.ID,
		}

		err = newAsset.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			passlog.L.Error().Err(err).Msg("Failed to insert new asset")
			return terror.Error(err, "Failed to get asset attributes")
		}

	}

	resp.UserID = user.ID
	resp.Username = user.Username
	resp.FactionID = user.FactionID
	resp.PublicAddress = user.PublicAddress

	return nil
}


