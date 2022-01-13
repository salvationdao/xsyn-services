package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"passport"
	"passport/db"
	"passport/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

// AuthController holds handlers for roles
type AuthController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewAuthController creates the role hub
func NewAuthController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *AuthController {
	authHub := &AuthController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "role_hub"),
		API:  api,
	}

	api.Command(HubKeyTwitchAuth, authHub.TwitchAuthHandler)
	api.Command(HubKeyAuthConnectFacebook, authHub.FacebookConnectHandler)

	return authHub
}

// TwitchAuthRequest requests an update for an existing user
type TwitchAuthRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TwitchToken string `json:"twitchToken"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyTwitchAuth, UserController.GetHandler)
const HubKeyTwitchAuth hub.HubCommandKey = "TWITCH:AUTH"

func (ac *AuthController) TwitchAuthHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchAuthRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.TwitchToken == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitch jwt is empty")
	}

	resp, err := ac.API.Auth.TwitchLogin(ctx, hubc, req.Payload.TwitchToken)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err := db.UserGet(ctx, ac.Conn, passport.UserID(resp.User.Fields().ID()))
	if err != nil {
		return terror.Error(err, "failed to query user")
	}

	reply(user)

	// send user changes to connected clients
	ac.API.SendToAllServerClient(&ServerClientMessage{
		Key: UserUpdated,
		Payload: struct {
			User *passport.User `json:"user"`
		}{
			User: user,
		},
	})

	return nil
}

type NewConnectionRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Token string `json:"token"`
	} `json:"payload"`
}

const HubKeyAuthConnectFacebook hub.HubCommandKey = "AUTH:FACEBOOK:CONNECT"

func (ac *AuthController) FacebookConnectHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &NewConnectionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Token == "" {
		return terror.Error(terror.ErrInvalidInput, "Facebook token is empty")
	}

	// Validate Facebook token
	errMsg := "There was a problem finding a user associated with the provided Facebook account, please check your details and try again."
	r, err := http.Get("https://graph.facebook.com/me?&access_token=" + url.QueryEscape(req.Payload.Token))
	if err != nil {
		return terror.Error(err)
	}
	defer r.Body.Close()
	resp := &struct {
		ID string `json:"id"`
	}{}
	err = json.NewDecoder(r.Body).Decode(resp)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not convert user ID to UUID")
	}

	// Get user
	user, err := db.UserGet(ctx, ac.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "failed to query user")
	}

	// Update user's Facebook ID
	user.FacebookID = null.StringFrom(resp.ID)

	// Update user
	err = db.UserUpdate(ctx, ac.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	// send user changes to connected clients
	ac.API.SendToAllServerClient(&ServerClientMessage{
		Key: UserUpdated,
		Payload: struct {
			User *passport.User `json:"user"`
		}{
			User: user,
		},
	})

	return nil
}
