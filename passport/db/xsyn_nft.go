package db

import (
	"context"
	"fmt"
	"xsyn-services/boiler"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// XsynMetadataInsert inserts a new item metadata
func XsynMetadataInsert(ctx context.Context, conn Conn, item *types.XsynMetadata, externalUrl string) error {
	// generate token id
	q := `SELECT coalesce(max(external_token_id), 0) from xsyn_metadata WHERE collection_id = $1`
	err := pgxscan.Get(ctx, conn, &item.ExternalTokenID, q, item.CollectionID)
	if err != nil {
		return terror.Error(err)
	}
	item.ExternalTokenID++
	// generate hash
	// TODO: get this to handle uint64
	item.Hash, err = helpers.GenerateMetadataHashID(item.CollectionID.String(), int(item.ExternalTokenID), false)
	if err != nil {
		return terror.Error(err)
	}

	q = `	INSERT INTO xsyn_metadata (hash, external_token_id, name, collection_id, game_object, description, image, animation_url, attributes, additional_metadata, external_url, image_avatar)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
			RETURNING hash, external_token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata, external_url, image_avatar`

	err = pgxscan.Get(ctx, conn, item, q,
		item.Hash,
		item.ExternalTokenID,
		item.Name,
		item.CollectionID,
		item.GameObject,
		item.Description,
		item.Image,
		item.AnimationURL,
		item.Attributes,
		item.AdditionalMetadata,
		fmt.Sprintf("%s/asset/%s", externalUrl, item.Hash),
		item.ImageAvatar,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// DefaultWarMachineGet return given amount of default war machines for given faction
func DefaultWarMachineGet(ctx context.Context, conn Conn, userID types.UserID) ([]*boiler.PurchasedItem, error) {
	return boiler.PurchasedItems(boiler.PurchasedItemWhere.IsDefault.EQ(true)).All(passdb.StdConn)

}
