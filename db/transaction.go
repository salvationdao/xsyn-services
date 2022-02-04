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

// ChainConfirmationsForUserID returns a list of chain confirmations by userID
func ChainConfirmationsForUserID(ctx context.Context, conn Conn, userID *passport.UserID) ([]*passport.ChainConfirmations, error) {
	var confirmations []*passport.ChainConfirmations

	q := `SELECT tx,
				 tx_id,
				 block,
				 confirmed_at,
				 deleted_at,
				 created_at 
		FROM chain_confirmations cc
		INNER JOIN transactions txs ON txs.id = cc.tx_id
		INNER JOIN users u ON u.id = tsx.credit OR u.id = tsx.debit
		WHERE confirmed_at IS NULL AND deleted_at IS NULL`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, terror.Error(err)
	}
	for rows.Next() {
		var confirmation *passport.ChainConfirmations
		err := rows.Scan(
			&confirmation.Tx,
			&confirmation.TxID,
			&confirmation.Block,
			&confirmation.ConfirmedAt,
			&confirmation.CreatedAt,
		)
		if err != nil {
			return nil, terror.Error(err)
		}
		confirmations = append(confirmations, confirmation)
	}

	return confirmations, nil
}

// PendingChainConfirmationsByChainID gets all chain confirmations that are not confirmed or deleted on the given chain ID
func PendingChainConfirmationsByChainID(ctx context.Context, conn Conn, chainID uint64) ([]*passport.ChainConfirmations, error) {
	var confirmations []*passport.ChainConfirmations

	q := `SELECT tx,
				 tx_id,
				 block,
				 confirmed_at,
				 chain_id,
				 created_at 
			FROM chain_confirmations WHERE chain_id = $1 AND confirmed_at IS NULL AND deleted_at IS NULL`

	rows, err := conn.Query(ctx, q, chainID)
	if err != nil {
		return nil, terror.Error(err)
	}
	for rows.Next() {
		confirmation := &passport.ChainConfirmations{}
		err := rows.Scan(
			&confirmation.Tx,
			&confirmation.TxID,
			&confirmation.Block,
			&confirmation.ConfirmedAt,
			&confirmation.ChainID,
			&confirmation.CreatedAt,
		)
		if err != nil {
			return nil, terror.Error(err)
		}
		confirmations = append(confirmations, confirmation)
	}

	return confirmations, nil
}

// PendingChainConfirmations gets all chain confirmations that are not confirmed or deleted
func PendingChainConfirmations(ctx context.Context, conn Conn) ([]*passport.ChainConfirmations, error) {
	var confirmations []*passport.ChainConfirmations

	q := `SELECT tx,
				 tx_id,
				 block,
				 confirmed_at,
				 created_at 
			FROM chain_confirmations WHERE confirmed_at IS NULL AND deleted_at IS NULL`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, terror.Error(err)
	}
	for rows.Next() {
		confirmation := &passport.ChainConfirmations{}
		err := rows.Scan(
			&confirmation.Tx,
			&confirmation.TxID,
			&confirmation.Block,
			&confirmation.ConfirmedAt,
			&confirmation.CreatedAt,
		)
		if err != nil {
			return nil, terror.Error(err)
		}
		confirmations = append(confirmations, confirmation)
	}

	return confirmations, nil
}

// ConfirmChainConfirmation sets a chain confirmation to confirmed
func ConfirmChainConfirmation(ctx context.Context, conn Conn, tx string) (*passport.ChainConfirmations, error) {
	confirmation := &passport.ChainConfirmations{}

	q := `UPDATE chain_confirmations
			SET confirmed_at = NOW()
			WHERE tx = $1
			RETURNING	tx,
					    tx_id,
					    block,
					    confirmed_at,
						chain_id,
					    created_at`

	err := conn.QueryRow(ctx, q, tx).Scan(
		&confirmation.Tx,
		&confirmation.TxID,
		&confirmation.Block,
		&confirmation.ConfirmedAt,
		&confirmation.ChainID,
		&confirmation.CreatedAt,
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return confirmation, nil
}

// CreateChainConfirmationEntry creates a chain confirmation record
func CreateChainConfirmationEntry(ctx context.Context, conn Conn, tx string, txRef int64, block uint64, chainID uint64) error {
	q := `INSERT INTO chain_confirmations (tx, tx_id, block, chain_id)
			VALUES($1, $2, $3, $4)`

	_, err := conn.Exec(ctx, q, tx, txRef, block, chainID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
