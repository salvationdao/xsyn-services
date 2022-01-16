package api

import (
	"context"
	"errors"
	"passport"
	"passport/db"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
)

type Transaction struct {
	From                 passport.UserID `json:"from"`
	To                   passport.UserID `json:"to"`
	Amount               passport.BigInt `json:"amount"`
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
		transaction := <-api.transaction

		// before we begin the db tx for the transaction, log that we are attempting a transaction (log success defaults to false, we set it to true if it succeeds)
		logID := uuid.UUID{}
		q := `INSERT INTO xsyn_transaction_log (from_id, to_id,  amount, transaction_reference) VALUES($1, $2, $3, $4) RETURNING id`

		err := pgxscan.Get(ctx, api.Conn, &logID, q, transaction.From, transaction.To, transaction.Amount.Int.String(), transaction.TransactionReference) // we can use any pgx connection for this, so we just use the pool
		if err != nil {
			api.Log.Fatal().Err(err).Msg("failed to log transaction")
			continue
		}

		err = func() error {
			// begin the db tx for the transaction
			tx, err := conn.Begin(ctx)
			if err != nil {
				return terror.Error(err)
			}
			defer func(tx pgx.Tx, ctx context.Context) {
				err := tx.Rollback(ctx)
				if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
					api.Log.Err(err).Msg("error rolling back")
				}
			}(tx, ctx)

			fromUser, err := db.UserGet(ctx, tx, transaction.From)
			if err != nil {
				return terror.Error(err)
			}

			q = `UPDATE users SET sups = sups - $2 WHERE id = $1`

			_, err = tx.Exec(ctx, q, fromUser.ID, transaction.Amount.Int.String())
			if err != nil {
				return terror.Error(err)
			}

			q = `UPDATE users SET sups = sups + $2 WHERE id = $1`

			_, err = tx.Exec(ctx, q, transaction.To, transaction.Amount.Int.String())
			if err != nil {
				return terror.Error(err)
			}

			err = tx.Commit(ctx)
			if err != nil {
				return terror.Error(err)
			}
			return nil
		}()

		if err != nil {
			api.Log.Err(err).Msg("failed to successfully transfer sups")
			continue
		}
		// update the transaction log to be successful
		q = `UPDATE xsyn_transaction_log SET status = 'success'	WHERE id = $1`

		_, err = api.Conn.Exec(ctx, q, logID)
		if err != nil {
			api.Log.Fatal().Err(err).Msg("failed to log transaction")
		}

		toUser, err := db.UserGet(ctx, api.Conn, transaction.To)
		if err != nil {
			api.Log.Fatal().Err(err).Msg("failed to get updated user")
		}
		if toUser.Username != passport.SupremacyGameUsername {
			api.SendToAllServerClient(&ServerClientMessage{
				Key: UserSupsUpdated,
				Payload: struct {
					UserID passport.UserID `json:"userID"`
					Sups   passport.BigInt `json:"sups"`
				}{
					UserID: toUser.ID,
					Sups:   toUser.Sups,
				},
			})
			// TODO: add sup subscription on passport
			//api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, toUser.ID)), toUser)
		}

		fromUser, err := db.UserGet(ctx, api.Conn, transaction.From)
		if err != nil {
			api.Log.Fatal().Err(err).Msg("failed to get updated user")
		}
		if fromUser.Username != passport.SupremacyGameUsername {
			api.SendToAllServerClient(&ServerClientMessage{
				Key: UserSupsUpdated,
				Payload: struct {
					UserID passport.UserID `json:"userID"`
					Sups   passport.BigInt `json:"sups"`
				}{
					UserID: fromUser.ID,
					Sups:   fromUser.Sups,
				},
			})
			// TODO: add sup subscription on passport
			//api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, fromUser.ID)), fromUser)
		}
	}
}
