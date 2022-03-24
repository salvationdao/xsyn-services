package api

import (
	"fmt"
	"net/http"
	"passport/db"
	"passport/passlog"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

func ModeratorRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/users", WithError(WithModerator(ListUsers)))
	r.Get("/users/{public_address}", WithError(WithModerator(UserHandler)))
	r.Get("/chat_timeout_username/{username}/{minutes}", WithError(WithModerator(ChatTimeoutUsername)))
	r.Get("/chat_timeout_userid/{userID}/{minutes}", WithError(WithModerator(ChatTimeoutUserID)))
	r.Get("/rename_ban_username/{username}/{banned}", WithError(WithModerator(RenameBanUsername)))
	r.Get("/rename_ban_userID/{userID}/{banned}", WithError(WithModerator(RenameBanUserID)))
	r.Get("/rename_asset/{hash}/{newName}", WithError(WithModerator(RenameAsset)))
	r.Get("/purchased_items", WithError(WithModerator(ListPurchasedItems)))
	r.Get("/store_items", WithError(WithModerator(ListStoreItems)))

	return r
}

// WithModerator checks that mod key is in the header.
func WithModerator(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		apiKeyIDStr := r.Header.Get("X-Authorization")
		apiKeyID, err := uuid.FromString(apiKeyIDStr)
		if err != nil {
			passlog.L.Warn().Err(err).Str("apiKeyID", apiKeyIDStr).Msg("unauthed attempted at mod rest end point")
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}
		apiKey, err := db.APIKey(apiKeyID)
		if err != nil {
			passlog.L.Warn().Err(err).Str("apiKeyID", apiKeyIDStr).Msg("unauthed attempted at mod rest end point")
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}
		if apiKey.Type != "ADMIN" && apiKey.Type != "MODERATOR" {
			return http.StatusUnauthorized, terror.Error(fmt.Errorf("not moderator key: %s", apiKey.Type), "Unauthorized.")
		}
		return next(w, r)
	}
	return fn
}
