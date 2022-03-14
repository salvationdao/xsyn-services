package api

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gofrs/uuid"
	"math/big"
	"passport"
	"passport/db"
	"passport/passdb"
	"passport/passlog"
	"time"

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
		return nil, err
	}

	for _, b := range balances {
		ucm.Store(b.ID.String(), b.Sups.Int)
	}
	return ucm, nil
}

var TransactionFailed = "TRANSACTION_FAILED"

func (ucm *UserCacheMap) Process(nt *passport.NewTransaction) (*big.Int, *big.Int, string, error) {
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
	if nt.To != passport.OnChainUserID && newFromBalance.Cmp(big.NewInt(0)) < 0 {
		return nil, nil, TransactionFailed, terror.Error(fmt.Errorf("from: not enough funds"), "Not enough funds.")
	}

	// do add
	newToBalance := big.NewInt(0)
	newToBalance.Add(newToBalance, &toBalance)
	newToBalance.Add(newToBalance, &nt.Amount)
	if nt.To != passport.OnChainUserID && newToBalance.Cmp(big.NewInt(0)) < 0 {
		return nil, nil, TransactionFailed, terror.Error(fmt.Errorf("to: not enough funds"), "Not enough funds.")
	}

	// store back to the map
	ucm.Store(nt.From.String(), *newFromBalance)
	ucm.Store(nt.To.String(), *newToBalance)

	transactionID := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	nt.ID = transactionID

	passlog.L.Info().Str("from", fromBalance.String()).Str("to", newToBalance.String()).Str("id", nt.ID).Msg("processing transaction")

	err = CreateTransactionEntry(passdb.StdConn, nt)
	if err != nil {
		passlog.L.Error().Err(err).Str("from", fromBalance.String()).Str("to", newToBalance.String()).Str("id", nt.ID).Msg("transaction failed")
		return nil, nil, TransactionFailed, terror.Error(err)
	}

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

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, nt *passport.NewTransaction) error {
	q := `INSERT INTO transactions(id ,description, transaction_reference, amount, credit, debit, "group", sub_group)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8);`

	_, err := conn.Exec(q, nt.ID, nt.Description, nt.TransactionReference, nt.Amount.String(), nt.To, nt.From, nt.Group, nt.SubGroup)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
