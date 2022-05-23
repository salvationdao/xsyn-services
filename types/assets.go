package types

import (
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/types"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passlog"
)

type UserAsset struct {
	ID              string      `json:"id"`
	CollectionID    string      `json:"collection_id"`
	TokenID         int64         `json:"token_id"`
	Tier            string      `json:"tier"`
	Hash            string      `json:"hash"`
	OwnerID         string      `json:"owner_id"`
	Data            types.JSON  `json:"data"`
	Attributes      []*Attribute  `json:"attributes"`
	Name            string      `json:"name"`
	ImageURL        null.String `json:"image_url,omitempty"`
	ExternalURL     null.String `json:"external_url,omitempty"`
	Description     null.String `json:"description,omitempty"`
	BackgroundColor null.String `json:"background_color,omitempty"`
	AnimationURL    null.String `json:"animation_url,omitempty"`
	YoutubeURL      null.String `json:"youtube_url,omitempty"`
	CardAnimationURL null.String `json:"card_animation_url,omitempty"`
	AvatarURL        null.String `json:"avatar_url,omitempty"`
	LargeImageURL    null.String `json:"large_image_url,omitempty"`
	UnlockedAt      time.Time   `json:"unlocked_at"`
	MintedAt        null.Time   `json:"minted_at,omitempty"`
	OnChainStatus   string      `json:"on_chain_status"`
	XsynLocked      null.Bool   `json:"xsyn_locked,omitempty"`
	DeletedAt       null.Time   `json:"deleted_at,omitempty"`
	DataRefreshedAt time.Time   `json:"data_refreshed_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	CreatedAt       time.Time   `json:"created_at"`
}


type Attribute struct {
	DisplayType DisplayType `json:"display_type,omitempty"`
	TraitType   string      `json:"trait_type"`
	TokenID     uint64      `json:"token_id,omitempty"`
	Value       interface{} `json:"value"` // string or number only
}

type DisplayType string

const (
	BoostNumber     DisplayType = "boost_number"
	BoostPercentage DisplayType = "boost_percentage"
	Number          DisplayType = "number"
	Date            DisplayType = "date"
)


func UserAssetsFromBoiler(us []*boiler.UserAsset) []*UserAsset {
var assets []*UserAsset
	for _, ass := range us {
		assets = append(assets, UserAssetFromBoiler(ass))
	}

return assets
}

func UserAssetFromBoiler(us *boiler.UserAsset) *UserAsset {
	var attribes []*Attribute
	 err := us.Attributes.Unmarshal(attribes)
	 if err != nil {
		 passlog.L.Error().Err(err).Interface("us.Attributes", us.Attributes).Msg("failed to unmarshall attributes")
	 }
	return &UserAsset{
		ID:us.ID,
		CollectionID:us.CollectionID,
		TokenID:us.TokenID,
		Tier:us.Tier,
		Hash:us.Hash,
		OwnerID:us.OwnerID,
		Data:us.Data,
		Attributes:attribes,
		Name:us.Name,
		ImageURL:us.ImageURL,
		ExternalURL:us.ExternalURL,
		Description:us.Description,
		BackgroundColor:us.BackgroundColor,
		AnimationURL:us.AnimationURL,
		YoutubeURL:us.YoutubeURL,
		UnlockedAt:us.UnlockedAt,
		MintedAt:us.MintedAt,
		OnChainStatus:us.OnChainStatus,
		XsynLocked:us.XsynLocked,
		DeletedAt:us.DeletedAt,
		DataRefreshedAt:us.DataRefreshedAt,
		UpdatedAt:us.UpdatedAt,
		CreatedAt:us.CreatedAt,
		CardAnimationURL: us.CardAnimationURL,
		AvatarURL: us.AvatarURL,
		LargeImageURL: us.LargeImageURL,
	}
}