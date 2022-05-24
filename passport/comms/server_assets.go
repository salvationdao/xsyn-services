package comms

import (
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/rpcclient"
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
			passlog.L.Error().Str("req.NewTemplateID", ass.NewTemplateID).Str("req.OldTemplateID",ass.OldTemplateID).Err(err).Msg("failed to update store item id")
			return err
		}
	}

	resp.Success = true
	return nil
}


type RegisterAssetReq struct {
	Asset *rpcclient.XsynAsset `json:"asset"`
}

type RegisterAssetResp struct {
	Success bool `json:"success"`
}


// AssetRegisterHandler registers a new asset
func (s *S) AssetRegisterHandler(req RegisterAssetReq, resp *RegisterAssetResp) error {
	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(req.Asset.CollectionSlug)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't get collection")
		return err
	}

	var attributeJson types.JSON
	if req.Asset.Attributes!= nil{
		err = attributeJson.Marshal(req.Asset.Attributes)
		if err != nil {
			passlog.L.Error().Interface("req", req).Interface("req.Asset.Attributes", req.Asset.Attributes).Err(err).Msg("failed to register new asset - can't marshall attributes")
			return err
		}
	}else {
		err = attributeJson.Marshal("{}")
		if err != nil {
			passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't marshall '{}' attributes")
			return err
		}
	}



		boilerAsset := &boiler.UserAsset{
		ID:              req.Asset.ID,
		CollectionID:    collection.ID,
		TokenID: req.Asset.TokenID,
		Tier:            req.Asset.Tier,
		Hash:            req.Asset.Hash,
		OwnerID:         req.Asset.OwnerID,
		Data:            req.Asset.Data,
		Attributes:      attributeJson,
		Name:            req.Asset.Name,
		ImageURL:        req.Asset.ImageURL,
		ExternalURL:     req.Asset.ExternalURL,
		Description:     req.Asset.Description,
		BackgroundColor: req.Asset.BackgroundColor,
		AnimationURL: req.Asset.AnimationURL,
		YoutubeURL: req.Asset.YoutubeURL,
		UnlockedAt: req.Asset.UnlockedAt,
		MintedAt: req.Asset.MintedAt,
		OnChainStatus: req.Asset.OnChainStatus,
		XsynLocked: req.Asset.XsynLocked,
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
	Assets []*rpcclient.XsynAsset `json:"assets"`
}

type RegisterAssetsResp struct {
	Success bool `json:"success"`
}

// AssetsRegisterHandler registers new assets
func (s *S) AssetsRegisterHandler(req RegisterAssetsReq, resp *RegisterAssetsResp) error {
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
				passlog.L.Error().Interface("req", req).Interface("asset.Attributes",asset.Attributes).Err(err).Msg("failed to register new asset - can't marshall attributes")
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
			ID:           asset.ID,
			CollectionID: collection.ID,
			TokenID: int64(int(asset.TokenID)),
			Tier:         asset.Tier,
			Hash:         asset.Hash,
			OwnerID:      asset.OwnerID,
			Data:            asset.Data,
			Attributes:      attributeJson,
			Name:            asset.Name,
			ImageURL:        asset.ImageURL,
			ExternalURL:     asset.ExternalURL,
			Description:     asset.Description,
			BackgroundColor: asset.BackgroundColor,
			AnimationURL: asset.AnimationURL,
			YoutubeURL: asset.YoutubeURL,
			UnlockedAt: asset.UnlockedAt,
			MintedAt: asset.MintedAt,
			OnChainStatus: asset.OnChainStatus,
			XsynLocked: asset.XsynLocked,
			ServiceLocked: null.StringFrom("Supremacy"), // TODO: hook this up from the service user
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


