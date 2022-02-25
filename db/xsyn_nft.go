package db

import (
	"context"
	"fmt"
	"passport"
	"passport/helpers"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// CollectionInsert inserts a new collection
func CollectionInsert(ctx context.Context, conn Conn, collection *passport.Collection) error {
	q := `INSERT INTO collections (name, mint_contract, stake_contract) VALUES($1, $2, $3)`
	_, err := conn.Exec(ctx, q, collection.Name, collection.MintContract, collection.StakeContract)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynMetadataInsert inserts a new item metadata
func XsynMetadataInsert(ctx context.Context, conn Conn, item *passport.XsynMetadata, externalUrl string) error {
	// generate token id
	q := `SELECT count(*) from xsyn_metadata WHERE collection_id = $1`
	err := pgxscan.Get(ctx, conn, &item.ExternalTokenID, q, item.CollectionID)
	if err != nil {
		return terror.Error(err)
	}

	// generate hash
	// TODO: get this to handle uint64
	item.Hash, err = helpers.GenerateMetadataHashID(item.CollectionID.String(), int(item.ExternalTokenID), false)
	if err != nil {
		return terror.Error(err)
	}

	q = `	INSERT INTO xsyn_metadata (hash, external_token_id, name, collection_id, game_object, description, image, animation_url, attributes, additional_metadata, external_url)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING hash, external_token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata, external_url`

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
		fmt.Sprintf("%s/asset/%s", externalUrl, item.Hash))
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

// XsynMetadataAvailableGet return a available nft for joining the battle queue
func XsynMetadataAvailableGet(ctx context.Context, conn Conn, userID passport.UserID, hash string) (*passport.XsynMetadata, error) {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT
			xnm.hash, xnm.minted, row_to_json(c) as collection, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes, xa.minting_signature
		FROM 
			xsyn_metadata xnm
		INNER JOIN
			xsyn_assets xa ON xa.metadata_hash = xnm.hash AND xa.user_id = $1 AND xa.metadata_hash = $2 AND xa.frozen_at ISNULL
		INNER JOIN
			collections c on c.id = xnm.collection_id
		WHERE
			xnm.durability = 100
	`
	err := pgxscan.Get(ctx, conn, nft, q,
		userID,
		hash,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return nft, nil
}

// XsynMetadataOwnerGet return metadata that owner by the given user id and token id
func XsynMetadataOwnerGet(ctx context.Context, conn Conn, userID passport.UserID, hash string) (*passport.XsynMetadata, error) {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT
			xnm.hash, xnm.minted, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes, xa.frozen_at, xa.locked_by_id, xa.minting_signature
		FROM 
			xsyn_metadata xnm
		INNER JOIN
			xsyn_assets xa ON xa.metadata_hash = xnm.hash
		WHERE xa.user_id = $1 AND xa.metadata_hash = $2
	`
	err := pgxscan.Get(ctx, conn, nft, q,
		userID,
		hash,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return nft, nil
}

// XsynMetadataGet return metadata from token id
func XsynMetadataGet(ctx context.Context, conn Conn, hash string) (*passport.XsynMetadata, error) {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT
			xnm.hash, xnm.minted, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM 
			xsyn_metadata xnm
		WHERE
			xnm.hash = $1
	`
	err := pgxscan.Get(ctx, conn, nft, q,
		hash,
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return nft, nil
}

// XsynAssetDurabilityBulkUpdate update xsyn NFT metadata durability
func XsynAssetDurabilityBulkUpdate(ctx context.Context, conn Conn, nfts []*passport.WarMachineMetadata) error {
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
		q += fmt.Sprintf("('%s', %d)", nft.Hash, nft.Durability)
		if i < len(nfts)-1 {
			q += ","
			continue
		}
		q += ") AS c(metadata_hash, durability)"
	}

	q += " WHERE c.metadata_hash = xnm.hash;"

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetDurabilityBulkIncrement update xsyn NFT metadata durability
func XsynAssetDurabilityBulkIncrement(ctx context.Context, conn Conn, assetHashes []string) ([]*passport.XsynMetadata, error) {
	nfts := []*passport.XsynMetadata{}
	q := `
		UPDATE
			xsyn_metadata
		SET
			durability = durability + 1
		WHERE
			durability < 100 AND hash IN (
	`
	for i, hash := range assetHashes {
		q += "'" + hash + "'"
		if i < len(assetHashes)-1 {
			q += ","
			continue
		}
		q += ")"
	}

	q += " RETURNING hash, durability;"

	err := pgxscan.Select(ctx, conn, &nfts, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return nfts, nil
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

// XsynAssetUnfreezeableCheck check whether the asset is unfreezeable
func XsynAssetUnfreezeableCheck(ctx context.Context, conn Conn, metadataHash string, userID passport.UserID) error {
	q := `
		SELECT 
			1 
		FROM 
			xsyn_assets
		WHERE
		metadata_hash = $1 AND user_id = $2 AND frozen_at NOTNULL AND locked_by_id ISNULL;
	`
	_, err := conn.Exec(ctx, q, metadataHash, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetBulkLock lock a list of xsyn nfts
func XsynAssetBulkLock(ctx context.Context, conn Conn, assetHashes []string, userID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			locked_by_id = $1
		WHERE
			metadata_hash IN (
	`
	for i, assetHast := range assetHashes {
		q += "'" + assetHast + "'"
		if i < len(assetHashes)-1 {
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

// XsynAssetBulkRelease freeze a xsyn nft
func XsynAssetBulkRelease(ctx context.Context, conn Conn, nfts []*passport.WarMachineMetadata, frozenByID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			frozen_at = NULL,
			frozen_by_id = NULL,
			locked_by_id = NULL
		WHERE
			frozen_by_id = $1 AND frozen_at NOTNULL AND metadata_hash IN (
	`

	for i, nft := range nfts {
		q += "'" + nft.Hash + "'"
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
		SELECT xnm.hash, xnm.minted, xnm.collection_id, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM xsyn_metadata xnm
	 	INNER JOIN xsyn_assets xa ON xa.token_id = xnm.token_id and xnm.collection_id = xa.collection_id
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

func AbilityAssetGet(ctx context.Context, conn Conn, abilityMetadata *passport.AbilityMetadata) error {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT 
			xnm.hash, xnm.minted, xnm.collection_id, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM 
			xsyn_metadata xnm
		WHERE 
			xnm.hash = $1 AND 
			xnm.attributes @> '[{"value": "Ability", "trait_type": "Asset Type"}]';
	`
	err := pgxscan.Get(ctx, conn, nft, q, abilityMetadata.Hash)
	if err != nil {
		return terror.Error(err)
	}

	// parse ability data
	passport.ParseAbilityMetadata(nft, abilityMetadata)

	return nil
}

// WarMachineGetByUserID return all the war machine owned by the user
func WarMachineGetByUserID(ctx context.Context, conn Conn, userID passport.UserID) ([]*passport.XsynMetadata, error) {
	xms := []*passport.XsynMetadata{}

	q := `
		SELECT xm.hash, xm.minted, xm.collection_id, xm.durability, xm.name, xm.description, xm.external_url, xm.image, xm.attributes, xa.minting_signature
		FROM xsyn_metadata xm
		INNER JOIN xsyn_assets xa ON xa.metadata_hash = xm.hash AND xa.user_id = $1
	`
	err := pgxscan.Select(ctx, conn, &xms, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return xms, nil
}

// WarMachineAbilityCostGet return the sups cost of the war machine ability
func WarMachineAbilityCostGet(ctx context.Context, conn Conn, warMachineTokenID, abilityTokenID uint64) (string, error) {
	supsCost := "0"

	q := `
		SELECT 
			sups_cost
		FROM 
			war_machine_ability_sups_cost
		WHERE
			war_machine_token_id = $1 AND ability_token_id = $2
	`

	err := pgxscan.Get(ctx, conn, &supsCost, q, warMachineTokenID, abilityTokenID)
	if err != nil {
		return "", terror.Error(err)
	}

	return supsCost, nil
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
