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
			factions (id, label, colour)
		VALUES
			($1, $2, $3)
		RETURNING 
			id, label
	`

	err := pgxscan.Get(ctx, conn, faction, q, faction.ID, faction.Label, faction.Colour)
	if err != nil {
		return terror.Error(err)
	}

	return nil
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
