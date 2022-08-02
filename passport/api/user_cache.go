package api

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"
	"xsyn-services/passport/benchmark"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/ws"

	"github.com/shopspring/decimal"

	"github.com/sasha-s/go-deadlock"
)

// do not buffer runner, no waitgroups in functions
type Transactor struct {
	deadlock.Map
	runner chan func() error
}

func NewTX() (*Transactor, error) {
	ucm := &Transactor{
		deadlock.Map{},
		make(chan func() error, 100),
	}
	balances, err := db.UserBalances()

	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to retrieve balances")
		return nil, err
	}

	for _, b := range balances {
		ucm.Store(b.ID.String(), b.Sups)
	}

	go ucm.Runner()

	return ucm, nil
}

var TransactionFailed = "TRANSACTION_FAILED"
var zero = decimal.New(0, 18)
var ErrNotEnoughFunds = fmt.Errorf("account does not have enough funds")

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
		nt.ID = transactionID
		tx := &types.Transaction{
			ID:                   transactionID,
			Credit:               nt.To,
			Debit:                nt.From,
			Amount:               nt.Amount,
			TransactionReference: string(nt.TransactionReference),
			Description:          nt.Description,
			CreatedAt:            nt.CreatedAt,
			Group:                nt.Group,
			SubGroup:             nt.SubGroup,
			RelatedTransactionID: nt.RelatedTransactionID,
			ServiceID:            nt.ServiceID,
		}
		bm := benchmark.New()
		bm.Start("Transact func CreateTransactionEntry")
		err = CreateTransactionEntry(passdb.StdConn, nt)
		if err != nil {
			passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("id", nt.ID).Msg("transaction failed")
			wg.Done()
			return err
		}
		bm.End("Transact func CreateTransactionEntry")
		bm.Alert(75)
		tx.CreatedAt = nt.CreatedAt

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

func (ucm *Transactor) BalanceUpdate(tx *types.Transaction) {
	fromBalance, err := ucm.Get(tx.Debit.String())
	if err == nil {
		newFromBalance := fromBalance.Sub(tx.Amount)
		ucm.Store(tx.Debit.String(), newFromBalance)

		if !tx.Debit.IsSystemUser() {
			ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", tx.Debit), HubKeyUserTransactionsSubscribe, []*types.Transaction{tx})
			ws.PublishMessage(fmt.Sprintf("/user/%s/sups", tx.Debit), HubKeyUserSupsSubscribe, newFromBalance.String())
		}
	}

	toBalance, err := ucm.Get(tx.Credit.String())
	if err == nil {
		newToBalance := toBalance.Add(tx.Amount)
		ucm.Store(tx.Credit.String(), newToBalance)

		if !tx.Credit.IsSystemUser() {
			ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", tx.Credit), HubKeyUserTransactionsSubscribe, []*types.Transaction{tx})
			ws.PublishMessage(fmt.Sprintf("/user/%s/sups", tx.Credit), HubKeyUserSupsSubscribe, newToBalance.String())
		}
	}

}

func (ucm *Transactor) SetAndGet(id string) (decimal.Decimal, error) {
	balance, err := db.UserBalance(id)
	if err != nil {
		return decimal.New(0, 18), err
	}

	ucm.Store(id, balance)
	return balance, err
}

func (ucm *Transactor) Get(id string) (decimal.Decimal, error) {
	result, ok := ucm.Load(id)
	if ok {
		return result.(decimal.Decimal), nil
	}

	return ucm.SetAndGet(id)
}

type UserCacheFunc func(userCacheList Transactor)

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, nt *types.NewTransaction) error {
	now := time.Now()
	q := `INSERT INTO transactions(
                         id, 
                         description, 
                         transaction_reference, 
                         amount, 
                         credit, 
                         debit, 
                         "group", 
                         sub_group, 
                         created_at, 
                         service_id, 
                         related_transaction_id
                         )
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);`

	_, err := conn.Exec(q,
		nt.ID,
		nt.Description,
		nt.TransactionReference,
		nt.Amount.String(),
		nt.To,
		nt.From,
		nt.Group,
		nt.SubGroup,
		now,
		nt.ServiceID,
		nt.RelatedTransactionID,
	)
	if err != nil {
		return err
	}
	nt.CreatedAt = now
	return nil
}
