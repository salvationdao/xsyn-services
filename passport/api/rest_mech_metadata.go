package api

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"net/http"
	"strconv"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/supremacy_rpcclient"
	"xsyn-services/types"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
)

/***
 *  # dev notes
 *
 *  test url:
 *  http://localhost:8086/api/asset/0x651d4424f34e6e918d8e4d2da4df3debdae83d0c/682
 *  https://opensea.io/assets/0x651d4424f34e6e918d8e4d2da4df3debdae83d0c/682
 *  https://api.xsyn.io/api/asset/0x651d4424f34e6e918d8e4d2da4df3debdae83d0c/682
 *
 */

// AssetGet grabs asset's metadata via token id
func (api *API) AssetGet(w http.ResponseWriter, r *http.Request) (int, error) {
	// Get token id
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("invalid asset hash"), "Invalid UserAsset Hash.")
	}

	// Get asset via token id

	item, err := db.PurchasedItemByHashDEPRECATE(hash)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset")
	}
	// Encode result
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to encode JSON")
	}

	return http.StatusOK, nil
}

// AssetGetByCollectionAndTokenID grabs asset's metadata via token id
func (api *API) AssetGetByCollectionAndTokenID(w http.ResponseWriter, r *http.Request) (int, error) {
	collectionAddress := chi.URLParam(r, "collection_address")
	if collectionAddress == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("collection_address not provided in URL"), "Invalid collection_address.")
	}
	tokenIDStr := chi.URLParam(r, "token_id")
	if tokenIDStr == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("token_id not provided in URL"), "Invalid token_id")
	}

	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		return http.StatusBadRequest, terror.Warn(err, "Invalid token_id")
	}

	collection, err := boiler.Collections(
		boiler.CollectionWhere.MintContract.EQ(null.StringFrom(common.HexToAddress(collectionAddress).Hex())),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Warn(err, "Failed to get collection from db")
	}

	var openseaAsset *openSeaMetaData

	asset, err := boiler.UserAssets(
		boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
		boiler.UserAssetWhere.TokenID.EQ(int64(tokenID)),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed find asset")
	}

	// if asset refreshed over 24 hours ago, pull it
	if asset.DataRefreshedAt.Before(time.Now().Add(-24 * time.Hour)) {
		xsynAsset, err := supremacy_rpcclient.AssetGet(asset.Hash)
		if err != nil {
			passlog.L.Error().Err(err).Str("asset.Hash", asset.Hash).Msg("failed to refresh metadata")
		} else {
			userAssetNew, err := db.UpdateUserAsset(xsynAsset, false)
			if err != nil {
				passlog.L.Error().Err(err).Str("asset.Hash", asset.Hash).Msg("failed to update metadata")
			} else {
				asset = userAssetNew
			}
		}
	}

	attribes := []*types.OpenSeaAttribute{}
	if asset.Attributes != nil {
		err := asset.Attributes.Unmarshal(&attribes)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed find asset")
		}
	}

	newAttributes := []*types.OpenSeaAttribute{}
	if len(attribes) > 0 {
		for _, attribute := range attribes {
			if attribute.TraitType == "Name" || attribute.TraitType == "name" {
				continue
			}
			newAttribute := &types.OpenSeaAttribute{
				DisplayType: attribute.DisplayType,
				TraitType:   attribute.TraitType,
				Value:       attribute.Value,
			}

			newAttributes = append(newAttributes, newAttribute)
		}
	}

	openseaAsset = &openSeaMetaData{
		Image:           asset.ImageURL.String,
		ExternalURL:     asset.ExternalURL.String,
		Description:     asset.Description.String,
		Name:            asset.Name,
		Attributes:      newAttributes,
		BackgroundColor: asset.BackgroundColor.String,
		AnimationURL:    asset.AnimationURL.String,
		YoutubeURL:      asset.YoutubeURL.String,
	}

	jsonObject, err := json.Marshal(openseaAsset)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed marshall asset")
	}

	_, err = w.Write(jsonObject)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to send metadata")
	}
	return http.StatusOK, nil
}

// openSeaMetaData data structure, reference https://docs.opensea.io/docs/metadata-standards
type openSeaMetaData struct {
	Image           string                    `json:"image,omitempty"`            // image url, to be cached by opensea
	ImageData       string                    `json:"image_data,omitempty"`       // raw image svg
	ExternalURL     string                    `json:"external_url,omitempty"`     // direct url link to image asset
	Description     string                    `json:"description,omitempty"`      // item description
	Name            string                    `json:"name,omitempty"`             // item name
	Attributes      []*types.OpenSeaAttribute `json:"attributes,omitempty"`       // item attributes
	BackgroundColor string                    `json:"background_color,omitempty"` // openseas page background
	AnimationURL    string                    `json:"animation_url,omitempty"`    // direct url link to video asset
	YoutubeURL      string                    `json:"youtube_url,omitempty"`      // url to youtube video
}

// purchasedItemMetaData shape of the purchased_items.metadata in the database
type purchasedItemMetaData struct {
	Mech    purchasedItemMetaDataMech         `json:"mech"`
	Chassis purchasedItemMetaDataChassis      `json:"chassis"`
	Modules purchasedItemMetaDataNestedModule `json:"modules"`
	Turrets purchasedItemMetaDataNestedTurret `json:"turrets"`
	Weapons purchasedItemMetaDataNestedWeapon `json:"weapons"`
}

// purchasedItemMetaDataNestedModule shape of module, object not array
type purchasedItemMetaDataNestedModule struct {
	Key0 purchasedItemMetaDataModule `json:"0"`
}

// purchasedItemMetaDataNestedTrurrent shape of turret, object not array
type purchasedItemMetaDataNestedTurret struct {
	Key0 purchasedItemMetaDataTurret `json:"0"`
	Key1 purchasedItemMetaDataTurret `json:"1"`
}

// purchasedItemMetaDataNestedWeapon shape of weapon, object not array
type purchasedItemMetaDataNestedWeapon struct {
	Key0 purchasedItemMetaDataWeapon `json:"0"`
	Key1 purchasedItemMetaDataWeapon `json:"1"`
}

// labels that we only need
type purchasedItemMetaDataMech struct {
	Name          string `json:"name"`
	Label         string `json:"label"`
	LargeImageURL string `json:"large_image_url"`
	ImageURL      string `json:"image_url"`
	AnimationURL  string `json:"animation_url"`
	AssetType     string `json:"asset_type"`
	Tier          string `json:"tier"`
	Slug          string `json:"slug"`
}
type purchasedItemMetaDataChassis struct {
	Label              string `json:"label"`
	Model              string `json:"model"`
	Skin               string `json:"skin"`
	ShieldRechargeRate int    `json:"shield_recharge_rate"`
	HealthRemaining    int    `json:"health_remaining"`
	WeaponHardpoints   int    `json:"weapon_hardpoints"`
	TurretHardpoints   int    `json:"turret_hardpoints"`
	UtilitySlots       int    `json:"utility_slots"`
	Speed              int    `json:"speed"`
	MaxHitpoints       int    `json:"max_hitpoints"`
	MaxShield          int    `json:"max_shield"`
}
type purchasedItemMetaDataModule struct {
	Label string `json:"label"`
}
type purchasedItemMetaDataTurret struct {
	Label string `json:"label"`
}
type purchasedItemMetaDataWeapon struct {
	Label string `json:"label"`
}
