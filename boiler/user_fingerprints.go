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

// UserFingerprint is an object representing the database table.
type UserFingerprint struct {
	UserID        string    `boiler:"user_id" boil:"user_id" json:"user_id" toml:"user_id" yaml:"user_id"`
	FingerprintID string    `boiler:"fingerprint_id" boil:"fingerprint_id" json:"fingerprint_id" toml:"fingerprint_id" yaml:"fingerprint_id"`
	DeletedAt     null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	CreatedAt     time.Time `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *userFingerprintR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L userFingerprintL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var UserFingerprintColumns = struct {
	UserID        string
	FingerprintID string
	DeletedAt     string
	CreatedAt     string
}{
	UserID:        "user_id",
	FingerprintID: "fingerprint_id",
	DeletedAt:     "deleted_at",
	CreatedAt:     "created_at",
}

var UserFingerprintTableColumns = struct {
	UserID        string
	FingerprintID string
	DeletedAt     string
	CreatedAt     string
}{
	UserID:        "user_fingerprints.user_id",
	FingerprintID: "user_fingerprints.fingerprint_id",
	DeletedAt:     "user_fingerprints.deleted_at",
	CreatedAt:     "user_fingerprints.created_at",
}

// Generated where

var UserFingerprintWhere = struct {
	UserID        whereHelperstring
	FingerprintID whereHelperstring
	DeletedAt     whereHelpernull_Time
	CreatedAt     whereHelpertime_Time
}{
	UserID:        whereHelperstring{field: "\"user_fingerprints\".\"user_id\""},
	FingerprintID: whereHelperstring{field: "\"user_fingerprints\".\"fingerprint_id\""},
	DeletedAt:     whereHelpernull_Time{field: "\"user_fingerprints\".\"deleted_at\""},
	CreatedAt:     whereHelpertime_Time{field: "\"user_fingerprints\".\"created_at\""},
}

// UserFingerprintRels is where relationship names are stored.
var UserFingerprintRels = struct {
	Fingerprint string
	User        string
}{
	Fingerprint: "Fingerprint",
	User:        "User",
}

// userFingerprintR is where relationships are stored.
type userFingerprintR struct {
	Fingerprint *Fingerprint `boiler:"Fingerprint" boil:"Fingerprint" json:"Fingerprint" toml:"Fingerprint" yaml:"Fingerprint"`
	User        *User        `boiler:"User" boil:"User" json:"User" toml:"User" yaml:"User"`
}

// NewStruct creates a new relationship struct
func (*userFingerprintR) NewStruct() *userFingerprintR {
	return &userFingerprintR{}
}

// userFingerprintL is where Load methods for each relationship are stored.
type userFingerprintL struct{}

var (
	userFingerprintAllColumns            = []string{"user_id", "fingerprint_id", "deleted_at", "created_at"}
	userFingerprintColumnsWithoutDefault = []string{"user_id", "fingerprint_id"}
	userFingerprintColumnsWithDefault    = []string{"deleted_at", "created_at"}
	userFingerprintPrimaryKeyColumns     = []string{"user_id", "fingerprint_id"}
	userFingerprintGeneratedColumns      = []string{}
)

type (
	// UserFingerprintSlice is an alias for a slice of pointers to UserFingerprint.
	// This should almost always be used instead of []UserFingerprint.
	UserFingerprintSlice []*UserFingerprint
	// UserFingerprintHook is the signature for custom UserFingerprint hook methods
	UserFingerprintHook func(boil.Executor, *UserFingerprint) error

	userFingerprintQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	userFingerprintType                 = reflect.TypeOf(&UserFingerprint{})
	userFingerprintMapping              = queries.MakeStructMapping(userFingerprintType)
	userFingerprintPrimaryKeyMapping, _ = queries.BindMapping(userFingerprintType, userFingerprintMapping, userFingerprintPrimaryKeyColumns)
	userFingerprintInsertCacheMut       sync.RWMutex
	userFingerprintInsertCache          = make(map[string]insertCache)
	userFingerprintUpdateCacheMut       sync.RWMutex
	userFingerprintUpdateCache          = make(map[string]updateCache)
	userFingerprintUpsertCacheMut       sync.RWMutex
	userFingerprintUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var userFingerprintAfterSelectHooks []UserFingerprintHook

var userFingerprintBeforeInsertHooks []UserFingerprintHook
var userFingerprintAfterInsertHooks []UserFingerprintHook

var userFingerprintBeforeUpdateHooks []UserFingerprintHook
var userFingerprintAfterUpdateHooks []UserFingerprintHook

var userFingerprintBeforeDeleteHooks []UserFingerprintHook
var userFingerprintAfterDeleteHooks []UserFingerprintHook

var userFingerprintBeforeUpsertHooks []UserFingerprintHook
var userFingerprintAfterUpsertHooks []UserFingerprintHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *UserFingerprint) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *UserFingerprint) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *UserFingerprint) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *UserFingerprint) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *UserFingerprint) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *UserFingerprint) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *UserFingerprint) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *UserFingerprint) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *UserFingerprint) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range userFingerprintAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddUserFingerprintHook registers your hook function for all future operations.
func AddUserFingerprintHook(hookPoint boil.HookPoint, userFingerprintHook UserFingerprintHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		userFingerprintAfterSelectHooks = append(userFingerprintAfterSelectHooks, userFingerprintHook)
	case boil.BeforeInsertHook:
		userFingerprintBeforeInsertHooks = append(userFingerprintBeforeInsertHooks, userFingerprintHook)
	case boil.AfterInsertHook:
		userFingerprintAfterInsertHooks = append(userFingerprintAfterInsertHooks, userFingerprintHook)
	case boil.BeforeUpdateHook:
		userFingerprintBeforeUpdateHooks = append(userFingerprintBeforeUpdateHooks, userFingerprintHook)
	case boil.AfterUpdateHook:
		userFingerprintAfterUpdateHooks = append(userFingerprintAfterUpdateHooks, userFingerprintHook)
	case boil.BeforeDeleteHook:
		userFingerprintBeforeDeleteHooks = append(userFingerprintBeforeDeleteHooks, userFingerprintHook)
	case boil.AfterDeleteHook:
		userFingerprintAfterDeleteHooks = append(userFingerprintAfterDeleteHooks, userFingerprintHook)
	case boil.BeforeUpsertHook:
		userFingerprintBeforeUpsertHooks = append(userFingerprintBeforeUpsertHooks, userFingerprintHook)
	case boil.AfterUpsertHook:
		userFingerprintAfterUpsertHooks = append(userFingerprintAfterUpsertHooks, userFingerprintHook)
	}
}

// One returns a single userFingerprint record from the query.
func (q userFingerprintQuery) One(exec boil.Executor) (*UserFingerprint, error) {
	o := &UserFingerprint{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for user_fingerprints")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all UserFingerprint records from the query.
func (q userFingerprintQuery) All(exec boil.Executor) (UserFingerprintSlice, error) {
	var o []*UserFingerprint

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to UserFingerprint slice")
	}

	if len(userFingerprintAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all UserFingerprint records in the query.
func (q userFingerprintQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count user_fingerprints rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q userFingerprintQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if user_fingerprints exists")
	}

	return count > 0, nil
}

// Fingerprint pointed to by the foreign key.
func (o *UserFingerprint) Fingerprint(mods ...qm.QueryMod) fingerprintQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.FingerprintID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Fingerprints(queryMods...)
	queries.SetFrom(query.Query, "\"fingerprints\"")

	return query
}

// User pointed to by the foreign key.
func (o *UserFingerprint) User(mods ...qm.QueryMod) userQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.UserID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := Users(queryMods...)
	queries.SetFrom(query.Query, "\"users\"")

	return query
}

// LoadFingerprint allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (userFingerprintL) LoadFingerprint(e boil.Executor, singular bool, maybeUserFingerprint interface{}, mods queries.Applicator) error {
	var slice []*UserFingerprint
	var object *UserFingerprint

	if singular {
		object = maybeUserFingerprint.(*UserFingerprint)
	} else {
		slice = *maybeUserFingerprint.(*[]*UserFingerprint)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &userFingerprintR{}
		}
		args = append(args, object.FingerprintID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &userFingerprintR{}
			}

			for _, a := range args {
				if a == obj.FingerprintID {
					continue Outer
				}
			}

			args = append(args, obj.FingerprintID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`fingerprints`),
		qm.WhereIn(`fingerprints.id in ?`, args...),
		qmhelper.WhereIsNull(`fingerprints.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Fingerprint")
	}

	var resultSlice []*Fingerprint
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Fingerprint")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for fingerprints")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for fingerprints")
	}

	if len(userFingerprintAfterSelectHooks) != 0 {
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
		object.R.Fingerprint = foreign
		if foreign.R == nil {
			foreign.R = &fingerprintR{}
		}
		foreign.R.UserFingerprints = append(foreign.R.UserFingerprints, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.FingerprintID == foreign.ID {
				local.R.Fingerprint = foreign
				if foreign.R == nil {
					foreign.R = &fingerprintR{}
				}
				foreign.R.UserFingerprints = append(foreign.R.UserFingerprints, local)
				break
			}
		}
	}

	return nil
}

// LoadUser allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (userFingerprintL) LoadUser(e boil.Executor, singular bool, maybeUserFingerprint interface{}, mods queries.Applicator) error {
	var slice []*UserFingerprint
	var object *UserFingerprint

	if singular {
		object = maybeUserFingerprint.(*UserFingerprint)
	} else {
		slice = *maybeUserFingerprint.(*[]*UserFingerprint)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &userFingerprintR{}
		}
		args = append(args, object.UserID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &userFingerprintR{}
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

	if len(userFingerprintAfterSelectHooks) != 0 {
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
		foreign.R.UserFingerprints = append(foreign.R.UserFingerprints, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.UserID == foreign.ID {
				local.R.User = foreign
				if foreign.R == nil {
					foreign.R = &userR{}
				}
				foreign.R.UserFingerprints = append(foreign.R.UserFingerprints, local)
				break
			}
		}
	}

	return nil
}

// SetFingerprint of the userFingerprint to the related item.
// Sets o.R.Fingerprint to related.
// Adds o to related.R.UserFingerprints.
func (o *UserFingerprint) SetFingerprint(exec boil.Executor, insert bool, related *Fingerprint) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"user_fingerprints\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"fingerprint_id"}),
		strmangle.WhereClause("\"", "\"", 2, userFingerprintPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.UserID, o.FingerprintID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.FingerprintID = related.ID
	if o.R == nil {
		o.R = &userFingerprintR{
			Fingerprint: related,
		}
	} else {
		o.R.Fingerprint = related
	}

	if related.R == nil {
		related.R = &fingerprintR{
			UserFingerprints: UserFingerprintSlice{o},
		}
	} else {
		related.R.UserFingerprints = append(related.R.UserFingerprints, o)
	}

	return nil
}

// SetUser of the userFingerprint to the related item.
// Sets o.R.User to related.
// Adds o to related.R.UserFingerprints.
func (o *UserFingerprint) SetUser(exec boil.Executor, insert bool, related *User) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"user_fingerprints\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"user_id"}),
		strmangle.WhereClause("\"", "\"", 2, userFingerprintPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.UserID, o.FingerprintID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.UserID = related.ID
	if o.R == nil {
		o.R = &userFingerprintR{
			User: related,
		}
	} else {
		o.R.User = related
	}

	if related.R == nil {
		related.R = &userR{
			UserFingerprints: UserFingerprintSlice{o},
		}
	} else {
		related.R.UserFingerprints = append(related.R.UserFingerprints, o)
	}

	return nil
}

// UserFingerprints retrieves all the records using an executor.
func UserFingerprints(mods ...qm.QueryMod) userFingerprintQuery {
	mods = append(mods, qm.From("\"user_fingerprints\""), qmhelper.WhereIsNull("\"user_fingerprints\".\"deleted_at\""))
	return userFingerprintQuery{NewQuery(mods...)}
}

// FindUserFingerprint retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindUserFingerprint(exec boil.Executor, userID string, fingerprintID string, selectCols ...string) (*UserFingerprint, error) {
	userFingerprintObj := &UserFingerprint{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"user_fingerprints\" where \"user_id\"=$1 AND \"fingerprint_id\"=$2 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, userID, fingerprintID)

	err := q.Bind(nil, exec, userFingerprintObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from user_fingerprints")
	}

	if err = userFingerprintObj.doAfterSelectHooks(exec); err != nil {
		return userFingerprintObj, err
	}

	return userFingerprintObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *UserFingerprint) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no user_fingerprints provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userFingerprintColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	userFingerprintInsertCacheMut.RLock()
	cache, cached := userFingerprintInsertCache[key]
	userFingerprintInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			userFingerprintAllColumns,
			userFingerprintColumnsWithDefault,
			userFingerprintColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(userFingerprintType, userFingerprintMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(userFingerprintType, userFingerprintMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"user_fingerprints\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"user_fingerprints\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into user_fingerprints")
	}

	if !cached {
		userFingerprintInsertCacheMut.Lock()
		userFingerprintInsertCache[key] = cache
		userFingerprintInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the UserFingerprint.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *UserFingerprint) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	userFingerprintUpdateCacheMut.RLock()
	cache, cached := userFingerprintUpdateCache[key]
	userFingerprintUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			userFingerprintAllColumns,
			userFingerprintPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update user_fingerprints, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"user_fingerprints\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, userFingerprintPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(userFingerprintType, userFingerprintMapping, append(wl, userFingerprintPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update user_fingerprints row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for user_fingerprints")
	}

	if !cached {
		userFingerprintUpdateCacheMut.Lock()
		userFingerprintUpdateCache[key] = cache
		userFingerprintUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q userFingerprintQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for user_fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for user_fingerprints")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o UserFingerprintSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userFingerprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"user_fingerprints\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, userFingerprintPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in userFingerprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all userFingerprint")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *UserFingerprint) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no user_fingerprints provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(userFingerprintColumnsWithDefault, o)

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

	userFingerprintUpsertCacheMut.RLock()
	cache, cached := userFingerprintUpsertCache[key]
	userFingerprintUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			userFingerprintAllColumns,
			userFingerprintColumnsWithDefault,
			userFingerprintColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			userFingerprintAllColumns,
			userFingerprintPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert user_fingerprints, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(userFingerprintPrimaryKeyColumns))
			copy(conflict, userFingerprintPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"user_fingerprints\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(userFingerprintType, userFingerprintMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(userFingerprintType, userFingerprintMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert user_fingerprints")
	}

	if !cached {
		userFingerprintUpsertCacheMut.Lock()
		userFingerprintUpsertCache[key] = cache
		userFingerprintUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single UserFingerprint record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *UserFingerprint) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no UserFingerprint provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), userFingerprintPrimaryKeyMapping)
		sql = "DELETE FROM \"user_fingerprints\" WHERE \"user_id\"=$1 AND \"fingerprint_id\"=$2"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"user_fingerprints\" SET %s WHERE \"user_id\"=$2 AND \"fingerprint_id\"=$3",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(userFingerprintType, userFingerprintMapping, append(wl, userFingerprintPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from user_fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for user_fingerprints")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q userFingerprintQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no userFingerprintQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from user_fingerprints")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for user_fingerprints")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o UserFingerprintSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(userFingerprintBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userFingerprintPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"user_fingerprints\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userFingerprintPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userFingerprintPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"user_fingerprints\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, userFingerprintPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from userFingerprint slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for user_fingerprints")
	}

	if len(userFingerprintAfterDeleteHooks) != 0 {
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
func (o *UserFingerprint) Reload(exec boil.Executor) error {
	ret, err := FindUserFingerprint(exec, o.UserID, o.FingerprintID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *UserFingerprintSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := UserFingerprintSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), userFingerprintPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"user_fingerprints\".* FROM \"user_fingerprints\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, userFingerprintPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in UserFingerprintSlice")
	}

	*o = slice

	return nil
}

// UserFingerprintExists checks if the UserFingerprint row exists.
func UserFingerprintExists(exec boil.Executor, userID string, fingerprintID string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"user_fingerprints\" where \"user_id\"=$1 AND \"fingerprint_id\"=$2 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, userID, fingerprintID)
	}
	row := exec.QueryRow(sql, userID, fingerprintID)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if user_fingerprints exists")
	}

	return exists, nil
}
