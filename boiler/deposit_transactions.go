// Code generated by SQLBoiler 4.8.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package boiler

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// DepositTransaction is an object representing the database table.
type DepositTransaction struct {
	ID        string          `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	UserID    string          `boiler:"user_id" boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	TXHash    string          `boiler:"tx_hash" boil:"tx_hash" json:"tx_hash" toml:"tx_hash" yaml:"tx_hash"`
	Amount    decimal.Decimal `boiler:"amount" boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	Status    string          `boiler:"status" boil:"status" json:"status" toml:"status" yaml:"status"`
	DeletedAt null.Time       `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt time.Time       `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt time.Time       `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *depositTransactionR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L depositTransactionL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DepositTransactionColumns = struct {
	ID        string
	UserID    string
	TXHash    string
	Amount    string
	Status    string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "id",
	UserID:    "user_id",
	TXHash:    "tx_hash",
	Amount:    "amount",
	Status:    "status",
	DeletedAt: "deleted_at",
	UpdatedAt: "updated_at",
	CreatedAt: "created_at",
}

var DepositTransactionTableColumns = struct {
	ID        string
	UserID    string
	TXHash    string
	Amount    string
	Status    string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "deposit_transactions.id",
	UserID:    "deposit_transactions.user_id",
	TXHash:    "deposit_transactions.tx_hash",
	Amount:    "deposit_transactions.amount",
	Status:    "deposit_transactions.status",
	DeletedAt: "deposit_transactions.deleted_at",
	UpdatedAt: "deposit_transactions.updated_at",
	CreatedAt: "deposit_transactions.created_at",
}

// Generated where

type whereHelperdecimal_Decimal struct{ field string }

func (w whereHelperdecimal_Decimal) EQ(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelperdecimal_Decimal) NEQ(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelperdecimal_Decimal) LT(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelperdecimal_Decimal) LTE(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelperdecimal_Decimal) GT(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelperdecimal_Decimal) GTE(x decimal.Decimal) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var DepositTransactionWhere = struct {
	ID        whereHelperstring
	UserID    whereHelperstring
	TXHash    whereHelperstring
	Amount    whereHelperdecimal_Decimal
	Status    whereHelperstring
	DeletedAt whereHelpernull_Time
	UpdatedAt whereHelpertime_Time
	CreatedAt whereHelpertime_Time
}{
	ID:        whereHelperstring{field: "\"deposit_transactions\".\"id\""},
	UserID:    whereHelperstring{field: "\"deposit_transactions\".\"user_id\""},
	TXHash:    whereHelperstring{field: "\"deposit_transactions\".\"tx_hash\""},
	Amount:    whereHelperdecimal_Decimal{field: "\"deposit_transactions\".\"amount\""},
	Status:    whereHelperstring{field: "\"deposit_transactions\".\"status\""},
	DeletedAt: whereHelpernull_Time{field: "\"deposit_transactions\".\"deleted_at\""},
	UpdatedAt: whereHelpertime_Time{field: "\"deposit_transactions\".\"updated_at\""},
	CreatedAt: whereHelpertime_Time{field: "\"deposit_transactions\".\"created_at\""},
}

// DepositTransactionRels is where relationship names are stored.
var DepositTransactionRels = struct {
	User string
}{
	User: "User",
}

// depositTransactionR is where relationships are stored.
type depositTransactionR struct {
	User *User `boiler:"User" boil:"User" json:"User" toml:"User" yaml:"User"`
}

// NewStruct creates a new relationship struct
func (*depositTransactionR) NewStruct() *depositTransactionR {
	return &depositTransactionR{}
}

// depositTransactionL is where Load methods for each relationship are stored.
type depositTransactionL struct{}

var (
	depositTransactionAllColumns            = []string{"id", "user_id", "tx_hash", "amount", "status", "deleted_at", "updated_at", "created_at"}
	depositTransactionColumnsWithoutDefault = []string{"user_id", "tx_hash", "amount", "deleted_at"}
	depositTransactionColumnsWithDefault    = []string{"id", "status", "updated_at", "created_at"}
	depositTransactionPrimaryKeyColumns     = []string{"id"}
)

type (
	// DepositTransactionSlice is an alias for a slice of pointers to DepositTransaction.
	// This should almost always be used instead of []DepositTransaction.
	DepositTransactionSlice []*DepositTransaction
	// DepositTransactionHook is the signature for custom DepositTransaction hook methods
	DepositTransactionHook func(boil.Executor, *DepositTransaction) error

	depositTransactionQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	depositTransactionType                 = reflect.TypeOf(&DepositTransaction{})
	depositTransactionMapping              = queries.MakeStructMapping(depositTransactionType)
	depositTransactionPrimaryKeyMapping, _ = queries.BindMapping(depositTransactionType, depositTransactionMapping, depositTransactionPrimaryKeyColumns)
	depositTransactionInsertCacheMut       sync.RWMutex
	depositTransactionInsertCache          = make(map[string]insertCache)
	depositTransactionUpdateCacheMut       sync.RWMutex
	depositTransactionUpdateCache          = make(map[string]updateCache)
	depositTransactionUpsertCacheMut       sync.RWMutex
	depositTransactionUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var depositTransactionBeforeInsertHooks []DepositTransactionHook
var depositTransactionBeforeUpdateHooks []DepositTransactionHook
var depositTransactionBeforeDeleteHooks []DepositTransactionHook
var depositTransactionBeforeUpsertHooks []DepositTransactionHook

var depositTransactionAfterInsertHooks []DepositTransactionHook
var depositTransactionAfterSelectHooks []DepositTransactionHook
var depositTransactionAfterUpdateHooks []DepositTransactionHook
var depositTransactionAfterDeleteHooks []DepositTransactionHook
var depositTransactionAfterUpsertHooks []DepositTransactionHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DepositTransaction) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DepositTransaction) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DepositTransaction) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DepositTransaction) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DepositTransaction) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DepositTransaction) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DepositTransaction) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DepositTransaction) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DepositTransaction) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range depositTransactionAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDepositTransactionHook registers your hook function for all future operations.
func AddDepositTransactionHook(hookPoint boil.HookPoint, depositTransactionHook DepositTransactionHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		depositTransactionBeforeInsertHooks = append(depositTransactionBeforeInsertHooks, depositTransactionHook)
	case boil.BeforeUpdateHook:
		depositTransactionBeforeUpdateHooks = append(depositTransactionBeforeUpdateHooks, depositTransactionHook)
	case boil.BeforeDeleteHook:
		depositTransactionBeforeDeleteHooks = append(depositTransactionBeforeDeleteHooks, depositTransactionHook)
	case boil.BeforeUpsertHook:
		depositTransactionBeforeUpsertHooks = append(depositTransactionBeforeUpsertHooks, depositTransactionHook)
	case boil.AfterInsertHook:
		depositTransactionAfterInsertHooks = append(depositTransactionAfterInsertHooks, depositTransactionHook)
	case boil.AfterSelectHook:
		depositTransactionAfterSelectHooks = append(depositTransactionAfterSelectHooks, depositTransactionHook)
	case boil.AfterUpdateHook:
		depositTransactionAfterUpdateHooks = append(depositTransactionAfterUpdateHooks, depositTransactionHook)
	case boil.AfterDeleteHook:
		depositTransactionAfterDeleteHooks = append(depositTransactionAfterDeleteHooks, depositTransactionHook)
	case boil.AfterUpsertHook:
		depositTransactionAfterUpsertHooks = append(depositTransactionAfterUpsertHooks, depositTransactionHook)
	}
}

// One returns a single depositTransaction record from the query.
func (q depositTransactionQuery) One(exec boil.Executor) (*DepositTransaction, error) {
	o := &DepositTransaction{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for deposit_transactions")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DepositTransaction records from the query.
func (q depositTransactionQuery) All(exec boil.Executor) (DepositTransactionSlice, error) {
	var o []*DepositTransaction

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to DepositTransaction slice")
	}

	if len(depositTransactionAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DepositTransaction records in the query.
func (q depositTransactionQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count deposit_transactions rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q depositTransactionQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if deposit_transactions exists")
	}

	return count > 0, nil
}

// User pointed to by the foreign key.
func (o *DepositTransaction) User(mods ...qm.QueryMod) userQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.UserID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Users(queryMods...)
	queries.SetFrom(query.Query, "\"users\"")

	return query
}

// LoadUser allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (depositTransactionL) LoadUser(e boil.Executor, singular bool, maybeDepositTransaction interface{}, mods queries.Applicator) error {
	var slice []*DepositTransaction
	var object *DepositTransaction

	if singular {
		object = maybeDepositTransaction.(*DepositTransaction)
	} else {
		slice = *maybeDepositTransaction.(*[]*DepositTransaction)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &depositTransactionR{}
		}
		args = append(args, object.UserID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &depositTransactionR{}
			}

			for _, a := range args {
				if a == obj.UserID {
					continue Outer
				}
			}

			args = append(args, obj.UserID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`users`),
		qm.WhereIn(`users.id in ?`, args...),
		qmhelper.WhereIsNull(`users.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load User")
	}

	var resultSlice []*User
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice User")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for users")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for users")
	}

	if len(depositTransactionAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.User = foreign
		if foreign.R == nil {
			foreign.R = &userR{}
		}
		foreign.R.DepositTransactions = append(foreign.R.DepositTransactions, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserID == foreign.ID {
				local.R.User = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.DepositTransactions = append(foreign.R.DepositTransactions, local)
				break
			}
		}
	}

	return nil
}

// SetUser of the depositTransaction to the related item.
// Sets o.R.User to related.
// Adds o to related.R.DepositTransactions.
func (o *DepositTransaction) SetUser(exec boil.Executor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"deposit_transactions\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_id"}),
		strmangle.WhereClause("\"", "\"", 2, depositTransactionPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.UserID = related.ID
	if o.R == nil {
		o.R = &depositTransactionR{
			User: related,
		}
	} else {
		o.R.User = related
	}

	if related.R == nil {
		related.R = &userR{
			DepositTransactions: DepositTransactionSlice{o},
		}
	} else {
		related.R.DepositTransactions = append(related.R.DepositTransactions, o)
	}

	return nil
}

// DepositTransactions retrieves all the records using an executor.
func DepositTransactions(mods ...qm.QueryMod) depositTransactionQuery {
	mods = append(mods, qm.From("\"deposit_transactions\""), qmhelper.WhereIsNull("\"deposit_transactions\".\"deleted_at\""))
	return depositTransactionQuery{NewQuery(mods...)}
}

// FindDepositTransaction retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDepositTransaction(exec boil.Executor, iD string, selectCols ...string) (*DepositTransaction, error) {
	depositTransactionObj := &DepositTransaction{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"deposit_transactions\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, depositTransactionObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from deposit_transactions")
	}

	if err = depositTransactionObj.doAfterSelectHooks(exec); err != nil {
		return depositTransactionObj, err
	}

	return depositTransactionObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DepositTransaction) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no deposit_transactions provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.UpdatedAt.IsZero() {
		o.UpdatedAt = currTime
	}
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(depositTransactionColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	depositTransactionInsertCacheMut.RLock()
	cache, cached := depositTransactionInsertCache[key]
	depositTransactionInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			depositTransactionAllColumns,
			depositTransactionColumnsWithDefault,
			depositTransactionColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(depositTransactionType, depositTransactionMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(depositTransactionType, depositTransactionMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"deposit_transactions\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"deposit_transactions\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "boiler: unable to insert into deposit_transactions")
	}

	if !cached {
		depositTransactionInsertCacheMut.Lock()
		depositTransactionInsertCache[key] = cache
		depositTransactionInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the DepositTransaction.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DepositTransaction) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	depositTransactionUpdateCacheMut.RLock()
	cache, cached := depositTransactionUpdateCache[key]
	depositTransactionUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			depositTransactionAllColumns,
			depositTransactionPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update deposit_transactions, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"deposit_transactions\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, depositTransactionPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(depositTransactionType, depositTransactionMapping, append(wl, depositTransactionPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	var result sql.Result
	result, err = exec.Exec(cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update deposit_transactions row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for deposit_transactions")
	}

	if !cached {
		depositTransactionUpdateCacheMut.Lock()
		depositTransactionUpdateCache[key] = cache
		depositTransactionUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q depositTransactionQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for deposit_transactions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for deposit_transactions")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DepositTransactionSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("boiler: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), depositTransactionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"deposit_transactions\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, depositTransactionPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in depositTransaction slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all depositTransaction")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DepositTransaction) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no deposit_transactions provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(depositTransactionColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	depositTransactionUpsertCacheMut.RLock()
	cache, cached := depositTransactionUpsertCache[key]
	depositTransactionUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			depositTransactionAllColumns,
			depositTransactionColumnsWithDefault,
			depositTransactionColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			depositTransactionAllColumns,
			depositTransactionPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert deposit_transactions, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(depositTransactionPrimaryKeyColumns))
			copy(conflict, depositTransactionPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"deposit_transactions\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(depositTransactionType, depositTransactionMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(depositTransactionType, depositTransactionMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, cache.query)
		fmt.Fprintln(boil.DebugWriter, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRow(cache.query, vals...).Scan(returns...)
		if err == sql.ErrNoRows {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.Exec(cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "boiler: unable to upsert deposit_transactions")
	}

	if !cached {
		depositTransactionUpsertCacheMut.Lock()
		depositTransactionUpsertCache[key] = cache
		depositTransactionUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single DepositTransaction record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DepositTransaction) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no DepositTransaction provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), depositTransactionPrimaryKeyMapping)
		sql = "DELETE FROM \"deposit_transactions\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"deposit_transactions\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(depositTransactionType, depositTransactionMapping, append(wl, depositTransactionPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), valueMapping)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from deposit_transactions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for deposit_transactions")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q depositTransactionQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no depositTransactionQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from deposit_transactions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for deposit_transactions")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DepositTransactionSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(depositTransactionBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), depositTransactionPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"deposit_transactions\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, depositTransactionPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), depositTransactionPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"deposit_transactions\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, depositTransactionPrimaryKeyColumns, len(o)),
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		args = append([]interface{}{currTime}, args...)
	}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from depositTransaction slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for deposit_transactions")
	}

	if len(depositTransactionAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *DepositTransaction) Reload(exec boil.Executor) error {
	ret, err := FindDepositTransaction(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DepositTransactionSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DepositTransactionSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), depositTransactionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"deposit_transactions\".* FROM \"deposit_transactions\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, depositTransactionPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in DepositTransactionSlice")
	}

	*o = slice

	return nil
}

// DepositTransactionExists checks if the DepositTransaction row exists.
func DepositTransactionExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"deposit_transactions\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if deposit_transactions exists")
	}

	return exists, nil
}
