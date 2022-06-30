package types

import (
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passlog"
)

type UserAsset struct {
	ID               string       `json:"id"`
	CollectionID     string       `json:"collection_id"`
	TokenID          int64        `json:"token_id"`
	Tier             string       `json:"tier"`
	Hash             string       `json:"hash"`
	OwnerID          string       `json:"owner_id"`
	Data             types.JSON   `json:"data"`
	Attributes       []*Attribute `json:"attributes"`
	Name             string       `json:"name"`
	AssetType        null.String  `json:"asset_type,omitempty"`
	ImageURL         null.String  `json:"image_url,omitempty"`
	ExternalURL      null.String  `json:"external_url,omitempty"`
	Description      null.String  `json:"description,omitempty"`
	BackgroundColor  null.String  `json:"background_color,omitempty"`
	AnimationURL     null.String  `json:"animation_url,omitempty"`
	YoutubeURL       null.String  `json:"youtube_url,omitempty"`
	CardAnimationURL null.String  `json:"card_animation_url,omitempty"`
	AvatarURL        null.String  `json:"avatar_url,omitempty"`
	LargeImageURL    null.String  `json:"large_image_url,omitempty"`
	UnlockedAt       time.Time    `json:"unlocked_at"`
	MintedAt         null.Time    `json:"minted_at,omitempty"`
	OnChainStatus    string       `json:"on_chain_status"`
	DeletedAt        null.Time    `json:"deleted_at,omitempty"`
	DataRefreshedAt  time.Time    `json:"data_refreshed_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	CreatedAt        time.Time    `json:"created_at"`
}

type User1155Asset struct {
	ID              string      `json:"id"`
	OwnerID         string      `json:"owner_id"`
	CollectionID    string      `json:"collection_id"`
	CollectionSlug  string      `json:"collection_slug"`
	ExternalTokenID int         `json:"external_token_id"`
	MintContract    null.String `json:"mint_contract"`
	Count           int         `json:"count"`
	Label           string      `json:"label"`
	Description     string      `json:"description"`
	ImageURL        string      `json:"image_url"`
	AnimationURL    null.String `json:"animation_url"`
	KeycardGroup    string      `json:"keycard_group"`
	Attributes      types.JSON  `json:"attributes"`
	ServiceID       null.String `json:"service_id"`
	CreatedAt       time.Time   `json:"created_at"`
}

type Attribute struct {
	DisplayType DisplayType `json:"display_type,omitempty"`
	TraitType   string      `json:"trait_type"`
	AssetHash   string      `json:"asset_hash,omitempty"`
	Value       interface{} `json:"value"` // string or number only
}

type DisplayType string

const (
	BoostNumber     DisplayType = "boost_number"
	BoostPercentage DisplayType = "boost_percentage"
	Number          DisplayType = "number"
	Date            DisplayType = "date"
)

func UserAssets721FromBoiler(us []*boiler.UserAsset) []*UserAsset {
	var assets []*UserAsset
	for _, ass := range us {
		assets = append(assets, UserAssetFromBoiler(ass))
	}

	return assets
}

func UserAssets1155FromBoiler(us []*boiler.UserAssets1155) []*User1155Asset {
	var assets []*User1155Asset

	// Goes through count and return individual items of 1155 assets
	for _, ass := range us {
		userAssets := UserAsset1155FromBoiler(ass)
		assets = append(assets, userAssets)
	}

	return assets
}

func UserAssetFromBoiler(us *boiler.UserAsset) *UserAsset {
	attribes := []*Attribute{}
	err := us.Attributes.Unmarshal(&attribes)
	if err != nil {
		passlog.L.Error().Err(err).Interface("us.Attributes", us.Attributes).Msg("failed to unmarshall attributes")
	}

	return &UserAsset{
		ID:               us.ID,
		CollectionID:     us.CollectionID,
		TokenID:          us.TokenID,
		Tier:             us.Tier,
		Hash:             us.Hash,
		OwnerID:          us.OwnerID,
		Data:             us.Data,
		Attributes:       attribes,
		AssetType:        us.AssetType,
		Name:             us.Name,
		ImageURL:         us.ImageURL,
		ExternalURL:      us.ExternalURL,
		Description:      us.Description,
		BackgroundColor:  us.BackgroundColor,
		AnimationURL:     us.AnimationURL,
		YoutubeURL:       us.YoutubeURL,
		UnlockedAt:       us.UnlockedAt,
		MintedAt:         us.MintedAt,
		OnChainStatus:    us.OnChainStatus,
		DeletedAt:        us.DeletedAt,
		DataRefreshedAt:  us.DataRefreshedAt,
		UpdatedAt:        us.UpdatedAt,
		CreatedAt:        us.CreatedAt,
		CardAnimationURL: us.CardAnimationURL,
		AvatarURL:        us.AvatarURL,
		LargeImageURL:    us.LargeImageURL,
	}
}

func UserAsset1155CountFromBoiler(us *boiler.UserAssets1155) *[]User1155Asset {
	var userAssets []User1155Asset

	for i := 0; i < us.Count; i++ {
		userAsset := User1155Asset{
			CollectionSlug:  us.R.Collection.Slug,
			ID:              us.ID,
			OwnerID:         us.OwnerID,
			CollectionID:    us.CollectionID,
			ExternalTokenID: us.ExternalTokenID,
			Label:           us.Label,
			Description:     us.Description,
			ImageURL:        us.ImageURL,
			AnimationURL:    us.AnimationURL,
			KeycardGroup:    us.KeycardGroup,
			Attributes:      us.Attributes,
			ServiceID:       us.ServiceID,
			CreatedAt:       us.CreatedAt,
			MintContract:    us.R.Collection.MintContract,
		}

		userAssets = append(userAssets, userAsset)
	}

	return &userAssets
}

func UserAsset1155FromBoiler(us *boiler.UserAssets1155) *User1155Asset {
	userAsset := User1155Asset{
		ID:              us.ID,
		OwnerID:         us.OwnerID,
		CollectionID:    us.CollectionID,
		ExternalTokenID: us.ExternalTokenID,
		Label:           us.Label,
		Description:     us.Description,
		ImageURL:        us.ImageURL,
		AnimationURL:    us.AnimationURL,
		KeycardGroup:    us.KeycardGroup,
		Attributes:      us.Attributes,
		ServiceID:       us.ServiceID,
		CreatedAt:       us.CreatedAt,
		Count:           us.Count,
	}
	if us.R != nil && us.R.Collection != nil {
		userAsset.CollectionSlug = us.R.Collection.Slug
		userAsset.MintContract = us.R.Collection.MintContract
	}
	return &userAsset
}

type SupremacyKeycardAttribute struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value,omitempty"`
}

type TransferEvent struct {
	TransferEventID int64       `json:"transfer_event_id"`
	AssetHash       string      `json:"asset_hash,omitempty"`
	FromUserID      string      `json:"from_user_id,omitempty"`
	ToUserID        string      `json:"to_user_id,omitempty"`
	TransferredAt   time.Time   `json:"transferred_at"`
	TransferTXID    null.String `json:"transfer_tx_id"`
	OwnedService    null.String `json:"owned_service"`
}
