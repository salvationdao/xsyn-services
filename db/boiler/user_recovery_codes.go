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
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// UserRecoveryCode is an object representing the database table.
type UserRecoveryCode struct {
	ID           string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	UserID       string    `boiler:"user_id" boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	RecoveryCode string    `boiler:"recovery_code" boil:"recovery_code" json:"recovery_code" toml:"recovery_code" yaml:"recovery_code"`
	UsedAt       null.Time `boiler:"used_at" boil:"used_at" json:"used_at,omitempty" toml:"used_at" yaml:"used_at,omitempty"`
	UpdatedAt    time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt    time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *userRecoveryCodeR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L userRecoveryCodeL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UserRecoveryCodeColumns = struct {
	ID           string
	UserID       string
	RecoveryCode string
	UsedAt       string
	UpdatedAt    string
	CreatedAt    string
}{
	ID:           "id",
	UserID:       "user_id",
	RecoveryCode: "recovery_code",
	UsedAt:       "used_at",
	UpdatedAt:    "updated_at",
	CreatedAt:    "created_at",
}

var UserRecoveryCodeTableColumns = struct {
	ID           string
	UserID       string
	RecoveryCode string
	UsedAt       string
	UpdatedAt    string
	CreatedAt    string
}{
	ID:           "user_recovery_codes.id",
	UserID:       "user_recovery_codes.user_id",
	RecoveryCode: "user_recovery_codes.recovery_code",
	UsedAt:       "user_recovery_codes.used_at",
	UpdatedAt:    "user_recovery_codes.updated_at",
	CreatedAt:    "user_recovery_codes.created_at",
}

// Generated where

var UserRecoveryCodeWhere = struct {
	ID           whereHelperstring
	UserID       whereHelperstring
	RecoveryCode whereHelperstring
	UsedAt       whereHelpernull_Time
	UpdatedAt    whereHelpertime_Time
	CreatedAt    whereHelpertime_Time
}{
	ID:           whereHelperstring{field: "\"user_recovery_codes\".\"id\""},
	UserID:       whereHelperstring{field: "\"user_recovery_codes\".\"user_id\""},
	RecoveryCode: whereHelperstring{field: "\"user_recovery_codes\".\"recovery_code\""},
	UsedAt:       whereHelpernull_Time{field: "\"user_recovery_codes\".\"used_at\""},
	UpdatedAt:    whereHelpertime_Time{field: "\"user_recovery_codes\".\"updated_at\""},
	CreatedAt:    whereHelpertime_Time{field: "\"user_recovery_codes\".\"created_at\""},
}

// UserRecoveryCodeRels is where relationship names are stored.
var UserRecoveryCodeRels = struct {
	User string
}{
	User: "User",
}

// userRecoveryCodeR is where relationships are stored.
type userRecoveryCodeR struct {
	User *User `boiler:"User" boil:"User" json:"User" toml:"User" yaml:"User"`
}

// NewStruct creates a new relationship struct
func (*userRecoveryCodeR) NewStruct() *userRecoveryCodeR {
	return &userRecoveryCodeR{}
}

// userRecoveryCodeL is where Load methods for each relationship are stored.
type userRecoveryCodeL struct{}

var (
	userRecoveryCodeAllColumns            = []string{"id", "user_id", "recovery_code", "used_at", "updated_at", "created_at"}
	userRecoveryCodeColumnsWithoutDefault = []string{"user_id", "recovery_code", "used_at"}
	userRecoveryCodeColumnsWithDefault    = []string{"id", "updated_at", "created_at"}
	userRecoveryCodePrimaryKeyColumns     = []string{"id"}
)

type (
	// UserRecoveryCodeSlice is an alias for a slice of pointers to UserRecoveryCode.
	// This should almost always be used instead of []UserRecoveryCode.
	UserRecoveryCodeSlice []*UserRecoveryCode
	// UserRecoveryCodeHook is the signature for custom UserRecoveryCode hook methods
	UserRecoveryCodeHook func(boil.Executor, *UserRecoveryCode) error

	userRecoveryCodeQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	userRecoveryCodeType                 = reflect.TypeOf(&UserRecoveryCode{})
	userRecoveryCodeMapping              = queries.MakeStructMapping(userRecoveryCodeType)
	userRecoveryCodePrimaryKeyMapping, _ = queries.BindMapping(userRecoveryCodeType, userRecoveryCodeMapping, userRecoveryCodePrimaryKeyColumns)
	userRecoveryCodeInsertCacheMut       sync.RWMutex
	userRecoveryCodeInsertCache          = make(map[string]insertCache)
	userRecoveryCodeUpdateCacheMut       sync.RWMutex
	userRecoveryCodeUpdateCache          = make(map[string]updateCache)
	userRecoveryCodeUpsertCacheMut       sync.RWMutex
	userRecoveryCodeUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var userRecoveryCodeBeforeInsertHooks []UserRecoveryCodeHook
var userRecoveryCodeBeforeUpdateHooks []UserRecoveryCodeHook
var userRecoveryCodeBeforeDeleteHooks []UserRecoveryCodeHook
var userRecoveryCodeBeforeUpsertHooks []UserRecoveryCodeHook

var userRecoveryCodeAfterInsertHooks []UserRecoveryCodeHook
var userRecoveryCodeAfterSelectHooks []UserRecoveryCodeHook
var userRecoveryCodeAfterUpdateHooks []UserRecoveryCodeHook
var userRecoveryCodeAfterDeleteHooks []UserRecoveryCodeHook
var userRecoveryCodeAfterUpsertHooks []UserRecoveryCodeHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UserRecoveryCode) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UserRecoveryCode) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UserRecoveryCode) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UserRecoveryCode) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UserRecoveryCode) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UserRecoveryCode) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UserRecoveryCode) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UserRecoveryCode) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UserRecoveryCode) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userRecoveryCodeAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUserRecoveryCodeHook registers your hook function for all future operations.
func AddUserRecoveryCodeHook(hookPoint boil.HookPoint, userRecoveryCodeHook UserRecoveryCodeHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		userRecoveryCodeBeforeInsertHooks = append(userRecoveryCodeBeforeInsertHooks, userRecoveryCodeHook)
	case boil.BeforeUpdateHook:
		userRecoveryCodeBeforeUpdateHooks = append(userRecoveryCodeBeforeUpdateHooks, userRecoveryCodeHook)
	case boil.BeforeDeleteHook:
		userRecoveryCodeBeforeDeleteHooks = append(userRecoveryCodeBeforeDeleteHooks, userRecoveryCodeHook)
	case boil.BeforeUpsertHook:
		userRecoveryCodeBeforeUpsertHooks = append(userRecoveryCodeBeforeUpsertHooks, userRecoveryCodeHook)
	case boil.AfterInsertHook:
		userRecoveryCodeAfterInsertHooks = append(userRecoveryCodeAfterInsertHooks, userRecoveryCodeHook)
	case boil.AfterSelectHook:
		userRecoveryCodeAfterSelectHooks = append(userRecoveryCodeAfterSelectHooks, userRecoveryCodeHook)
	case boil.AfterUpdateHook:
		userRecoveryCodeAfterUpdateHooks = append(userRecoveryCodeAfterUpdateHooks, userRecoveryCodeHook)
	case boil.AfterDeleteHook:
		userRecoveryCodeAfterDeleteHooks = append(userRecoveryCodeAfterDeleteHooks, userRecoveryCodeHook)
	case boil.AfterUpsertHook:
		userRecoveryCodeAfterUpsertHooks = append(userRecoveryCodeAfterUpsertHooks, userRecoveryCodeHook)
	}
}

// One returns a single userRecoveryCode record from the query.
func (q userRecoveryCodeQuery) One(exec boil.Executor) (*UserRecoveryCode, error) {
	o := &UserRecoveryCode{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for user_recovery_codes")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UserRecoveryCode records from the query.
func (q userRecoveryCodeQuery) All(exec boil.Executor) (UserRecoveryCodeSlice, error) {
	var o []*UserRecoveryCode

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to UserRecoveryCode slice")
	}

	if len(userRecoveryCodeAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UserRecoveryCode records in the query.
func (q userRecoveryCodeQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count user_recovery_codes rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q userRecoveryCodeQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if user_recovery_codes exists")
	}

	return count > 0, nil
}

// User pointed to by the foreign key.
func (o *UserRecoveryCode) User(mods ...qm.QueryMod) userQuery {
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
func (userRecoveryCodeL) LoadUser(e boil.Executor, singular bool, maybeUserRecoveryCode interface{}, mods queries.Applicator) error {
	var slice []*UserRecoveryCode
	var object *UserRecoveryCode

	if singular {
		object = maybeUserRecoveryCode.(*UserRecoveryCode)
	} else {
		slice = *maybeUserRecoveryCode.(*[]*UserRecoveryCode)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &userRecoveryCodeR{}
		}
		args = append(args, object.UserID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &userRecoveryCodeR{}
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

	if len(userRecoveryCodeAfterSelectHooks) != 0 {
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
		foreign.R.UserRecoveryCodes = append(foreign.R.UserRecoveryCodes, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserID == foreign.ID {
				local.R.User = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.UserRecoveryCodes = append(foreign.R.UserRecoveryCodes, local)
				break
			}
		}
	}

	return nil
}

// SetUser of the userRecoveryCode to the related item.
// Sets o.R.User to related.
// Adds o to related.R.UserRecoveryCodes.
func (o *UserRecoveryCode) SetUser(exec boil.Executor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"user_recovery_codes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_id"}),
		strmangle.WhereClause("\"", "\"", 2, userRecoveryCodePrimaryKeyColumns),
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
		o.R = &userRecoveryCodeR{
			User: related,
		}
	} else {
		o.R.User = related
	}

	if related.R == nil {
		related.R = &userR{
			UserRecoveryCodes: UserRecoveryCodeSlice{o},
		}
	} else {
		related.R.UserRecoveryCodes = append(related.R.UserRecoveryCodes, o)
	}

	return nil
}

// UserRecoveryCodes retrieves all the records using an executor.
func UserRecoveryCodes(mods ...qm.QueryMod) userRecoveryCodeQuery {
	mods = append(mods, qm.From("\"user_recovery_codes\""))
	return userRecoveryCodeQuery{NewQuery(mods...)}
}

// FindUserRecoveryCode retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUserRecoveryCode(exec boil.Executor, iD string, selectCols ...string) (*UserRecoveryCode, error) {
	userRecoveryCodeObj := &UserRecoveryCode{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"user_recovery_codes\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, userRecoveryCodeObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from user_recovery_codes")
	}

	if err = userRecoveryCodeObj.doAfterSelectHooks(exec); err != nil {
		return userRecoveryCodeObj, err
	}

	return userRecoveryCodeObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UserRecoveryCode) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no user_recovery_codes provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(userRecoveryCodeColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	userRecoveryCodeInsertCacheMut.RLock()
	cache, cached := userRecoveryCodeInsertCache[key]
	userRecoveryCodeInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			userRecoveryCodeAllColumns,
			userRecoveryCodeColumnsWithDefault,
			userRecoveryCodeColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(userRecoveryCodeType, userRecoveryCodeMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(userRecoveryCodeType, userRecoveryCodeMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"user_recovery_codes\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"user_recovery_codes\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into user_recovery_codes")
	}

	if !cached {
		userRecoveryCodeInsertCacheMut.Lock()
		userRecoveryCodeInsertCache[key] = cache
		userRecoveryCodeInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the UserRecoveryCode.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UserRecoveryCode) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	userRecoveryCodeUpdateCacheMut.RLock()
	cache, cached := userRecoveryCodeUpdateCache[key]
	userRecoveryCodeUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			userRecoveryCodeAllColumns,
			userRecoveryCodePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update user_recovery_codes, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"user_recovery_codes\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, userRecoveryCodePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(userRecoveryCodeType, userRecoveryCodeMapping, append(wl, userRecoveryCodePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update user_recovery_codes row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for user_recovery_codes")
	}

	if !cached {
		userRecoveryCodeUpdateCacheMut.Lock()
		userRecoveryCodeUpdateCache[key] = cache
		userRecoveryCodeUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q userRecoveryCodeQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for user_recovery_codes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for user_recovery_codes")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UserRecoveryCodeSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userRecoveryCodePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"user_recovery_codes\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, userRecoveryCodePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in userRecoveryCode slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all userRecoveryCode")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UserRecoveryCode) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no user_recovery_codes provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userRecoveryCodeColumnsWithDefault, o)

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

	userRecoveryCodeUpsertCacheMut.RLock()
	cache, cached := userRecoveryCodeUpsertCache[key]
	userRecoveryCodeUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			userRecoveryCodeAllColumns,
			userRecoveryCodeColumnsWithDefault,
			userRecoveryCodeColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			userRecoveryCodeAllColumns,
			userRecoveryCodePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert user_recovery_codes, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(userRecoveryCodePrimaryKeyColumns))
			copy(conflict, userRecoveryCodePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"user_recovery_codes\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(userRecoveryCodeType, userRecoveryCodeMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(userRecoveryCodeType, userRecoveryCodeMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert user_recovery_codes")
	}

	if !cached {
		userRecoveryCodeUpsertCacheMut.Lock()
		userRecoveryCodeUpsertCache[key] = cache
		userRecoveryCodeUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single UserRecoveryCode record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UserRecoveryCode) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no UserRecoveryCode provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), userRecoveryCodePrimaryKeyMapping)
	sql := "DELETE FROM \"user_recovery_codes\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from user_recovery_codes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for user_recovery_codes")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q userRecoveryCodeQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no userRecoveryCodeQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from user_recovery_codes")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for user_recovery_codes")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UserRecoveryCodeSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(userRecoveryCodeBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userRecoveryCodePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"user_recovery_codes\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userRecoveryCodePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from userRecoveryCode slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for user_recovery_codes")
	}

	if len(userRecoveryCodeAfterDeleteHooks) != 0 {
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
func (o *UserRecoveryCode) Reload(exec boil.Executor) error {
	ret, err := FindUserRecoveryCode(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UserRecoveryCodeSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UserRecoveryCodeSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userRecoveryCodePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"user_recovery_codes\".* FROM \"user_recovery_codes\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userRecoveryCodePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in UserRecoveryCodeSlice")
	}

	*o = slice

	return nil
}

// UserRecoveryCodeExists checks if the UserRecoveryCode row exists.
func UserRecoveryCodeExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"user_recovery_codes\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if user_recovery_codes exists")
	}

	return exists, nil
}
