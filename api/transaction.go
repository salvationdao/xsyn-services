package api

import (
	"database/sql"
	"math/big"
	"passport"
	"sync"

	"github.com/ninja-software/terror/v2"
)

type TransactionReference string

type NewTransaction struct {
	To                   passport.UserID      `json:"credit" db:"credit"`
	From                 passport.UserID      `json:"debit" db:"debit"`
	Amount               big.Int              `json:"amount" db:"amount"`
	TransactionReference TransactionReference `json:"transactionReference" db:"transaction_reference"`
	Description          string               `json:"description" db:"description"`
	ResultChan           chan *passport.Transaction
}

// HandleTransactions listens to the handle transaction channel and processes transactions, this is to ensure they happen asynchronously
func (api *API) HandleTransactions() {
	for {
		transaction := <-api.transaction
		resultTx, err := CreateTransactionEntry(
			api.TxConn,
			transaction.Amount,
			transaction.To,
			transaction.From,
			"",
			transaction.TransactionReference,
		)

		if err != nil {
			api.Log.Err(err).Msg("failed to transfer sups")
		}
		if transaction.ResultChan != nil {
			transaction.ResultChan <- resultTx
		}
	}
}

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, amount big.Int, to, from passport.UserID, description string, txRef TransactionReference) (*passport.Transaction, error) {
	result := &passport.Transaction{
		Description:          description,
		TransactionReference: string(txRef),
		Amount:               passport.BigInt{Int: amount},
		Credit:               to,
		Debit:                from,
	}
	q := `INSERT INTO transactions(description, transaction_reference, amount, credit, debit)
				VALUES($1, $2, $3, $4, $5)
				RETURNING status, reason;`

	err := conn.QueryRow(q, description, txRef, amount.String(), to, from).Scan(&result.Status, &result.Reason)
	if err != nil {
		return nil, terror.Error(err)
	}

	return result, nil
}

// HandleHeldTransactions is where the held transactions live
func (api *API) HandleHeldTransactions() {
	heldTransactions := make(map[TransactionReference]*NewTransaction)
	for {
		heldTxFunc := <-api.heldTransactions
		heldTxFunc(heldTransactions)
	}
}

// HeldTransactions accepts a function that loops over the held transaction map
func (api *API) HeldTransactions(fn func(heldTxList map[TransactionReference]*NewTransaction)) {
	var wg sync.WaitGroup
	wg.Add(1)
	api.heldTransactions <- func(heldTxList map[TransactionReference]*NewTransaction) {
		fn(heldTxList)
		wg.Done()
	}
	wg.Wait()
}

// ReleaseHeldTransaction removes a held transaction and update the users sups in the user cache
func (api *API) ReleaseHeldTransaction(txRefs ...TransactionReference) {
	for _, txRef := range txRefs {
		api.HeldTransactions(func(heldTxList map[TransactionReference]*NewTransaction) {
			tx, ok := heldTxList[txRef]
			if ok {
				errChan := make(chan error, 10)
				api.UpdateUserCacheRemoveSups(tx.To, tx.Amount, errChan)
				err := <-errChan
				if err != nil {
					api.Log.Info().Err(err)
					return
				}
				api.UpdateUserCacheAddSups(tx.From, tx.Amount)
				delete(heldTxList, txRef)
			}
		})
	}
}

// HoldTransaction adds a new transaction to the hold transaction map and updates the user cache sups accordingly
func (api *API) HoldTransaction(holdErrChan chan error, txs ...*NewTransaction) {
	// Here we take the sups away from the user in their cache and hold the transactions in a slice
	// So later we can fire the commit command and put all the transactions into the database
	// HERE SHIT ISNT WORKING
	api.HeldTransactions(func(heldTxList map[TransactionReference]*NewTransaction) {
		for _, tx := range txs {
			errChan := make(chan error, 10)
			api.UpdateUserCacheRemoveSups(tx.From, tx.Amount, errChan)
			err := <-errChan
			if err != nil {
				api.Log.Err(err).Msg(err.Error())
				holdErrChan <- err
				return
			}
			api.UpdateUserCacheAddSups(tx.To, tx.Amount)
			heldTxList[tx.TransactionReference] = tx
		}
		holdErrChan <- nil
	})
}

// CommitTransactions goes through and commits the given transactions, returning their status
func (api *API) CommitTransactions(resultChan chan []*passport.Transaction, txRefs ...TransactionReference) {
	results := []*passport.Transaction{}
	// we loop the transactions, and see the results!

	api.HeldTransactions(func(heldTxList map[TransactionReference]*NewTransaction) {
		for _, txRef := range txRefs {
			if tx, ok := heldTxList[txRef]; ok {
				tx.ResultChan = make(chan *passport.Transaction, 1)
				api.transaction <- tx
				result := <-tx.ResultChan
				// if result is failed, update the cache map
				if result.Status == passport.TransactionFailed {
					errChan := make(chan error, 10)
					api.UpdateUserCacheRemoveSups(tx.To, tx.Amount, errChan)
					err := <-errChan
					if err != nil {
						api.Log.Err(err).Msg(err.Error())
						continue
					}
					api.UpdateUserCacheAddSups(tx.From, tx.Amount)
				}
				results = append(results, result)

				// remove cached transaction after it is committed
				delete(heldTxList, txRef)
			}
		}
		resultChan <- results
	})
}
