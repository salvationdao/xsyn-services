package api

import (
	"passport/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

// AuthController holds handlers for roles
type AuthController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewAuthController creates the role hub
func NewAuthController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *AuthController {
	authHub := &AuthController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "role_hub"),
		API:  api,
	}

	return authHub
}
