package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v3"
	"github.com/ninja-software/hub/v3/ext/messagebus"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// GamebarController holds handlers for authentication
type GamebarController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

func NewGamebarController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *GamebarController {
	gamebarHub := &GamebarController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "user_hub"),
		API:  api,
	}

	api.Command(HubKeyGamebarSessionIDGet, gamebarHub.GetSessionIDHandler)
	api.SubscribeCommand(HubKeyGamebarUserSubscribe, gamebarHub.UserUpdatedSubscribeHandler)

	return gamebarHub
}

// 	rootHub.SecureCommand(HubKeyUserGet, UserController.GetHandler)
const HubKeyGamebarSessionIDGet hub.HubCommandKey = "GAMEBAR:SESSION:ID:GET"

func (gc *GamebarController) GetSessionIDHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	reply(hubc.SessionID)

	return nil
}

const HubKeyGamebarUserSubscribe hub.HubCommandKey = "GAMEBAR:USER:SUBSCRIBE"

// UserUpdatedSubscribeRequest to subscribe to user updates
type UserUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SessionID string `json:"sessionID"`
	} `json:"payload"`
}

func (gc *GamebarController) UserUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &UserUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyGamebarUserSubscribe, req.Payload.SessionID)), nil
}
