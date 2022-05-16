package api

import (
	"database/sql"
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"

	"github.com/ninja-software/terror/v2"
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

func (ucm *Transactor) Transact(nt *types.NewTransaction) (decimal.Decimal, decimal.Decimal, string, error) {
	transactionID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	nt.ID = transactionID
	fromUser, toUser, msg, err := (func() (*boiler.User, *boiler.User, string, error) {
		if nt.Amount.LessThanOrEqual(zero) {
			return nil, nil, TransactionFailed, terror.Error(fmt.Errorf("amount should be a positive number: %s", nt.Amount.String()), "Amount should be greater than zero")
		}

		fromUser, err := boiler.FindUser(passdb.StdConn, nt.From.String())
		if err != nil {
			passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("reason", "failed to retrieve user from database").Str("id", nt.ID).Msg("transaction failed")
			return nil, nil, TransactionFailed, terror.Error(err, "Failed to process transaction.")
		}

		toUser, err := boiler.FindUser(passdb.StdConn, nt.To.String())
		if err != nil {
			passlog.L.Error().Err(err).Str("from", nt.From.String()).Str("to", nt.To.String()).Str("reason", "failed to retrieve user from database").Str("id", nt.ID).Msg("transaction failed")
			return nil, nil, TransactionFailed, terror.Error(err, "Failed to process transaction.")
		}

		if fromUser.ID != types.OnChainUserID.String() {
			remaining := fromUser.Sups.Sub(nt.Amount)
			if remaining.LessThan(zero) {
				passlog.L.Info().Str("from_id", fromUser.ID).Str("to_user", nt.To.String()).Msg("account would go into negative")
				return fromUser, toUser, TransactionFailed, terror.Error(ErrNotEnoughFunds, "Not enough funds.")
			}
		}

		return fromUser, toUser, "", nil
	})()

	if err != nil {
		if toUser != nil || fromUser != nil {
			failedTx := &boiler.FailedTransaction{
				ID:              transactionID,
				Credit:          toUser.ID,
				Debit:           fromUser.ID,
				Amount:          nt.Amount,
				FailedReference: string(nt.TransactionReference),
				Description:     nt.Description,
				CreatedAt:       nt.CreatedAt,
				Group:           null.StringFrom(string(nt.Group)),
				SubGroup:        null.StringFrom(string(nt.SubGroup)),
				ServiceID:       null.StringFrom(nt.ServiceID.String()),
			}

			errFailedTx := failedTx.Insert(passdb.StdConn, boil.Infer())
			if errFailedTx != nil {
				passlog.L.Error().Err(err).Msg("failed to insert failed transaction")
			}
		}
		return decimal.Zero, decimal.Zero, msg, err
	}

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
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", from.ID), HubKeyUserLatestTransactionSubscribe, []*types.Transaction{tx})
				//go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserLatestTransactionSubscribe, from.ID)), []*types.Transaction{tx})
			}
			ws.PublishMessage(fmt.Sprintf("/user/%s/sups", from.ID), HubKeyUserSupsSubscribe, from.Sups.String())
			// go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, from.ID)), from.Sups.String())
		}
		if !nt.To.IsSystemUser() {
			if success {
				ws.PublishMessage(fmt.Sprintf("/user/%s/transactions", to.ID), HubKeyUserLatestTransactionSubscribe, []*types.Transaction{tx})
				// go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserLatestTransactionSubscribe, to.ID)), []*types.Transaction{tx})
			}
			ws.PublishMessage(fmt.Sprintf("/user/%s/sups", to.ID), HubKeyUserSupsSubscribe, from.Sups.String())
			// go ucm.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, to.ID)), to.Sups.String())
		}
	}

	err = CreateTransactionEntry(passdb.StdConn, nt)
	if err != nil {
		ucm.Store(fromUser.ID, fromUser.Sups)
		ucm.Store(toUser.ID, toUser.Sups)
		blast(fromUser, toUser, false)
		passlog.L.Error().Err(err).Str("from", fromUser.ID).Str("to", toUser.ID).Str("id", nt.ID).Msg("transaction failed")
		return decimal.Zero, decimal.Zero, TransactionFailed, err
	}
	tx.CreatedAt = nt.CreatedAt

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
