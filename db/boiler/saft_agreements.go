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

// SaftAgreement is an object representing the database table.
type SaftAgreement struct {
	ID                string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	UserPublicAddress string      `boiler:"user_public_address" boil:"user_public_address" json:"user_public_address" toml:"user_public_address" yaml:"user_public_address"`
	Message           null.String `boiler:"message" boil:"message" json:"message,omitempty" toml:"message" yaml:"message,omitempty"`
	MessageHex        null.String `boiler:"message_hex" boil:"message_hex" json:"message_hex,omitempty" toml:"message_hex" yaml:"message_hex,omitempty"`
	SignatureHex      null.String `boiler:"signature_hex" boil:"signature_hex" json:"signature_hex,omitempty" toml:"signature_hex" yaml:"signature_hex,omitempty"`
	SignerAddressHex  null.String `boiler:"signer_address_hex" boil:"signer_address_hex" json:"signer_address_hex,omitempty" toml:"signer_address_hex" yaml:"signer_address_hex,omitempty"`
	Agree             null.Bool   `boiler:"agree" boil:"agree" json:"agree,omitempty" toml:"agree" yaml:"agree,omitempty"`
	SignedAt          time.Time   `boiler:"signed_at" boil:"signed_at" json:"signed_at" toml:"signed_at" yaml:"signed_at"`
	DeletedAt         null.Time   `boiler:"deleted_at" boil:"deleted_at" json:"deleted_at,omitempty" toml:"deleted_at" yaml:"deleted_at,omitempty"`
	UpdatedAt         time.Time   `boiler:"updated_at" boil:"updated_at" json:"updated_at" toml:"updated_at" yaml:"updated_at"`
	CreatedAt         time.Time   `boiler:"created_at" boil:"created_at" json:"created_at" toml:"created_at" yaml:"created_at"`

	R *saftAgreementR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L saftAgreementL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var SaftAgreementColumns = struct {
	ID                string
	UserPublicAddress string
	Message           string
	MessageHex        string
	SignatureHex      string
	SignerAddressHex  string
	Agree             string
	SignedAt          string
	DeletedAt         string
	UpdatedAt         string
	CreatedAt         string
}{
	ID:                "id",
	UserPublicAddress: "user_public_address",
	Message:           "message",
	MessageHex:        "message_hex",
	SignatureHex:      "signature_hex",
	SignerAddressHex:  "signer_address_hex",
	Agree:             "agree",
	SignedAt:          "signed_at",
	DeletedAt:         "deleted_at",
	UpdatedAt:         "updated_at",
	CreatedAt:         "created_at",
}

var SaftAgreementTableColumns = struct {
	ID                string
	UserPublicAddress string
	Message           string
	MessageHex        string
	SignatureHex      string
	SignerAddressHex  string
	Agree             string
	SignedAt          string
	DeletedAt         string
	UpdatedAt         string
	CreatedAt         string
}{
	ID:                "saft_agreements.id",
	UserPublicAddress: "saft_agreements.user_public_address",
	Message:           "saft_agreements.message",
	MessageHex:        "saft_agreements.message_hex",
	SignatureHex:      "saft_agreements.signature_hex",
	SignerAddressHex:  "saft_agreements.signer_address_hex",
	Agree:             "saft_agreements.agree",
	SignedAt:          "saft_agreements.signed_at",
	DeletedAt:         "saft_agreements.deleted_at",
	UpdatedAt:         "saft_agreements.updated_at",
	CreatedAt:         "saft_agreements.created_at",
}

// Generated where

type whereHelpernull_Bool struct{ field string }

func (w whereHelpernull_Bool) EQ(x null.Bool) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, false, x)
}
func (w whereHelpernull_Bool) NEQ(x null.Bool) qm.QueryMod {
	return qmhelper.WhereNullEQ(w.field, true, x)
}
func (w whereHelpernull_Bool) LT(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpernull_Bool) LTE(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpernull_Bool) GT(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpernull_Bool) GTE(x null.Bool) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

func (w whereHelpernull_Bool) IsNull() qm.QueryMod    { return qmhelper.WhereIsNull(w.field) }
func (w whereHelpernull_Bool) IsNotNull() qm.QueryMod { return qmhelper.WhereIsNotNull(w.field) }

var SaftAgreementWhere = struct {
	ID                whereHelperstring
	UserPublicAddress whereHelperstring
	Message           whereHelpernull_String
	MessageHex        whereHelpernull_String
	SignatureHex      whereHelpernull_String
	SignerAddressHex  whereHelpernull_String
	Agree             whereHelpernull_Bool
	SignedAt          whereHelpertime_Time
	DeletedAt         whereHelpernull_Time
	UpdatedAt         whereHelpertime_Time
	CreatedAt         whereHelpertime_Time
}{
	ID:                whereHelperstring{field: "\"saft_agreements\".\"id\""},
	UserPublicAddress: whereHelperstring{field: "\"saft_agreements\".\"user_public_address\""},
	Message:           whereHelpernull_String{field: "\"saft_agreements\".\"message\""},
	MessageHex:        whereHelpernull_String{field: "\"saft_agreements\".\"message_hex\""},
	SignatureHex:      whereHelpernull_String{field: "\"saft_agreements\".\"signature_hex\""},
	SignerAddressHex:  whereHelpernull_String{field: "\"saft_agreements\".\"signer_address_hex\""},
	Agree:             whereHelpernull_Bool{field: "\"saft_agreements\".\"agree\""},
	SignedAt:          whereHelpertime_Time{field: "\"saft_agreements\".\"signed_at\""},
	DeletedAt:         whereHelpernull_Time{field: "\"saft_agreements\".\"deleted_at\""},
	UpdatedAt:         whereHelpertime_Time{field: "\"saft_agreements\".\"updated_at\""},
	CreatedAt:         whereHelpertime_Time{field: "\"saft_agreements\".\"created_at\""},
}

// SaftAgreementRels is where relationship names are stored.
var SaftAgreementRels = struct {
}{}

// saftAgreementR is where relationships are stored.
type saftAgreementR struct {
}

// NewStruct creates a new relationship struct
func (*saftAgreementR) NewStruct() *saftAgreementR {
	return &saftAgreementR{}
}

// saftAgreementL is where Load methods for each relationship are stored.
type saftAgreementL struct{}

var (
	saftAgreementAllColumns            = []string{"id", "user_public_address", "message", "message_hex", "signature_hex", "signer_address_hex", "agree", "signed_at", "deleted_at", "updated_at", "created_at"}
	saftAgreementColumnsWithoutDefault = []string{"user_public_address", "message", "message_hex", "signature_hex", "signer_address_hex", "agree", "deleted_at"}
	saftAgreementColumnsWithDefault    = []string{"id", "signed_at", "updated_at", "created_at"}
	saftAgreementPrimaryKeyColumns     = []string{"id"}
)

type (
	// SaftAgreementSlice is an alias for a slice of pointers to SaftAgreement.
	// This should almost always be used instead of []SaftAgreement.
	SaftAgreementSlice []*SaftAgreement
	// SaftAgreementHook is the signature for custom SaftAgreement hook methods
	SaftAgreementHook func(boil.Executor, *SaftAgreement) error

	saftAgreementQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	saftAgreementType                 = reflect.TypeOf(&SaftAgreement{})
	saftAgreementMapping              = queries.MakeStructMapping(saftAgreementType)
	saftAgreementPrimaryKeyMapping, _ = queries.BindMapping(saftAgreementType, saftAgreementMapping, saftAgreementPrimaryKeyColumns)
	saftAgreementInsertCacheMut       sync.RWMutex
	saftAgreementInsertCache          = make(map[string]insertCache)
	saftAgreementUpdateCacheMut       sync.RWMutex
	saftAgreementUpdateCache          = make(map[string]updateCache)
	saftAgreementUpsertCacheMut       sync.RWMutex
	saftAgreementUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var saftAgreementBeforeInsertHooks []SaftAgreementHook
var saftAgreementBeforeUpdateHooks []SaftAgreementHook
var saftAgreementBeforeDeleteHooks []SaftAgreementHook
var saftAgreementBeforeUpsertHooks []SaftAgreementHook

var saftAgreementAfterInsertHooks []SaftAgreementHook
var saftAgreementAfterSelectHooks []SaftAgreementHook
var saftAgreementAfterUpdateHooks []SaftAgreementHook
var saftAgreementAfterDeleteHooks []SaftAgreementHook
var saftAgreementAfterUpsertHooks []SaftAgreementHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *SaftAgreement) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *SaftAgreement) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *SaftAgreement) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *SaftAgreement) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *SaftAgreement) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *SaftAgreement) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *SaftAgreement) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *SaftAgreement) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *SaftAgreement) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range saftAgreementAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddSaftAgreementHook registers your hook function for all future operations.
func AddSaftAgreementHook(hookPoint boil.HookPoint, saftAgreementHook SaftAgreementHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		saftAgreementBeforeInsertHooks = append(saftAgreementBeforeInsertHooks, saftAgreementHook)
	case boil.BeforeUpdateHook:
		saftAgreementBeforeUpdateHooks = append(saftAgreementBeforeUpdateHooks, saftAgreementHook)
	case boil.BeforeDeleteHook:
		saftAgreementBeforeDeleteHooks = append(saftAgreementBeforeDeleteHooks, saftAgreementHook)
	case boil.BeforeUpsertHook:
		saftAgreementBeforeUpsertHooks = append(saftAgreementBeforeUpsertHooks, saftAgreementHook)
	case boil.AfterInsertHook:
		saftAgreementAfterInsertHooks = append(saftAgreementAfterInsertHooks, saftAgreementHook)
	case boil.AfterSelectHook:
		saftAgreementAfterSelectHooks = append(saftAgreementAfterSelectHooks, saftAgreementHook)
	case boil.AfterUpdateHook:
		saftAgreementAfterUpdateHooks = append(saftAgreementAfterUpdateHooks, saftAgreementHook)
	case boil.AfterDeleteHook:
		saftAgreementAfterDeleteHooks = append(saftAgreementAfterDeleteHooks, saftAgreementHook)
	case boil.AfterUpsertHook:
		saftAgreementAfterUpsertHooks = append(saftAgreementAfterUpsertHooks, saftAgreementHook)
	}
}

// One returns a single saftAgreement record from the query.
func (q saftAgreementQuery) One(exec boil.Executor) (*SaftAgreement, error) {
	o := &SaftAgreement{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for saft_agreements")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all SaftAgreement records from the query.
func (q saftAgreementQuery) All(exec boil.Executor) (SaftAgreementSlice, error) {
	var o []*SaftAgreement

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to SaftAgreement slice")
	}

	if len(saftAgreementAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all SaftAgreement records in the query.
func (q saftAgreementQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count saft_agreements rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q saftAgreementQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if saft_agreements exists")
	}

	return count > 0, nil
}

// SaftAgreements retrieves all the records using an executor.
func SaftAgreements(mods ...qm.QueryMod) saftAgreementQuery {
	mods = append(mods, qm.From("\"saft_agreements\""), qmhelper.WhereIsNull("\"saft_agreements\".\"deleted_at\""))
	return saftAgreementQuery{NewQuery(mods...)}
}

// FindSaftAgreement retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindSaftAgreement(exec boil.Executor, iD string, selectCols ...string) (*SaftAgreement, error) {
	saftAgreementObj := &SaftAgreement{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"saft_agreements\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, saftAgreementObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from saft_agreements")
	}

	if err = saftAgreementObj.doAfterSelectHooks(exec); err != nil {
		return saftAgreementObj, err
	}

	return saftAgreementObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *SaftAgreement) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no saft_agreements provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(saftAgreementColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	saftAgreementInsertCacheMut.RLock()
	cache, cached := saftAgreementInsertCache[key]
	saftAgreementInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			saftAgreementAllColumns,
			saftAgreementColumnsWithDefault,
			saftAgreementColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(saftAgreementType, saftAgreementMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(saftAgreementType, saftAgreementMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"saft_agreements\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"saft_agreements\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into saft_agreements")
	}

	if !cached {
		saftAgreementInsertCacheMut.Lock()
		saftAgreementInsertCache[key] = cache
		saftAgreementInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the SaftAgreement.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *SaftAgreement) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	saftAgreementUpdateCacheMut.RLock()
	cache, cached := saftAgreementUpdateCache[key]
	saftAgreementUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			saftAgreementAllColumns,
			saftAgreementPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update saft_agreements, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"saft_agreements\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, saftAgreementPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(saftAgreementType, saftAgreementMapping, append(wl, saftAgreementPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update saft_agreements row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for saft_agreements")
	}

	if !cached {
		saftAgreementUpdateCacheMut.Lock()
		saftAgreementUpdateCache[key] = cache
		saftAgreementUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q saftAgreementQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for saft_agreements")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for saft_agreements")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o SaftAgreementSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), saftAgreementPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"saft_agreements\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, saftAgreementPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in saftAgreement slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all saftAgreement")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *SaftAgreement) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no saft_agreements provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(saftAgreementColumnsWithDefault, o)

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

	saftAgreementUpsertCacheMut.RLock()
	cache, cached := saftAgreementUpsertCache[key]
	saftAgreementUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			saftAgreementAllColumns,
			saftAgreementColumnsWithDefault,
			saftAgreementColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			saftAgreementAllColumns,
			saftAgreementPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert saft_agreements, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(saftAgreementPrimaryKeyColumns))
			copy(conflict, saftAgreementPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"saft_agreements\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(saftAgreementType, saftAgreementMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(saftAgreementType, saftAgreementMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert saft_agreements")
	}

	if !cached {
		saftAgreementUpsertCacheMut.Lock()
		saftAgreementUpsertCache[key] = cache
		saftAgreementUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single SaftAgreement record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *SaftAgreement) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no SaftAgreement provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), saftAgreementPrimaryKeyMapping)
		sql = "DELETE FROM \"saft_agreements\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"saft_agreements\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(saftAgreementType, saftAgreementMapping, append(wl, saftAgreementPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from saft_agreements")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for saft_agreements")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q saftAgreementQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no saftAgreementQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from saft_agreements")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for saft_agreements")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o SaftAgreementSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(saftAgreementBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), saftAgreementPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"saft_agreements\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, saftAgreementPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), saftAgreementPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"saft_agreements\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, saftAgreementPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from saftAgreement slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for saft_agreements")
	}

	if len(saftAgreementAfterDeleteHooks) != 0 {
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
func (o *SaftAgreement) Reload(exec boil.Executor) error {
	ret, err := FindSaftAgreement(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *SaftAgreementSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := SaftAgreementSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), saftAgreementPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"saft_agreements\".* FROM \"saft_agreements\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, saftAgreementPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in SaftAgreementSlice")
	}

	*o = slice

	return nil
}

// SaftAgreementExists checks if the SaftAgreement row exists.
func SaftAgreementExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"saft_agreements\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if saft_agreements exists")
	}

	return exists, nil
}
