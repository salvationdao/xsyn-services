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

	"github.com/friendsofgo/errors"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

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
	users, err := boiler.Users(
		qm.Load(boiler.UserRels.Account),
	).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to retrieve user account balances")
		return nil, err
	}

	for _, u := range users {
		if u.R.Account == nil {
			passlog.L.Warn().Str("user id", u.ID).Msg("User missing account")
			return nil, fmt.Errorf("user missing account")
		}

		ucm.Put(u.ID, u.R.Account)
		ucm.PutAccountLookup(u.R.Account.ID, u.ID, true)
	}

	syndicates, err := boiler.Syndicates(
		qm.Load(boiler.SyndicateRels.Account),
	).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to retrieve syndicate account balances")
		return nil, err
	}

	for _, s := range syndicates {
		if s.R.Account == nil {
			passlog.L.Warn().Str("syndicate id", s.ID).Msg("Syndicate missing account")
			return nil, fmt.Errorf("syndicate missing account")
		}

		ucm.Put(s.ID, s.R.Account)
		ucm.PutAccountLookup(s.R.Account.ID, s.ID, false)
	}

	go ucm.RunBalanceUpdate()

	return ucm, nil
}

var TransactionFailed = "TRANSACTION_FAILED"
var zero = decimal.New(0, 18)
var ErrNotEnoughFunds = fmt.Errorf("account does not have enough funds")

func (ucm *Transactor) Transact(nt *types.NewTransaction) (string, error) {
	creditAccount, err := ucm.Get(nt.Credit)
	if err != nil {
		return "", err
	}
	debitAccount, err := ucm.Get(nt.Debit)
	if err != nil {
		return "", err
	}

	transactionID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	nt.ID = transactionID

	tx := &boiler.Transaction{
		ID:                   transactionID,
		CreditAccountID:      creditAccount.ID,
		DebitAccountID:       debitAccount.ID,
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
	err = CreateTransactionEntry(passdb.StdConn, tx)
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
	var account *boiler.Account
	isUser := false

	u, err := boiler.Users(
		boiler.UserWhere.ID.EQ(ownerID),
		qm.Load(boiler.UserRels.Account),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if u != nil {
		// check account is exist
		if u.R.Account == nil {
			return nil, fmt.Errorf("user does not have an account")
		}

		account = u.R.Account
		isUser = true

	} else {
		// try getting account from syndicate table
		s, err := boiler.Syndicates(
			boiler.SyndicateWhere.ID.EQ(ownerID),
			qm.Load(boiler.SyndicateRels.Account),
		).One(passdb.StdConn)
		if err != nil {
			return nil, err
		}

		// check account exist
		if s.R.Account == nil {
			return nil, fmt.Errorf("syndicate does not have an account")
		}

		account = s.R.Account
	}

	// store data to the map
	ucm.Put(ownerID, account)
	ucm.PutAccountLookup(account.ID, ownerID, isUser)

	return account, nil
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

type AccountLookup struct {
	OwnerID string
	IsUser  bool
}

func (ucm *Transactor) PutAccountLookup(accountID string, ownerID string, isUser bool) {
	ucm.accountLookup.Store(accountID, &AccountLookup{ownerID, isUser})
}

func (ucm *Transactor) PutAndGetAccountLookup(accountID string) (string, bool, error) {
	id := ""
	isUser := false
	user, err := boiler.Users(
		boiler.UserWhere.AccountID.EQ(accountID),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}

	if user != nil {
		id = user.ID
		isUser = true
	} else {
		// check account id from syndicate
		syndicate, err := boiler.Syndicates(
			boiler.SyndicateWhere.AccountID.EQ(accountID),
		).One(passdb.StdConn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", false, err
		}

		id = syndicate.ID
	}

	ucm.PutAccountLookup(accountID, id, isUser)
	return id, isUser, nil
}

func (ucm *Transactor) GetAccountLookup(accountID string) (string, bool, error) {
	result, ok := ucm.accountLookup.Load(accountID)
	if ok {
		al := result.(*AccountLookup)
		return al.OwnerID, al.IsUser, nil
	}

	return ucm.PutAndGetAccountLookup(accountID)
}

func (ucm *Transactor) IsNormalUser(ownerID string) bool {
	acc, err := ucm.Get(ownerID)
	if err != nil || acc.Type != boiler.AccountTypeUSER {
		return false
	}

	return !types.IsSystemUser(ownerID)
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
