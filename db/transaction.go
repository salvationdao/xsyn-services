package db

import (
	"context"
	"fmt"
	"passport"
	"time"

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

func UserBalances(ctx context.Context, conn Conn) ([]*passport.UserBalance, error) {
	q := `SELECT id, sups FROM users`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	balances := []*passport.UserBalance{}

	for rows.Next() {
		balance := &passport.UserBalance{
			ID:   passport.UserID{},
			Sups: &passport.BigInt{},
		}
		err := rows.Scan(
			&balance.ID,
			balance.Sups,
		)
		if err != nil {
			return balances, terror.Error(err)
		}
		balance.Sups.Init()
		balances = append(balances, balance)
	}

	return balances, nil
}

func UserBalance(ctx context.Context, conn Conn, userID string) (*passport.BigInt, error) {
	var wrap struct {
		Sups passport.BigInt `db:"sups"`
	}
	q := `SELECT sups FROM users WHERE id = $1`

	err := pgxscan.Get(ctx, conn, &wrap, q, userID)
	if err != nil {
		return nil, terror.Error(err)
	}
	wrap.Sups.Init()
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
func PendingChainConfirmationsByChainID(ctx context.Context, conn Conn, chainID int64) ([]*passport.ChainConfirmations, error) {
	var confirmations []*passport.ChainConfirmations

	q := `SELECT cc.tx,
				 cc.tx_id,
				 cc.block,
				 cc.confirmed_at,
				 cc.chain_id,
				 cc.created_at,
				 t.credit as user_id
			FROM chain_confirmations cc 
			INNER JOIN transactions t ON t.id = tx_id
			WHERE chain_id = $1 AND confirmed_at IS NULL AND deleted_at IS NULL`

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
			&confirmation.UserID,
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
					    created_at,
						confirmation_amount,
						(SELECT credit FROM transactions WHERE id = tx_id) as user_id`

	err := conn.QueryRow(ctx, q, tx).Scan(
		&confirmation.Tx,
		&confirmation.TxID,
		&confirmation.Block,
		&confirmation.ConfirmedAt,
		&confirmation.ChainID,
		&confirmation.CreatedAt,
		&confirmation.ConfirmationAmount,
		&confirmation.UserID,
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return confirmation, nil
}

func UpdateConfirmationAmount(ctx context.Context, conn Conn, tx string, confirmedBlocks int) (*passport.ChainConfirmations, error) {
	confirmation := &passport.ChainConfirmations{}

	q := `UPDATE chain_confirmations
		SET confirmation_amount = $1
		WHERE tx = $2
		RETURNING	tx,
					tx_id,
					block,
					confirmed_at,
					chain_id,
					created_at,
					confirmation_amount,
					(SELECT credit FROM transactions WHERE id = tx_id) as user_id`

	err := conn.QueryRow(ctx, q, confirmedBlocks, tx).Scan(
		&confirmation.Tx,
		&confirmation.TxID,
		&confirmation.Block,
		&confirmation.ConfirmedAt,
		&confirmation.ChainID,
		&confirmation.CreatedAt,
		&confirmation.ConfirmationAmount,
		&confirmation.UserID,
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return confirmation, nil
}

// CreateChainConfirmationEntry creates a chain confirmation record
func CreateChainConfirmationEntry(ctx context.Context, conn Conn, tx string, txRef string, block uint64, chainID int64) (*passport.ChainConfirmations, error) {
	conf := &passport.ChainConfirmations{}
	//
	q := `INSERT INTO chain_confirmations (tx, tx_id, block, chain_id)
			VALUES($1, $2, $3, $4)
			RETURNING 	tx,
						tx_id,
						block,
						chain_id,
						confirmed_at,
						confirmation_amount`

	err := pgxscan.Get(ctx, conn, conf, q,
		tx, txRef, block, chainID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return conf, nil
}

func BattleArenaSupsTopContributors(ctx context.Context, conn Conn, startTime, endTime time.Time) ([]*passport.User, error) {
	uss := []*passport.User{}
	q := `
		SELECT 
			u.id,
			u.username, 
			u.avatar_id,
			u.faction_id 
		FROM 
			transactions t
		INNER JOIN 
			users u ON u.id = t.debit AND u.faction_id NOTNULL
		WHERE
			t.credit = $1 AND t.created_at >= $2 AND t.created_at <= $3
		GROUP BY 
			u.id, t.debit
		ORDER BY 
			SUM(t.amount) DESC
		LIMIT 5
	`
	err := pgxscan.Select(ctx, conn, &uss, q, passport.SupremacyBattleUserID, startTime, endTime)
	if err != nil {
		return nil, terror.Error(err)
	}

	return uss, nil
}

func BattleArenaSupsTopContributeFaction(ctx context.Context, conn Conn, startTime, endTime time.Time) ([]*passport.Faction, error) {
	fss := []*passport.Faction{}
	q := `
		SELECT 
			f.id,
			f.label,
			f.logo_blob_id,
			f.theme
		FROM 
			transactions t
		INNER JOIN 
			users u ON u.id = t.debit AND u.faction_id NOTNULL
		INNER JOIN 
			factions f ON f.id = u.faction_id 
		WHERE
			t.credit = $1 AND t.created_at >= $2 AND t.created_at <= $3
		GROUP BY 
			f.id
		ORDER BY 
			SUM(t.amount) DESC
		LIMIT 3
	`
	err := pgxscan.Select(ctx, conn, &fss, q, passport.SupremacyBattleUserID, startTime, endTime)
	if err != nil {
		return nil, terror.Error(err)
	}

	return fss, nil
}
