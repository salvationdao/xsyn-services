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
	"time"

	goaway "github.com/TwiN/go-away"

	"github.com/ninja-software/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	leakybucket "github.com/kevinms/leakybucket-go"
	"github.com/microcosm-cc/bluemonday"
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

var profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(Profanities, []string{}, []string{})
var bm = bluemonday.StrictPolicy()

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
	api.SecureCommand(HubKeyChatMessage, factionHub.ChatMessageHandler)

	api.SubscribeCommand(HubKeyFactionUpdatedSubscribe, factionHub.FactionUpdatedSubscribeHandler)
	api.SubscribeCommand(HubKeyFactionStatUpdatedSubscribe, factionHub.FactionStatUpdatedSubscribeHandler)
	api.SubscribeCommand(HubKeyGlobalChatSubscribe, factionHub.GlobalChatUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyFactionChatSubscribe, factionHub.FactionChatUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyFactionContractRewardSubscribe, factionHub.FactionContractRewardUpdateSubscriber)

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
	go fc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	var resp struct {
		IsSuccess bool `json:"isSuccess"`
	}

	err = fc.API.GameserverRequest(http.MethodPost, "/user_enlist_faction", struct {
		UserID    passport.UserID    `json:"userID"`
		FactionID passport.FactionID `json:"factionID"`
	}{
		UserID:    userID,
		FactionID: req.Payload.FactionID,
	}, &resp)
	if err != nil {
		return terror.Error(err)
	}

	if !resp.IsSuccess {
		return terror.Error(fmt.Errorf("failed to enlist faction"))
	}

	reply(true)

	return nil
}

// FactionChatRequest sends chat message to specific faction.
type FactionChatRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID    passport.FactionID `json:"factionID"`
		MessageColor string             `json:"messageColor"`
		Message      string             `json:"message"`
	} `json:"payload"`
}

// ChatMessageSend contains chat message data to send.
type ChatMessageSend struct {
	Message           string           `json:"message"`
	MessageColor      string           `json:"messageColor"`
	FromUserID        passport.UserID  `json:"fromUserID"`
	FromUsername      string           `json:"fromUsername"`
	FactionColour     *string          `json:"factionColour,omitempty"`
	FactionLogoBlobID *passport.BlobID `json:"factionLogoBlobID,omitempty"`
	AvatarID          *passport.BlobID `json:"avatarID,omitempty"`
	SentAt            time.Time        `json:"sentAt"`
}

// rootHub.SecureCommand(HubKeyFactionChat, factionHub.ChatMessageHandler)
const HubKeyChatMessage hub.HubCommandKey = "CHAT:MESSAGE"

func firstN(s string, n int) string {
	i := 0
	for j := range s {
		if i == n {
			return s[:j]
		}
		i++
	}
	return s
}

var bucket = leakybucket.NewCollector(2, 10, true)
var minuteBucket = leakybucket.NewCollector(0.5, 30, true)

// ChatMessageHandler sends chat message from user
func (fc *FactionController) ChatMessageHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	b1 := bucket.Add(hubc.Identifier(), 1)
	b2 := minuteBucket.Add(hubc.Identifier(), 1)

	if b1 == 0 || b2 == 0 {
		return terror.Error(errors.New("too many messages"), "too many message")
	}

	req := &FactionChatRequest{}
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

	// get faction primary colour from faction
	var (
		factionColour     *string
		factionLogoBlobID *passport.BlobID
	)
	if user.FactionID != nil {
		faction, err := db.FactionGet(ctx, fc.Conn, *user.FactionID)
		if err != nil {
			return terror.Error(err)
		}
		factionColour = &faction.Theme.Primary
		factionLogoBlobID = &faction.LogoBlobID
	}

	msg := bm.Sanitize(req.Payload.Message)
	msg = profanityDetector.Censor(req.Payload.Message)
	if len(msg) > 280 {
		msg = firstN(msg, 280)
	}

	// check if the faction id is provided
	if !req.Payload.FactionID.IsNil() {
		if user.FactionID == nil || user.FactionID.IsNil() {
			return terror.Error(terror.ErrInvalidInput, "Require to join a faction to send message")
		}

		if *user.FactionID != req.Payload.FactionID {
			return terror.Error(terror.ErrForbidden, "Users are not allow to join the faction chat which they are not belong to")
		}

		// send message
		fc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, user.FactionID)), &ChatMessageSend{
			Message:           msg,
			MessageColor:      req.Payload.MessageColor,
			FromUserID:        user.ID,
			FromUsername:      user.Username,
			AvatarID:          user.AvatarID,
			SentAt:            time.Now(),
			FactionColour:     factionColour,
			FactionLogoBlobID: factionLogoBlobID,
		})
		reply(true)
		return nil
	}

	// global message
	fc.API.MessageBus.Send(ctx, messagebus.BusKey(HubKeyGlobalChatSubscribe), &ChatMessageSend{
		Message:           req.Payload.Message,
		MessageColor:      req.Payload.MessageColor,
		FromUserID:        user.ID,
		FromUsername:      user.Username,
		AvatarID:          user.AvatarID,
		SentAt:            time.Now(),
		FactionColour:     factionColour,
		FactionLogoBlobID: factionLogoBlobID,
	})
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
	req := &FactionUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	if req.Payload.FactionID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput, "Faction id is empty")
	}

	// get faction detail
	faction, err := db.FactionGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	reply(faction)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionUpdatedSubscribe, req.Payload.FactionID)), nil
}

const HubKeyFactionStatUpdatedSubscribe hub.HubCommandKey = "FACTION:STAT:SUBSCRIBE"

func (fc *FactionController) FactionStatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &FactionUpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	if req.Payload.FactionID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput, "Faction id is empty")
	}

	factionStat := &passport.FactionStat{}
	err = fc.API.GameserverRequest(http.MethodPost, "/faction_stat", struct {
		FactionID passport.FactionID `json:"factionID"`
	}{
		FactionID: req.Payload.FactionID,
	}, factionStat)
	if err != nil {
		return "", "", terror.Error(err)
	}

	user, err := db.FactionMvpGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", terror.Error(err)
	}
	voteSup, err := db.FactionSupsVotedGet(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return "", "", terror.Error(err)
	}
	num, err := db.FactionGetRecruitNumber(ctx, fc.Conn, req.Payload.FactionID)
	if err != nil {
		return "", "", terror.Error(err)
	}
	factionStat.RecruitNumber = num
	factionStat.MVP = user
	factionStat.SupsVoted = voteSup.String()

	reply(factionStat)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionStatUpdatedSubscribe, req.Payload.FactionID)), nil
}

const HubKeyGlobalChatSubscribe hub.HubCommandKey = "GLOBAL:CHAT:SUBSCRIBE"

func (fc *FactionController) GlobalChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}
	return req.TransactionID, messagebus.BusKey(HubKeyGlobalChatSubscribe), nil
}

const HubKeyFactionChatSubscribe hub.HubCommandKey = "FACTION:CHAT:SUBSCRIBE"

func (fc *FactionController) FactionChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	// get user in valid faction
	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}
	user, err := db.UserGet(ctx, fc.Conn, userID)
	if err != nil {
		return "", "", terror.Error(err)
	}
	if user.FactionID == nil || user.FactionID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput, "Require to join faction to receive")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, user.FactionID)), nil
}

const HubKeyFactionContractRewardSubscribe hub.HubCommandKey = "FACTION:CONTRACT:REWARD:SUBSCRIBE"

func (fc *FactionController) FactionContractRewardUpdateSubscriber(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err)
	}

	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	// get user faction
	faction, err := db.FactionGetByUserID(ctx, fc.Conn, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", terror.Error(err)
	}

	var resp struct {
		ContractReward string `json:"contractReward"`
	}
	// get contract reward from web hook
	err = fc.API.GameserverRequest(http.MethodPost, "/faction_contract_reward", struct {
		FactionID passport.FactionID `json:"factionID"`
	}{
		FactionID: faction.ID,
	}, &resp)
	if err != nil {
		return "", "", terror.Error(err, err.Error())
	}

	reply(resp.ContractReward)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionContractRewardSubscribe, faction.ID)), nil
}
