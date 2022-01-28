package db

import (
	"context"
	"passport"

	"github.com/georgysavva/scany/pgxscan"
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
				description,
				velocity,
				share_percent,
				recruit_number,
				win_count,
				loss_count,
				kill_count,
				death_count,
				mvp
			)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
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
		faction.Velocity,
		faction.SharePercent,
		faction.RecruitNumber,
		faction.WinCount,
		faction.LossCount,
		faction.KillCount,
		faction.DeathCount,
		faction.MVP,
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
