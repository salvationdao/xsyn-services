package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

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
	userID string,
) (map[string][]string, error) {
	// Get all transactions with group IDs
	q := `--sql
		SELECT transactions.group, transactions.sub_group
		from transactions
		WHERE transactions.group is not null
		AND (transactions.credit = $1 OR transactions.debit = $1)
	`
	var args []interface{}
	args = append(args, userID)

	rows := make([]struct {
		Group    string
		SubGroup string
	}, 0)
	r, err := passdb.StdConn.Query(q, args...)
	if err != nil {
		return nil, err
	}

	for r.Next() {
		tg := struct {
			Group    string
			SubGroup string
		}{}

		err = r.Scan(
			&tg.Group,
			&tg.SubGroup,
		)
		if err != nil {
			return nil, err
		}

		rows = append(rows, tg)
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
	userID *string, // if user id is provided, returns transactions that only matter to this user
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
			column := TransactionColumn(f.Column)
			err := column.IsValid()
			if err != nil {
				return 0, nil, err
			}

			condition, value := GenerateListFilterSQL(f.Column, f.Value, f.Operator, argIndex)
			if condition != "" {
				switch f.Operator {
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
		args = append(args, *userID)
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
	err := passdb.StdConn.QueryRow(countQ, args...).Scan(&totalRows)
	if err != nil {
		return 0, nil, err
	}
	if totalRows == 0 {
		return 0, make([]*types.Transaction, 0), nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, nil, err
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
	r, err := passdb.StdConn.Query(q, args...)
	if err != nil {
		return 0, nil, err
	}
	for r.Next() {
		ts := &types.Transaction{}

		err = r.Scan(
			&ts.To,
			&ts.From,
			&ts.ID,
			&ts.Description,
			&ts.TransactionReference,
			&ts.Amount,
			&ts.Credit,
			&ts.Debit,
			&ts.Reason,
			&ts.ServiceID,
			&ts.RelatedTransactionID,
			&ts.CreatedAt,
			&ts.Group,
			&ts.SubGroup,
		)
		if err != nil {
			return 0, nil, err
		}

		result = append(result, ts)
	}

	//fmt.Println(totalRows)
	//fmt.Println(len(result))

	return totalRows, result, nil
}

// TransactionGet get store item by id
func TransactionGet(transactionID string) (*boiler.Transaction, error) {
	transaction, err := boiler.Transactions(
		boiler.TransactionWhere.ID.EQ(transactionID),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// TransactionAddRelatedTransaction adds a refund transaction ID to a transaction
func TransactionAddRelatedTransaction(transactionID string, refundTransactionID string) error {
	_, err := boiler.Transactions(
		boiler.TransactionWhere.ID.EQ(transactionID),
	).UpdateAll(passdb.StdConn, boiler.M{"related_transaction_id": refundTransactionID})
	if err != nil {
		return err
	}

	return nil
}

func TransactionExists(txhash string) (bool, error) {
	tx, err := boiler.Transactions(
		boiler.TransactionWhere.TransactionReference.EQ(txhash),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, terror.Error(err)
	}

	return tx != nil, nil
}

func UserBalances() ([]*types.UserBalance, error) {
	q := `SELECT id, sups FROM users`

	rows, err := passdb.StdConn.Query(q)
	if err != nil {
		return nil, err
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
			return balances, err
		}
		balances = append(balances, balance)
	}

	return balances, nil
}

func UserBalance(userID string) (decimal.Decimal, error) {
	var sups decimal.Decimal
	q := `SELECT sups FROM users WHERE id = $1`
	row := passdb.StdConn.QueryRow(q, userID)

	err := row.Scan(&sups)

	if err != nil {
		return decimal.New(0, 18), err
	}
	return sups, nil
}
