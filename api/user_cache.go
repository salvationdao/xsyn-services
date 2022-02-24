package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"passport"
	"passport/db"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

type UserCacheMap struct {
	sync.Map
	conn             *pgxpool.Pool
	TransactionCache *TransactionCache
	MessageBus       *messagebus.MessageBus
}

func NewUserCacheMap(conn *pgxpool.Pool, tc *TransactionCache, msgBus *messagebus.MessageBus) (*UserCacheMap, error) {
	ucm := &UserCacheMap{
		sync.Map{},
		conn,
		tc,
		msgBus,
	}
	balances, err := db.UserBalances(context.Background(), ucm.conn)

	if err != nil {
		return nil, err
	}

	for _, b := range balances {
		ucm.Store(b.ID.String(), b.Sups.Int)
	}
	return ucm, nil
}

var TransactionFailed = "TRANSACTION_FAILED"

func (ucm *UserCacheMap) Process(nt *passport.NewTransaction) (*big.Int, *big.Int, string, error) {
	// load balance first
	fromBalance, err := ucm.Get(nt.From.String())
	if err != nil {
		return nil, nil, TransactionFailed, terror.Error(err, "Failed to read debit balance. Please contact support if this problem persists.")
	}

	toBalance, err := ucm.Get(nt.To.String())
	if err != nil {
		return nil, nil, TransactionFailed, terror.Error(err, "Failed to read credit balance. Please contact support if this problem persists.")
	}

	// do subtract
	newFromBalance := big.NewInt(0)
	newFromBalance.Add(newFromBalance, &fromBalance)
	newFromBalance.Sub(newFromBalance, &nt.Amount)
	if newFromBalance.Cmp(big.NewInt(0)) < 0 {
		fromamt, _ := ucm.Get(nt.From.String())
		fmt.Println(fromamt.Int64(), nt.Amount)
		return nil, nil, TransactionFailed, terror.Error(errors.New("not enough funds"), "Not enough funds.")
	}

	// do add
	newToBalance := big.NewInt(0)
	newToBalance.Add(newToBalance, &toBalance)
	newToBalance.Add(newToBalance, &nt.Amount)
	if newToBalance.Cmp(big.NewInt(0)) < 0 {
		return nil, nil, TransactionFailed, terror.Error(errors.New("not enough funds"), "Not enough funds.")
	}

	// store back to the map
	ucm.Store(nt.From.String(), *newFromBalance)
	ucm.Store(nt.To.String(), *newToBalance)

	transactonID := ucm.TransactionCache.Process(nt)

	ctx := context.Background()
	if !nt.From.IsSystemUser() {
		go ucm.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, nt.From)), newFromBalance.String())
	}
	if !nt.To.IsSystemUser() {
		go ucm.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, nt.To)), newToBalance.String())
	}

	return newFromBalance, newToBalance, transactonID, nil
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
