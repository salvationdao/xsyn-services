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

		// TODO: add passport user sup subscribe

		// // broadcast the update to server clients
		// api.SendToAllServerClient(&ServerClientMessage{
		// 	Key: UserSupsUpdated,
		// 	Payload: struct {
		// 		UserID passport.UserID `json:"userID"`
		// 		Sups   passport.BigInt `json:"sups"`
		// 	}{
		// 		UserID: user.ID,
		// 		Sups:   user.Sups,
		// 	},
		// })

		// broadcast to game bar
		resp := &UserWalletDetail{
			OnChainSups: "0",
			OnWorldSups: user.Sups.Int.String(),
		}

		api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), resp)
	})
}

// UpdateUserInCache updates a user in the cache, if user doesn't exist it does nothing
func (api *API) UpdateUserInCache(user *passport.User) {

	api.UserCache(func(userMap UserCacheMap) {
		// TODO: Do NOT create a user if not exists, leaving it for now until gamebar is in twitch ui
		// currently we dont know what user comes online if they auth through the twitchui/gameserver and no point adding it for now
		if _, ok := userMap[user.ID]; !ok {
			userMap[user.ID] = &UserCache{User: user, CacheLastUpdated: time.Now()}
		}

		// TODO: check the held tx map
		api.HeldTransactions(func(heldTxList map[TransactionReference]*NewTransaction) {
			for _, tx := range heldTxList {
				if tx.To == user.ID {
					user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &tx.Amount)
				} else if tx.From == user.ID {
					user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &tx.Amount)
				}
			}
		})

		supsChanged := userMap[user.ID].Sups.Int.Cmp(&user.Sups.Int) != 0
		userMap[user.ID].User = user
		userMap[user.ID].CacheLastUpdated = time.Now()

		// broadcast the update to the users connected directly to passport
		// api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID)), user)
		// broadcast the update to connect client servers
		//api.SendToAllServerClient(&ServerClientMessage{
		//	Key: UserUpdated,
		//	Payload: struct {
		//		User *passport.User `json:"user"`
		//	}{
		//		User: user,
		//	},
		//})
		if supsChanged {
			// TODO: add passport user sup subscribe
			//api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupSubscribe, user.ID)), user)

			// // broadcast the update to server clients
			// api.SendToAllServerClient(&ServerClientMessage{
			// 	Key: UserSupsUpdated,
			// 	Payload: struct {
			// 		UserID passport.UserID `json:"userID"`
			// 		Sups   passport.BigInt `json:"sups"`
			// 	}{
			// 		UserID: user.ID,
			// 		Sups:   user.Sups,
			// 	},
			// })

			// broadcast to game bar
			resp := &UserWalletDetail{
				OnChainSups: "0",
				OnWorldSups: user.Sups.Int.String(),
			}

			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), resp)
		}
		//}
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

			// TODO: add passport user sup subscribe
			//api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupSubscribe, user.ID)), user)

			// // broadcast the update to server clients
			// api.SendToAllServerClient(&ServerClientMessage{
			// 	Key: UserSupsUpdated,
			// 	Payload: struct {
			// 		UserID passport.UserID `json:"userID"`
			// 		Sups   passport.BigInt `json:"sups"`
			// 	}{
			// 		UserID: user.ID,
			// 		Sups:   user.Sups,
			// 	},
			// })

			// broadcast to game bar
			resp := &UserWalletDetail{
				OnChainSups: "0",
				OnWorldSups: user.Sups.Int.String(),
			}

			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), resp)
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

			// // broadcast the update to server clients
			// api.SendToAllServerClient(&ServerClientMessage{
			// 	Key: UserSupsUpdated,
			// 	Payload: struct {
			// 		UserID passport.UserID `json:"userID"`
			// 		Sups   passport.BigInt `json:"sups"`
			// 	}{
			// 		UserID: user.ID,
			// 		Sups:   user.Sups,
			// 	},
			// })

			// broadcast to game bar
			resp := &UserWalletDetail{
				OnChainSups: "0",
				OnWorldSups: user.Sups.Int.String(),
			}

			api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), resp)

		}
		errChan <- nil
	})
}
