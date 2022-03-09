package api

import (
	"context"
	"fmt"
	"passport"
	"passport/db"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

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
	go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyGamebarUserSubscribe, client.SessionID)), nil)

	go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, client.Identifier())), false)
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
	go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, user.ID.String())), true)

	// broadcast user to gamebar
	go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyGamebarUserSubscribe, client.SessionID)), user)

	return nil
}
