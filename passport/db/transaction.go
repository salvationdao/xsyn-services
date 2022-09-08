package db

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type TransactionDetailed struct {
	boiler.Transaction `boil:",bind"`
	To                 string `json:"to"`
	From               string `json:"from"`
}

func IsValidColumn(column string, columnStruct interface{}) bool {
	v := reflect.ValueOf(columnStruct)
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).String() == column {
			return true
		}
	}

	return false
}

func ColumnsToString(columnStruct interface{}) string {
	v := reflect.ValueOf(columnStruct)
	result := ""
	for i := 0; i < v.NumField(); i++ {
		result += v.Field(i).String()
		if i == v.NumField()-1 {
			break
		}
		result += ",\n"
	}
	return result
}

var TransactionGetQuery = fmt.Sprintf(`
SELECT 
%s,
t.%s as to,
f.%s as from
`,
	ColumnsToString(boiler.TransactionTableColumns),
	boiler.UserColumns.Username,
	boiler.UserColumns.Username,
) + TransactionGetQueryFrom

var TransactionGetQueryFrom = fmt.Sprintf(`
FROM %s 
INNER JOIN %s t ON %s = t.%s
INNER JOIN %s f ON %s = f.%s
`,
	boiler.TableNames.Transactions,
	boiler.TableNames.Users,
	boiler.TransactionTableColumns.Credit,
	boiler.UserColumns.ID,
	boiler.TableNames.Users,
	boiler.TransactionTableColumns.Debit,
	boiler.UserColumns.ID,
)

// UsersTransactionGroups returns details about the user's transactions that have group IDs
func UsersTransactionGroups(
	userID string,
) (map[string][]string, error) {
	// Get all transactions with group IDs
	q := fmt.Sprintf(`--sql
		SELECT %s, %s
		from %s
		WHERE %s is not null
		AND (%s = $1 OR %s = $1)
	`,
		boiler.TransactionTableColumns.Group,
		boiler.TransactionTableColumns.SubGroup,
		boiler.TableNames.Transactions,
		boiler.TransactionTableColumns.Group,
		boiler.TransactionTableColumns.Credit,
		boiler.TransactionTableColumns.Debit,
	)
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
	sortBy string,
	sortDir SortByDir,
) (int, []*TransactionDetailed, error) {
	var args []interface{}

	// Prepare Filters
	filterConditionsString := ""
	filterConditions := []string{}
	argIndex := 1
	if filter != nil {
		for _, f := range filter.Items {
			valid := IsValidColumn(f.Column, boiler.TransactionColumns)
			if !valid {
				return 0, nil, fmt.Errorf("invalid transaction column")
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
			filterConditionsString = strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	if userID != nil {
		args = append(args, *userID)
		if len(filterConditions) > 0 {
			filterConditionsString += " AND "
		}
		filterConditionsString += fmt.Sprintf("(%[2]s = $%[1]d OR %[3]s = $%[1]d) ", len(args), boiler.TransactionColumns.Credit, boiler.TransactionColumns.Debit)
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			if len(filterConditions) > 0 {
				filterConditionsString += " AND "
			}
			searchCondition = fmt.Sprintf("((to_tsvector('english', %[2]s) @@ to_tsquery($%[1]d)) OR (to_tsvector('english', %[3]s) @@ to_tsquery($%[1]d)))",
				len(args),
				boiler.TransactionTableColumns.Description,
				boiler.TransactionTableColumns.TransactionReference)
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(%s)
		%s
		WHERE %s
		%s
		`,
		boiler.TransactionTableColumns.ID,
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
		return 0, []*TransactionDetailed{}, nil
	}

	// Order and Limit
	orderBy := fmt.Sprintf(" ORDER BY %s desc", boiler.TransactionTableColumns.CreatedAt)
	if sortBy != "" {
		valid := IsValidColumn(sortBy, boiler.TransactionColumns)
		if !valid {
			return 0, nil, fmt.Errorf("invalid transaction column")
		}
		orderBy = fmt.Sprintf(" ORDER BY transactions.%s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(`--sql
		%s
		WHERE %s
		%s
		%s
		%s`,
		TransactionGetQuery,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)
	scanned := []*TransactionDetailed{}
	err = boiler.NewQuery(
		qm.SQL(q, args...),
	).Bind(nil, passdb.StdConn, &scanned)
	if err != nil {
		return 0, nil, err
	}

	return totalRows, scanned, nil
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
	).UpdateAll(passdb.StdConn, boiler.M{boiler.TransactionColumns.RelatedTransactionID: refundTransactionID})
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
