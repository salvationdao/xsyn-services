package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"passport"
	"passport/db"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/sasha-s/go-deadlock"
)

type UserCacheMap struct {
	deadlock.Map
	conn             *pgxpool.Pool
	TransactionCache *TransactionCache
	MessageBus       *messagebus.MessageBus
}

func NewUserCacheMap(conn *pgxpool.Pool, tc *TransactionCache, msgBus *messagebus.MessageBus) (*UserCacheMap, error) {
	ucm := &UserCacheMap{
		deadlock.Map{},
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
	ucm.TransactionCache.IsLocked.RLock()
	if ucm.TransactionCache.IsLocked.isLocked {
		ucm.TransactionCache.IsLocked.RUnlock()
		return nil, nil, TransactionFailed, terror.Error(fmt.Errorf("transactions are locked"), "Unable to process payment, please contact support.")
	}
	ucm.TransactionCache.IsLocked.RUnlock()
	if nt.Amount.Cmp(big.NewInt(0)) < 1 {
		return nil, nil, TransactionFailed, terror.Error(fmt.Errorf("amount should be a positive number: %s", nt.Amount.String()), "Amount should be greater than zero")
	}

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
		return nil, nil, TransactionFailed, terror.Error(errors.New("from: not enough funds"), "Not enough funds.")
	}

	// do add
	newToBalance := big.NewInt(0)
	newToBalance.Add(newToBalance, &toBalance)
	newToBalance.Add(newToBalance, &nt.Amount)
	if newToBalance.Cmp(big.NewInt(0)) < 0 {
		return nil, nil, TransactionFailed, terror.Error(errors.New("to: not enough funds"), "Not enough funds.")
	}

	// store back to the map
	ucm.Store(nt.From.String(), *newFromBalance)
	ucm.Store(nt.To.String(), *newToBalance)

	transactionID := ucm.TransactionCache.Process(nt)

	tx := &passport.Transaction{
		ID:     transactionID,
		Credit: nt.To,
		Debit:  nt.From,
		Amount: passport.BigInt{
			Int: nt.Amount,
		},
		TransactionReference: string(nt.TransactionReference),
		Description:          nt.Description,
		CreatedAt:            nt.CreatedAt,
		Group:                nt.Group,
		SubGroup:             nt.SubGroup,
	}

	ctx := context.Background()
	if !nt.From.IsSystemUser() {
		go ucm.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserLatestTransactionSubscribe, nt.From)), []*passport.Transaction{tx})
		go ucm.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, nt.From)), newFromBalance.String())
	}
	if !nt.To.IsSystemUser() {
		go ucm.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserLatestTransactionSubscribe, nt.To)), []*passport.Transaction{tx})
		go ucm.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, nt.To)), newToBalance.String())
	}

	return newFromBalance, newToBalance, transactionID, nil
}

func (ucm *UserCacheMap) Get(id string) (big.Int, error) {
	result, ok := ucm.Load(id)
	if ok {
		return result.(big.Int), nil
	}

	balance, err := db.UserBalance(context.Background(), ucm.conn, id)
	if err != nil {
		return *big.NewInt(0), err
	}

	ucm.Store(id, balance.Int)
	return balance.Int, err
}

type UserCacheFunc func(userCacheList UserCacheMap)
