package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-software/log_helpers"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

var Profanities = []string{
	"fag",
	"fuck",
	"nigga",
	"nigger",
	"rape",
	"retard",
}

// FactionController holds handlers for roles
type FactionController struct {
	Log *zerolog.Logger
	API *API
}

// NewFactionController creates the role hub
func NewFactionController(log *zerolog.Logger, api *API) *FactionController {
	factionHub := &FactionController{
		Log: log_helpers.NamedLogger(log, "role_hub"),
		API: api,
	}

	api.Command(HubKeyFactionAll, factionHub.FactionAllHandler)
	api.SecureCommand(HubKeyFactionEnlist, factionHub.FactionEnlistHandler)

	//api.SubscribeCommand(HubKeyFactionUpdatedSubscribe, factionHub.FactionUpdatedSubscribeHandler)
	// api.SecureUserSubscribeCommand(HubKeyFactionContractRewardSubscribe, factionHub.FactionContractRewardUpdateSubscriber)

	return factionHub
}

// 	rootHub.SecureCommand(HubKeyFactionAll, UserController.GetHandler)
const HubKeyFactionAll = "FACTION:ALL"

func (fc *FactionController) FactionAllHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to query factions, try again or contact support."

	// Get all factions
	factions, err := boiler.Factions().All(passdb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(factions)

	return nil
}

// FactionEnlistRequest enlist a faction
type FactionEnlistRequest struct {
	Payload struct {
		FactionID string `json:"faction_id"`
	} `json:"payload"`
}

// rootHub.SecureCommand(HubKeyFactionEnlist, factionHub.FactionEnlistHandler)
const HubKeyFactionEnlist = "FACTION:ENLIST"

// FactionEnlistHandler assign faction to user
func (fc *FactionController) FactionEnlistHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to enlist user into faction, try again or contact support."
	req := &FactionEnlistRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if user.FactionID.Valid {
		return terror.Error(terror.ErrInvalidInput, "User has already enlisted in a faction.")
	}

	// record old user state
	oldUser := *user

	faction, err := boiler.FindFaction(passdb.StdConn, req.Payload.FactionID)
	if err != nil {
		return terror.Error(err, "Failed to get faction")
	}

	// assign faction to current user
	user.FactionID = null.StringFrom(faction.ID)

	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.FactionID))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// record user activity
	fc.API.RecordUserActivity(ctx,
		user.ID,
		"Enlist faction",
		types.ObjectTypeUser,
		helpers.StringPointer(user.ID),
		&user.Username,
		helpers.StringPointer(user.FirstName.String+" "+user.LastName.String),
		&types.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	// assign faction to user
	user.Faction = faction

	// broadcast updated user to gamebar user
	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	var resp struct {
		IsSuccess bool `json:"is_success"`
	}

	err = fc.API.GameserverRequest(http.MethodPost, "/user_enlist_faction", struct {
		UserID    string `json:"user_id"`
		FactionID string `json:"faction_id"`
	}{
		UserID:    user.ID,
		FactionID: user.FactionID.String,
	}, &resp)
	if err != nil {
		passlog.L.Error().Err(err).Msg("gameserver request failed")
		return terror.Error(err, "Error requesting game server, try again or contact support.")
	}

	if !resp.IsSuccess {
		return terror.Error(fmt.Errorf("failed to enlist faction"), errMsg)
	}

	// TODO: generate a new token and reply
	reply(true)

	return nil
}

// FactionUpdatedSubscribeRequest subscribe to faction updates
type FactionUpdatedSubscribeRequest struct {
	Payload struct {
		FactionID types.FactionID `json:"faction_id"`
	} `json:"payload"`
}
