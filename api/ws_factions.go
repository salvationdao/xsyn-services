package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/helpers"
	"time"

	"github.com/ninja-software/log_helpers"

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
	api.SecureCommand(HubKeyChatMessage, factionHub.ChatMessageHandler)

	api.SubscribeCommand(HubKeyFactionUpdatedSubscribe, factionHub.FactionUpdatedSubscribeHandler)
	api.SubscribeCommand(HubKeyFactionStatUpdatedSubscribe, factionHub.FactionStatUpdatedSubscribeHandler)
	api.SubscribeCommand(HubKeyGlobalChatSubscribe, factionHub.GlobalChatUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyFactionChatSubscribe, factionHub.FactionChatUpdatedSubscribeHandler)

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

	// broadcast updated user to server client
	fc.API.SendToAllServerClient(ctx, &ServerClientMessage{
		Key: UserEnlistFaction,
		Payload: struct {
			UserID    passport.UserID    `json:"userID"`
			FactionID passport.FactionID `json:"factionID"`
		}{
			UserID:    userID,
			FactionID: req.Payload.FactionID,
		},
	})

	// send faction stat request to game server
	fc.API.SendToServerClient(
		ctx,
		SupremacyGameServer,
		&ServerClientMessage{
			Key: FactionStatGet,
			Payload: struct {
				FactionID passport.FactionID `json:"factionID,omitempty"`
			}{
				FactionID: req.Payload.FactionID,
			},
		},
	)

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
	FactionLogoBlobID *passport.BlobID `json:"factionLogoBlobID,omitempty"`
	AvatarID          *passport.BlobID `json:"avatarID,omitempty"`
	SentAt            time.Time        `json:"sentAt"`
}

// rootHub.SecureCommand(HubKeyFactionChat, factionHub.ChatMessageHandler)
const HubKeyChatMessage hub.HubCommandKey = "CHAT:MESSAGE"

// ChatMessageHandler sends chat message from user
func (fc *FactionController) ChatMessageHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

	var (
		factionLogoBlobID *passport.BlobID
	)
	// get faction primary colour from faction
	if user.FactionID != nil {
		for _, faction := range passport.Factions {
			if faction.ID == *user.FactionID {
				factionLogoBlobID = &faction.LogoBlobID
				break
			}
		}
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
			Message:           req.Payload.Message,
			MessageColor:      req.Payload.MessageColor,
			FromUserID:        user.ID,
			FromUsername:      user.Username,
			AvatarID:          user.AvatarID,
			SentAt:            time.Now(),
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

	var userID *passport.UserID
	var sessionID *hub.SessionID

	if client.Identifier() != "" {
		uid := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
		userID = &uid
	} else {
		sessionID = &client.SessionID
	}

	// send faction stat request to game server
	fc.API.SendToServerClient(ctx,
		SupremacyGameServer,
		&ServerClientMessage{
			Key: FactionStatGet,
			Payload: struct {
				UserID    *passport.UserID   `json:"userID,omitempty"`
				SessionID *hub.SessionID     `json:"sessionID,omitempty"`
				FactionID passport.FactionID `json:"factionID,omitempty"`
			}{
				UserID:    userID,
				SessionID: sessionID,
				FactionID: req.Payload.FactionID,
			},
		},
	)

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
