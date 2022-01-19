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
			factions (id, label, theme)
		VALUES
			($1, $2, $3)
		RETURNING 
			id, label
	`

	err := pgxscan.Get(ctx, conn, faction, q, faction.ID, faction.Label, faction.Theme)
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
