package db

import (
	"context"
	"fmt"
	"passport"
	"passport/helpers"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// XsynMetadataInsert inserts a new item metadata
func XsynMetadataInsert(ctx context.Context, conn Conn, item *passport.XsynMetadata, externalUrl string) error {
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

// XsynMetadataAssignUser assign a nft metadata to a user
func XsynMetadataAssignUser(ctx context.Context, conn Conn, metadataHash string, userID passport.UserID, collectionID passport.CollectionID, externalTokenID uint64) error {
	q := `
		INSERT INTO 
			xsyn_assets (metadata_hash, user_id, collection_id, external_token_id)
		VALUES
			($1, $2, $3, $4);
	`

	_, err := conn.Exec(ctx, q, metadataHash, userID, collectionID, externalTokenID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// XsynAssetFreeze freeze a xsyn nft
func XsynAssetFreeze(ctx context.Context, conn Conn, assetHash string, userID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			frozen_at = NOW(),
			frozen_by_id = $2
		WHERE
			metadata_hash = $1 AND frozen_at ISNULL;
	`
	_, err := conn.Exec(ctx, q, assetHash, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// DefaultWarMachineGet return given amount of default war machines for given faction
func DefaultWarMachineGet(ctx context.Context, conn Conn, userID passport.UserID) ([]*passport.XsynMetadata, error) {
	nft := []*passport.XsynMetadata{}
	q := `
		SELECT xnm.hash, xnm.minted, xnm.collection_id, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM xsyn_metadata xnm
	 	INNER JOIN xsyn_assets xa ON xa.metadata_hash = xnm.hash and xnm.collection_id = xa.collection_id
		WHERE xa.user_id = $1
		AND xnm.attributes @> '[{"value": "War Machine", "trait_type": "Asset Type"}]'
	`

	// TODO: find a better way to get the default warmachines out
	err := pgxscan.Select(ctx, conn, &nft, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return nft, nil
}

// XsynAssetLock locks a asset
func XsynAssetLock(ctx context.Context, conn Conn, assetHash string, userID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			locked_by_id = $1
		WHERE metadata_hash = $2`

	_, err := conn.Exec(ctx, q, userID, assetHash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetMintLock sets minting_signature of an asset
func XsynAssetMintLock(ctx context.Context, conn Conn, assetHash string, sig string, expiry string) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			minting_signature = $1, signature_expiry = $2
		WHERE metadata_hash = $3`

	_, err := conn.Exec(ctx, q, sig, expiry, assetHash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetMintUnLock removed all mint locks from a users assets, should only be called when they complete another mint
func XsynAssetMintUnLock(ctx context.Context, conn Conn, userID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			minting_signature = ''
		WHERE user_id = $1`

	_, err := conn.Exec(ctx, q, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetMinted marked as a
func XsynAssetMinted(ctx context.Context, conn Conn, assetHash string) error {
	q := `
		UPDATE 
			xsyn_metadata
		SET
			minted = true
		WHERE hash = $1`

	_, err := conn.Exec(ctx, q, assetHash)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
