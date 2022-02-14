package db

import (
	"context"
	"errors"
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
)

// FactionCreate create a new faction
func FactionCreate(ctx context.Context, conn Conn, faction *passport.Faction) error {
	q := `
		INSERT INTO 
			factions (
				id, 
				label, 
				theme, 
				logo_blob_id,
				background_blob_id,
				description
			)
		VALUES
			($1, $2, $3, $4, $5, $6)
		RETURNING 
			id, label
	`

	err := pgxscan.Get(ctx, conn, faction, q,
		faction.ID,
		faction.Label,
		faction.Theme,
		faction.LogoBlobID,
		faction.BackgroundBlobID,
		faction.Description,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func FactionGet(ctx context.Context, conn Conn, factionID passport.FactionID) (*passport.Faction, error) {
	result := &passport.Faction{}

	q := `
		SELECT * FROM factions WHERE id = $1
	`

	err := pgxscan.Get(ctx, conn, result, q, factionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

func FactionAll(ctx context.Context, conn Conn) ([]*passport.Faction, error) {
	result := []*passport.Faction{}

	q := `
		SELECT * FROM factions
	`

	err := pgxscan.Select(ctx, conn, &result, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// FactionGetByUserID
func FactionGetByUserID(ctx context.Context, conn Conn, userID passport.UserID) (*passport.Faction, error) {
	result := &passport.Faction{}
	q := `
		SELECT
			f.*
		FROM 
			factions f
		INNER JOIN 
			users u ON u.faction_id = f.id AND u.id = $1
	`

	err := pgxscan.Get(ctx, conn, result, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// FactionGetRecruitNumber return the number of users recruit the faction
func FactionGetRecruitNumber(ctx context.Context, conn Conn, factionID passport.FactionID) (int64, error) {
	recruitNumber := int64(0)
	q := `
		SELECT
			COUNT(id)
		FROM 
			users u
		WHERE
			u.faction_id = $1
	`

	err := pgxscan.Get(ctx, conn, &recruitNumber, q, factionID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return 0, terror.Error(err)
	}

	return recruitNumber, nil
}
