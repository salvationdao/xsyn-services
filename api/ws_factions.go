package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/helpers"
	"passport/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
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
	api.SecureCommand(HubKeyFactionEnlist, factionHub.FactionEnlistHandler)

	api.SubscribeCommand(HubKeyFactionUpdatedSubscribe, factionHub.FactionUpdatedSubscribeHandler)

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

// FactionEnlistRequest enlist a faction
type FactionEnlistRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"factionID"`
	} `json:"payload"`
}

// rootHub.SecureCommand(HubKeyFactionEnlist, factionHub.FactionEnlistHandler)
const HubKeyFactionEnlist hub.HubCommandKey = "FACTION:ENLIST"

// FactionEnlistHandler assign faction to user
func (fc *FactionController) FactionEnlistHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &FactionEnlistRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	// get user
	user, err := db.UserGet(ctx, fc.Conn, userID)
	if err != nil {
		return terror.Error(err)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User has already enlisted in a faction")
	}

	// record old user state
	oldUser := *user

	faction, err := db.FactionGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return terror.Error(err)
	}
	user.Faction = faction

	// assign faction to current user
	user.FactionID = &req.Payload.FactionID

	err = db.UserFactionEnlist(ctx, fc.Conn, user)
	if err != nil {
		return terror.Error(err)
	}

	// record user activity
	fc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Enlist faction",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	// broadcast updated user to gamebar user
	fc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	// broadcast updated user to server client
	fc.API.SendToAllServerClient(&ServerClientMessage{
		Key: UserEnlistFaction,
		Payload: struct {
			UserID    passport.UserID    `json:"userID"`
			FactionID passport.FactionID `json:"factionID"`
		}{
			UserID:    userID,
			FactionID: req.Payload.FactionID,
		},
	})

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
