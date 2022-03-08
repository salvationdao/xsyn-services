package api

import (
	"database/sql"
	"fmt"
	"passport"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"github.com/sasha-s/go-deadlock"
)

type IsLocked struct {
	deadlock.RWMutex
	isLocked bool
}

type TransactionCache struct {
	deadlock.RWMutex
	conn         *sql.DB
	log          *zerolog.Logger
	transactions []*passport.NewTransaction
	IsLocked     *IsLocked
}

func NewTransactionCache(conn *sql.DB, log *zerolog.Logger) *TransactionCache {
	tc := &TransactionCache{
		deadlock.RWMutex{},
		conn,
		log,
		[]*passport.NewTransaction{},
		&IsLocked{
			isLocked: false,
		},
	}

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			<-ticker.C
			tc.commit()
		}
	}()

	return tc
}

func (tc *TransactionCache) commit() {
	tc.Lock()
	ctrans := make([]*passport.NewTransaction, len(tc.transactions))
	copy(ctrans, tc.transactions)
	tc.transactions = []*passport.NewTransaction{}
	tc.Unlock()
	for _, tx := range ctrans {
		err := CreateTransactionEntry(
			tc.conn,
			tx,
		)

		if err != nil {
			tc.
				log.
				Err(err).
				Str("amt", tx.Amount.String()).
				Str("from", tx.From.String()).
				Str("to", tx.To.String()).
				Str("txref", string(tx.TransactionReference)).
				Msg("transaction cache lock")
			if tx.NotSafe {
				tc.IsLocked.Lock()
				tc.IsLocked.isLocked = true
				tc.IsLocked.Unlock()
				tc.Lock() //grind to a halt if transactions fail to save to database
			}
			return
		}
	}
}

func (tc *TransactionCache) Process(t *passport.NewTransaction) string {

	if t.Processed {
		return t.ID
	}
	t.ID = fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	t.CreatedAt = time.Now()
	t.Processed = true

	tc.Lock()
	defer func() {
		tc.Unlock()
		if !t.NotSafe {
			tc.commit()
		}
	}()
	tc.transactions = append(tc.transactions, t)

	return t.ID
}

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, nt *passport.NewTransaction) error {
	q := `INSERT INTO transactions(id ,description, transaction_reference, amount, credit, debit, "group", sub_group, created_at)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	_, err := conn.Exec(q, nt.ID, nt.Description, nt.TransactionReference, nt.Amount.String(), nt.To, nt.From, nt.Group, nt.SubGroup, nt.CreatedAt)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
