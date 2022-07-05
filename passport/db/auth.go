package db

import (
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

func UserHasPassword(userID string) (*bool, error) {
	count := 0
	q := `
		SELECT COUNT(*)
		FROM password_hashes
		WHERE user_id = $1
	`
	err := passdb.StdConn.QueryRow(q, userID).Scan(&count)
	if err != nil {
		return nil, err
	}
	hasPassword := count > 0
	return &hasPassword, nil
}

// HashByUserID returns a user's password hash
func HashByUserID(userID string) (string, error) {
	result := ""
	q := `
		SELECT password_hash
		FROM password_hashes
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 1`
	err := passdb.StdConn.QueryRow(q, userID).Scan(&result)
	if err != nil {
		return "", err
	}
	
	return result, nil
}

// AuthSetPasswordHash will set the user's password
func AuthSetPasswordHash(tx boil.Executor, userID string, passwordHashedB64 string) error {
	q := `--sql
		INSERT INTO password_hashes (user_id, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET password_hash = EXCLUDED.password_hash
		`
	_, err := tx.Exec(q, userID, passwordHashedB64)
	if err != nil {
		return err
	}
	return nil
}

// AuthRemoveTokenWithID will remove a token with given id
func AuthRemoveTokenWithID(id types.IssueTokenID) error {
	q := `
		DELETE FROM issue_tokens
		WHERE id = $1 `
	_, err := passdb.StdConn.Exec(q, id)
	if err != nil {
		return err
	}
	return nil
}
