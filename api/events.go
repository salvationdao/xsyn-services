package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"passport"
	"passport/auth"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type ClientAuth struct {
	User     passport.User `json:"user"`
	JWTToken string        `json:"jwt_token"`
}

// ClientOnline gets trigger on connection online
func (api *API) ClientOnline(ctx context.Context, client *hub.Client) error {
	return nil
}

// ClientOffline gets trigger on connection offline
func (api *API) ClientOffline(ctx context.Context, client *hub.Client) error {
	// // if they are level 5, they are server client. So lets remove them
	// if client.Level == 5 {
	// 	api.ServerClientOffline(client)
	// }
	return nil
}

func (api *API) ClientLogout(ctx context.Context, client *hub.Client) error {

	// broadcast logout to gamebar
	go api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyGamebarUserSubscribe, client.SessionID)), nil)

	go api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, client.Identifier())), false)
	api.MessageBus.Unsub("", client, "")

	return nil
}

// ClientAuth gets triggered on auth and handles setting the clients permissions and levels
func (api *API) ClientAuth(ctx context.Context, client *hub.Client) error {

	if client.Level == passport.ServerClientLevel {
		return nil
	}

	userUuidString := client.Identifier()
	// client identifier gets set on auth so this shouldn't be empty
	if userUuidString == "" {
		api.Log.Err(fmt.Errorf("missing user uuid"))
	}
	userUuid, err := uuid.FromString(userUuidString)
	if err != nil {
		api.Log.Err(err)
	}

	user, err := db.UserGet(ctx, api.Conn, passport.UserID(userUuid))
	if err != nil {
		api.Log.Err(err)
	}

	if user.DeletedAt != nil {
		api.Log.Warn().Msgf("deleted user tried to login %s", user.ID)
	}

	// set their level to role tier
	client.SetLevel(user.Role.Tier)
	// set their perms
	client.SetPermissions(user.Role.Permissions)

	// add online user to our user cache
	// go api.InsertUserToCache(ctx, user)
	// broadcast user online status
	go api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, user.ID.String())), true)

	// create jwt for users to auth gameserver
	tokenID := uuid.Must(uuid.NewV4())
	jwt, sign, err := auth.GenerateJWT(tokenID.String(), *user, client.Request.UserAgent(), api.Tokens.tokenExpirationDays, api.JWTKey)
	if err != nil {
		passlog.L.Error().Str("user_id", user.ID.String()).Err(err).Msg("Unable to generate jwt")
		return terror.Error(err, "Unable to generate jwt. Please try again")
	}

	jwtSigned, err := sign(jwt, api.Tokens.encryptToken, api.Tokens.EncryptTokenKey())
	if err != nil {
		return terror.Error(err, "Unable to sign JWT. Please try again")
	}

	tokenEncoded := base64.StdEncoding.EncodeToString(jwtSigned)
	// store in db
	it := boiler.IssueToken{
		ID:        tokenID.String(),
		UserID:    user.ID.String(),
		UserAgent: client.Request.UserAgent(),
		ExpiresAt: null.TimeFrom(time.Now().AddDate(0, 0, api.Tokens.tokenExpirationDays)),
	}
	err = it.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert into issue token table")
	}

	// broadcast user to gamebar
	go api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyGamebarUserSubscribe, client.SessionID)), &ClientAuth{
		User:     *user,
		JWTToken: tokenEncoded,
	})

	return nil
}
