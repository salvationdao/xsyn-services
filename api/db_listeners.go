package api

import (
	"context"
	"encoding/json"
	"fmt"
	"passport"

	"github.com/ninja-syndicate/hub/ext/messagebus"

	"github.com/jackc/pgconn"
)

// DBListenForUserUpdateEvent creates a listener for the user_notify_event event in the db and then runs a UpsertUserToCache with the updated user
func (api *API) DBListenForUserUpdateEvent() {
	ctx := context.Background()
	conn, err := api.Conn.Acquire(ctx)
	if err != nil {
		if !pgconn.Timeout(err) {
			api.Log.Err(err).Msg("failed to acquire database connection to listen for user changes")
		}
		return
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "listen user_update_event")
	if err != nil {
		if !pgconn.Timeout(err) {
			api.Log.Err(err).Msg("failed to listen to user_notify_event")
		}
		return
	}

	for {
		userUpdate, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			if !pgconn.Timeout(err) {
				api.Log.Err(err).Msg("failed while waiting for notification of user_notify_event")
			}
			return
		}

		user := &passport.User{}
		err = json.Unmarshal([]byte(userUpdate.Payload), user)
		if err != nil {
			api.Log.Err(err).Msg("failed to parse postgres notification to user struct")
		}

		if !user.ID.IsSystemUser() {
			inCache := api.UpdateUserInCache(ctx, user)
			if !inCache {
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
			}
		}

	}
}
