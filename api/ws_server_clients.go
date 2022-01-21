package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/log_helpers"

	"github.com/ninja-software/terror/v2"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v3"
	"github.com/rs/zerolog"
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

	api.Command(HubKeyElevateAsServerClient, serverClientHub.Handler)

	api.ServerClientCommand(HubKeyCheckTransactionList, serverClientHub.CheckTransactionsHandler)

	return serverClientHub
}

const HubKeyElevateAsServerClient = hub.HubCommandKey("AUTH:SERVERCLIENT")

type ElevateAsServerClientRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"payload"`
}

func (ch *ServerClientControllerWS) Handler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &ElevateAsServerClientRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// TODO: add some sorta auth
	if req.Payload.ClientID == "" {
		return terror.Error(fmt.Errorf("missing client id"))
	}
	if req.Payload.ClientSecret == "" {
		return terror.Error(fmt.Errorf("missing client secret"))
	}

	// TODO: get the client IDs name from a db table then get the relative game user and set details on the hub client
	supremacyUser, err := db.UserIDFromUsername(ctx, ch.Conn, passport.SupremacyGameUsername)
	if err != nil {
		return terror.Error(err)
	}
	// setting level and identifier
	hubc.SetLevel(passport.ServerClientLevel)
	hubc.SetIdentifier(supremacyUser.String())
	hubc.LockClient = true // lock the client so it cannot be updated

	// TODO: get the matching server name
	serverName := SupremacyGameServer

	// add this connection to our server client map
	ch.API.ServerClientOnline(serverName, hubc)

	reply(true)
	ch.API.SendToServerClient(serverName, &ServerClientMessage{
		Key:     Authed,
		Payload: nil,
	})
	return nil
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
