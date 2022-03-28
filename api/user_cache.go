package api

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/shopspring/decimal"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"time"

	"github.com/gofrs/uuid"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/sasha-s/go-deadlock"
)

type UserCacheMap struct {
	deadlock.Map
	conn       *pgxpool.Pool
	MessageBus *messagebus.MessageBus
}

func NewUserCacheMap(conn *pgxpool.Pool, msgBus *messagebus.MessageBus) (*UserCacheMap, error) {
	ucm := &UserCacheMap{
		deadlock.Map{},
		conn,
		msgBus,
	}
	balances, err := db.UserBalances(context.Background(), ucm.conn)

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

func (ucm *UserCacheMap) Process(nt *passport.NewTransaction) (decimal.Decimal, decimal.Decimal, string, error) {
	if nt.Amount.LessThanOrEqual(zero) {
		return decimal.Zero, decimal.Zero, TransactionFailed, terror.Error(fmt.Errorf("amount should be a positive number: %s", nt.Amount.String()), "Amount should be greater than zero")
	}

	transactionID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	nt.ID = transactionID

	fromUser, err := boiler.FindUser(passdb.StdConn, nt.From.String())
	if err != nil {
		passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("reason", "failed to retrieve user from database").Str("id", nt.ID).Msg("transaction failed")
		return decimal.Zero, decimal.Zero, TransactionFailed, terror.Error(err, "failed to process transaction")
	}

	remaining := fromUser.Sups.Sub(nt.Amount)
	if remaining.LessThan(zero) {
		passlog.L.Info().Str("from_id", fromUser.ID).Str("to_user", nt.To.String()).Msg("account would go into negative")
		return decimal.Zero, decimal.Zero, TransactionFailed, terror.Error(ErrNotEnoughFunds, "not enough funds")
	}

	toUser, err := boiler.FindUser(passdb.StdConn, nt.To.String())
	if err != nil {
		passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("reason", "failed to retrieve user from database").Str("id", nt.ID).Msg("transaction failed")
		return decimal.Zero, decimal.Zero, TransactionFailed, terror.Error(err, "failed to process transaction")
	}

	tx := &passport.Transaction{
		ID:                   transactionID,
		Credit:               nt.To,
		Debit:                nt.From,
		Amount:               nt.Amount,
		TransactionReference: string(nt.TransactionReference),
		Description:          nt.Description,
		CreatedAt:            nt.CreatedAt,
		Group:                nt.Group,
		SubGroup:             nt.SubGroup,
	}

	//passlog.L.Info().
	//	Str("id", nt.ID).
	//	Str("from.id", fromUser.ID).
	//	Str("from.address", fromUser.PublicAddress.String).
	//	Str("to.id", toUser.ID).
	//	Str("to.address", toUser.PublicAddress.String).
	//	Str("amount", tx.Amount.Shift(-18).StringFixed(4)).
	//	Str("txref", tx.TransactionReference).
	//	Msg("processing transaction")

	blast := func(from *boiler.User, to *boiler.User, success bool) {
		if !nt.From.IsSystemUser() {
			if success {
				go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserLatestTransactionSubscribe, from.ID)), []*passport.Transaction{tx})
			}
			go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, from.ID)), from.Sups.String())
		}
		if !nt.To.IsSystemUser() {
			if success {
				go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserLatestTransactionSubscribe, to.ID)), []*passport.Transaction{tx})
			}
			go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, to.ID)), to.Sups.String())
		}
	}

	err = CreateTransactionEntry(passdb.StdConn, nt)
	if err != nil {
		ucm.Store(fromUser.ID, fromUser.Sups)
		ucm.Store(toUser.ID, toUser.Sups)
		blast(fromUser, toUser, false)
		passlog.L.Error().Err(err).Str("from", fromUser.ID).Str("to", toUser.ID).Str("id", nt.ID).Msg("transaction failed")
		return decimal.Zero, decimal.Zero, TransactionFailed, terror.Error(err)
	}

	didErr := false
	fromUser, err = boiler.FindUser(passdb.StdConn, nt.From.String())
	if err != nil {
		passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("reason", "failed to retrieve user from database").Str("id", nt.ID).Msg("transaction failed")
		didErr = true
	}

	toUser, err = boiler.FindUser(passdb.StdConn, nt.To.String())
	if err != nil {
		passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("reason", "failed to retrieve user from database").Str("id", nt.ID).Msg("transaction failed")
		didErr = true
	}

	if !didErr {
		// store back to the map
		ucm.Store(fromUser.ID, fromUser.Sups)
		ucm.Store(toUser.ID, toUser.Sups)
		blast(fromUser, toUser, true)
	}

	return fromUser.Sups, toUser.Sups, transactionID, nil
}

func (ucm *UserCacheMap) Get(id string) (decimal.Decimal, error) {
	result, ok := ucm.Load(id)
	if ok {
		return result.(decimal.Decimal), nil
	}

	balance, err := db.UserBalance(context.Background(), ucm.conn, id)
	if err != nil {
		return decimal.New(0, 18), err
	}

	ucm.Store(id, balance)
	return balance, err
}

type UserCacheFunc func(userCacheList UserCacheMap)

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, nt *passport.NewTransaction) error {
	now := time.Now()
	q := `INSERT INTO transactions(id ,description, transaction_reference, amount, credit, debit, "group", sub_group, created_at)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	_, err := conn.Exec(q, nt.ID, nt.Description, nt.TransactionReference, nt.Amount.String(), nt.To, nt.From, nt.Group, nt.SubGroup, now)
	if err != nil {
		return terror.Error(err)
	}
	nt.CreatedAt = now
	return nil
}
