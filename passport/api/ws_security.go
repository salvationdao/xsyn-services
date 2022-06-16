package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/go-chi/chi/v5"
	"github.com/ninja-syndicate/ws"

	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

func (api *API) Command(key string, fn ws.CommandFunc) {
	api.Commander.Command(key, func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		return fn(ctx, key, payload, reply)
	})
}

func (api *API) WSURLParam(URI string, params ...string) func(r *http.Request) (string, error) {
	return func(r *http.Request) (string, error) {
		uri := URI
		for _, paramName := range params {
			param := chi.URLParam(r, paramName)

			uri = strings.ReplaceAll(uri, fmt.Sprintf("{%s}", paramName), param)

			if param == "" {
				return "", fmt.Errorf("%s not matched", paramName)
			}
		}
		return uri, nil
	}
}

type SecureCommandFunc func(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error

var factions map[string]*boiler.Faction
var rlock sync.RWMutex
var once sync.Once

func Faction(id string) *boiler.Faction {
	if id == "" {
		return nil
	}
	once.Do(func() {
		factions = map[string]*boiler.Faction{}
		factionsAll, err := boiler.Factions().All(passdb.StdConn)
		if err != nil {
			passlog.L.Fatal().Err(err).Msg("unable to load factions from database")
		}

		for _, f := range factionsAll {
			factions[f.ID] = f
		}
	})
	rlock.RLock()
	defer rlock.RUnlock()

	return factions[id]
}

func RetrieveUser(ctx context.Context) (*types.User, error) {

	userID := ctx.Value("user_id")

	userIDStr, ok := userID.(string)

	if !ok || userIDStr == "" {
		return nil, fmt.Errorf("can not retrieve user id")
	}

	user, err := boiler.FindUser(passdb.StdConn, userIDStr)
	if err != nil {
		return nil, fmt.Errorf("not authorized to access this endpoint")
	}

	var faction *boiler.Faction
	if user.FactionID.Valid {
		faction = Faction(user.FactionID.String)
	}

	return &types.User{
		User:    user,
		Faction: faction,
	}, nil
}

func (api *API) MustSecure(fn SecureCommandFunc) ws.CommandFunc {
	return func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return err
		}

		return fn(ctx, user, key, payload, reply)
	}
}

func (api *API) SecureCommand(key string, fn SecureCommandFunc) {
	api.Commander.Command(key, api.MustSecure(fn))
}

// SecureCommandWithPerm registers a command to the hub that will only run if the websocket has authenticated and the user has the specified permission
func (api *API) SecureCommandWithPerm(key string, fn SecureCommandFunc, perm types.Perm) {
	api.Commander.Command(key, func(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
		user, err := RetrieveUser(ctx)
		if err != nil {
			return fmt.Errorf("can not retrieve user")
		}

		err = user.L.LoadRole(passdb.StdConn, true, user, nil)
		p := perm.String()
		for _, prm := range user.R.Role.Permissions {
			if prm == p {
				return fn(ctx, user, key, payload, reply)
			}
		}

		return fmt.Errorf("user does not have permission")
	})
}

// HubSubscribeCommandFunc is a registered handler for the hub to route to for subscriptions (returns sessionID and arguments)
type HubSubscribeCommandFunc func(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error)

func (api *API) AuthWS(required bool, userIDMustMatch bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			var token string
			var ok bool

			cookie, err := r.Cookie("xsyn-token")
			if err != nil {
				token = r.URL.Query().Get("token")
				if token == "" {
					token, ok = r.Context().Value("token").(string)
					if !ok || token == "" {
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
				}
			} else {
				if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
					if required {
						http.Error(w, "Unauthorized: cookie error", http.StatusUnauthorized)
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}
			resp, err := api.TokenLogin(token, "")
			if err != nil {

				// delete cookies, in case there is one
				api.DeleteCookie(w, r)

				if required {
					http.Error(w, "Unauthorized: token login failed", http.StatusUnauthorized)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			if userIDMustMatch {
				userID := chi.URLParam(r, "userId")
				if userID == "" || userID != resp.User.ID {
					http.Error(w, "user id not match", http.StatusUnauthorized)
					return
				}
			}

			ctx := context.WithValue(r.Context(), "user_id", resp.User.ID)
			*r = *r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}
		return http.HandlerFunc(fn)
	}
}
