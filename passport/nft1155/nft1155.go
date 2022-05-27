package nft1155

import (
	"database/sql"
	"github.com/friendsofgo/errors"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/rpcclient"
	types2 "xsyn-services/types"
)

type AttributeType struct {
	Attributes []*AttributeInner
}

type AttributeInner struct {
	TraitType string `json:"trait_type"`
	Value     string `json:"value"`
}

func CreateOrGet1155Asset(externalTokenID int, user *types2.User, collectionSlug string) (*boiler.UserAssets1155, error) {
	collection, err := db.CollectionBySlug(collectionSlug)
	if err != nil {
		return nil, err
	}

	asset, err := boiler.UserAssets1155S(
		boiler.UserAssets1155Where.ExternalTokenID.EQ(externalTokenID),
		boiler.UserAssets1155Where.OwnerID.EQ(user.ID),
		boiler.UserAssets1155Where.CollectionID.EQ(collection.ID),
		boiler.UserAssets1155Where.ServiceID.IsNull(),
	).One(passdb.StdConn)
	if errors.Is(err, sql.ErrNoRows) {

		assetDetail, err := rpcclient.Get1155Details(externalTokenID, collection.Slug)
		if err != nil {
			return nil, err
		}

		var assetJson types.JSON

		if !assetDetail.Syndicate.Valid {
			assetDetail.Syndicate.String = "N/A"
		}

		inner := &AttributeInner{
			TraitType: "Syndicate",
			Value:     assetDetail.Syndicate.String,
		}

		var inners []*AttributeInner

		inners = append(inners, inner)

		aType := &AttributeType{
			Attributes: inners,
		}

		err = assetJson.Marshal(aType)
		if err != nil {
			return nil, err
		}

		asset = &boiler.UserAssets1155{
			OwnerID:         user.ID,
			CollectionID:    collection.ID,
			ExternalTokenID: externalTokenID,
			Label:           assetDetail.Label,
			Description:     assetDetail.Description,
			ImageURL:        assetDetail.ImageURL,
			AnimationURL:    assetDetail.AnimationUrl,
			KeycardGroup:    assetDetail.Group,
			Attributes:      assetJson,
		}

		err = asset.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return nil, err
		}
	}

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to get asset")
	}

	return asset, nil
}
