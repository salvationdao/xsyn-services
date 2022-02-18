package api

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"passport"
	"strings"
	"sync"

	"github.com/ninja-software/terror/v2"
)

// HandleTransactions listens to the handle transaction channel and processes transactions, this is to ensure they happen asynchronously
func (api *API) HandleTransactions() {
	for {
		transaction := <-api.transaction
		transactionResult := &passport.TransactionResult{}

		resultTx, err := CreateTransactionEntry(
			api.TxConn,
			transaction.Amount,
			transaction.To,
			transaction.From,
			transaction.Description,
			transaction.TransactionReference,
		)
		if err != nil {
			transactionResult.Error = terror.Error(err, "failed to transfer sups")
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				api.Log.Debug().Err(err).Msg("failed to transfer sups")
			} else {
				api.Log.Err(err).Msg("failed to transfer sups")
			}
		}

		transactionResult.Transaction = resultTx
		if transaction.ResultChan != nil {
			transaction.ResultChan <- transactionResult
		}
	}
}

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, amount big.Int, to, from passport.UserID, description string, txRef passport.TransactionReference) (*passport.Transaction, error) {
	result := &passport.Transaction{
		Description:          description,
		TransactionReference: string(txRef),
		Amount:               passport.BigInt{Int: amount},
		Credit:               to,
		Debit:                from,
	}
	q := `INSERT INTO transactions(description, transaction_reference, amount, credit, debit)
				VALUES($1, $2, $3, $4, $5)
				RETURNING id, status, reason;`

	err := conn.QueryRow(q, description, txRef, amount.String(), to, from).Scan(&result.ID, &result.Status, &result.Reason)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// HandleHeldTransactions is where the held transactions live
func (api *API) HandleHeldTransactions() {
	heldTransactions := make(map[passport.TransactionReference]*passport.NewTransaction)
	for {
		heldTxFunc := <-api.heldTransactions
		heldTxFunc(heldTransactions)
	}
}

// HeldTransactions accepts a function that loops over the held transaction map
func (api *API) HeldTransactions(fn func(heldTxList map[passport.TransactionReference]*passport.NewTransaction)) {
	var wg sync.WaitGroup
	wg.Add(1)
	api.heldTransactions <- func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		fn(heldTxList)
		wg.Done()
	}
	wg.Wait()
}

// ReleaseHeldTransaction removes a held transaction and update the users sups in the user cache
func (api *API) ReleaseHeldTransaction(ctx context.Context, txRefs ...passport.TransactionReference) {
	for _, txRef := range txRefs {
		api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
			tx, ok := heldTxList[txRef]
			if ok {
				errChan := make(chan error, 10)
				api.UpdateUserCacheRemoveSups(ctx, tx.To, tx.Amount, errChan)
				err := <-errChan
				if err != nil {
					api.Log.Info().Err(err)
					return
				}
				api.UpdateUserCacheAddSups(ctx, tx.From, tx.Amount)
				delete(heldTxList, txRef)
			}
		})
	}
}

// HoldTransaction adds a new transaction to the hold transaction map and updates the user cache sups accordingly
func (api *API) HoldTransaction(ctx context.Context, holdErrChan chan error, txs ...*passport.NewTransaction) {
	// Here we take the sups away from the user in their cache and hold the transactions in a slice
	// So later we can fire the commit command and put all the transactions into the database
	api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		for _, tx := range txs {
			errChan := make(chan error, 10)
			fmt.Println("start remove sup")
			api.UpdateUserCacheRemoveSups(ctx, tx.From, tx.Amount, errChan)
			err := <-errChan
			fmt.Println("end remove sup")
			if err != nil {
				holdErrChan <- err
				return
			}
			api.UpdateUserCacheAddSups(ctx, tx.To, tx.Amount)
			heldTxList[tx.TransactionReference] = tx
		}
		holdErrChan <- nil
	})
}

// CommitTransactions goes through and commits the given transactions, returning their status
func (api *API) CommitTransactions(ctx context.Context, resultChan chan []*passport.Transaction, txRefs ...passport.TransactionReference) {
	results := []*passport.Transaction{}
	// we loop the transactions, and see the results!
	api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		for _, txRef := range txRefs {
			if tx, ok := heldTxList[txRef]; ok {
				tx.ResultChan = make(chan *passport.TransactionResult, 1)
				api.transaction <- tx
				result := <-tx.ResultChan
				// if result is failed, update the cache map
				if result.Error != nil || result.Transaction.Status == passport.TransactionFailed {
					errChan := make(chan error, 10)
					api.UpdateUserCacheRemoveSups(ctx, tx.To, tx.Amount, errChan)
					err := <-errChan
					if err != nil {
						api.Log.Err(err).Msg(err.Error())
						continue
					}
					api.UpdateUserCacheAddSups(ctx, tx.From, tx.Amount)
				}
				results = append(results, result.Transaction)

				// remove cached transaction after it is committed
				delete(heldTxList, txRef)
			} else {
				api.Log.Warn().Msgf("unable to find tx ref %s in held transactions", txRef)
			}
		}
		resultChan <- results
	})
}
