package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"passport"
	"passport/db"
	"passport/helpers"

	"github.com/ninja-software/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
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
	api.SubscribeCommand(HubKeyFactionStatUpdatedSubscribe, factionHub.FactionStatUpdatedSubscribeHandler)
	// api.SecureUserSubscribeCommand(HubKeyFactionContractRewardSubscribe, factionHub.FactionContractRewardUpdateSubscriber)

	return factionHub
}

// 	rootHub.SecureCommand(HubKeyFactionAll, UserController.GetHandler)
const HubKeyFactionAll hub.HubCommandKey = "FACTION:ALL"

func (fc *FactionController) FactionAllHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Failed to query factions, try again or contact support."
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Get all factions
	factions, err := db.FactionAll(ctx, fc.Conn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(factions)

	return nil
}

// FactionEnlistRequest enlist a faction
type FactionEnlistRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"faction_id"`
	} `json:"payload"`
}

// rootHub.SecureCommand(HubKeyFactionEnlist, factionHub.FactionEnlistHandler)
const HubKeyFactionEnlist hub.HubCommandKey = "FACTION:ENLIST"

// FactionEnlistHandler assign faction to user
func (fc *FactionController) FactionEnlistHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Failed to enlist user into faction, try again or contact support."
	req := &FactionEnlistRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden, "User is not logged in, access forbidden.")
	}

	// get user
	user, err := db.UserGet(ctx, fc.Conn, userID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User has already enlisted in a faction.")
	}

	// record old user state
	oldUser := *user

	faction, err := db.FactionGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	user.Faction = faction

	// assign faction to current user
	user.FactionID = &req.Payload.FactionID

	err = db.UserFactionEnlist(ctx, fc.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
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
	go fc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	var resp struct {
		IsSuccess bool `json:"is_success"`
	}

	err = fc.API.GameserverRequest(http.MethodPost, "/user_enlist_faction", struct {
		UserID    passport.UserID    `json:"user_id"`
		FactionID passport.FactionID `json:"faction_id"`
	}{
		UserID:    userID,
		FactionID: req.Payload.FactionID,
	}, &resp)
	if err != nil {
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
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:""`
	} `json:"payload"`
}

const HubKeyFactionUpdatedSubscribe hub.HubCommandKey = "FACTION:SUBSCRIBE"

func (fc *FactionController) FactionUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	errMsg := "Failed to subscribe to faction updates, try again or contact support."
	req := &FactionUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	if req.Payload.FactionID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput, errMsg)
	}

	// get faction detail
	faction, err := db.FactionGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	reply(faction)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionUpdatedSubscribe, req.Payload.FactionID)), nil
}

const HubKeyFactionStatUpdatedSubscribe hub.HubCommandKey = "FACTION:STAT:SUBSCRIBE"

func (fc *FactionController) FactionStatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	errMsg := "Failed to subscribe to faction stat updates, try again or contact support."
	req := &FactionUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	if req.Payload.FactionID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput, errMsg)
	}

	factionStat := &passport.FactionStat{}
	err = fc.API.GameserverRequest(http.MethodPost, "/faction_stat", struct {
		FactionID passport.FactionID `json:"faction_id"`
	}{
		FactionID: req.Payload.FactionID,
	}, factionStat)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	user, err := db.FactionMvpGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", terror.Error(err, errMsg)
	}
	voteSup, err := db.FactionSupsVotedGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}
	num, err := db.FactionGetRecruitNumber(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}
	factionStat.RecruitNumber = num
	factionStat.MVP = user
	factionStat.SupsVoted = voteSup.String()

	reply(factionStat)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionStatUpdatedSubscribe, req.Payload.FactionID)), nil
}

const HubKeyFactionContractRewardSubscribe hub.HubCommandKey = "FACTION:CONTRACT:REWARD:SUBSCRIBE"

func (fc *FactionController) FactionContractRewardUpdateSubscriber(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	errMsg := "There was a problem fetching the reward, try again or contact support."
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden, "User is not logged in, access forbidden.")
	}

	// get user faction
	faction, err := db.FactionGetByUserID(ctx, fc.Conn, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", terror.Error(err, errMsg)
	}

	var resp struct {
		ContractReward string `json:"contract_reward"`
	}
	// get contract reward from web hook
	err = fc.API.GameserverRequest(http.MethodPost, "/faction_contract_reward", struct {
		FactionID passport.FactionID `json:"faction_id"`
	}{
		FactionID: faction.ID,
	}, &resp)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	reply(resp.ContractReward)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionContractRewardSubscribe, faction.ID)), nil
}
