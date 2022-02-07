package db

import (
	"context"
	"fmt"
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// Collection inserts a new collection
func CollectionInsert(ctx context.Context, conn Conn, collection *passport.Collection) error {
	q := `INSERT INTO collections (name) VALUES($1)`
	_, err := conn.Exec(ctx, q, collection.Name)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynMetadataInsert inserts a new item metadata
func XsynMetadataInsert(ctx context.Context, conn Conn, nft *passport.XsynMetadata) error {
	q := `	INSERT INTO xsyn_metadata (token_id, name, collection_id, game_object, description, external_url, image, attributes, additional_metadata)
			VALUES((SELECT nextval('token_id_seq')),$1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING token_id,  name, description, external_url, image, attributes`

	err := pgxscan.Get(ctx, conn, nft, q,
		nft.Name,
		nft.CollectionID,
		nft.GameObject,
		nft.Description,
		nft.ExternalUrl,
		nft.Image,
		nft.Attributes,
		nft.AdditionalMetadata,
	)

	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynMetadataAssignUser assign a nft metadata to a user
func XsynMetadataAssignUser(ctx context.Context, conn Conn, nftTokenID uint64, userID passport.UserID) error {
	q := `
		INSERT INTO 
			xsyn_assets (token_id, user_id)
		VALUES
			($1, $2);
	`

	_, err := conn.Exec(ctx, q, nftTokenID, userID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// XsynMetadataAvailableGet return a available nft for joining the battle queue
func XsynMetadataAvailableGet(ctx context.Context, conn Conn, userID passport.UserID, nftTokenID uint64) (*passport.XsynMetadata, error) {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT
			xnm.token_id, row_to_json(c) as collection, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM 
			xsyn_metadata xnm
		INNER JOIN
			xsyn_assets xa ON xa.token_id = xnm.token_id AND xa.user_id = $1 AND xa.token_id = $2 AND xa.frozen_at ISNULL
		INNER JOIN
			collections c on c.id = xnm.collection_id
		WHERE
			xnm.durability = 100
	`
	err := pgxscan.Get(ctx, conn, nft, q,
		userID,
		nftTokenID,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return nft, nil
}

// XsynAsseetDurabilityBulkUpdate update xsyn NFT metadata durability
func XsynAsseetDurabilityBulkUpdate(ctx context.Context, conn Conn, nfts []*passport.WarMachineMetadata) error {
	q := `
		UPDATE 
			xsyn_metadata xnm
		SET
			durability = c.durability
		FROM 
			(
				VALUES

	`

	for i, nft := range nfts {
		q += fmt.Sprintf("(%d, %d)", nft.TokenID, nft.Durability)
		if i < len(nfts)-1 {
			q += ","
			continue
		}
		q += ") AS c(token_id, durability)"
	}

	q += " WHERE c.token_id = xnm.token_id;"

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetDurabilityBulkIncrement update xsyn NFT metadata durability
func XsynAssetDurabilityBulkIncrement(ctx context.Context, conn Conn) error {
	q := `
		UPDATE
			xsyn_metadata
		SET
			durability = durability + 1
		WHERE
			durability < 100
	`
	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetFreeze freeze a xsyn nft
func XsynAssetFreeze(ctx context.Context, conn Conn, nftTokenID uint64, userID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			frozen_at = NOW(),
			frozen_by_id = $2
		WHERE
			token_id = $1 AND frozen_at ISNULL;
	`
	_, err := conn.Exec(ctx, q, nftTokenID, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetUnfreezeableCheck check whether the asset is unfreezeable
func XsynAssetUnfreezeableCheck(ctx context.Context, conn Conn, nftTokenID uint64, userID passport.UserID) error {
	q := `
		SELECT 
			1 
		FROM 
			xsyn_assets
		WHERE
			token_id = $1 AND user_id = $2 AND frozen_at NOTNULL AND locked_by_id ISNULL;
	`
	_, err := conn.Exec(ctx, q, nftTokenID, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetBulkLock lock a list of xsyn nfts
func XsynAssetBulkLock(ctx context.Context, conn Conn, nftTokenIDs []uint64, userID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			locked_by_id = $1
		WHERE
			token_id IN (
	`
	for i, nftTokenID := range nftTokenIDs {
		q += fmt.Sprintf("%d", nftTokenID)
		if i < len(nftTokenIDs)-1 {
			q += ","
			continue
		}

		q += ")"
	}

	// don't lock if owned by faction account
	q += `AND user_id NOT IN ('1a657a32-778e-4612-8cc1-14e360665f2b', '305da475-53dc-4973-8d78-a30d390d3de5','15f29ee9-e834-4f76-aff8-31e39faabe2d')`

	_, err := conn.Exec(ctx, q, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetRelease freeze a xsyn nft
func XsynAssetBulkRelease(ctx context.Context, conn Conn, nfts []*passport.WarMachineMetadata, frozenByID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			frozen_at = NULL,
			frozen_by_id = NULL,
			locked_by_id = NULL
		WHERE
			frozen_by_id = $1 AND frozen_at NOTNULL AND token_id IN (
	`

	for i, nft := range nfts {
		q += fmt.Sprintf("%d", nft.TokenID)
		if i < len(nfts)-1 {
			q += ","
			continue
		}
		q += ");"
	}

	_, err := conn.Exec(ctx, q, frozenByID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// DefaultWarMachineGet return given amount of default war machines for given faction
func DefaultWarMachineGet(ctx context.Context, conn Conn, userID passport.UserID, amount int) ([]*passport.XsynMetadata, error) {
	nft := []*passport.XsynMetadata{}
	q := `
		SELECT xnm.token_id, xnm.collection_id, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM xsyn_metadata xnm
				 INNER JOIN xsyn_assets xa ON xa.token_id = xnm.token_id
		WHERE xa.user_id = $1
		AND xnm.attributes @> '[{"value": "War Machine", "trait_type": "Asset Type"}]'
		LIMIT $2
	`

	// TODO: find a better way to get the default warmachines out
	err := pgxscan.Select(ctx, conn, &nft, q,
		userID,
		amount,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return nft, nil
}
