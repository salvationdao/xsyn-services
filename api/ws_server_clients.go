package api

import (
	"context"
	"encoding/json"
	"passport"
	"passport/db"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"

	"github.com/ninja-syndicate/hub"
)

// ServerClientControllerWS holds handlers for serverClienting serverClient status
type ServerClientControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewServerClientController creates the serverClient hub
func NewServerClientController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *ServerClientControllerWS {
	serverClientHub := &ServerClientControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "serverClient_hub"),
		API:  api,
	}

	// api.Command(HubKeyElevateAsServerClient, serverClientHub.Handler)

	api.ServerClientCommand(HubKeyCheckTransactionList, serverClientHub.CheckTransactionsHandler)

	return serverClientHub
}

const HubKeyCheckTransactionList = hub.HubCommandKey("TRANSACTION:CHECK_LIST")

type CheckTransactionsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TransactionReferences []string `json:"transactionReferences"`
	} `json:"payload"`
}

// CheckTransactionsHandler takes a list of transaction references and returns failed transaction references
func (ch *ServerClientControllerWS) CheckTransactionsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &CheckTransactionsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	list, err := db.TransactionGetList(ctx, ch.Conn, req.Payload.TransactionReferences)
	if err != nil {
		return terror.Error(err)
	}

	reply(struct {
		Transactions []*passport.Transaction `json:"transactions"`
	}{
		Transactions: list,
	})
	return nil
}
