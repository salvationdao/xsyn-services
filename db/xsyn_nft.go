package db

import (
	"context"
	"fmt"
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// CollectionInsert inserts a new collection
func CollectionInsert(ctx context.Context, conn Conn, collection *passport.Collection) error {
	q := `INSERT INTO collections (name) VALUES($1)`
	_, err := conn.Exec(ctx, q, collection.Name)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynMetadataInsert inserts a new item metadata
func XsynMetadataInsert(ctx context.Context, conn Conn, item *passport.XsynMetadata, externalUrl string) error {
	q := `	INSERT INTO xsyn_metadata (token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata)
			VALUES((SELECT nextval('token_id_seq')),$1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata`

	err := pgxscan.Get(ctx, conn, item, q, item.Name, item.CollectionID, item.GameObject, item.Description, item.Image, item.AnimationURL, item.Attributes, item.AdditionalMetadata)
	if err != nil {
		return terror.Error(err)
	}
	updateQ := `UPDATE xsyn_metadata 
				SET external_url = $1 
				WHERE token_id = $2
				RETURNING external_url`
	err = pgxscan.Get(ctx, conn, item, updateQ , fmt.Sprintf("%s/asset/%d", externalUrl, item.TokenID), item.TokenID)
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
			xnm.token_id, xnm.minted, row_to_json(c) as collection, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes, xa.minting_signature
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

// XsynMetadataOwnerGet return metadata that owner by the given user id and token id
func XsynMetadataOwnerGet(ctx context.Context, conn Conn, userID passport.UserID, nftTokenID uint64) (*passport.XsynMetadata, error) {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT
			xnm.token_id, xnm.minted, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes, xa.frozen_at, xa.locked_by_id, xa.minting_signature
		FROM 
			xsyn_metadata xnm
		INNER JOIN
			xsyn_assets xa ON xa.token_id = xnm.token_id AND xa.user_id = $1 AND xa.token_id = $2
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

// XsynMetadataGet return metadata from token id
func XsynMetadataGet(ctx context.Context, conn Conn, nftTokenID uint64) (*passport.XsynMetadata, error) {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT
			xnm.token_id, xnm.minted, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM 
			xsyn_metadata xnm
		WHERE
			xnm.token_id = $1
	`
	err := pgxscan.Get(ctx, conn, nft, q,
		nftTokenID,
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
func XsynAssetDurabilityBulkIncrement(ctx context.Context, conn Conn, tokenIDs []uint64) ([]*passport.XsynMetadata, error) {
	nfts := []*passport.XsynMetadata{}
	q := `
		UPDATE
			xsyn_metadata
		SET
			durability = durability + 1
		WHERE
			durability < 100 AND token_id IN (
	`
	for i, tokenID := range tokenIDs {
		q += fmt.Sprintf("%v", tokenID)
		if i < len(tokenIDs)-1 {
			q += ","
			continue
		}
		q += ")"
	}

	q += " RETURNING token_id, durability;"

	err := pgxscan.Select(ctx, conn, &nfts, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return nfts, nil
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
		SELECT xnm.token_id, xnm.minted, xnm.collection_id, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
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

// AbilityAssetGet
func AbilityAssetGet(ctx context.Context, conn Conn, abilityMetadata *passport.AbilityMetadata) error {
	nft := &passport.XsynMetadata{}
	q := `
		SELECT 
			xnm.token_id, xnm.minted, xnm.collection_id, xnm.durability, xnm.name, xnm.description, xnm.external_url, xnm.image, xnm.attributes
		FROM 
			xsyn_metadata xnm
		WHERE 
			xnm.token_id = $1 AND 
			xnm.attributes @> '[{"value": "Ability", "trait_type": "Asset Type"}]';
	`
	err := pgxscan.Get(ctx, conn, nft, q, abilityMetadata.TokenID)
	if err != nil {
		return terror.Error(err)
	}

	// parse ability data
	passport.ParseAbilityMetadata(nft, abilityMetadata)

	return nil
}

// WarMachineAbilitySet
func WarMachineAbilitySet(ctx context.Context, conn Conn, warMachineTokenID uint64, abilityTokenID uint64, warMachineAbilitySlot passport.WarMachineAttField) error {
	if warMachineAbilitySlot != passport.WarMachineAttFieldAbility01 && warMachineAbilitySlot != passport.WarMachineAttFieldAbility02 {
		return terror.Error(fmt.Errorf("invalid attribute slot"))
	}

	q := fmt.Sprintf(`
	UPDATE 	
		xsyn_metadata xm
	SET
		attributes = (
			select xm3.att || '{"trait_type": "%s", "value": "none", "token_id": %d}'
			from (
				SELECT xm2.token_id ,to_jsonb(array_agg(elem)) AS att
				FROM   xsyn_metadata xm2, jsonb_array_elements(xm2."attributes") elem
				WHERE  xm2.token_id = $1 and elem ->> 'trait_type' <> '%s'
				group by token_id
			) xm3
		)
	WHERE 
		xm.token_id = $1;
	`, warMachineAbilitySlot, abilityTokenID, warMachineAbilitySlot)
	_, err := conn.Exec(ctx, q, warMachineTokenID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// WarMachineGetByUserID return all the war machine owned by the user
func WarMachineGetByUserID(ctx context.Context, conn Conn, userID passport.UserID) ([]*passport.XsynMetadata, error) {
	xms := []*passport.XsynMetadata{}

	q := `
		SELECT xm.token_id, xm.minted, xm.collection_id, xm.durability, xm.name, xm.description, xm.external_url, xm.image, xm.attributes, xa.minting_signature
		FROM xsyn_metadata xm
		INNER JOIN xsyn_assets xa ON xa.token_id = xm.token_id AND xa.user_id = $1
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

// WarMachineAbilityCostUpsert Upsert the sups cost of a war machine ability
func WarMachineAbilityCostUpsert(ctx context.Context, conn Conn, warMachineTokenID, abilityTokenID uint64, supsCost string) error {
	q := `
		INSERT INTO 
			war_machine_ability_sups_cost (war_machine_token_id, ability_token_id, sups_cost)
		VALUES
			($1, $2, $3)
		ON CONFLICT
			(war_machine_token_id, ability_token_id)
		DO UPDATE SET
			sups_cost = $3
	`

	_, err := conn.Exec(ctx, q, warMachineTokenID, abilityTokenID, supsCost)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetLock locks a asset
func XsynAssetLock(ctx context.Context, conn Conn, tokenID uint64, userID passport.UserID) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			locked_by_id = $1
		WHERE token_id = $2`

	_, err := conn.Exec(ctx, q, userID, tokenID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// XsynAssetMintLock sets minting_signature of an asset
func XsynAssetMintLock(ctx context.Context, conn Conn, tokenID uint64, sig string) error {
	q := `
		UPDATE 
			xsyn_assets
		SET
			minting_signature = $1
		WHERE token_id = $2`

	_, err := conn.Exec(ctx, q, sig, tokenID)
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
func XsynAssetMinted(ctx context.Context, conn Conn, tokenID uint64) error {
	q := `
		UPDATE 
			xsyn_metadata
		SET
			minted = true
		WHERE token_id = $1`

	_, err := conn.Exec(ctx, q, tokenID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
