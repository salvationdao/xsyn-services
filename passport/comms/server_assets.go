package comms

import (
	"fmt"
	"github.com/gofrs/uuid"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/nft1155"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
	"xsyn-services/passport/supremacy_rpcclient"
	xsynTypes "xsyn-services/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
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
	ApiKey string                           `json:"api_key"`
	Asset  []*supremacy_rpcclient.XsynAsset `json:"asset"`
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

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, ass := range req.Asset {
		_, err = db.RegisterUserAsset(ass, serviceID, tx)
		if err != nil {
			passlog.L.Error().Err(err).Interface("req.Asset", req.Asset).Msg("failed to register new asset")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
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
		passlog.L.Error().Err(err).Msg("failed to register new assets")
		return err
	}

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, asset := range req.Assets {
		_, err = db.RegisterUserAsset(asset, serviceID, tx)
		if err != nil {
			passlog.L.Error().Err(err).Interface("asset", asset).Int("index of fail", i).Msg("failed to register new assets")
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	resp.Success = true
	return nil
}

type UpdateUser1155AssetReq struct {
	ApiKey        string               `json:"api_key"`
	PublicAddress string               `json:"public_address"`
	AssetData     []Supremacy1155Asset `json:"asset_data"`
}

type Supremacy1155Asset struct {
	Label          string                                `json:"label"`
	Description    string                                `json:"description"`
	CollectionSlug string                                `json:"collection_slug"`
	TokenID        int                                   `json:"token_id"`
	Count          int                                   `json:"count"`
	ImageURL       string                                `json:"image_url"`
	AnimationURL   string                                `json:"animation_url"`
	KeycardGroup   string                                `json:"keycard_group"`
	Attributes     []xsynTypes.SupremacyKeycardAttribute `json:"attributes"`
}

type UpdateUser1155AssetResp struct {
	UserID        string      `json:"user_id"`
	Username      string      `json:"username"`
	FactionID     null.String `json:"faction_id"`
	PublicAddress null.String `json:"public_address"`
}

//InsertUser1155AssetHandler inserts keycards
func (s *S) InsertUser1155AssetHandler(req UpdateUser1155AssetReq, resp *UpdateUser1155AssetResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}
	user, err := payments.CreateOrGetUser(common.HexToAddress(req.PublicAddress), s.API.Environment)
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
			ServiceID:       null.StringFrom(serviceID),
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

type Asset1155CountUpdateSupremacyReq struct {
	ApiKey         string      `json:"api_key"`
	TokenID        int         `json:"token_id"`
	Address        string      `json:"address"`
	CollectionSlug string      `json:"collection_slug"`
	Amount         int         `json:"amount"`
	ImageURL       string      `json:"image_url"`
	AnimationURL   null.String `json:"animation_url"`
	KeycardGroup   string      `json:"keycard_group"`
	Attributes     types.JSON  `json:"attributes"`
	IsAdd          bool        `json:"is_add"`
}

type Asset1155CountUpdateSupremacyResp struct {
	Count int `json:"count"`
}

// AssetKeycardCountUpdateSupremacy update keycard count
func (s *S) AssetKeycardCountUpdateSupremacy(req Asset1155CountUpdateSupremacyReq, resp *Asset1155CountUpdateSupremacyResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - Asset1155CountUpdateSupremacy")
		return err
	}
	user, err := payments.CreateOrGetUser(common.HexToAddress(req.Address), s.API.Environment)
	if err != nil {
		return terror.Error(err, "Failed to get user")
	}

	asset, err := nft1155.CreateOrGet1155AssetWithService(req.TokenID, user, req.CollectionSlug, xsynTypes.SupremacyGameUserID.String())
	if err != nil {
		return terror.Error(err, "Failed to create or get asset with service id")
	}

	if req.IsAdd {
		asset.Count += req.Amount
	} else {
		asset.Count -= req.Amount
		if asset.Count < 0 {
			return terror.Error(fmt.Errorf("total after taking is less than 0"), "Total after taking is less than 0")
		}
	}

	asset.ImageURL = req.ImageURL
	asset.AnimationURL = req.AnimationURL
	asset.KeycardGroup = req.KeycardGroup
	asset.Attributes = req.Attributes

	_, err = asset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.Count))
	if err != nil {
		return terror.Error(err, "Failed to  service id")
	}

	resp.Count = asset.Count

	return nil
}

type DeleteAssetHandlerReq struct {
	ApiKey  string `json:"api_key"`
	AssetID string `json:"asset_id"`
}

type DeleteAssetHandlerResp struct {
}

func (s *S) DeleteAssetHandler(req DeleteAssetHandlerReq, resp *DeleteAssetHandlerResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - DeleteAssetHandler")
		return err
	}

	_, err = boiler.UserAssets(
		boiler.UserAssetWhere.ID.EQ(req.AssetID),
	).DeleteAll(passdb.StdConn, false)
	if err != nil {
		passlog.L.Error().Err(err).Str("req.AssetID", req.AssetID).Msg("failed to delete asset - DeleteAssetHandler")
		return err
	}

	return nil
}

type AssignTemplateReq struct {
	ApiKey      string   `json:"api_key"`
	UserID      string   `json:"user_id"`
	TemplateIDs []string `json:"template_ids"`
}

type AssignTemplateResp struct {
}

func (s *S) AssignTemplateHandler(req AssignTemplateReq, resp *AssignTemplateResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - DeleteAssetHandler")
		return err
	}

	for _, tempID := range req.TemplateIDs {
		_, err = db.PurchasedItemRegister(uuid.Must(uuid.FromString(tempID)), uuid.Must(uuid.FromString(req.UserID)))
		if err != nil {
			passlog.L.Error().Err(err).Msg("failed to PurchasedItemRegister")
		}
	}

	return nil
}

type AssetUpdateReq struct {
	ApiKey string                         `json:"api_key"`
	Asset  *supremacy_rpcclient.XsynAsset `json:"asset"`
}

type AssetUpdateResp struct {
}

func (s *S) AssetUpdateHandler(req AssetUpdateReq, resp *AssetUpdateResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - AssetUpdateHandler")
		return err
	}

	_, err = db.UpdateUserAsset(req.Asset, true)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to UpdateUserAsset")
	}

	return nil
}
