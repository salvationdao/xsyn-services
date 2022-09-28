package api

import (
	"errors"
	"fmt"
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
	accounts, err := boiler.Users().All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to retrieve user account balances")
		return nil, err
	}

	ucm.Lock()
	for _, acc := range accounts {
		ucm.m[acc.ID] = acc.Sups
		//if acc.Type == boiler.AccountTypeSYNDICATE {
		//	ucm.syndicates[acc.ID] = 1
		//}
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
	var err error = nil
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
			Credit:               nt.Credit,
			Debit:                nt.Debit,
			Amount:               nt.Amount,
			TransactionReference: string(nt.TransactionReference),
			Description:          nt.Description,
			CreatedAt:            time.Now(),
			Group:                string(nt.Group),
			SubGroup:             null.StringFrom(nt.SubGroup),
			RelatedTransactionID: nt.RelatedTransactionID,
			ServiceID:            serviceID,
		}

		bm := benchmark.New()
		bm.Start("Transact func CreateTransactionEntry")
		_, err = passdb.StdConn.Exec(`
					INSERT INTO transactions (id, description, transaction_reference, amount, credit, debit, created_at, "group", sub_group, service_id, related_transaction_id)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			tx.ID,
			tx.Description,
			tx.TransactionReference,
			tx.Amount,
			tx.Credit,
			tx.Debit,
			tx.CreatedAt,
			tx.Group,
			tx.SubGroup,
			tx.ServiceID,
			tx.RelatedTransactionID,
		)
		if err != nil {
			passlog.L.Error().Err(err).Str("from", tx.Debit).Str("to", tx.Credit).Str("id", nt.ID).Msg("transaction failed")
			wg.Done()
			return err
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

	return transactionID, err
}

func (ucm *Transactor) BalanceUpdate(tx *boiler.Transaction) {
	supsFromAccount, err := ucm.Get(tx.Debit)
	if err != nil {
		passlog.L.Error().Err(err).Interface("tx", tx).Msg("error updating balance")
	}
	if err == nil {
		supsFromAccount = supsFromAccount.Sub(tx.Amount)
		ucm.Put(tx.Debit, supsFromAccount)

		ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", tx.Debit), HubKeyUserTransactionsSubscribe, []*boiler.Transaction{tx})
		ws.PublishMessage(fmt.Sprintf("/user/%s/sups", tx.Debit), HubKeyUserSupsSubscribe, supsFromAccount.String())
	}

	supsToAccount, err := ucm.Get(tx.Credit)
	if err != nil {
		passlog.L.Error().Err(err).Interface("tx", tx).Msg("error updating balance")
	}
	if err == nil {
		supsToAccount = supsToAccount.Add(tx.Amount)
		ucm.Put(tx.Credit, supsToAccount)

		ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", tx.Credit), HubKeyUserTransactionsSubscribe, []*boiler.Transaction{tx})
		ws.PublishMessage(fmt.Sprintf("/user/%s/sups", tx.Credit), HubKeyUserSupsSubscribe, supsToAccount.String())
	}
}

func (ucm *Transactor) GetAndSet(ownerID string) (decimal.Decimal, error) {
	a, err := boiler.Users(
		boiler.UserWhere.ID.EQ(ownerID),
	).One(passdb.StdConn)
	if err != nil {
		return decimal.Zero, err
	}

	ucm.m[a.ID] = a.Sups

	return a.Sups, nil
}

func (ucm *Transactor) Get(ownerID string) (decimal.Decimal, error) {
	ucm.RLock()
	defer ucm.RUnlock()

	result, ok := ucm.m[ownerID]
	if ok {
		return result, nil
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
