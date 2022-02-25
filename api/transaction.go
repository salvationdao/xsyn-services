package api

import (
	"database/sql"
	"fmt"
	"passport"
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

type TransactionCache struct {
	sync.RWMutex
	conn         *sql.DB
	log          *zerolog.Logger
	transactions []*passport.NewTransaction
}

func NewTransactionCache(conn *sql.DB, log *zerolog.Logger) *TransactionCache {
	tc := &TransactionCache{
		sync.RWMutex{},
		conn,
		log,
		[]*passport.NewTransaction{},
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
			tx.TransactionReference,
		)
		if err != nil {
			tc.log.Err(err)
			if !tx.Safe {
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
		if t.Safe {
			tc.commit()
		}
	}()
	tc.transactions = append(tc.transactions, t)

	return t.ID
}

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, nt *passport.NewTransaction, txRef passport.TransactionReference) error {
	q := `INSERT INTO transactions(id ,description, transaction_reference, amount, credit, debit, group_id , created_at)
				VALUES($1, $2, $3, $4, $5, $6, $7);`

	_, err := conn.Exec(q, nt.ID, nt.Description, txRef, nt.Amount.String(), nt.To, nt.From, nt.GroupID, nt.CreatedAt)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
