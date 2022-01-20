package api

import (
	"context"
	"fmt"
	"passport"
	"passport/db"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/messagebus"
)

// ClientOnline gets trigger on connection online
func (api *API) ClientOnline(ctx context.Context, client *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
}

// ClientOffline gets trigger on connection offline
func (api *API) ClientOffline(ctx context.Context, client *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	// if they are level 5, they are server client. So lets remove them
	if client.Level == 5 {
		api.ServerClientOffline(client)
	}
}

func (api *API) ClientLogout(ctx context.Context, client *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, client.Identifier())), false)
	api.MessageBus.Unsub("", client, "")
	// broadcast user online status to server clients
	api.SendToAllServerClient(&ServerClientMessage{
		Key: UserOnlineStatus,
		Payload: struct {
			UserID string `json:"userID"`
			Status bool   `json:"status"`
		}{
			UserID: client.Identifier(),
			Status: true,
		},
	})
}

// ClientAuth gets triggered on auth and handles setting the clients permissions and levels
func (api *API) ClientAuth(ctx context.Context, client *hub.Client, clients hub.ClientsList, ch hub.TriggerChan) {
	if client.Level == passport.ServerClientLevel {
		return
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

	// broadcast user online status
	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, user.ID.String())), true)

	// broadcast user to gamebar
	api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyGamebarUserSubscribe, client.SessionID)), user)

	// broadcast user online status to server clients
	api.SendToAllServerClient(&ServerClientMessage{
		Key: UserOnlineStatus,
		Payload: struct {
			UserID passport.UserID `json:"userID"`
			Status bool            `json:"status"`
		}{
			UserID: user.ID,
			Status: true,
		},
	})
}
