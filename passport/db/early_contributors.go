package db

import (
	"context"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
)

type EarlyContributor struct {
	ID                uuid.UUID   `json:"id" db:"id"`
	UserPublicAddress null.String `json:"user_public_address" db:"user_public_address"`
	Message           null.String `json:"message" db:"message"`
	MessageHex        null.String `json:"message_hex" db:"message_hex"`
	SignatureHex      null.String `json:"signature_hex" db:"signature_hex"`
	SignerAddressHex  null.String `json:"signer_address_hex" db:"signer_address_hex"`
	Agree             null.Bool   `json:"agree" db:"agree"`
	SignedAt          time.Time   `json:"signed_at" db:"signed_at"`
	CreatedAt         time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at" db:"updated_at"`
	DeletedAt         *time.Time  `json:"deleted_at" db:"deleted_at"`
}

func IsUserEarlyContributor(ctx context.Context, conn Conn, address string) (bool, EarlyContributor, error) {
	earlyContributor := EarlyContributor{}
	q := `SELECT * FROM saft_agreements WHERE user_public_address ILIKE $1;`
	err := pgxscan.Get(ctx, conn, &earlyContributor, q, address)
	notFoundError := pgxscan.NotFound(err)
	if notFoundError {
		return false, earlyContributor, err //This error can't be wrapped in terror
	}
	if err != nil {
		return false, earlyContributor, terror.Error(err, "error finding user")
	}
	return true, earlyContributor, nil
}

func UserSignMessage(ctx context.Context, conn Conn, address, message, signature, messageHex string, agree bool) error {
	q := `UPDATE saft_agreements SET message = $2, signature_hex = $3, signer_address_hex = $4, agree = $5, message_hex = $6, signed_at = $7 WHERE user_public_address ILIKE $1;`
	_, err := conn.Exec(ctx, q, address, message, signature, address, agree, messageHex, time.Now())
	if err != nil {
		return err
	}
	return nil
}