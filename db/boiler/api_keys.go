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

// APIKey is an object representing the database table.
type APIKey struct {
	ID        string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	UserID    string    `boiler:"user_id" boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	Type      string    `boiler:"type" boil:"type" json:"type" toml:"type" yaml:"type"`
	DeletedAt null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt time.Time `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *apiKeyR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L apiKeyL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var APIKeyColumns = struct {
	ID        string
	UserID    string
	Type      string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "id",
	UserID:    "user_id",
	Type:      "type",
	DeletedAt: "deleted_at",
	UpdatedAt: "updated_at",
	CreatedAt: "created_at",
}

var APIKeyTableColumns = struct {
	ID        string
	UserID    string
	Type      string
	DeletedAt string
	UpdatedAt string
	CreatedAt string
}{
	ID:        "api_keys.id",
	UserID:    "api_keys.user_id",
	Type:      "api_keys.type",
	DeletedAt: "api_keys.deleted_at",
	UpdatedAt: "api_keys.updated_at",
	CreatedAt: "api_keys.created_at",
}

// Generated where

type whereHelperstring struct{ field string }

func (w whereHelperstring) EQ(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperstring) NEQ(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperstring) LT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperstring) LTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperstring) GT(x string) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperstring) GTE(x string) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperstring) IN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperstring) NIN(slice []string) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpernull_Time struct{ field string }

func (w whereHelpernull_Time) EQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Time) NEQ(x null.Time) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Time) LT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Time) LTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Time) GT(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Time) GTE(x null.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Time) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Time) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var APIKeyWhere = struct {
	ID        whereHelperstring
	UserID    whereHelperstring
	Type      whereHelperstring
	DeletedAt whereHelpernull_Time
	UpdatedAt whereHelpertime_Time
	CreatedAt whereHelpertime_Time
}{
	ID:        whereHelperstring{field: "\"api_keys\".\"id\""},
	UserID:    whereHelperstring{field: "\"api_keys\".\"user_id\""},
	Type:      whereHelperstring{field: "\"api_keys\".\"type\""},
	DeletedAt: whereHelpernull_Time{field: "\"api_keys\".\"deleted_at\""},
	UpdatedAt: whereHelpertime_Time{field: "\"api_keys\".\"updated_at\""},
	CreatedAt: whereHelpertime_Time{field: "\"api_keys\".\"created_at\""},
}

// APIKeyRels is where relationship names are stored.
var APIKeyRels = struct {
	User string
}{
	User: "User",
}

// apiKeyR is where relationships are stored.
type apiKeyR struct {
	User *User `boiler:"User" boil:"User" json:"User" toml:"User" yaml:"User"`
}

// NewStruct creates a new relationship struct
func (*apiKeyR) NewStruct() *apiKeyR {
	return &apiKeyR{}
}

// apiKeyL is where Load methods for each relationship are stored.
type apiKeyL struct{}

var (
	apiKeyAllColumns            = []string{"id", "user_id", "type", "deleted_at", "updated_at", "created_at"}
	apiKeyColumnsWithoutDefault = []string{"user_id", "type", "deleted_at"}
	apiKeyColumnsWithDefault    = []string{"id", "updated_at", "created_at"}
	apiKeyPrimaryKeyColumns     = []string{"id"}
)

type (
	// APIKeySlice is an alias for a slice of pointers to APIKey.
	// This should almost always be used instead of []APIKey.
	APIKeySlice []*APIKey
	// APIKeyHook is the signature for custom APIKey hook methods
	APIKeyHook func(boil.Executor, *APIKey) error

	apiKeyQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	apiKeyType                 = reflect.TypeOf(&APIKey{})
	apiKeyMapping              = queries.MakeStructMapping(apiKeyType)
	apiKeyPrimaryKeyMapping, _ = queries.BindMapping(apiKeyType, apiKeyMapping, apiKeyPrimaryKeyColumns)
	apiKeyInsertCacheMut       sync.RWMutex
	apiKeyInsertCache          = make(map[string]insertCache)
	apiKeyUpdateCacheMut       sync.RWMutex
	apiKeyUpdateCache          = make(map[string]updateCache)
	apiKeyUpsertCacheMut       sync.RWMutex
	apiKeyUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var apiKeyBeforeInsertHooks []APIKeyHook
var apiKeyBeforeUpdateHooks []APIKeyHook
var apiKeyBeforeDeleteHooks []APIKeyHook
var apiKeyBeforeUpsertHooks []APIKeyHook

var apiKeyAfterInsertHooks []APIKeyHook
var apiKeyAfterSelectHooks []APIKeyHook
var apiKeyAfterUpdateHooks []APIKeyHook
var apiKeyAfterDeleteHooks []APIKeyHook
var apiKeyAfterUpsertHooks []APIKeyHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *APIKey) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *APIKey) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *APIKey) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *APIKey) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *APIKey) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *APIKey) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *APIKey) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *APIKey) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *APIKey) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range apiKeyAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddAPIKeyHook registers your hook function for all future operations.
func AddAPIKeyHook(hookPoint boil.HookPoint, apiKeyHook APIKeyHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		apiKeyBeforeInsertHooks = append(apiKeyBeforeInsertHooks, apiKeyHook)
	case boil.BeforeUpdateHook:
		apiKeyBeforeUpdateHooks = append(apiKeyBeforeUpdateHooks, apiKeyHook)
	case boil.BeforeDeleteHook:
		apiKeyBeforeDeleteHooks = append(apiKeyBeforeDeleteHooks, apiKeyHook)
	case boil.BeforeUpsertHook:
		apiKeyBeforeUpsertHooks = append(apiKeyBeforeUpsertHooks, apiKeyHook)
	case boil.AfterInsertHook:
		apiKeyAfterInsertHooks = append(apiKeyAfterInsertHooks, apiKeyHook)
	case boil.AfterSelectHook:
		apiKeyAfterSelectHooks = append(apiKeyAfterSelectHooks, apiKeyHook)
	case boil.AfterUpdateHook:
		apiKeyAfterUpdateHooks = append(apiKeyAfterUpdateHooks, apiKeyHook)
	case boil.AfterDeleteHook:
		apiKeyAfterDeleteHooks = append(apiKeyAfterDeleteHooks, apiKeyHook)
	case boil.AfterUpsertHook:
		apiKeyAfterUpsertHooks = append(apiKeyAfterUpsertHooks, apiKeyHook)
	}
}

// One returns a single apiKey record from the query.
func (q apiKeyQuery) One(exec boil.Executor) (*APIKey, error) {
	o := &APIKey{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for api_keys")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all APIKey records from the query.
func (q apiKeyQuery) All(exec boil.Executor) (APIKeySlice, error) {
	var o []*APIKey

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to APIKey slice")
	}

	if len(apiKeyAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all APIKey records in the query.
func (q apiKeyQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count api_keys rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q apiKeyQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if api_keys exists")
	}

	return count > 0, nil
}

// User pointed to by the foreign key.
func (o *APIKey) User(mods ...qm.QueryMod) userQuery {
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
func (apiKeyL) LoadUser(e boil.Executor, singular bool, maybeAPIKey interface{}, mods queries.Applicator) error {
	var slice []*APIKey
	var object *APIKey

	if singular {
		object = maybeAPIKey.(*APIKey)
	} else {
		slice = *maybeAPIKey.(*[]*APIKey)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &apiKeyR{}
		}
		args = append(args, object.UserID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &apiKeyR{}
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

	if len(apiKeyAfterSelectHooks) != 0 {
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
		foreign.R.APIKeys = append(foreign.R.APIKeys, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserID == foreign.ID {
				local.R.User = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.APIKeys = append(foreign.R.APIKeys, local)
				break
			}
		}
	}

	return nil
}

// SetUser of the apiKey to the related item.
// Sets o.R.User to related.
// Adds o to related.R.APIKeys.
func (o *APIKey) SetUser(exec boil.Executor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"api_keys\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_id"}),
		strmangle.WhereClause("\"", "\"", 2, apiKeyPrimaryKeyColumns),
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
		o.R = &apiKeyR{
			User: related,
		}
	} else {
		o.R.User = related
	}

	if related.R == nil {
		related.R = &userR{
			APIKeys: APIKeySlice{o},
		}
	} else {
		related.R.APIKeys = append(related.R.APIKeys, o)
	}

	return nil
}

// APIKeys retrieves all the records using an executor.
func APIKeys(mods ...qm.QueryMod) apiKeyQuery {
	mods = append(mods, qm.From("\"api_keys\""), qmhelper.WhereIsNull("\"api_keys\".\"deleted_at\""))
	return apiKeyQuery{NewQuery(mods...)}
}

// FindAPIKey retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindAPIKey(exec boil.Executor, iD string, selectCols ...string) (*APIKey, error) {
	apiKeyObj := &APIKey{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"api_keys\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, apiKeyObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from api_keys")
	}

	if err = apiKeyObj.doAfterSelectHooks(exec); err != nil {
		return apiKeyObj, err
	}

	return apiKeyObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *APIKey) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no api_keys provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(apiKeyColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	apiKeyInsertCacheMut.RLock()
	cache, cached := apiKeyInsertCache[key]
	apiKeyInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			apiKeyAllColumns,
			apiKeyColumnsWithDefault,
			apiKeyColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(apiKeyType, apiKeyMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(apiKeyType, apiKeyMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"api_keys\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"api_keys\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into api_keys")
	}

	if !cached {
		apiKeyInsertCacheMut.Lock()
		apiKeyInsertCache[key] = cache
		apiKeyInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the APIKey.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *APIKey) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	apiKeyUpdateCacheMut.RLock()
	cache, cached := apiKeyUpdateCache[key]
	apiKeyUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			apiKeyAllColumns,
			apiKeyPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update api_keys, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"api_keys\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, apiKeyPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(apiKeyType, apiKeyMapping, append(wl, apiKeyPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update api_keys row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for api_keys")
	}

	if !cached {
		apiKeyUpdateCacheMut.Lock()
		apiKeyUpdateCache[key] = cache
		apiKeyUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q apiKeyQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for api_keys")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for api_keys")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o APIKeySlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), apiKeyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"api_keys\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, apiKeyPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in apiKey slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all apiKey")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *APIKey) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no api_keys provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(apiKeyColumnsWithDefault, o)

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

	apiKeyUpsertCacheMut.RLock()
	cache, cached := apiKeyUpsertCache[key]
	apiKeyUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			apiKeyAllColumns,
			apiKeyColumnsWithDefault,
			apiKeyColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			apiKeyAllColumns,
			apiKeyPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert api_keys, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(apiKeyPrimaryKeyColumns))
			copy(conflict, apiKeyPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"api_keys\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(apiKeyType, apiKeyMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(apiKeyType, apiKeyMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert api_keys")
	}

	if !cached {
		apiKeyUpsertCacheMut.Lock()
		apiKeyUpsertCache[key] = cache
		apiKeyUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single APIKey record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *APIKey) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no APIKey provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), apiKeyPrimaryKeyMapping)
		sql = "DELETE FROM \"api_keys\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"api_keys\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(apiKeyType, apiKeyMapping, append(wl, apiKeyPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from api_keys")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for api_keys")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q apiKeyQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no apiKeyQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from api_keys")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for api_keys")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o APIKeySlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(apiKeyBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), apiKeyPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"api_keys\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, apiKeyPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), apiKeyPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"api_keys\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, apiKeyPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from apiKey slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for api_keys")
	}

	if len(apiKeyAfterDeleteHooks) != 0 {
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
func (o *APIKey) Reload(exec boil.Executor) error {
	ret, err := FindAPIKey(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *APIKeySlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := APIKeySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), apiKeyPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"api_keys\".* FROM \"api_keys\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, apiKeyPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in APIKeySlice")
	}

	*o = slice

	return nil
}

// APIKeyExists checks if the APIKey row exists.
func APIKeyExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"api_keys\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if api_keys exists")
	}

	return exists, nil
}