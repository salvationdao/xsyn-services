package db

import (
	"context"
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func UserHasPassword(ctx context.Context, conn Conn, user *types.User) (*bool, error) {
	count := 0
	q := `
		SELECT COUNT(*)
		FROM password_hashes
		WHERE user_id = $1
	`
	err := pgxscan.Get(ctx, conn, &count, q, user.ID)
	if err != nil {
		return nil, terror.Error(err)
	}
	hasPassword := count > 0
	return &hasPassword, nil
}

// HashByUserID returns a user's password hash
func HashByUserID(ctx context.Context, conn Conn, userID types.UserID) (string, error) {
	result := ""
	q := `
		SELECT password_hash
		FROM password_hashes
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1`
	err := pgxscan.Get(ctx, conn, &result, q, userID)
	if err != nil {
		return "", terror.Error(err)
	}
	return result, nil
}

// AuthRegister will create a new user and insert password hash
func AuthRegister(ctx context.Context, conn Conn, user *types.User, passwordHashedB64 string) error {
	usernameOK, err := UsernameAvailable(ctx, conn, user.Username, nil)
	if err != nil {
		return terror.Error(err)
	}
	if !usernameOK {
		return terror.Error(fmt.Errorf("username is taken: %s", user.Username))
	}

	q := `--sql
		WITH new_user AS (
 			INSERT INTO users (first_name, last_name, email, role_id, username)
				VALUES ($2, $3, $4, $5, $6)
				RETURNING id, first_name, last_name, email, role_id, verified, old_password_required, username
		),
			new_hash AS (INSERT INTO password_hashes (user_id, password_hash)
				VALUES ((SELECT id FROM new_user), $1))
		SELECT * from new_user`
	err = pgxscan.Get(
		ctx,
		conn,
		user,
		q,
		passwordHashedB64,
		user.FirstName,
		user.LastName,
		user.Email,
		user.RoleID,
		user.Username,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// AuthDeactivateUserPasswordHashes will deactivate a user's password
func AuthDeactivateUserPasswordHashes(ctx context.Context, conn Conn, userID types.UserID) error {
	q := `
		DELETE FROM password_hashes
		WHERE user_id = $1`
	_, err := conn.Exec(ctx, q, userID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// AuthSetPasswordHash will set the user's password
func AuthSetPasswordHash(ctx context.Context, conn Conn, userID types.UserID, passwordHashedB64 string) error {
	q := `--sql
		INSERT INTO password_hashes (user_id, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET password_hash = EXCLUDED.password_hash
		`
	_, err := conn.Exec(ctx, q, userID, passwordHashedB64)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// AuthFindToken will get a token via token uuid
func AuthFindToken(ctx context.Context, conn Conn, tokenID types.IssueTokenID) (*types.IssueToken, error) {
	result := &types.IssueToken{}

	q := `
		SELECT id, user_id
		FROM issue_tokens
		WHERE id = $1`
	err := pgxscan.Get(ctx, conn, result, q, tokenID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

// AuthRemoveTokensFromUserID will remove tokens connected to a users id
func AuthRemoveTokensFromUserID(ctx context.Context, conn Conn, userID types.UserID) error {
	q := `
		DELETE FROM issue_tokens
		WHERE user_id = $1 `
	_, err := conn.Exec(ctx, q, userID.String())
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// AuthRemoveTokenWithID will remove a token with given id
func AuthRemoveTokenWithID(ctx context.Context, conn Conn, id types.IssueTokenID) error {
	q := `
		DELETE FROM issue_tokens
		WHERE id = $1 `
	_, err := conn.Exec(ctx, q, id)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// AuthSaveToken adds token to DB
func AuthSaveToken(ctx context.Context, conn Conn, tokenID types.IssueTokenID, userID types.UserID, userAgent string, expiredAt time.Time) error {
	it := boiler.IssueToken{
		ID:        tokenID.String(),
		UserID:    userID.String(),
		UserAgent: userAgent,
		ExpiresAt: null.TimeFrom(expiredAt),
	}

	err := it.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err)
	}
	return nil
}
