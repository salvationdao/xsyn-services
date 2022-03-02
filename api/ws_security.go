package api

import (
	"context"
	"encoding/json"
	"passport"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func (api *API) Command(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	api.Hub.Handle(key, func(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		span := tracer.StartSpan("ws.Command", tracer.ResourceName(string(key)))
		defer span.Finish()
		return fn(ctx, hubc, payload, reply)
	})
}

func (api *API) SecureCommand(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	api.Hub.Handle(key, func(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		span := tracer.StartSpan("ws.SecureCommand", tracer.ResourceName(string(key)))
		defer span.Finish()
		uuidString := hubc.Identifier() // identifier gets set on auth by default, so no ident = not authed
		if uuidString == "" {
			return terror.Error(terror.ErrUnauthorised)
		}

		return fn(ctx, hubc, payload, reply)
	})
}

// SecureCommandWithPerm registers a command to the hub that will only run if the websocket has authenticated and the user has the specified permission
func (api *API) SecureCommandWithPerm(key hub.HubCommandKey, fn hub.HubCommandFunc, perm passport.Perm) {
	api.Hub.Handle(key, func(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		span := tracer.StartSpan("ws.SecureCommandWithPerm", tracer.ResourceName(string(key)))
		defer span.Finish()
		allowed := hubc.HasPermission(perm.String())
		if !allowed {
			return terror.Error(terror.ErrUnauthorised)
		}

		return fn(ctx, hubc, payload, reply)
	})
}

// HubSubscribeCommandFunc is a registered handler for the hub to route to for subscriptions (returns sessionID and arguments)
type HubSubscribeCommandFunc func(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error)

// SubscribeCommand registers a subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) SubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	api.SubscribeCommandWithPermission(key, fn, "")
}

// SubscribeCommandWithPermission registers a subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) SubscribeCommandWithPermission(key hub.HubCommandKey, fn HubSubscribeCommandFunc, perm passport.Perm) {
	busKey := messagebus.BusKey("")

	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if perm != "" && !wsc.HasPermission(string(perm)) {
			return terror.Error(terror.ErrForbidden)
		}

		tx, bskey, err := fn(ctx, wsc, payload, reply)
		if err != nil {
			return terror.Error(err)
		}
		busKey = bskey

		// add subscription to the message bus
		api.MessageBus.Sub(busKey, wsc, tx)
		api.Log.Trace().Msgf("subscribed to %s - %s ", busKey, tx)
		return err
	})

	// Unsubscribe
	unsubscribeKey := key + ":UNSUBSCRIBE"
	api.Hub.Handle(unsubscribeKey, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		req := &hub.HubCommandRequest{}
		err := json.Unmarshal(payload, req)
		if err != nil {
			return terror.Error(err, "Invalid request received")
		}

		// remove subscription from message bus
		api.MessageBus.Unsub(busKey, wsc, req.TransactionID)
		api.Log.Trace().Msgf("unsubscribed from %s - %s ", busKey, req.TransactionID)
		return err
	})
}

// SecureUserSubscribeCommand registers a subscription command to the hub that will only run if the websocket has authenticated
//
// If fn is not provided, will use default
func (api *API) SecureUserSubscribeCommand(key hub.HubCommandKey, fn HubSubscribeCommandFunc) {
	api.SubscribeCommandWithAuthCheck(key, fn, func(wsc *hub.Client) bool {
		if wsc.Identifier() == "" || wsc.Level < 1 {
			return false
		}
		return true
	})
}

// SubscribeCommandWithAuthCheck registers a subscription command to the hub
//
// If fn is not provided, will use default
func (api *API) SubscribeCommandWithAuthCheck(key hub.HubCommandKey, fn HubSubscribeCommandFunc, authIsValid func(wsc *hub.Client) bool) {
	var err error
	busKey := messagebus.BusKey("")
	transactionID := ""

	api.Hub.Handle(key, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if !authIsValid(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		transactionID, busKey, err = fn(ctx, wsc, payload, reply)
		if err != nil {
			return terror.Error(err)
		}

		// add subscription to the message bus
		api.MessageBus.Sub(busKey, wsc, transactionID)

		return err
	})

	// Unsubscribe
	unsubscribeKey := hub.HubCommandKey(key + ":UNSUBSCRIBE")
	api.Hub.Handle(unsubscribeKey, func(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if !authIsValid(wsc) {
			return terror.Error(terror.ErrForbidden)
		}

		req := &hub.HubCommandRequest{}
		err := json.Unmarshal(payload, req)
		if err != nil {
			return terror.Error(err, "Invalid request received")
		}

		// remove subscription if buskey not empty from message bus
		if busKey != "" {
			api.MessageBus.Unsub(busKey, wsc, req.TransactionID)
		}

		return err
	})
}

// SupremacyCommand is a check to make sure the client is authed a supremacy game server
func (api *API) SupremacyCommand(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	api.Hub.Handle(key, func(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if hubc.Level != passport.ServerClientLevel {
			return terror.Error(terror.ErrForbidden)
		}

		if hubc.Identifier() != passport.SupremacyGameUserID.String() {
			return terror.Error(terror.ErrForbidden)
		}

		return fn(ctx, hubc, payload, reply)
	})
}

// ServerClientCommand is a check to make sure the client is a server client
func (api *API) ServerClientCommand(key hub.HubCommandKey, fn hub.HubCommandFunc) {
	api.Hub.Handle(key, func(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
		if hubc.Level != passport.ServerClientLevel {
			return terror.Error(terror.ErrForbidden)
		}

		// TODO: add a check for server client more than hubc level
		//supremacyUser, err := db.UserIDFromUsername(ctx, api.Conn, passport.SupremacyGameUsername)
		//if err != nil {
		//	return terror.Error(err)
		//}
		//
		//if hubc.Identifier() != supremacyUser.String() {
		//	return terror.Error(terror.ErrForbidden)
		//}

		return fn(ctx, hubc, payload, reply)
	})
}
