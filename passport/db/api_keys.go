package db

import (
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"

	"github.com/gofrs/uuid"
)

func APIKey(apiKeyID uuid.UUID) (*boiler.APIKey, error) {
	return boiler.FindAPIKey(passdb.StdConn, apiKeyID.String())
}
