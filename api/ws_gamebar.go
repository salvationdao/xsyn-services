package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"
	"passport/db"
	"passport/log_helpers"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
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
	api.SecureCommand(HubKeyGamebarAuthRingCheck, gamebarHub.AuthTwitchRingCheck)
	api.SubscribeCommand(HubKeyGamebarUserSubscribe, gamebarHub.UserUpdatedSubscribeHandler)

	return gamebarHub
}

// 	rootHub.SecureCommand(HubKeyUserGet, UserController.GetHandler)
const HubKeyGamebarSessionIDGet hub.HubCommandKey = "GAMEBAR:SESSION:ID:GET"

func (gc *GamebarController) GetSessionIDHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	reply(hubc.SessionID)

	return nil
}

// AuthTwitchRingCheckRequest to bind twitch ui with current gamebar user
type AuthTwitchRingCheckRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TwitchExtensionJWT  string `json:"twitchExtensionJWT"`
		GameserverSessionID string `json:"gameserverSessionID"`
	} `json:"payload"`
}

const HubKeyGamebarAuthRingCheck hub.HubCommandKey = "GAMEBAR:AUTH:RING:CHECK"

func (gc *GamebarController) AuthTwitchRingCheck(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AuthTwitchRingCheckRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput)
	}

	user, err := db.UserGet(ctx, gc.Conn, userID, gc.API.HostUrl)
	if err != nil {
		return terror.Error(terror.ErrInvalidInput)
	}

	if req.Payload.TwitchExtensionJWT != "" {
		claims, err := gc.API.Auth.GetClaimsFromTwitchExtensionToken(req.Payload.TwitchExtensionJWT)
		if err != nil {
			return terror.Error(err)
		}

		if !strings.HasPrefix(claims.OpaqueUserID, "U") {
			return terror.Error(terror.ErrInvalidInput, "Twitch user is not login")
		}

		if claims.TwitchAccountID == "" {
			return terror.Error(terror.ErrInvalidInput, "No twitch account id is provided")
		}

		if user.TwitchID.Valid {
			// check twitch id match current passport user twitch account
			if claims.TwitchAccountID != user.TwitchID.String {
				return terror.Error(terror.ErrInvalidInput, "twitch account id does not match to current user")
			}
		} else {
			// associate current twitch id with
			err := db.UserAddTwitch(ctx, gc.Conn, user, claims.TwitchAccountID)
			if err != nil {
				return terror.Error(terror.ErrInvalidInput, "This Twitch account is already associated with a user")
			}

			user, err = db.UserGet(ctx, gc.Conn, user.ID, gc.API.HostUrl)
			if err != nil {
				return terror.Error(err)
			}

			// broadcast the update
			gc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
		}

		reply(true)

		// send to supremacy server
		gc.API.SendToServerClient(SupremacyGameServer, &ServerClientMessage{
			Key: "AUTH:RING:CHECK",
			Payload: struct {
				User               *passport.User `json:"user"`
				TwitchExtensionJWT string         `json:"twitchExtensionJWT"`
				SessionID          hub.SessionID  `json:"sessionID"`
			}{
				User:               user,
				TwitchExtensionJWT: req.Payload.TwitchExtensionJWT,
				SessionID:          hubc.SessionID,
			},
		})

		return nil

	} else if req.Payload.GameserverSessionID != "" {
		reply(true)

		fmt.Println(req.Key)
		// send to supremacy server
		gc.API.SendToServerClient(SupremacyGameServer, &ServerClientMessage{
			Key: "AUTH:RING:CHECK",
			Payload: struct {
				User                *passport.User `json:"user"`
				GameserverSessionID string         `json:"gameserverSessionID"`
				SessionID           hub.SessionID  `json:"sessionID"`
			}{
				User:                user,
				GameserverSessionID: req.Payload.GameserverSessionID,
				SessionID:           hubc.SessionID,
			},
		})

		return nil
	}

	return terror.Error(terror.ErrInvalidInput)
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
