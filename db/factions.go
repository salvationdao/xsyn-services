package db

import (
	"context"
	"errors"
	"fmt"
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

func FactionMvpMaterialisedViewCreate(ctx context.Context, conn Conn) error {
	q := fmt.Sprintf(`
		CREATE MATERIALIZED VIEW faction_mvp AS
		SELECT f1.id AS faction_id,f2.sups_voted, f3.id AS mvp_user_id FROM (
			SELECT f.id FROM factions f
		)f1 LEFT JOIN LATERAL(
			SELECT SUM(t.amount) AS sups_voted FROM transactions t
			INNER JOIN (
					SELECT id, faction_id FROM users u
				) u ON u.id = t.debit AND u.faction_id = f1.id
			WHERE t.credit = '%[1]s' AND t.status = 'success'
		)f2 ON true LEFT JOIN LATERAL(
			SELECT u.id FROM transactions t 
			INNER JOIN (
					SELECT id, faction_id FROM users u
				) u ON u.id = t.debit AND u.faction_id = f1.id
			WHERE t.credit = '%[1]s' AND t.status = 'success'
			GROUP BY u.id
			ORDER BY SUM(t.amount) desc
			LIMIT 1
		)f3 ON true
	`, passport.SupremacyBattleUserID)
	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	q = `
		CREATE UNIQUE INDEX faction_id ON faction_mvp (faction_id);
	`
	_, err = conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func FactionMvpMaterialisedViewRefresh(ctx context.Context, conn Conn) error {
	q := `
		REFRESH MATERIALIZED VIEW faction_mvp;
	`
	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func FactionMvpGet(ctx context.Context, conn Conn, factionID passport.FactionID) (*passport.UserBrief, error) {
	user := &passport.UserBrief{}
	q := `
		SELECT u.id, u.username, u.avatar_id FROM users u 
		INNER JOIN faction_mvp fm ON fm.mvp_user_id = u.id AND fm.faction_id = $1
	`
	err := pgxscan.Get(ctx, conn, user, q, factionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return user, nil
}

func FactionSupsVotedGet(ctx context.Context, conn Conn, factionID passport.FactionID) (*passport.BigInt, error) {
	var wrap struct {
		Sups passport.BigInt `db:"sups_voted"`
	}
	q := `
		SELECT COALESCE(sups_voted,0) AS sups_voted FROM faction_mvp WHERE faction_id = $1
	`
	err := pgxscan.Get(ctx, conn, &wrap, q, factionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return &wrap.Sups, nil
}
