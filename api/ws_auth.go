package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"passport"
	"passport/db"
	"passport/log_helpers"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/auth"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"google.golang.org/api/idtoken"
)

// AuthController holds handlers for roles
type AuthController struct {
	Conn               *pgxpool.Pool
	Log                *zerolog.Logger
	API                *API
	Google             *auth.GoogleConfig
	TwitchClientID     string
	TwitchClientSecret string
}

// NewAuthController creates the role hub
func NewAuthController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, googleConfig *auth.GoogleConfig, twitchClientID string, twitchClientSecret string) *AuthController {
	authHub := &AuthController{
		Conn:               conn,
		Log:                log_helpers.NamedLogger(log, "role_hub"),
		API:                api,
		Google:             googleConfig,
		TwitchClientID:     twitchClientID,
		TwitchClientSecret: twitchClientSecret,
	}

	api.Command(HubKeyAuthFacebookConnect, authHub.FacebookConnectHandler)
	api.Command(HubKeyAuthGoogleConnect, authHub.GoogleConnectHandler)
	api.Command(HubKeyAuthConnectTwitch, authHub.TwitchConnectHandler)

	return authHub
}

type NewConnectionRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Token string `json:"token"`
	} `json:"payload"`
}

const HubKeyAuthFacebookConnect hub.HubCommandKey = "AUTH:FACEBOOK:CONNECT"

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

const HubKeyAuthGoogleConnect hub.HubCommandKey = "AUTH:GOOGLE:CONNECT"

func (ac *AuthController) GoogleConnectHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &NewConnectionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Token == "" {
		return terror.Error(terror.ErrInvalidInput, "Google token is empty")
	}

	// Validate Google token
	errMsg := "There was a problem finding a user associated with the provided Google account, please check your details and try again."
	resp, err := idtoken.Validate(ctx, req.Payload.Token, ac.Google.ClientID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	googleID, ok := resp.Claims["sub"].(string)
	if !ok {
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

	// Update user's Google ID
	user.GoogleID = null.StringFrom(googleID)

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

type TwitchConnectionRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Code        string `json:"code"`
		RedirectURI string `json:"redirectURI"`
	} `json:"payload"`
}

const HubKeyAuthConnectTwitch hub.HubCommandKey = "AUTH:TWITCH:CONNECT"

func (ac *AuthController) TwitchConnectHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &TwitchConnectionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Code == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitch code is empty")
	}

	// Get Twitch access token from code
	requestUri := fmt.Sprintf(`https://id.twitch.tv/oauth2/token
    ?client_id=%s
    &client_secret=%s
    &code=%s
    &grant_type=authorization_code
    &redirect_uri=%s`,
		ac.TwitchClientID,
		ac.TwitchClientSecret,
		req.Payload.Code,
		req.Payload.RedirectURI)
	r, err := http.Post(requestUri, "application/json", nil)
	if err != nil {
		return terror.Error(err, "Failed to get Twitch access token")
	}
	defer r.Body.Close()

	respBody := &struct {
		AccessToken  string   `json:"access_token"`
		RefreshToken string   `json:"refresh_token"`
		ExpiresIn    int64    `json:"expires_in"`
		Scope        []string `json:"scope"`
		TokenType    string   `json:"token_type"`
	}{}
	err = json.NewDecoder(r.Body).Decode(respBody)
	if err != nil {
		return terror.Error(err, "Failed to get Twitch access token")
	}

	// Get Twitch user using access token
	reqUser, err := http.NewRequest("GET", "https://api.twitch.tv/kraken/user", nil)
	reqUser.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	reqUser.Header.Set("Client-ID", ac.TwitchClientID)
	reqUser.Header.Set("Authorization", fmt.Sprintf("OAuth %s", respBody.AccessToken))
	client := &http.Client{}
	respUser, err := client.Do(reqUser)
	if err != nil {
		return terror.Error(err, "Failed to get Twitch user")
	}
	respUserBody := &struct {
		ID string `json:"_id"`
	}{}
	err = json.NewDecoder(respUser.Body).Decode(respUserBody)
	if err != nil {
		return terror.Error(err, "Failed to get Twitch user")
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
	user.TwitchID = null.StringFrom(respUserBody.ID)

	// Update user
	err = db.UserUpdate(ctx, ac.Conn, user)
	if err != nil {
		return terror.Error(err, "Failed to update user with Twitch ID")
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
