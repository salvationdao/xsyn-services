package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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
		return http.StatusBadRequest, terror.Error(fmt.Errorf("invalid asset hash"), "Invalid Asset Hash.")
	}

	// Get asset via token id

	item, err := db.PurchasedItemByHash(hash)
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

// AssetGet grabs asset's metadata via token id
func (api *API) AssetGetByCollectionAndTokenID(w http.ResponseWriter, r *http.Request) (int, error) {
	collectionAddress := chi.URLParam(r, "collection_address")
	if collectionAddress == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("collection_address not provided in URL"), "metadata")
	}
	tokenIDStr := chi.URLParam(r, "token_id")
	if tokenIDStr == "" {
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("token_id not provided in URL"), "metadata")
	}
	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		return http.StatusBadRequest, terror.Warn(err, "get asset from db")
	}
	item, err := db.PurchasedItemByMintContractAndTokenID(common.HexToAddress(collectionAddress), tokenID)
	if err != nil {
		return http.StatusBadRequest, terror.Warn(err, "get asset from db")
	}

	b, err := purchasedItemToOpenseaMetaData(api, item)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to convert to opensea metadata")
	}

	_, err = w.Write(b)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to send metadata")
	}

	return http.StatusOK, nil
}

// openSeaMetaData data structure, reference https://docs.opensea.io/docs/metadata-standards
type openSeaMetaData struct {
	Image           string               `json:"image,omitempty"`            // image url, to be cached by opensea
	ImageData       string               `json:"image_data,omitempty"`       // raw image svg
	ExternalURL     string               `json:"external_url,omitempty"`     // direct url link to image asset
	Description     string               `json:"description,omitempty"`      // item description
	Name            string               `json:"name,omitempty"`             // item name
	Attributes      []passport.Attribute `json:"attributes,omitempty"`       // item attributes, custom    TODO
	BackgroundColor string               `json:"background_color,omitempty"` // openseas page background
	AnimationURL    string               `json:"animation_url,omitempty"`    // direct url link to video asset
	YoutubeURL      string               `json:"youtube_url,omitempty"`      // url to youtube video
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
	Label        string `json:"label"`
	ImageURL     string `json:"image_url"`
	AnimationURL string `json:"animation_url"`
	AssetType    string `json:"asset_type"`
	Tier         string `json:"tier"`
	Slug         string `json:"slug"`
}
type purchasedItemMetaDataChassis struct {
	Label string `json:"label"`
	Model string `json:"model"`
	Skin  string `json:"skin"`
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

func purchasedItemToOpenseaMetaData(api *API, item *boiler.PurchasedItem) (jb []byte, err error) {
	if item == nil {
		return nil, terror.Error(fmt.Errorf("item is nil"))
	}

	itemMeta := purchasedItemMetaData{}
	err = item.Data.Unmarshal(&itemMeta)
	if err != nil {
		return nil, terror.Error(err)
	}

	datOpensea := openSeaMetaData{}
	datOpensea.Image = itemMeta.Mech.ImageURL
	datOpensea.Description = itemMeta.Mech.Label // TODO bring back when decided what to put
	datOpensea.Name = itemMeta.Mech.Name
	datOpensea.AnimationURL = itemMeta.Mech.AnimationURL

	// prepare attributes adding
	attributes := []passport.Attribute{}
	var str string
	var atr passport.Attribute

	// asset type
	str = itemMeta.Mech.AssetType
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Asset Type",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// brand
	// HACK: cheat to do quick brand lookup
	// TODO: may need to do db or rpc call in the future
	// hint: itemMeta.Chassis.BrandID == "id"
	if strings.Contains(itemMeta.Mech.Slug, "zaibatsu") {
		str = "Zaibatsu Heavy Industries"
	} else if strings.Contains(itemMeta.Mech.Slug, "mountain") {
		str = "Red Mountain Offworld Mining Corporation"
	} else if strings.Contains(itemMeta.Mech.Slug, "boston") {
		str = "Boston Cybernetics"
	}
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Brand",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// model
	str = itemMeta.Chassis.Model
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Model",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// rarity
	str = itemMeta.Mech.Tier
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Rarity",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// submodel
	str = itemMeta.Chassis.Skin
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Submodel",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// torrent 1
	str = itemMeta.Turrets.Key0.Label
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Turret One",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// torrent 2
	str = itemMeta.Turrets.Key1.Label
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Turret Two",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// utility 1
	str = itemMeta.Modules.Key0.Label
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Utility One",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// weapon 1
	str = itemMeta.Weapons.Key0.Label
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Weapon One",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// weapon 2
	str = itemMeta.Weapons.Key1.Label
	if len(str) > 0 {
		atr = passport.Attribute{
			TraitType: "Weapon Two",
			Value:     str,
		}
		attributes = append(attributes, atr)
	}

	// insert attributes
	if len(attributes) < 10 {
		api.Log.Warn().Err(fmt.Errorf("invalid opensea attributes length")).Msg("opensea attributes less than 10")
	}
	datOpensea.Attributes = attributes

	// turn into json string
	jb, err = json.Marshal(datOpensea)
	if err != nil {
		return nil, terror.Error(err)
	}

	return
}
