package api

import (
	"database/sql"
	"fmt"
	"time"
	"xsyn-services/passport/benchmark"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"

	"github.com/sasha-s/go-deadlock"
)

type Transactor struct {
	deadlock.Map
}

func NewTX() (*Transactor, error) {
	ucm := &Transactor{
		deadlock.Map{},
	}
	balances, err := db.UserBalances()

	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to retrieve balances")
		return nil, err
	}

	for _, b := range balances {
		ucm.Store(b.ID.String(), b.Sups)
	}
	return ucm, nil
}

var TransactionFailed = "TRANSACTION_FAILED"
var zero = decimal.New(0, 18)
var ErrNotEnoughFunds = fmt.Errorf("account does not have enough funds")

func (ucm *Transactor) Transact(nt *types.NewTransaction) (string, error) {


	transactionID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
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
	err := CreateTransactionEntry(passdb.StdConn, nt)
	if err != nil {
		passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("id", nt.ID).Msg("transaction failed")
		return TransactionFailed, err
	}
	bm.End("Transact func CreateTransactionEntry")
	bm.Alert(75)
	tx.CreatedAt = nt.CreatedAt

	go func(fromID, toID types.UserID, amount decimal.Decimal, _tx *types.Transaction) {
		fromBalance, err := ucm.Get(fromID.String())
		if err == nil {
			newFromBalance := fromBalance.Sub(amount)
			ucm.Store(fromID, newFromBalance)

			if !fromID.IsSystemUser() {
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", fromID), HubKeyUserTransactionsSubscribe, []*types.Transaction{_tx})
				ws.PublishMessage(fmt.Sprintf("/user/%s/sups", fromID), HubKeyUserSupsSubscribe, newFromBalance.String())
			}
		}

		toBalance, err := ucm.Get(toID.String())
		if err == nil {
			newToBalance := toBalance.Add(amount)
			ucm.Store(toID, newToBalance)

			if !toID.IsSystemUser() {
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", toID), HubKeyUserTransactionsSubscribe, []*types.Transaction{_tx})
				ws.PublishMessage(fmt.Sprintf("/user/%s/sups", toID), HubKeyUserSupsSubscribe, newToBalance.String())
			}
		}
	}(nt.From, nt.To, nt.Amount, tx)


	return transactionID, nil
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
