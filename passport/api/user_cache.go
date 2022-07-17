package api

import (
	"database/sql"
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/benchmark"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/volatiletech/null/v8"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-syndicate/ws"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"

	"github.com/sasha-s/go-deadlock"
)

type Transactor struct {
	deadlock.Map
	updateBalance chan *types.NewTransaction
	accountLookup deadlock.Map
}

func NewTX() (*Transactor, error) {
	ucm := &Transactor{
		deadlock.Map{},
		make(chan *types.NewTransaction, 100),
		deadlock.Map{},
	}
	accounts, err := boiler.Accounts().All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to retrieve user account balances")
		return nil, err
	}

	for _, acc := range accounts {
		ucm.Put(acc.ID, acc)
	}

	go ucm.RunBalanceUpdate()

	return ucm, nil
}

var TransactionFailed = "TRANSACTION_FAILED"
var zero = decimal.New(0, 18)
var ErrNotEnoughFunds = fmt.Errorf("account does not have enough funds")

func (ucm *Transactor) Transact(nt *types.NewTransaction) (string, error) {
	transactionID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	nt.ID = transactionID

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
		ServiceID:            null.StringFrom(nt.ServiceID.String()),
	}
	bm := benchmark.New()
	bm.Start("Transact func CreateTransactionEntry")
	err := CreateTransactionEntry(passdb.StdConn, tx)
	if err != nil {
		passlog.L.Error().Err(err).Str("from", nt.Debit).Str("to", nt.Credit).Str("id", nt.ID).Msg("transaction failed")
		return TransactionFailed, err
	}
	bm.End("Transact func CreateTransactionEntry")
	bm.Alert(75)
	tx.CreatedAt = nt.CreatedAt

	ucm.updateBalance <- nt

	return transactionID, nil
}

func (ucm *Transactor) RunBalanceUpdate() {
	for {
		nt := <-ucm.updateBalance

		fromAccount, err := ucm.Get(nt.Debit)
		if err == nil {
			fromAccount.Sups = fromAccount.Sups.Sub(nt.Amount)
			ucm.Put(nt.Debit, fromAccount)

			if ucm.IsNormalUser(nt.Debit) {
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", nt.Debit), HubKeyUserTransactionsSubscribe, []*types.NewTransaction{nt})
				ws.PublishMessage(fmt.Sprintf("/user/%s/sups", nt.Debit), HubKeyUserSupsSubscribe, fromAccount.Sups.String())
			}
		}

		toAccount, err := ucm.Get(nt.Credit)
		if err == nil {
			toAccount.Sups = toAccount.Sups.Add(nt.Amount)
			ucm.Put(nt.Credit, toAccount)

			if ucm.IsNormalUser(nt.Credit) {
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", nt.Credit), HubKeyUserTransactionsSubscribe, []*types.NewTransaction{nt})
				ws.PublishMessage(fmt.Sprintf("/user/%s/sups", nt.Credit), HubKeyUserSupsSubscribe, toAccount.Sups.String())
			}
		}
	}
}

func (ucm *Transactor) SetAndGet(ownerID string) (*boiler.Account, error) {
	a, err := boiler.Accounts(
		boiler.AccountWhere.ID.EQ(ownerID),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	// store data to the map
	ucm.Put(ownerID, a)

	return a, nil
}

func (ucm *Transactor) Get(ownerID string) (*boiler.Account, error) {
	result, ok := ucm.Load(ownerID)
	if !ok {
		return result.(*boiler.Account), nil
	}

	return ucm.SetAndGet(ownerID)
}

func (ucm *Transactor) Put(ownerID string, account *boiler.Account) {
	ucm.Store(ownerID, account)
}

func (ucm *Transactor) IsNormalUser(ownerID string) bool {
	acc, err := ucm.Get(ownerID)
	if err != nil || acc.Type != boiler.AccountTypeUSER {
		return false
	}

	return !types.IsSystemUser(ownerID)
}

func (ucm *Transactor) IsSyndicate(ownerID string) bool {
	acc, err := ucm.Get(ownerID)
	if err != nil || acc.Type != boiler.AccountTypeSYNDICATE {
		return false
	}

	return true
}

type UserCacheFunc func(userCacheList Transactor)

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, nt *boiler.Transaction) error {
	now := time.Now()
	nt.CreatedAt = now
	err := nt.Insert(conn, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}
