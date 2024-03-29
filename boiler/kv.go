// Code generated by SQLBoiler 4.8.6 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// KV is an object representing the database table.
type KV struct {
	ID        string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	Key       string    `boiler:"key" boil:"key" json:"key" toml:"key" yaml:"key"`
	Value     string    `boiler:"value" boil:"value" json:"value" toml:"value" yaml:"value"`
	DeletedAt null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *kvR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L kvL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var KVColumns = struct {
	ID        string
	Key       string
	Value     string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "id",
	Key:       "key",
	Value:     "value",
	DeletedAt: "deleted_at",
	UpdatedAt: "updated_at",
	CreatedAt: "created_at",
}

var KVTableColumns = struct {
	ID        string
	Key       string
	Value     string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "kv.id",
	Key:       "kv.key",
	Value:     "kv.value",
	DeletedAt: "kv.deleted_at",
	UpdatedAt: "kv.updated_at",
	CreatedAt: "kv.created_at",
}

// Generated where

var KVWhere = struct {
	ID        whereHelperstring
	Key       whereHelperstring
	Value     whereHelperstring
	DeletedAt whereHelpernull_Time
	UpdatedAt whereHelpertime_Time
	CreatedAt whereHelpertime_Time
}{
	ID:        whereHelperstring{field: "\"kv\".\"id\""},
	Key:       whereHelperstring{field: "\"kv\".\"key\""},
	Value:     whereHelperstring{field: "\"kv\".\"value\""},
	DeletedAt: whereHelpernull_Time{field: "\"kv\".\"deleted_at\""},
	UpdatedAt: whereHelpertime_Time{field: "\"kv\".\"updated_at\""},
	CreatedAt: whereHelpertime_Time{field: "\"kv\".\"created_at\""},
}

// KVRels is where relationship names are stored.
var KVRels = struct {
}{}

// kvR is where relationships are stored.
type kvR struct {
}

// NewStruct creates a new relationship struct
func (*kvR) NewStruct() *kvR {
	return &kvR{}
}

// kvL is where Load methods for each relationship are stored.
type kvL struct{}

var (
	kvAllColumns            = []string{"id", "key", "value", "deleted_at", "updated_at", "created_at"}
	kvColumnsWithoutDefault = []string{}
	kvColumnsWithDefault    = []string{"id", "key", "value", "deleted_at", "updated_at", "created_at"}
	kvPrimaryKeyColumns     = []string{"id"}
	kvGeneratedColumns      = []string{}
)

type (
	// KVSlice is an alias for a slice of pointers to KV.
	// This should almost always be used instead of []KV.
	KVSlice []*KV
	// KVHook is the signature for custom KV hook methods
	KVHook func(boil.Executor, *KV) error

	kvQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	kvType                 = reflect.TypeOf(&KV{})
	kvMapping              = queries.MakeStructMapping(kvType)
	kvPrimaryKeyMapping, _ = queries.BindMapping(kvType, kvMapping, kvPrimaryKeyColumns)
	kvInsertCacheMut       sync.RWMutex
	kvInsertCache          = make(map[string]insertCache)
	kvUpdateCacheMut       sync.RWMutex
	kvUpdateCache          = make(map[string]updateCache)
	kvUpsertCacheMut       sync.RWMutex
	kvUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var kvAfterSelectHooks []KVHook

var kvBeforeInsertHooks []KVHook
var kvAfterInsertHooks []KVHook

var kvBeforeUpdateHooks []KVHook
var kvAfterUpdateHooks []KVHook

var kvBeforeDeleteHooks []KVHook
var kvAfterDeleteHooks []KVHook

var kvBeforeUpsertHooks []KVHook
var kvAfterUpsertHooks []KVHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *KV) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range kvAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *KV) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range kvBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *KV) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range kvAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *KV) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range kvBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *KV) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range kvAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *KV) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range kvBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *KV) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range kvAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *KV) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range kvBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *KV) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range kvAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddKVHook registers your hook function for all future operations.
func AddKVHook(hookPoint boil.HookPoint, kvHook KVHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		kvAfterSelectHooks = append(kvAfterSelectHooks, kvHook)
	case boil.BeforeInsertHook:
		kvBeforeInsertHooks = append(kvBeforeInsertHooks, kvHook)
	case boil.AfterInsertHook:
		kvAfterInsertHooks = append(kvAfterInsertHooks, kvHook)
	case boil.BeforeUpdateHook:
		kvBeforeUpdateHooks = append(kvBeforeUpdateHooks, kvHook)
	case boil.AfterUpdateHook:
		kvAfterUpdateHooks = append(kvAfterUpdateHooks, kvHook)
	case boil.BeforeDeleteHook:
		kvBeforeDeleteHooks = append(kvBeforeDeleteHooks, kvHook)
	case boil.AfterDeleteHook:
		kvAfterDeleteHooks = append(kvAfterDeleteHooks, kvHook)
	case boil.BeforeUpsertHook:
		kvBeforeUpsertHooks = append(kvBeforeUpsertHooks, kvHook)
	case boil.AfterUpsertHook:
		kvAfterUpsertHooks = append(kvAfterUpsertHooks, kvHook)
	}
}

// One returns a single kv record from the query.
func (q kvQuery) One(exec boil.Executor) (*KV, error) {
	o := &KV{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for kv")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all KV records from the query.
func (q kvQuery) All(exec boil.Executor) (KVSlice, error) {
	var o []*KV

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to KV slice")
	}

	if len(kvAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all KV records in the query.
func (q kvQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count kv rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q kvQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if kv exists")
	}

	return count > 0, nil
}

// KVS retrieves all the records using an executor.
func KVS(mods ...qm.QueryMod) kvQuery {
	mods = append(mods, qm.From("\"kv\""), qmhelper.WhereIsNull("\"kv\".\"deleted_at\""))
	return kvQuery{NewQuery(mods...)}
}

// FindKV retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindKV(exec boil.Executor, iD string, selectCols ...string) (*KV, error) {
	kvObj := &KV{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"kv\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, kvObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from kv")
	}

	if err = kvObj.doAfterSelectHooks(exec); err != nil {
		return kvObj, err
	}

	return kvObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *KV) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no kv provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(kvColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	kvInsertCacheMut.RLock()
	cache, cached := kvInsertCache[key]
	kvInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			kvAllColumns,
			kvColumnsWithDefault,
			kvColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(kvType, kvMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(kvType, kvMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"kv\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"kv\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into kv")
	}

	if !cached {
		kvInsertCacheMut.Lock()
		kvInsertCache[key] = cache
		kvInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the KV.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *KV) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	kvUpdateCacheMut.RLock()
	cache, cached := kvUpdateCache[key]
	kvUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			kvAllColumns,
			kvPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update kv, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"kv\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, kvPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(kvType, kvMapping, append(wl, kvPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update kv row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for kv")
	}

	if !cached {
		kvUpdateCacheMut.Lock()
		kvUpdateCache[key] = cache
		kvUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q kvQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for kv")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for kv")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o KVSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), kvPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"kv\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, kvPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in kv slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all kv")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *KV) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no kv provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(kvColumnsWithDefault, o)

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

	kvUpsertCacheMut.RLock()
	cache, cached := kvUpsertCache[key]
	kvUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			kvAllColumns,
			kvColumnsWithDefault,
			kvColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			kvAllColumns,
			kvPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert kv, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(kvPrimaryKeyColumns))
			copy(conflict, kvPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"kv\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(kvType, kvMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(kvType, kvMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert kv")
	}

	if !cached {
		kvUpsertCacheMut.Lock()
		kvUpsertCache[key] = cache
		kvUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single KV record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *KV) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no KV provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), kvPrimaryKeyMapping)
		sql = "DELETE FROM \"kv\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"kv\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(kvType, kvMapping, append(wl, kvPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from kv")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for kv")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q kvQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no kvQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from kv")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for kv")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o KVSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(kvBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), kvPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"kv\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, kvPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), kvPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"kv\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, kvPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from kv slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for kv")
	}

	if len(kvAfterDeleteHooks) != 0 {
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
func (o *KV) Reload(exec boil.Executor) error {
	ret, err := FindKV(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *KVSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := KVSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), kvPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"kv\".* FROM \"kv\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, kvPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in KVSlice")
	}

	*o = slice

	return nil
}

// KVExists checks if the KV row exists.
func KVExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"kv\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if kv exists")
	}

	return exists, nil
}
