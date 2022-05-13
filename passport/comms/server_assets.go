package comms

import (
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"

	"github.com/ninja-software/terror/v2"
)

func (s *S) AssetOnChainStatusHandler(req AssetOnChainStatusReq, resp *AssetOnChainStatusResp) error {
	item, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.ID.EQ(req.AssetID)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetID", req.AssetID).Err(err).Msg("failed to get asset")
		return terror.Error(err)
	}

	resp.OnChainStatus = item.OnChainStatus
	return nil
}

func (s *S) AssetsOnChainStatusHandler(req AssetsOnChainStatusReq, resp *AssetsOnChainStatusResp) error {
	items, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.ID.IN(req.AssetIDs)).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetIDs", strings.Join(req.AssetIDs, ", ")).Err(err).Msg("failed to get assets")
		return terror.Error(err)
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
		return terror.Error(err)
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
			return terror.Error(err)
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
			return terror.Error(err)
		}
	}

	resp.Success = true
	return nil
}
