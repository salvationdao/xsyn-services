package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"passport"
	"passport/db"
	"time"

	"github.com/ninja-software/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
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

	if req.Payload.GameserverSessionID == "" {
		return terror.Error(terror.ErrInvalidInput, "Missing game site session id")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User is not logged in")
	}

	user, err := db.UserGet(ctx, gc.Conn, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return terror.Error(terror.ErrInvalidInput, "Failed to find the user detail")
	}

	if user == nil {
		return terror.Error(fmt.Errorf("user not found"), "User not found")
	}

	if os.Getenv("PASSPORT_ENVIRONMENT") == "development" || os.Getenv("PASSPORT_ENVIRONMENT") == "staging" {
		oneSups := big.NewInt(1000000000000000000)
		oneSups.Mul(oneSups, big.NewInt(100000))
		_, _, _, err := gc.API.userCacheMap.Process(&passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynSaleUserID,
			Amount:               *oneSups,
			NotSafe: true,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())),
			Description:          "Give away for testing",
		})

		if err != nil {
			gc.API.Log.Err(err).Msg("NO SUPS FOR YOU :p")
		}

	}

	reply(true)

	// send to supremacy server
	gc.API.SendToServerClient(ctx, SupremacyGameServer, &ServerClientMessage{
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
