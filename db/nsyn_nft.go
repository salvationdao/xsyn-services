package db

import (
	"context"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"passport"
)

// NsynNftMetadataInsert inserts a new nft metadata
func NsynNftMetadataInsert(ctx context.Context, conn Conn,
	name,
	game string,
	description,
	externalUrl,
	imageUrl string,
	gameObject interface{},
	attributes []*passport.Attribute,
	additionalMetadata []*passport.AdditionalMetadata,
) (*passport.NsynNftMetadata, error) {
	var result passport.NsynNftMetadata
	q := `	INSERT INTO nsyn_nft_metadata (token_id, name,game, game_object,  description, external_url, image, attributes, additional_metadata)
			VALUES((SELECT nextval('token_id_seq')),$1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING token_id, name, description, external_url, image, attributes`

	err := pgxscan.Get(ctx, conn, &result, q,
		name,
		game,
		gameObject,
		description,
		externalUrl,
		imageUrl,
		attributes,
		additionalMetadata,
	)

	if err != nil {
		return nil, terror.Error(err)
	}

	return &result, nil
}
