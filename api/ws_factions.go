package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/messagebus"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// FactionController holds handlers for roles
type FactionController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewFactionController creates the role hub
func NewFactionController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *FactionController {
	factionHub := &FactionController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "role_hub"),
		API:  api,
	}

	api.Command(HubKeyFactionAll, factionHub.FactionAllHandler)

	api.SecureUserSubscribeCommand(HubKeyFactionUpdatedSubscribe, factionHub.FactionUpdatedSubscribeHandler)

	return factionHub
}

// 	rootHub.SecureCommand(HubKeyFactionAll, UserController.GetHandler)
const HubKeyFactionAll hub.HubCommandKey = "FACTION:ALL"

func (fc *FactionController) FactionAllHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// Get all factions
	factions, err := db.FactionAll(ctx, fc.Conn)
	if err != nil {
		return terror.Error(err, "failed to query factions")
	}

	reply(factions)

	return nil
}

// FactionUpdatedSubscribeRequest subscribe to faction updates
type FactionUpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:""`
	} `json:"payload"`
}

const HubKeyFactionUpdatedSubscribe hub.HubCommandKey = "FACTION:SUBSCRIBE"

func (fc *FactionController) FactionUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &FactionUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	// get faction detail
	faction, err := db.FactionGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	reply(faction)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionUpdatedSubscribe, req.Payload.FactionID)), nil
}
