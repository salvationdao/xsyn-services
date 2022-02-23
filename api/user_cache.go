package api

import (
	"context"
	"math/big"
	"passport/db"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
)

type UserCacheMap struct {
	sync.Map
	conn *pgxpool.Pool
}

func NewUserCacheMap(conn *pgxpool.Pool) *UserCacheMap {
	ucm := &UserCacheMap{
		sync.Map{},
		conn,
	}

	return ucm
}

func (ucm *UserCacheMap) Initialise() error {
	balances, err := db.UserBalances(context.Background(), ucm.conn)

	if err != nil {
		return err
	}

	for _, b := range balances {
		ucm.Store(b.ID.String(), b.Sups.Int)
	}

	return nil
}

func (ucm *UserCacheMap) Process(fromID string, toID string, amount big.Int) (*big.Int, *big.Int, error) {
	// load balance first
	fromBalance, err := ucm.Get(fromID)
	if err != nil {
		return nil, nil, terror.Error(err, "failed to read debit balance")
	}

	toBalance, err := ucm.Get(toID)
	if err != nil {
		return nil, nil, terror.Error(err, "failed to read credit balance")
	}

	// do subtract
	newFromBalance := big.NewInt(0)
	newFromBalance.Add(newFromBalance, &fromBalance)
	newFromBalance.Sub(newFromBalance, &amount)
	if newFromBalance.Cmp(big.NewInt(0)) < 0 {
		return nil, nil, terror.Error(err, "no enough fund")
	}

	// do add
	newToBalance := big.NewInt(0)
	newToBalance.Add(newToBalance, &toBalance)
	newToBalance.Add(newToBalance, &amount)
	if newToBalance.Cmp(big.NewInt(0)) < 0 {
		return nil, nil, terror.Error(err, "no enough fund")
	}

	// store back to the map
	ucm.Store(fromID, *newFromBalance)
	ucm.Store(toID, *newToBalance)

	return newFromBalance, newToBalance, nil
}

func (ucm *UserCacheMap) Get(id string) (big.Int, error) {
	result, ok := ucm.Load(id)
	if ok {
		return result.(big.Int), nil
	}

	balance, err := db.UserBalance(context.Background(), ucm.conn, id)
	if err != nil {
		return balance.Int, err
	}

	ucm.Store(id, balance.Int)
	return balance.Int, err
}

type UserCacheFunc func(userCacheList UserCacheMap)

// HandleUserCache is where the user cache map lives and where we pass through the updates to it
// func (api *API) HandleUserCache() {
// 	var userCacheMap UserCacheMap
// 	for {
// 		userFunc := <-api.users
// 		userFunc(userCacheMap)
// 	}
// }

// // UserCache accepts a function that loops over the user cache map
// func (api *API) UserCache(fn UserCacheFunc, stuff ...string) {
// 	if len(stuff) > 0 {
// 		fmt.Printf("users cache start %s\n", stuff[0])
// 	}
// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	select {
// 	case api.users <- func(userCacheList UserCacheMap) {
// 		fn(userCacheList)
// 		wg.Done()
// 	}:

// 	case <-time.After(10 * time.Second):
// 		api.Log.Err(errors.New("timeout on channel send exceeded"))
// 		panic("User Cache")
// 	}

// 	wg.Wait()
// 	if len(stuff) > 0 {
// 		fmt.Printf("users cache end %s\n", stuff[0])
// 	}
// }

// // InsertUserToCache adds a user to the cache
// func (api *API) InsertUserToCache(ctx context.Context, user *passport.User) {
// 	api.UserCache(func(userMap UserCacheMap) {
// 		userMap[user.ID] = &UserCache{User: user, CacheLastUpdated: time.Now()}

// 		// broadcast the update to the users connected directly to passport
// 		go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID)), user)
// 		// broadcast the update to connect client servers
// 		go api.SendToAllServerClient(ctx, &ServerClientMessage{
// 			Key: UserUpdated,
// 			Payload: struct {
// 				User *passport.User `json:"user"`
// 			}{
// 				User: user,
// 			},
// 		})

// 		if !user.ID.IsSystemUser() {
// 			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
// 		}
// 	}, "InsertUserToCache")
// }

// // UpdateUserInCache updates a user in the cache, if user doesn't exist it does nothing and returns false
// func (api *API) UpdateUserInCache(ctx context.Context, user *passport.User) {
// 	api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
// 		api.UserCache(func(userMap UserCacheMap) {
// 			// add user to cache if not exist
// 			if _, ok := userMap[user.ID]; !ok {
// 				userMap[user.ID] = &UserCache{}
// 			}

// 			for _, tx := range heldTxList {
// 				if tx.To == user.ID {
// 					user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &tx.Amount)
// 				} else if tx.From == user.ID {
// 					user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &tx.Amount)
// 				}
// 			}

// 			// update cache
// 			userMap[user.ID].User = user
// 			userMap[user.ID].CacheLastUpdated = time.Now()

// 			go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
// 		}, "HeldTransactions - UserCache")
// 	}, "UpdateUserInCache")
// }

// // RemoveUserFromCache removes a user from the cache
// func (api *API) RemoveUserFromCache(userID passport.UserID) {
// 	api.UserCache(func(userMap UserCacheMap) {
// 		delete(userMap, userID)
// 	}, "RemoveUserFromCache")
// }

// // UpdateUserCacheAddSups updates a users sups in the cache and adds the given amount
// func (api *API) UpdateUserCacheAddSups(ctx context.Context, userID passport.UserID, amount big.Int) {
// 	api.UserCache(func(userMap UserCacheMap) {
// 		user, ok := userMap[userID]
// 		if ok {
// 			user.Sups.Int = *user.Sups.Int.Add(&user.Sups.Int, &amount)

// 			if !user.ID.IsSystemUser() {
// 				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
// 			}
// 		}
// 	}, "UpdateUserCacheAddSups")
// }

// // UpdateUserCacheRemoveSups updates a users sups in the cache and removes the given amount, returns error if not enough
// func (api *API) UpdateUserCacheRemoveSups(ctx context.Context, userID passport.UserID, amount big.Int) error {
// 	var err error = nil
// 	api.UserCache(func(userMap UserCacheMap) {
// 		user, ok := userMap[userID]
// 		if ok {
// 			enoughFunds := user.Sups.Int.Cmp(&amount) >= 0

// 			if !enoughFunds {
// 				err = fmt.Errorf("not enough funds")
// 				return
// 			}

// 			user.Sups.Int = *user.Sups.Int.Sub(&user.Sups.Int, &amount)

// 			if !user.ID.IsSystemUser() {
// 				go api.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, user.ID)), user.Sups.Int.String())
// 			}

// 		}
// 	}, "UpdateUserCacheRemoveSups")
// 	return err
// }
