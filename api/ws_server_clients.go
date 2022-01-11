package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport/log_helpers"

	"github.com/ninja-software/terror/v2"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
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

	// setting level, not sure if its needed but may be handy
	hubc.SetLevel(5)

	// TODO: get the client IDs name
	serverName := SupremacyGameServer

	// add this connection to our server client map
	ch.API.ServerClientOnline(serverName, hubc)

	reply(true)
	ch.API.SendToAllServerClient(&ServerClientMessage{
		Key:     Authed,
		Payload: nil,
	})
	return nil
}
