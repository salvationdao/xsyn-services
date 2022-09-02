package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"

	"github.com/ninja-software/terror/v2"
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

// TransactionIDList
func TransactionIDList(
	userID *string, // if user id is provided, returns transactions that only matter to this user
	search string,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy TransactionColumn,
	sortDir SortByDir,
) (int, []string, error) {
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
	err := passdb.StdConn.QueryRow(countQ, args...).Scan(&totalRows)
	if err != nil {
		return 0, nil, err
	}
	if totalRows == 0 {
		return 0, []string{}, nil
	}

	// Order and Limit
	orderBy := " ORDER BY transactions.created_at desc"
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, nil, err
		}
		orderBy = fmt.Sprintf(" ORDER BY transactions.%s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(`--sql
		SELECT 
		transactions.id
		%s
		WHERE transactions.id IS NOT NULL
		%s
		%s
		%s
		%s`,
		TransactionGetQueryFrom,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := make([]string, 0)
	r, err := passdb.StdConn.Query(q, args...)
	if err != nil {
		return 0, nil, err
	}
	for r.Next() {
		txid := ""

		err = r.Scan(
			&txid,
		)
		if err != nil {
			return 0, nil, err
		}

		result = append(result, txid)
	}

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

func UserBalance(userID string) (*boiler.User, error) {
	user, err := boiler.Users(
		boiler.UserWhere.ID.EQ(userID),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	return user, nil
}
