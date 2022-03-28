package api

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	leakybucket "github.com/kevinms/leakybucket-go"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
	"github.com/sasha-s/go-deadlock"
)

const PersistChatMessageLimit = 20

var profanityDetector = goaway.NewProfanityDetector().WithCustomDictionary(Profanities, []string{}, []string{})
var bm = bluemonday.StrictPolicy()

// ChatMessageSend contains chat message data to send.
type ChatMessageSend struct {
	Message           string    `json:"message"`
	MessageColor      string    `json:"message_color"`
	FromUserID        string    `json:"from_user_id"`
	FromUsername      string    `json:"from_username"`
	FactionColour     *string   `json:"faction_colour,omitempty"`
	FactionLogoBlobID *string   `json:"faction_logo_blob_id,omitempty"`
	AvatarID          *string   `json:"avatar_id,omitempty"`
	SentAt            time.Time `json:"sent_at"`
}

// Chatroom holds a specific chat room
type Chatroom struct {
	deadlock.RWMutex
	factionID *passport.FactionID
	messages  []*ChatMessageSend
}

func (c *Chatroom) AddMessage(message *ChatMessageSend) {
	c.Lock()
	c.messages = append(c.messages, message)
	if len(c.messages) >= PersistChatMessageLimit {
		c.messages = c.messages[1:]
	}
	c.Unlock()
}

func (c *Chatroom) Range(fn func(chatMessage *ChatMessageSend) bool) {
	c.RLock()
	for _, message := range c.messages {
		if !fn(message) {
			break
		}
	}
	c.RUnlock()
}

func NewChatroom(factionID *passport.FactionID) *Chatroom {
	chatroom := &Chatroom{
		factionID: factionID,
		messages:  []*ChatMessageSend{},
	}
	return chatroom
}

// ChatController holds handlers for chat
type ChatController struct {
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	GlobalChat      *Chatroom
	RedMountainChat *Chatroom
	BostonChat      *Chatroom
	ZaibatsuChat    *Chatroom
}

// NewChatController creates the role hub
func NewChatController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, globalChat, redMountainChat, bostonChat, zaibatsuChat *Chatroom) *ChatController {
	chatHub := &ChatController{
		Conn:            conn,
		Log:             log_helpers.NamedLogger(log, "chat_hub"),
		API:             api,
		GlobalChat:      globalChat,
		RedMountainChat: redMountainChat,
		BostonChat:      bostonChat,
		ZaibatsuChat:    zaibatsuChat,
	}

	api.Command(HubKeyChatPastMessages, chatHub.ChatPastMessagesHandler)
	api.SecureCommand(HubKeyChatMessage, chatHub.ChatMessageHandler)

	api.SubscribeCommand(HubKeyGlobalChatSubscribe, chatHub.GlobalChatUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyFactionChatSubscribe, chatHub.FactionChatUpdatedSubscribeHandler)

	return chatHub
}

// FactionChatRequest sends chat message to specific faction.
type FactionChatRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID    passport.FactionID `json:"faction_id"`
		MessageColor string             `json:"message_color"`
		Message      string             `json:"message"`
	} `json:"payload"`
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
func (fc *ChatController) ChatMessageHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Issue sending message in chat, try again or contact support."
	b1 := bucket.Add(hubc.Identifier(), 1)
	b2 := minuteBucket.Add(hubc.Identifier(), 1)

	if b1 == 0 || b2 == 0 {
		return terror.Error(fmt.Errorf("too many messages"), "Too many messages.")
	}

	req := &FactionChatRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	user, err := boiler.FindUser(passdb.StdConn, hubc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// if chat banned just return
	if user.ChatBannedUntil.Valid && user.ChatBannedUntil.Time.After(time.Now()) {
		reply(true)
		return nil
	}

	// get faction primary colour from faction
	var (
		factionColour     *string
		factionLogoBlobID *string
	)
	if user.FactionID.Valid {
		faction, err := user.Faction().One(passdb.StdConn)
		if err != nil {
			return terror.Error(err, errMsg)
		}
		theme := &passport.FactionTheme{}

		err = faction.Theme.Unmarshal(theme)
		if err != nil {
			return terror.Error(err, errMsg)
		}
		factionColour = &theme.Primary
		factionLogoBlobID = &faction.LogoBlobID
	}

	msg := html.UnescapeString(bm.Sanitize(req.Payload.Message))
	msg = profanityDetector.Censor(msg)
	if len(msg) > 280 {
		msg = firstN(msg, 280)
	}

	// check if the faction id is provided
	if !req.Payload.FactionID.IsNil() {
		if !user.FactionID.Valid || user.FactionID.String == "" {
			return terror.Error(terror.ErrInvalidInput, "Required to join a faction to send message, please enlist in a faction.")
		}

		if user.FactionID.String != req.Payload.FactionID.String() {
			return terror.Error(terror.ErrForbidden, "Users are not allow to join the faction chat which they are not belong to.")
		}

		chatMessage := &ChatMessageSend{
			Message:           msg,
			MessageColor:      req.Payload.MessageColor,
			FromUserID:        user.ID,
			FromUsername:      user.Username,
			AvatarID:          &user.AvatarID.String,
			SentAt:            time.Now(),
			FactionColour:     factionColour,
			FactionLogoBlobID: factionLogoBlobID,
		}

		switch user.FactionID.String {
		case passport.RedMountainFactionID.String():
			fc.RedMountainChat.AddMessage(chatMessage)
		case passport.BostonCyberneticsFactionID.String():
			fc.BostonChat.AddMessage(chatMessage)
		case passport.ZaibatsuFactionID.String():
			fc.ZaibatsuChat.AddMessage(chatMessage)
		}

		// send message
		fc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, user.FactionID.String)), chatMessage)
		reply(true)
		return nil
	}

	// global message
	chatMessage := &ChatMessageSend{
		Message:           msg,
		MessageColor:      req.Payload.MessageColor,
		FromUserID:        user.ID,
		FromUsername:      user.Username,
		AvatarID:          &user.AvatarID.String,
		SentAt:            time.Now(),
		FactionColour:     factionColour,
		FactionLogoBlobID: factionLogoBlobID,
	}
	fc.GlobalChat.AddMessage(chatMessage)
	fc.API.MessageBus.Send(messagebus.BusKey(HubKeyGlobalChatSubscribe), chatMessage)
	reply(true)

	return nil
}

// ChatPastMessagesRequest sends chat message to specific faction.
type ChatPastMessagesRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"faction_id"`
	} `json:"payload"`
}

const HubKeyChatPastMessages hub.HubCommandKey = "CHAT:PAST_MESSAGES"

func (fc *ChatController) ChatPastMessagesHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &ChatPastMessagesRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if !req.Payload.FactionID.IsNil() {
		uuidString := hubc.Identifier() // identifier gets set on auth by default, so no ident = not authed
		if uuidString == "" {
			return terror.Error(terror.ErrUnauthorised, "Unauthorised access.")
		}
	}

	resp := []*ChatMessageSend{}
	chatRangeHandler := func(message *ChatMessageSend) bool {
		resp = append(resp, message)
		return true
	}

	switch req.Payload.FactionID {
	case passport.RedMountainFactionID:
		fc.RedMountainChat.Range(chatRangeHandler)
	case passport.BostonCyberneticsFactionID:
		fc.BostonChat.Range(chatRangeHandler)
	case passport.ZaibatsuFactionID:
		fc.ZaibatsuChat.Range(chatRangeHandler)
	default:
		fc.GlobalChat.Range(chatRangeHandler)
	}

	reply(resp)

	return nil
}

const HubKeyFactionChatSubscribe hub.HubCommandKey = "FACTION:CHAT:SUBSCRIBE"

func (fc *ChatController) FactionChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	errMsg := "Could not subscribe to faction chat updates, try again or contact support."
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	// get user in valid faction
	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden, "User is not logged in, access forbidden.")
	}
	user, err := db.UserGet(ctx, fc.Conn, userID)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}
	if user.FactionID == nil || user.FactionID.IsNil() {
		return "", "", terror.Error(terror.ErrInvalidInput, "Require to join faction to receive messages.")
	}
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionChatSubscribe, user.FactionID)), nil
}

const HubKeyGlobalChatSubscribe hub.HubCommandKey = "GLOBAL:CHAT:SUBSCRIBE"

func (fc *ChatController) GlobalChatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Could not subscribe to global chat updates, try again or contact support.")
	}
	return req.TransactionID, messagebus.BusKey(HubKeyGlobalChatSubscribe), nil
}
