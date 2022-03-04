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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// DeathAddress is an object representing the database table.
type DeathAddress struct {
	WalletAddress string `boiler:"wallet_address" boil:"wallet_address" json:"walletAddress" toml:"walletAddress" yaml:"walletAddress"`

	R *deathAddressR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L deathAddressL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DeathAddressColumns = struct {
	WalletAddress string
}{
	WalletAddress: "wallet_address",
}

var DeathAddressTableColumns = struct {
	WalletAddress string
}{
	WalletAddress: "death_addresses.wallet_address",
}

// Generated where

var DeathAddressWhere = struct {
	WalletAddress whereHelperstring
}{
	WalletAddress: whereHelperstring{field: "\"death_addresses\".\"wallet_address\""},
}

// DeathAddressRels is where relationship names are stored.
var DeathAddressRels = struct {
}{}

// deathAddressR is where relationships are stored.
type deathAddressR struct {
}

// NewStruct creates a new relationship struct
func (*deathAddressR) NewStruct() *deathAddressR {
	return &deathAddressR{}
}

// deathAddressL is where Load methods for each relationship are stored.
type deathAddressL struct{}

var (
	deathAddressAllColumns            = []string{"wallet_address"}
	deathAddressColumnsWithoutDefault = []string{"wallet_address"}
	deathAddressColumnsWithDefault    = []string{}
	deathAddressPrimaryKeyColumns     = []string{"wallet_address"}
)

type (
	// DeathAddressSlice is an alias for a slice of pointers to DeathAddress.
	// This should almost always be used instead of []DeathAddress.
	DeathAddressSlice []*DeathAddress
	// DeathAddressHook is the signature for custom DeathAddress hook methods
	DeathAddressHook func(boil.Executor, *DeathAddress) error

	deathAddressQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	deathAddressType                 = reflect.TypeOf(&DeathAddress{})
	deathAddressMapping              = queries.MakeStructMapping(deathAddressType)
	deathAddressPrimaryKeyMapping, _ = queries.BindMapping(deathAddressType, deathAddressMapping, deathAddressPrimaryKeyColumns)
	deathAddressInsertCacheMut       sync.RWMutex
	deathAddressInsertCache          = make(map[string]insertCache)
	deathAddressUpdateCacheMut       sync.RWMutex
	deathAddressUpdateCache          = make(map[string]updateCache)
	deathAddressUpsertCacheMut       sync.RWMutex
	deathAddressUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var deathAddressBeforeInsertHooks []DeathAddressHook
var deathAddressBeforeUpdateHooks []DeathAddressHook
var deathAddressBeforeDeleteHooks []DeathAddressHook
var deathAddressBeforeUpsertHooks []DeathAddressHook

var deathAddressAfterInsertHooks []DeathAddressHook
var deathAddressAfterSelectHooks []DeathAddressHook
var deathAddressAfterUpdateHooks []DeathAddressHook
var deathAddressAfterDeleteHooks []DeathAddressHook
var deathAddressAfterUpsertHooks []DeathAddressHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DeathAddress) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DeathAddress) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DeathAddress) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DeathAddress) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DeathAddress) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DeathAddress) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DeathAddress) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DeathAddress) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DeathAddress) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range deathAddressAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDeathAddressHook registers your hook function for all future operations.
func AddDeathAddressHook(hookPoint boil.HookPoint, deathAddressHook DeathAddressHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		deathAddressBeforeInsertHooks = append(deathAddressBeforeInsertHooks, deathAddressHook)
	case boil.BeforeUpdateHook:
		deathAddressBeforeUpdateHooks = append(deathAddressBeforeUpdateHooks, deathAddressHook)
	case boil.BeforeDeleteHook:
		deathAddressBeforeDeleteHooks = append(deathAddressBeforeDeleteHooks, deathAddressHook)
	case boil.BeforeUpsertHook:
		deathAddressBeforeUpsertHooks = append(deathAddressBeforeUpsertHooks, deathAddressHook)
	case boil.AfterInsertHook:
		deathAddressAfterInsertHooks = append(deathAddressAfterInsertHooks, deathAddressHook)
	case boil.AfterSelectHook:
		deathAddressAfterSelectHooks = append(deathAddressAfterSelectHooks, deathAddressHook)
	case boil.AfterUpdateHook:
		deathAddressAfterUpdateHooks = append(deathAddressAfterUpdateHooks, deathAddressHook)
	case boil.AfterDeleteHook:
		deathAddressAfterDeleteHooks = append(deathAddressAfterDeleteHooks, deathAddressHook)
	case boil.AfterUpsertHook:
		deathAddressAfterUpsertHooks = append(deathAddressAfterUpsertHooks, deathAddressHook)
	}
}

// One returns a single deathAddress record from the query.
func (q deathAddressQuery) One(exec boil.Executor) (*DeathAddress, error) {
	o := &DeathAddress{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for death_addresses")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DeathAddress records from the query.
func (q deathAddressQuery) All(exec boil.Executor) (DeathAddressSlice, error) {
	var o []*DeathAddress

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to DeathAddress slice")
	}

	if len(deathAddressAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DeathAddress records in the query.
func (q deathAddressQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count death_addresses rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q deathAddressQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if death_addresses exists")
	}

	return count > 0, nil
}

// DeathAddresses retrieves all the records using an executor.
func DeathAddresses(mods ...qm.QueryMod) deathAddressQuery {
	mods = append(mods, qm.From("\"death_addresses\""))
	return deathAddressQuery{NewQuery(mods...)}
}

// FindDeathAddress retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDeathAddress(exec boil.Executor, walletAddress string, selectCols ...string) (*DeathAddress, error) {
	deathAddressObj := &DeathAddress{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"death_addresses\" where \"wallet_address\"=$1", sel,
	)

	q := queries.Raw(query, walletAddress)

	err := q.Bind(nil, exec, deathAddressObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from death_addresses")
	}

	if err = deathAddressObj.doAfterSelectHooks(exec); err != nil {
		return deathAddressObj, err
	}

	return deathAddressObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DeathAddress) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no death_addresses provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(deathAddressColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	deathAddressInsertCacheMut.RLock()
	cache, cached := deathAddressInsertCache[key]
	deathAddressInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			deathAddressAllColumns,
			deathAddressColumnsWithDefault,
			deathAddressColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(deathAddressType, deathAddressMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(deathAddressType, deathAddressMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"death_addresses\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"death_addresses\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into death_addresses")
	}

	if !cached {
		deathAddressInsertCacheMut.Lock()
		deathAddressInsertCache[key] = cache
		deathAddressInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the DeathAddress.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DeathAddress) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	deathAddressUpdateCacheMut.RLock()
	cache, cached := deathAddressUpdateCache[key]
	deathAddressUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			deathAddressAllColumns,
			deathAddressPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update death_addresses, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"death_addresses\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, deathAddressPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(deathAddressType, deathAddressMapping, append(wl, deathAddressPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update death_addresses row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for death_addresses")
	}

	if !cached {
		deathAddressUpdateCacheMut.Lock()
		deathAddressUpdateCache[key] = cache
		deathAddressUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q deathAddressQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for death_addresses")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for death_addresses")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DeathAddressSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deathAddressPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"death_addresses\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, deathAddressPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in deathAddress slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all deathAddress")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DeathAddress) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no death_addresses provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(deathAddressColumnsWithDefault, o)

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

	deathAddressUpsertCacheMut.RLock()
	cache, cached := deathAddressUpsertCache[key]
	deathAddressUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			deathAddressAllColumns,
			deathAddressColumnsWithDefault,
			deathAddressColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			deathAddressAllColumns,
			deathAddressPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert death_addresses, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(deathAddressPrimaryKeyColumns))
			copy(conflict, deathAddressPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"death_addresses\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(deathAddressType, deathAddressMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(deathAddressType, deathAddressMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert death_addresses")
	}

	if !cached {
		deathAddressUpsertCacheMut.Lock()
		deathAddressUpsertCache[key] = cache
		deathAddressUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single DeathAddress record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DeathAddress) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no DeathAddress provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), deathAddressPrimaryKeyMapping)
	sql := "DELETE FROM \"death_addresses\" WHERE \"wallet_address\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from death_addresses")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for death_addresses")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q deathAddressQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no deathAddressQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from death_addresses")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for death_addresses")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DeathAddressSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(deathAddressBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deathAddressPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"death_addresses\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deathAddressPrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from deathAddress slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for death_addresses")
	}

	if len(deathAddressAfterDeleteHooks) != 0 {
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
func (o *DeathAddress) Reload(exec boil.Executor) error {
	ret, err := FindDeathAddress(exec, o.WalletAddress)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DeathAddressSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DeathAddressSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), deathAddressPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"death_addresses\".* FROM \"death_addresses\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, deathAddressPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in DeathAddressSlice")
	}

	*o = slice

	return nil
}

// DeathAddressExists checks if the DeathAddress row exists.
func DeathAddressExists(exec boil.Executor, walletAddress string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"death_addresses\" where \"wallet_address\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, walletAddress)
	}
	row := exec.QueryRow(sql, walletAddress)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if death_addresses exists")
	}

	return exists, nil
}
