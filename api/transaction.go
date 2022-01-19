package api

import (
	"context"
	"database/sql"
	"math/big"
	"passport"
	"passport/db"

	"github.com/ninja-software/terror/v2"
)

type NewTransaction struct {
	To                   passport.UserID `json:"credit" db:"credit"`
	From                 passport.UserID `json:"debit" db:"debit"`
	Amount               big.Int         `json:"amount" db:"amount"`
	TransactionReference string          `json:"transactionReference" db:"transaction_reference"`
	Description          string          `json:"description" db:"description"`
}

// HandleTransactions listens to the handle transaction channel and processes transactions, this is to ensure they happen asynchronously
func (api *API) HandleTransactions() {
	ctx := context.Background()

	for {
		transaction := <-api.transaction

		err := CreateTransactionEntry(
			api.TxConn,
			transaction.Amount,
			transaction.To,
			transaction.From,
			"",
			transaction.TransactionReference,
		)

		if err != nil {
			api.Log.Err(err).Msg("failed to successfully transfer sups")
			continue
		}
		if transaction.To != passport.SupremacyGameUserID {
			balanceTo, err := db.UserBalance(ctx, api.Conn, transaction.To)
			if err != nil {
				api.Log.Fatal().Err(err).Msg("failed to get updated user")
			}
			api.SendToAllServerClient(&ServerClientMessage{
				Key: UserSupsUpdated,
				Payload: struct {
					UserID passport.UserID `json:"userID"`
					Sups   passport.BigInt `json:"sups"`
				}{
					UserID: transaction.To,
					Sups:   *balanceTo,
				},
			})
			// TODO: add sup subscription on passport
			//api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, toUser.ID)), toUser)
		}

		if transaction.From != passport.SupremacyGameUserID {
			balanceFrom, err := db.UserBalance(ctx, api.Conn, transaction.From)
			if err != nil {
				api.Log.Fatal().Err(err).Msg("failed to get updated user")
			}
			api.SendToAllServerClient(&ServerClientMessage{
				Key: UserSupsUpdated,
				Payload: struct {
					UserID passport.UserID `json:"userID"`
					Sups   passport.BigInt `json:"sups"`
				}{
					UserID: transaction.From,
					Sups:   *balanceFrom,
				},
			})
			// TODO: add sup subscription on passport
			//api.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, fromUser.ID)), fromUser)
		}
	}
}

// CreateTransactionEntry adds an entry to the transaction entry table
func CreateTransactionEntry(conn *sql.DB, amount big.Int, to, from passport.UserID, description, txRef string) error {
	q := `INSERT INTO transactions(description, transaction_reference, amount, credit, debit)
				VALUES($1, $2, $3, $4, $5);`

	_, err := conn.Exec(q, description, txRef, amount.String(), to, from)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
