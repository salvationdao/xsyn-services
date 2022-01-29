package api

import (
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
func (api *API) UserCache(fn UserCacheFunc) {
	var wg sync.WaitGroup
	wg.Add(1)
	api.users <- func(userCacheList UserCacheMap) {
		fn(userCacheList)
		wg.Done()
	}
	wg.Wait()
}

// InsertUserToCache adds a user to the cache
func (api *API) InsertUserToCache(user *passport.User) {
	api.UserCache(func(userMap UserCacheMap) {
		userMap[user.ID] = &UserCache{User: user, CacheLastUpdated: time.Now()}

		// broadcast the update to the users connected directly to passport
		api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID)), user)
		// broadcast the update to connect client servers
		api.SendToAllServerClient(&ServerClientMessage{
			Key: UserUpdated,
			Payload: struct {
				User *passport.User `json:"user"`
			}{
				User: user,
			},
		})

		if !user.ID.IsSystemUser() {
			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), &UserWalletDetail{
				OnChainSups: "0",
				OnWorldSups: user.Sups.Int.String(),
			})
		}
	})
}

// UpdateUserInCache updates a user in the cache, if user doesn't exist it does nothing
func (api *API) UpdateUserInCache(user *passport.User) {

	api.UserCache(func(userMap UserCacheMap) {
		// cache map should have the latest user sups detail
		// so skip if user is already in the cache map
		if _, ok := userMap[user.ID]; ok {
			return
		}

		// otherwise process user uncommitted transactions
		api.HeldTransactions(func(heldTxList map[TransactionReference]*NewTransaction) {
			for _, tx := range heldTxList {
				if tx.To == user.ID {
					user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &tx.Amount)
				} else if tx.From == user.ID {
					user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &tx.Amount)
				}
			}
		})

		// add user to cache map
		userMap[user.ID] = &UserCache{User: user, CacheLastUpdated: time.Now()}

		// broadcast user sups, if user is not the system user
		if !user.ID.IsSystemUser() {
			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), &UserWalletDetail{
				OnChainSups: "0",
				OnWorldSups: user.Sups.Int.String(),
			})
		}
	})
}

// RemoveUserFromCache removes a user from the cache
func (api *API) RemoveUserFromCache(userID passport.UserID) {
	api.UserCache(func(userMap UserCacheMap) {
		delete(userMap, userID)
	})
}

// UpdateUserCacheAddSups updates a users sups in the cache and adds the given amount
func (api *API) UpdateUserCacheAddSups(userID passport.UserID, amount big.Int) {
	api.UserCache(func(userMap UserCacheMap) {
		user, ok := userMap[userID]
		if ok {
			user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &amount)

			if !user.ID.IsSystemUser() {
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), &UserWalletDetail{
					OnChainSups: "0",
					OnWorldSups: user.Sups.Int.String(),
				})
			}
		}
	})
}

// UpdateUserCacheRemoveSups updates a users sups in the cache and removes the given amount, returns error if not enough
func (api *API) UpdateUserCacheRemoveSups(userID passport.UserID, amount big.Int, errChan chan error) {
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
				api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), &UserWalletDetail{
					OnChainSups: "0",
					OnWorldSups: user.Sups.Int.String(),
				})
			}

		}
		errChan <- nil
	})
}
