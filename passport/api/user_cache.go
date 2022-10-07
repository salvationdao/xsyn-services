package api

import (
	"errors"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"sync"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/benchmark"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/volatiletech/null/v8"

	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"

	"github.com/sasha-s/go-deadlock"
)

// do not buffer runner, no waitgroups in functions
type Transactor struct {
	m          map[string]decimal.Decimal
	syndicates map[string]byte
	runner     chan func() error
	deadlock.RWMutex
}

func NewTX() (*Transactor, error) {
	ucm := &Transactor{
		make(map[string]decimal.Decimal),
		make(map[string]byte),
		make(chan func() error, 100),
		deadlock.RWMutex{},
	}
	accounts, err := boiler.Accounts().All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to retrieve user account balances")
		return nil, err
	}

	ucm.Lock()
	for _, acc := range accounts {
		ucm.m[acc.ID] = acc.Sups
		if acc.Type == boiler.AccountTypeSYNDICATE {
			ucm.syndicates[acc.ID] = 1
		}
	}
	ucm.Unlock()

	go ucm.Runner()

	return ucm, nil
}

var ErrTimeToClose = errors.New("closing")
var ErrQueueFull = errors.New("transaction queue is full")

func (ucm *Transactor) Runner() {
	for {
		select {
		case fn := <-ucm.runner:
			if fn == nil {
				return
			}
			err := fn()
			if errors.Is(err, ErrTimeToClose) {
				return
			}
		}
	}
}

func (ucm *Transactor) Close() {
	wg := sync.WaitGroup{}
	wg.Add(1)

	fn := func() error {
		wg.Done()
		return ErrTimeToClose
	}

	select {
	case ucm.runner <- fn: //queue close
	default: //unless it's full!
		passlog.L.Error().Msg("Transaction queue is blocked! Exiting.")
		return
	}
	wg.Wait()

}

func (ucm *Transactor) Transact(nt *types.NewTransaction) (string, error) {
	var trasnactionError error = nil
	transactionID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	wg := sync.WaitGroup{}
	wg.Add(1)
	fn := func() error {
		serviceID := null.StringFrom(nt.ServiceID.String())
		if nt.ServiceID.IsNil() || nt.ServiceID.String() == "" {
			serviceID.Valid = false
		}
		tx := &boiler.Transaction{
			ID:                   transactionID,
			CreditAccountID:      nt.Credit,
			DebitAccountID:       nt.Debit,
			Amount:               nt.Amount,
			TransactionReference: string(nt.TransactionReference),
			Description:          nt.Description,
			CreatedAt:            nt.CreatedAt,
			Group:                string(nt.Group),
			SubGroup:             null.StringFrom(nt.SubGroup),
			RelatedTransactionID: nt.RelatedTransactionID,
			ServiceID:            serviceID,
		}

		bm := benchmark.New()
		bm.Start("Transact func CreateTransactionEntry")
		trasnactionError = tx.Insert(passdb.StdConn, boil.Infer())
		if trasnactionError != nil {
			passlog.L.Error().Err(trasnactionError).Str("from", tx.DebitAccountID).Str("to", tx.CreditAccountID).Str("id", tx.ID).Str("amount", tx.Amount.String()).Msg("transaction failed")
			wg.Done()
			return trasnactionError
		}
		bm.End("Transact func CreateTransactionEntry")
		bm.Alert(75)

		ucm.BalanceUpdate(tx)
		wg.Done()
		return nil
	}
	select {
	case ucm.runner <- fn: //put in channel
	default: //unless it's full!
		passlog.L.Error().Msg("Transaction queue is blocked! 100 transactions waiting to be processed.")
		return transactionID, ErrQueueFull
	}
	wg.Wait()

	return transactionID, trasnactionError
}

func (ucm *Transactor) BalanceUpdate(tx *boiler.Transaction) {
	supsFromAccount, accType, err := ucm.Get(tx.DebitAccountID)
	if err != nil {
		passlog.L.Error().Err(err).Interface("tx", tx).Msg("error updating balance")
	}
	if err == nil {
		supsFromAccount = supsFromAccount.Sub(tx.Amount)
		ucm.Put(tx.DebitAccountID, supsFromAccount)

		if accType == boiler.AccountTypeUSER {
			ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", tx.DebitAccountID), HubKeyUserTransactionsSubscribe, []*boiler.Transaction{tx})
			ws.PublishMessage(fmt.Sprintf("/user/%s/sups", tx.DebitAccountID), HubKeyUserSupsSubscribe, supsFromAccount.String())
		}
	}

	supsToAccount, accType, err := ucm.Get(tx.CreditAccountID)
	if err != nil {
		passlog.L.Error().Err(err).Interface("tx", tx).Msg("error updating balance")
	}
	if err == nil {
		supsToAccount = supsToAccount.Add(tx.Amount)
		ucm.Put(tx.CreditAccountID, supsToAccount)

		if accType == boiler.AccountTypeUSER {
			ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", tx.CreditAccountID), HubKeyUserTransactionsSubscribe, []*boiler.Transaction{tx})
			ws.PublishMessage(fmt.Sprintf("/user/%s/sups", tx.CreditAccountID), HubKeyUserSupsSubscribe, supsToAccount.String())
		}
	}
}

func (ucm *Transactor) GetAndSet(ownerID string) (decimal.Decimal, string, error) {
	a, err := boiler.Accounts(
		boiler.AccountWhere.ID.EQ(ownerID),
	).One(passdb.StdConn)
	if err != nil {
		return decimal.Zero, "", err
	}

	ucm.m[a.ID] = a.Sups

	if a.Type == boiler.AccountTypeSYNDICATE {
		ucm.syndicates[ownerID] = 1
		return a.Sups, boiler.AccountTypeSYNDICATE, nil
	}
	return a.Sups, boiler.AccountTypeUSER, nil
}

func (ucm *Transactor) Get(ownerID string) (decimal.Decimal, string, error) {
	ucm.RLock()
	defer ucm.RUnlock()

	result, ok := ucm.m[ownerID]
	if ok {
		if _, isSyndicate := ucm.syndicates[ownerID]; !isSyndicate {
			return result, boiler.AccountTypeUSER, nil
		}
		return result, boiler.AccountTypeSYNDICATE, nil
	}

	return ucm.GetAndSet(ownerID)
}

func (ucm *Transactor) Put(ownerID string, sups decimal.Decimal) {
	ucm.Lock()
	ucm.m[ownerID] = sups
	ucm.Unlock()
}

func (ucm *Transactor) IsNormalUser(ownerID string) bool {
	ucm.RLock()
	defer ucm.RUnlock()

	_, ok := ucm.syndicates[ownerID]
	if ok {
		return false
	}

	return !types.IsSystemUser(ownerID)
}

func (ucm *Transactor) IsSyndicate(ownerID string) bool {
	ucm.RLock()
	defer ucm.RUnlock()
	_, ok := ucm.syndicates[ownerID]
	return ok
}

type UserCacheFunc func(userCacheList Transactor)
