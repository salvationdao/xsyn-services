package api

import (
	"fmt"
	"net/http"
	"passport/db"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/auth"
)

// AuthController holds connection data for auth handlers
type AuthController struct {
	Conn    db.Conn
	API     *API
	Twitter *auth.TwitterConfig
}

// AuthRouter returns a new router for handling OAuth requests (currently just Twitter)
func AuthRouter(conn *pgxpool.Pool, api *API, twitterConfig *auth.TwitterConfig) chi.Router {
	c := &AuthController{
		Conn:    conn,
		API:     api,
		Twitter: twitterConfig,
	}

	r := chi.NewRouter()
	r.Get("/twitter", WithError(c.TwitterAuth))

	return r
}

// The TwitterAuth endpoint kicks off the OAuth 1.0a flow
// https://developer.twitter.com/en/docs/authentication/oauth-1-0a/obtaining-user-access-tokens
func (c *AuthController) TwitterAuth(w http.ResponseWriter, r *http.Request) (int, error) {
	oauthCallback := r.URL.Query().Get("oauth_callback")
	if oauthCallback == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("Invalid OAuth callback url provided"))
	}

	oauthConfig := oauth1.Config{
		ConsumerKey:    c.Twitter.APIKey,
		ConsumerSecret: c.Twitter.APISecret,
		CallbackURL:    oauthCallback,
		Endpoint:       twitter.AuthorizeEndpoint,
	}

	requestToken, _, err := oauthConfig.RequestToken()
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	http.Redirect(w, r, fmt.Sprintf("https://api.twitter.com/oauth/authorize?oauth_token=%s", requestToken), http.StatusSeeOther)

	return http.StatusOK, nil
}
