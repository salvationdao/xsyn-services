package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"xsyn-services/types"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

type TransactionColumn string

const (
	TransactionColumnID                   TransactionColumn = "id"
	TransactionColumnDescription          TransactionColumn = "description"
	TransactionColumnTransactionReference TransactionColumn = "transaction_reference"
	TransactionColumnAmount               TransactionColumn = "amount"
	TransactionColumnCredit               TransactionColumn = "credit"
	TransactionColumnDebit                TransactionColumn = "debit"
	TransactionColumnStatus               TransactionColumn = "status"
	TransactionColumnReason               TransactionColumn = "reason"
	TransactionColumnCreatedAt            TransactionColumn = "created_at"
	TransactionColumnGroup                TransactionColumn = "group"
	TransactionColumnSubGroup             TransactionColumn = "sub_group"
)

func (ic TransactionColumn) IsValid() error {
	switch ic {
	case TransactionColumnID,
		TransactionColumnDescription,
		TransactionColumnTransactionReference,
		TransactionColumnAmount,
		TransactionColumnCredit,
		TransactionColumnDebit,
		TransactionColumnStatus,
		TransactionColumnReason,
		TransactionColumnCreatedAt,
		TransactionColumnGroup,
		TransactionColumnSubGroup:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid transaction column type"))
}

const TransactionGetQuery string = `
SELECT 
row_to_json(t) as to,
row_to_json(f) as from,
transactions.id,
transactions.description,
transactions.transaction_reference,
transactions.amount,
transactions.credit,
transactions.debit,
transactions.status,
transactions.reason,
transactions.service_id,
transactions.related_transaction_id,
transactions.created_at,
transactions.group,
transactions.sub_group
` + TransactionGetQueryFrom

const TransactionGetQueryFrom = `
FROM transactions 
INNER JOIN users t ON transactions.credit = t.id
INNER JOIN users f ON transactions.debit = f.id
`

// UsersTransactionGroups returns details about the user's transactions that have group IDs
func UsersTransactionGroups(
	userID types.UserID,
	ctx context.Context,
	conn Conn,
) (map[string][]string, error) {
	// Get all transactions with group IDs
	q := `--sql
		SELECT transactions.group, transactions.sub_group
		from transactions
		WHERE transactions.group is not null
		AND (transactions.credit = $1 OR transactions.debit = $1)
	`
	var args []interface{}
	args = append(args, userID.String())

	rows := make([]struct {
		Group    string
		SubGroup string
	}, 0)
	err := pgxscan.Select(ctx, conn, &rows, q, args...)
	if err != nil {
		return nil, terror.Error(err)
	}

	m := make(map[string]map[string]struct{})
	for _, g := range rows {
		set := m[g.Group]
		if set != nil {
			set[g.SubGroup] = struct{}{}
		} else {
			set = map[string]struct{}{g.SubGroup: {}}
		}
		m[g.Group] = set
	}

	result := make(map[string][]string, 0)
	for key, e := range m {
		result[key] = make([]string, 0)
		for s := range e {
			result[key] = append(result[key], s)
		}
	}

	return result, nil
}

// TransactionList gets a list of Transactions depending on the filters
func TransactionList(
	ctx context.Context,
	conn Conn,
	userID *types.UserID, // if user id is provided, returns transactions that only matter to this user
	search string,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy TransactionColumn,
	sortDir SortByDir,
) (int, []*types.Transaction, error) {
	var args []interface{}

	// Prepare Filters
	filterConditionsString := ""
	argIndex := 1
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := TransactionColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			condition, value := GenerateListFilterSQL(f.ColumnField, f.Value, f.OperatorValue, argIndex)
			if condition != "" {
				switch f.OperatorValue {
				case OperatorValueTypeIsNull, OperatorValueTypeIsNotNull:
					break
				default:
					argIndex += 1
					args = append(args, value)
				}
				filterConditions = append(filterConditions, condition)
			}
		}
		if len(filterConditions) > 0 {
			filterConditionsString = " AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	if userID != nil {
		args = append(args, userID)
		filterConditionsString += fmt.Sprintf(" AND (credit = $%[1]d OR debit = $%[1]d) ", len(args))
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			searchCondition = fmt.Sprintf(" AND ((to_tsvector('english', transactions.description) @@ to_tsquery($%[1]d)) OR (to_tsvector('english', transactions.transaction_reference) @@ to_tsquery($%[1]d)))", len(args))
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT transactions.id)
		%s
		WHERE transactions.id IS NOT NULL
			%s
			%s
		`,
		TransactionGetQueryFrom,
		filterConditionsString,
		searchCondition,
	)

	var totalRows int

	err := pgxscan.Get(ctx, conn, &totalRows, countQ, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if totalRows == 0 {
		return 0, make([]*types.Transaction, 0), nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(
		TransactionGetQuery+`--sql
		WHERE transactions.id IS NOT NULL
			%s
			%s
		%s
		%s`,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := make([]*types.Transaction, 0)
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	return totalRows, result, nil
}

// TransactionGet get store item by id
func TransactionGet(ctx context.Context, conn Conn, transactionID string) (*types.Transaction, error) {
	transaction := &types.Transaction{}
	q := TransactionGetQuery + "WHERE transactions.id = $1"

	err := pgxscan.Get(ctx, conn, transaction, q, transactionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return transaction, nil
}

// TransactionAddRelatedTransaction adds a refund transaction ID to a transaction
func TransactionAddRelatedTransaction(ctx context.Context, conn Conn, transactionID string, refundTransactionID string) error {
	q := "UPDATE transactions SET related_transaction_id = $2 WHERE id = $1"

	_, err := conn.Exec(ctx, q, transactionID, refundTransactionID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func TransactionExists(ctx context.Context, conn Conn, txhash string) (bool, error) {
	q := "SELECT count(id) FROM transactions WHERE transaction_reference = $1"
	var count int
	err := pgxscan.Get(ctx, conn, &count, q, txhash)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// TransactionGetList returns list of transactions from a list of transaction references
func TransactionGetList(ctx context.Context, conn Conn, transactionList []string) ([]*types.Transaction, error) {
	var transactions []*types.Transaction
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

func UserBalances(ctx context.Context, conn Conn) ([]*types.UserBalance, error) {
	q := `SELECT id, sups FROM users`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, terror.Error(err)
	}

	balances := []*types.UserBalance{}

	for rows.Next() {
		balance := &types.UserBalance{
			ID:   types.UserID{},
			Sups: decimal.New(0, 18),
		}
		err := rows.Scan(
			&balance.ID,
			&balance.Sups,
		)
		if err != nil {
			return balances, terror.Error(err)
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

func UserBalance(ctx context.Context, conn Conn, userID string) (decimal.Decimal, error) {
	var wrap struct {
		Sups decimal.Decimal `db:"sups"`
	}
	q := `SELECT sups FROM users WHERE id = $1`

	err := pgxscan.Get(ctx, conn, &wrap, q, userID)
	if err != nil {
		return decimal.New(0, 18), terror.Error(err)
	}
	return wrap.Sups, nil
}
