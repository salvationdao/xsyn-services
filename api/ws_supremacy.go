package api

import (
	"context"
	"encoding/json"
	"passport"
	"passport/db"
	"passport/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// SupremacyControllerWS holds handlers for supremacying supremacy status
type SupremacyControllerWS struct {
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	SupremacyUserID passport.UserID
}

// NewSupremacyController creates the supremacy hub
func NewSupremacyController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *SupremacyControllerWS {
	supremacyHub := &SupremacyControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "supremacy"),
		API:  api,
	}

	supID, err := db.UserIDFromUsername(context.Background(), conn, passport.SupremacyGameUsername)
	if err != nil {
		log.Panic().Err(err).Msgf("unable to find supremacy user")
	}

	supremacyHub.SupremacyUserID = *supID

	api.SupremacyCommand(HubKeySupremacyTakeSups, supremacyHub.SupremacyTakeSupsHandler)

	return supremacyHub
}

const HubKeySupremacyTakeSups = hub.HubCommandKey("SUPREMACY:TAKE_SUPS")

type SupremacyTakeSupsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount               int64           `json:"amount"`
		FromUserID           passport.UserID `json:"userId"`
		TransactionReference string          `json:"transactionReference"`
	} `json:"payload"`
}

func (ctrlr *SupremacyControllerWS) SupremacyTakeSupsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyTakeSupsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	tx := &Transaction{
		From:                 req.Payload.FromUserID,
		To:                   ctrlr.SupremacyUserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount,
	}

	ctrlr.API.transaction <- tx

	reply(true)
	return nil
}
