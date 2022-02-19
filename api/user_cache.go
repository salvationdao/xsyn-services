package api

import (
	"context"
	"fmt"
	"math/big"
	"passport"
	"time"

	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type UserCache struct {
	*passport.User
	CacheLastUpdated time.Time
}

type UserCacheMap map[passport.UserID]*UserCache
type UserCacheFunc func(userCacheList UserCacheMap)

// HandleUserCache is where the user cache map lives and where we pass through the updates to it
func (api *API) HandleUserCache() {
	var userCacheMap UserCacheMap = map[passport.UserID]*UserCache{}
	for userFunc := range api.users {
		userFunc(userCacheMap)
	}
}

// UserCache accepts a function that loops over the user cache map
func (api *API) UserCache(fn UserCacheFunc) {
	api.users <- func(userCacheList UserCacheMap) {
		fn(userCacheList)
	}
}

// InsertUserToCache adds a user to the cache
func (api *API) InsertUserToCache(ctx context.Context, user *passport.User) {
	api.UserCache(func(userMap UserCacheMap) {
		userMap[user.ID] = &UserCache{User: user, CacheLastUpdated: time.Now()}

		// broadcast the update to the users connected directly to passport
		go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID)), user)
		// broadcast the update to connect client servers
		api.SendToAllServerClient(ctx, &ServerClientMessage{
			Key: UserUpdated,
			Payload: struct {
				User *passport.User `json:"user"`
			}{
				User: user,
			},
		})

		if !user.ID.IsSystemUser() {
			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
		}
	})
}

// UpdateUserInCache updates a user in the cache, if user doesn't exist it does nothing and returns false
func (api *API) UpdateUserInCache(ctx context.Context, user *passport.User) {
	api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		api.UserCache(func(userMap UserCacheMap) {
			// skip if user is system user
			if user.ID.IsSystemUser() {
				return
			}

			// add user to cache if not exist
			if _, ok := userMap[user.ID]; !ok {
				userMap[user.ID] = &UserCache{User: user, CacheLastUpdated: time.Now()}
			}

			for _, tx := range heldTxList {
				if tx.To == user.ID {
					user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &tx.Amount)
				} else if tx.From == user.ID {
					user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &tx.Amount)
				}
			}

			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
		})
	})
}

// RemoveUserFromCache removes a user from the cache
func (api *API) RemoveUserFromCache(userID passport.UserID) {
	api.UserCache(func(userMap UserCacheMap) {
		delete(userMap, userID)
	})
}

// UpdateUserCacheAddSups updates a users sups in the cache and adds the given amount
func (api *API) UpdateUserCacheAddSups(ctx context.Context, userID passport.UserID, amount big.Int) {
	api.UserCache(func(userMap UserCacheMap) {
		user, ok := userMap[userID]
		if ok {
			user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &amount)

			if !user.ID.IsSystemUser() {
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
			}
		}
	})
}

// UpdateUserCacheRemoveSups updates a users sups in the cache and removes the given amount, returns error if not enough
func (api *API) UpdateUserCacheRemoveSups(ctx context.Context, userID passport.UserID, amount big.Int, errChan chan error) {
	api.UserCache(func(userMap UserCacheMap) {
		user, ok := userMap[userID]
		if ok {
			enoughFunds := user.Sups.Int.Cmp(&amount) >= 0

			if !enoughFunds {
				errChan <- fmt.Errorf("not enough funds")
				return
			}

			user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &amount)

			if !user.ID.IsSystemUser() {
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
			}

		}
		errChan <- nil
	})
}
