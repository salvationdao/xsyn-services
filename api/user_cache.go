package api

import (
	"context"
	"fmt"
	"math/big"
	"passport"
	"sync"
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
	for {
		userFunc := <-api.users
		userFunc(userCacheMap)
	}
}

// UserCache accepts a function that loops over the user cache map
func (api *API) UserCache(fn UserCacheFunc, stuff ...string) {
	if len(stuff) > 0 {
		fmt.Printf("users cache start %s\n", stuff[0])
	}
	var wg sync.WaitGroup
	wg.Add(1)
	api.users <- func(userCacheList UserCacheMap) {
		fn(userCacheList)
		wg.Done()
	}
	wg.Wait()
	if len(stuff) > 0 {
		fmt.Printf("users cache end %s\n", stuff[0])
	}
}

// InsertUserToCache adds a user to the cache
func (api *API) InsertUserToCache(ctx context.Context, user *passport.User) {
	api.UserCache(func(userMap UserCacheMap) {
		userMap[user.ID] = &UserCache{User: user, CacheLastUpdated: time.Now()}

		// broadcast the update to the users connected directly to passport
		go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID)), user)
		// broadcast the update to connect client servers
		go api.SendToAllServerClient(ctx, &ServerClientMessage{
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
	}, "InsertUserToCache")
}

// UpdateUserInCache updates a user in the cache, if user doesn't exist it does nothing and returns false
func (api *API) UpdateUserInCache(ctx context.Context, user *passport.User) {
	api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		api.UserCache(func(userMap UserCacheMap) {
			// add user to cache if not exist
			if _, ok := userMap[user.ID]; !ok {
				userMap[user.ID] = &UserCache{}
			}

			for _, tx := range heldTxList {
				if tx.To == user.ID {
					user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &tx.Amount)
				} else if tx.From == user.ID {
					user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &tx.Amount)
				}
			}

			// update cache
			userMap[user.ID].User = user
			userMap[user.ID].CacheLastUpdated = time.Now()

			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
		}, "HeldTransactions - UserCache")
	}, "UpdateUserInCache")
}

// RemoveUserFromCache removes a user from the cache
func (api *API) RemoveUserFromCache(userID passport.UserID) {
	api.UserCache(func(userMap UserCacheMap) {
		delete(userMap, userID)
	}, "RemoveUserFromCache")
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
	}, "UpdateUserCacheAddSups")
}

// UpdateUserCacheRemoveSups updates a users sups in the cache and removes the given amount, returns error if not enough
func (api *API) UpdateUserCacheRemoveSups(ctx context.Context, userID passport.UserID, amount big.Int) error {
	var err error = nil
	api.UserCache(func(userMap UserCacheMap) {
		user, ok := userMap[userID]
		if ok {
			enoughFunds := user.Sups.Int.Cmp(&amount) >= 0

			if !enoughFunds {
				err = fmt.Errorf("not enough funds")
				return
			}

			user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &amount)

			if !user.ID.IsSystemUser() {
				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
			}

		}
	}, "UpdateUserCacheRemoveSups")
	return err
}
