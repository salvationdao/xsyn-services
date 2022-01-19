package db

import (
	"context"
	"fmt"
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

// TransactionGetList returns list of transactions from a list of transaction references
func TransactionGetList(ctx context.Context, conn Conn, transactionList []string) ([]*passport.Transaction, error) {
	var transactions []*passport.Transaction
	args := []interface{}{}

	whereCondition := ""
	for i, s := range transactionList {
		args = append(args, s)
		if i == 0 {
			whereCondition = fmt.Sprintf("$%d", i+1)
			continue
		}
		whereCondition = fmt.Sprintf("%s, $%d", whereCondition, i+1)
	}

	q := fmt.Sprintf(`--sql
		SELECT *
		FROM transactions
		WHERE transaction_reference IN (%s)`, whereCondition)
	err := pgxscan.Select(ctx, conn, &transactions, q, args...)
	if err != nil {
		return nil, terror.Error(err)
	}
	return transactions, nil
}

// UserBalance gets a users balance from the materialized view
func UserBalance(ctx context.Context, conn Conn, userID passport.UserID) (*passport.BigInt, error) {
	var wrap struct {
		Sups passport.BigInt `db:"sups"`
	}
	q := `SELECT sups FROM users WHERE id = $1`

	err := pgxscan.Get(ctx, conn, &wrap, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return &wrap.Sups, nil
}
