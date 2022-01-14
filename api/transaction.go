package api

import (
	"context"
	"passport"
	"passport/db"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

type Transaction struct {
	From                 passport.UserID `json:"from"`
	To                   passport.UserID `json:"to"`
	Amount               int64           `json:"amount"`
	TransactionReference string          `json:"transactionReference"`
}

// HandleTransactions listens to the handle transaction channel and processes transactions, this is to ensure they happen asynchronously
func (api *API) HandleTransactions() {
	ctx := context.Background()
	conn, err := api.Conn.Acquire(ctx)
	if err != nil {
		api.Log.Panic().Err(err).Msg("unable to acquire connection for transactions")
	}

	for {
		select {
		case transaction := <-api.transaction:

			// before we begin the db tx for the transaction, log that we are attempting a transaction (log success defaults to false, we set it to true if it succeeds)
			logID := uuid.UUID{}
			q := `INSERT INTO xsyn_transaction_log (from_id, to_id,  amount, transaction_reference) ($1, $2, $3, $4)`

			err := pgxscan.Get(ctx, api.Conn, &logID, q, transaction.From, transaction.To, transaction.Amount, transaction.TransactionReference) // we can use any pgx connection for this, so we just use the pool
			if err != nil {
				api.Log.Panic().Err(err).Msg("failed to log transaction")
				continue
			}

			err = func() error {
				// begin the db tx for the transaction
				tx, err := conn.Begin(ctx)
				if err != nil {
					return terror.Error(err, "failed to start transaction transaction")
				}
				defer tx.Rollback(ctx)

				fromUser, err := db.UserGet(ctx, tx, transaction.From)
				if err != nil {
					return terror.Error(err, "failed to get from user from ID")
				}

				q = `UPDATE users SET sups = sups - $2 WHERE id = $1`

				_, err = tx.Exec(ctx, q, fromUser.ID, transaction.Amount)
				if err != nil {
					return terror.Error(err, "failed to get user from ID")
				}

				toUser, err := db.UserGet(ctx, tx, transaction.To)
				if err != nil {
					return terror.Error(err, "failed to get to user from ID")
				}

				q = `UPDATE users SET sups = sups + $2 WHERE id = $1`

				_, err = tx.Exec(ctx, q, toUser.ID, transaction.Amount)
				if err != nil {
					return terror.Error(err, "failed to get user from ID")
				}

				err = tx.Commit(ctx)
				if err != nil {
					return terror.Error(err, "failed to commit transaction")
				}
				return nil
			}()

			if err != nil {
				api.Log.Panic().Err(err)
				continue
			}
			// update the transaction log to be successful
			q = `UPDATE xsyn_transaction_log SET success = true	WHERE id = $1`

			_, err = api.Conn.Exec(ctx, q, logID)
			if err != nil {
				api.Log.Panic().Err(err).Msg("failed to log transaction")
			}
		}
	}
}
