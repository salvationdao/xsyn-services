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
	m             map[string]decimal.Decimal
	syndicates    map[string]byte
	updateBalance chan *types.NewTransaction
	deadlock.RWMutex
}

func NewTX() (*Transactor, error) {
	ucm := &Transactor{
		make(map[string]decimal.Decimal),
		make(map[string]byte),
		make(chan *types.NewTransaction, 100),
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

		supsFromAccount, accType, err := ucm.Get(nt.Debit)
		if err == nil {
			supsFromAccount = supsFromAccount.Sub(nt.Amount)
			ucm.Put(nt.Debit, supsFromAccount)

			if accType == boiler.AccountTypeUSER {
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", nt.Debit), HubKeyUserTransactionsSubscribe, []*types.NewTransaction{nt})
				ws.PublishMessage(fmt.Sprintf("/user/%s/sups", nt.Debit), HubKeyUserSupsSubscribe, supsFromAccount.String())
			}
		}

		supsToAccount, accType, err := ucm.Get(nt.Credit)
		if err == nil {
			supsToAccount = supsToAccount.Add(nt.Amount)
			ucm.Put(nt.Credit, supsToAccount)

			if accType == boiler.AccountTypeUSER {
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", nt.Credit), HubKeyUserTransactionsSubscribe, []*types.NewTransaction{nt})
				ws.PublishMessage(fmt.Sprintf("/user/%s/sups", nt.Credit), HubKeyUserSupsSubscribe, supsToAccount.String())
			}
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

	// store data to the map
	return ucm.PutFromAccount(ownerID, a)
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

func (ucm *Transactor) PutFromAccount(ownerID string, acc *boiler.Account) (decimal.Decimal, string, error) {
	ucm.m[acc.ID] = acc.Sups
	if acc.Type == boiler.AccountTypeSYNDICATE {
		ucm.syndicates[ownerID] = 1
		return acc.Sups, boiler.AccountTypeSYNDICATE, nil
	}
	return acc.Sups, boiler.AccountTypeUSER, nil
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
