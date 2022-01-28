package db

import (
	"context"
	"fmt"
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// XsynNftMetadataInsert inserts a new nft metadata
func XsynNftMetadataInsert(ctx context.Context, conn Conn, nft *passport.XsynNftMetadata) error {
	q := `	INSERT INTO xsyn_nft_metadata (token_id, name, collection, game_object,  description, external_url, image, attributes, additional_metadata)
			VALUES((SELECT nextval('token_id_seq')),$1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING token_id,  name, description, external_url, image, attributes`

	err := pgxscan.Get(ctx, conn, nft, q,
		nft.Name,
		nft.Collection,
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

// XsynNftMetadataAssignUser assign a nft metadata to a user
func XsynNftMetadataAssignUser(ctx context.Context, conn Conn, nftTokenID uint64, userID passport.UserID) error {
	q := `
		INSERT INTO 
			xsyn_assets (token_id, owner_id)
		VALUES
			($1, $2);
	`

	_, err := conn.Exec(ctx, q, nftTokenID, userID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// XsynNftMetadataAvailableGet return a available nft for joining the battle queue
func XsynNftMetadataAvailableGet(ctx context.Context, conn Conn, userID passport.UserID, nftTokenID uint64) (*passport.XsynNftMetadata, error) {
	nft := &passport.XsynNftMetadata{}
	q := `
		SELECT
			xnm.token_id, xnm.collection, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM 
			xsyn_nft_metadata xnm
		INNER JOIN
			xsyn_assets xa ON xa.token_id = xnm.token_id AND xa.user_id = $1 AND xa.token_id = $2 AND xa.frozen_at ISNULL
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

// XsynNftMetadataDurabilityUpdate update xsyn NFT metadata durability
func XsynNftMetadataDurabilityBulkUpdate(ctx context.Context, conn Conn, nfts []*passport.WarMachineNFT) error {
	q := `
		UPDATE 
			xsyn_nft_metadata xnm
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

// XsynNftMetadataDurabilityBulkIncrement update xsyn NFT metadata durability
func XsynNftMetadataDurabilityBulkIncrement(ctx context.Context, conn Conn) error {
	q := `
		UPDATE
			xsyn_nft_metadata
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
			token_id = $1 AND owner_id = $2 AND frozen_at NOTNULL AND locked_by_id ISNULL;
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

	_, err := conn.Exec(ctx, q, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetRelease freeze a xsyn nft
func XsynAssetBulkRelease(ctx context.Context, conn Conn, nfts []*passport.WarMachineNFT, frozenByID passport.UserID) error {
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
