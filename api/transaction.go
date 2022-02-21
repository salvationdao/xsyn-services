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
	for heldTxFunc := range api.heldTransactions {
		heldTxFunc(heldTransactions)
	}
}

// HeldTransactions accepts a function that loops over the held transaction map
func (api *API) HeldTransactions(fn func(heldTxList map[passport.TransactionReference]*passport.NewTransaction), stuff ...string) {
	if len(stuff) > 0 {
		fmt.Printf("start %s\n", stuff[0])
	}
	var wg sync.WaitGroup
	wg.Add(1)

	api.heldTransactions <- func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		fn(heldTxList)
		wg.Done()
	}
	wg.Wait()
	if len(stuff) > 0 {
		fmt.Printf("end %s\n", stuff[0])
	}
}

// ReleaseHeldTransaction removes a held transaction and update the users sups in the user cache
func (api *API) ReleaseHeldTransaction(ctx context.Context, txRefs ...passport.TransactionReference) {
	for _, txRef := range txRefs {
		api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
			tx, ok := heldTxList[txRef]
			if ok {
				err := api.UpdateUserCacheRemoveSups(ctx, tx.To, tx.Amount)
				if err != nil {
					api.Log.Info().Err(err)
					return
				}
				api.UpdateUserCacheAddSups(ctx, tx.From, tx.Amount)
				delete(heldTxList, txRef)
			}
		}, "ReleaseHeldTransaction")
	}
}

// HoldTransaction adds a new transaction to the hold transaction map and updates the user cache sups accordingly
func (api *API) HoldTransaction(ctx context.Context, tx *passport.NewTransaction) error {
	var err error = nil
	// Here we take the sups away from the user in their cache and hold the transactions in a slice
	// So later we can fire the commit command and put all the transactions into the database
	api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		//for _, tx := range txs {
		fmt.Println("START UpdateUserCacheRemoveSups")
		err = api.UpdateUserCacheRemoveSups(ctx, tx.From, tx.Amount)
		fmt.Println("END UpdateUserCacheRemoveSups")
		if err != nil {
			return
		}
		fmt.Println("START UpdateUserCacheAddSups")
		api.UpdateUserCacheAddSups(ctx, tx.To, tx.Amount)
		fmt.Println("END UpdateUserCacheAddSups")

		heldTxList[tx.TransactionReference] = tx
		//}
	}, "HoldTransaction")
	return err
}

// CommitTransactions goes through and commits the given transactions, returning their status
func (api *API) CommitTransactions(ctx context.Context, txRefs ...passport.TransactionReference) []*passport.Transaction {
	results := []*passport.Transaction{}
	// we loop the transactions, and see the results!
	api.HeldTransactions(func(heldTxList map[passport.TransactionReference]*passport.NewTransaction) {
		for _, txRef := range txRefs {
			if tx, ok := heldTxList[txRef]; ok {
				tx.ResultChan = make(chan *passport.TransactionResult)
				api.transaction <- tx
				result := <-tx.ResultChan
				// if result is failed, update the cache map
				if result.Error != nil || result.Transaction.Status == passport.TransactionFailed {
					api.Log.Debug().Msg("START UpdateUserCacheRemoveSups in CommitTransactions")
					err := api.UpdateUserCacheRemoveSups(ctx, tx.To, tx.Amount)
					api.Log.Debug().Msg("FINISH UpdateUserCacheRemoveSups in CommitTransactions")
					if err != nil {
						api.Log.Err(err).Msg(err.Error())
						results = append(results, result.Transaction)
						delete(heldTxList, txRef)
						continue
					}
				}
				results = append(results, result.Transaction)
				// remove cached transaction after it is committed
				delete(heldTxList, txRef)
			} else {
				api.Log.Warn().Msgf("unable to find tx ref %s in held transactions", txRef)
			}
		}

	}, "CommitTransactions")
	return results
}
