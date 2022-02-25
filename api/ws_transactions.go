package api

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
)

// TransactionController holds handlers for transaction endpoints
type TransactionController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewTransactionController creates the user hub
func NewTransactionController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *TransactionController {
	transactionHub := &TransactionController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "user_hub"),
		API:  api,
	}

	// api.Command(HubKeyUserGet, transactionHub.GetHandler) // Perm check inside handler (users can get themselves; need UserRead permission to get other users)
	// api.SecureCommand(HubKeyUserUpdate, transactionHub.UpdateHandler)
	// api.SecureCommandWithPerm(HubKeyUserForceDisconnect, transactionHub.ForceDisconnectHandler, passport.PermUserForceDisconnect)

	// api.SubscribeCommand(HubKeyUserForceDisconnected, transactionHub.ForceDisconnectedHandler)
	// api.SubscribeCommand(HubKeyUserSubscribe, transactionHub.UpdatedSubscribeHandler)
	// api.SubscribeCommand(HubKeyUserOnlineStatus, transactionHub.OnlineStatusSubscribeHandler)
	// api.SubscribeCommand(HubKeySUPSRemainingSubscribe, transactionHub.TotalSupRemainingHandler)
	// api.SubscribeCommand(HubKeySUPSExchangeRates, transactionHub.ExchangeRatesHandler)

	// api.SecureUserSubscribeCommand(HubKeyBlockConfirmation, transactionHub.BlockConfirmationHandler)

	return transactionHub
}
