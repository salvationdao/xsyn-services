package db

import (
	"passport/db/boiler"
	"passport/passdb"

	"github.com/gofrs/uuid"
)

func APIKey(apiKeyID uuid.UUID) (*boiler.APIKey, error) {
	return boiler.FindAPIKey(passdb.StdConn, apiKeyID.String())
}
