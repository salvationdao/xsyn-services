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
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// BattleWarMachineQueue is an object representing the database table.
type BattleWarMachineQueue struct {
	WarMachineMetadata types.JSON  `boiler:"war_machine_metadata" boil:"war_machine_metadata" json:"warMachineMetadata" toml:"warMachineMetadata" yaml:"warMachineMetadata"`
	QueuedAt           time.Time   `boiler:"queued_at" boil:"queued_at" json:"queuedAt" toml:"queuedAt" yaml:"queuedAt"`
	DeletedAt          null.Time   `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	IsInsured          bool        `boiler:"is_insured" boil:"is_insured" json:"isInsured" toml:"isInsured" yaml:"isInsured"`
	ContractReward     string      `boiler:"contract_reward" boil:"contract_reward" json:"contractReward" toml:"contractReward" yaml:"contractReward"`
	Fee                string      `boiler:"fee" boil:"fee" json:"fee" toml:"fee" yaml:"fee"`
	WarMachineHash     null.String `boiler:"war_machine_hash" boil:"war_machine_hash" json:"warMachineHash,omitempty" toml:"warMachineHash" yaml:"warMachineHash,omitempty"`
	FactionID          null.String `boiler:"faction_id" boil:"faction_id" json:"factionID,omitempty" toml:"factionID" yaml:"factionID,omitempty"`
	CreatedAt          null.Time   `boiler:"created_at" boil:"created_at" json:"createdAt,omitempty" toml:"createdAt" yaml:"createdAt,omitempty"`
	ID                 string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`

	R *battleWarMachineQueueR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L battleWarMachineQueueL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BattleWarMachineQueueColumns = struct {
	WarMachineMetadata string
	QueuedAt           string
	DeletedAt          string
	IsInsured          string
	ContractReward     string
	Fee                string
	WarMachineHash     string
	FactionID          string
	CreatedAt          string
	ID                 string
}{
	WarMachineMetadata: "war_machine_metadata",
	QueuedAt:           "queued_at",
	DeletedAt:          "deleted_at",
	IsInsured:          "is_insured",
	ContractReward:     "contract_reward",
	Fee:                "fee",
	WarMachineHash:     "war_machine_hash",
	FactionID:          "faction_id",
	CreatedAt:          "created_at",
	ID:                 "id",
}

var BattleWarMachineQueueTableColumns = struct {
	WarMachineMetadata string
	QueuedAt           string
	DeletedAt          string
	IsInsured          string
	ContractReward     string
	Fee                string
	WarMachineHash     string
	FactionID          string
	CreatedAt          string
	ID                 string
}{
	WarMachineMetadata: "battle_war_machine_queues.war_machine_metadata",
	QueuedAt:           "battle_war_machine_queues.queued_at",
	DeletedAt:          "battle_war_machine_queues.deleted_at",
	IsInsured:          "battle_war_machine_queues.is_insured",
	ContractReward:     "battle_war_machine_queues.contract_reward",
	Fee:                "battle_war_machine_queues.fee",
	WarMachineHash:     "battle_war_machine_queues.war_machine_hash",
	FactionID:          "battle_war_machine_queues.faction_id",
	CreatedAt:          "battle_war_machine_queues.created_at",
	ID:                 "battle_war_machine_queues.id",
}

// Generated where

var BattleWarMachineQueueWhere = struct {
	WarMachineMetadata whereHelpertypes_JSON
	QueuedAt           whereHelpertime_Time
	DeletedAt          whereHelpernull_Time
	IsInsured          whereHelperbool
	ContractReward     whereHelperstring
	Fee                whereHelperstring
	WarMachineHash     whereHelpernull_String
	FactionID          whereHelpernull_String
	CreatedAt          whereHelpernull_Time
	ID                 whereHelperstring
}{
	WarMachineMetadata: whereHelpertypes_JSON{field: "\"battle_war_machine_queues\".\"war_machine_metadata\""},
	QueuedAt:           whereHelpertime_Time{field: "\"battle_war_machine_queues\".\"queued_at\""},
	DeletedAt:          whereHelpernull_Time{field: "\"battle_war_machine_queues\".\"deleted_at\""},
	IsInsured:          whereHelperbool{field: "\"battle_war_machine_queues\".\"is_insured\""},
	ContractReward:     whereHelperstring{field: "\"battle_war_machine_queues\".\"contract_reward\""},
	Fee:                whereHelperstring{field: "\"battle_war_machine_queues\".\"fee\""},
	WarMachineHash:     whereHelpernull_String{field: "\"battle_war_machine_queues\".\"war_machine_hash\""},
	FactionID:          whereHelpernull_String{field: "\"battle_war_machine_queues\".\"faction_id\""},
	CreatedAt:          whereHelpernull_Time{field: "\"battle_war_machine_queues\".\"created_at\""},
	ID:                 whereHelperstring{field: "\"battle_war_machine_queues\".\"id\""},
}

// BattleWarMachineQueueRels is where relationship names are stored.
var BattleWarMachineQueueRels = struct {
}{}

// battleWarMachineQueueR is where relationships are stored.
type battleWarMachineQueueR struct {
}

// NewStruct creates a new relationship struct
func (*battleWarMachineQueueR) NewStruct() *battleWarMachineQueueR {
	return &battleWarMachineQueueR{}
}

// battleWarMachineQueueL is where Load methods for each relationship are stored.
type battleWarMachineQueueL struct{}

var (
	battleWarMachineQueueAllColumns            = []string{"war_machine_metadata", "queued_at", "deleted_at", "is_insured", "contract_reward", "fee", "war_machine_hash", "faction_id", "created_at", "id"}
	battleWarMachineQueueColumnsWithoutDefault = []string{"war_machine_metadata", "deleted_at", "war_machine_hash", "faction_id"}
	battleWarMachineQueueColumnsWithDefault    = []string{"queued_at", "is_insured", "contract_reward", "fee", "created_at", "id"}
	battleWarMachineQueuePrimaryKeyColumns     = []string{"id"}
)

type (
	// BattleWarMachineQueueSlice is an alias for a slice of pointers to BattleWarMachineQueue.
	// This should almost always be used instead of []BattleWarMachineQueue.
	BattleWarMachineQueueSlice []*BattleWarMachineQueue
	// BattleWarMachineQueueHook is the signature for custom BattleWarMachineQueue hook methods
	BattleWarMachineQueueHook func(boil.Executor, *BattleWarMachineQueue) error

	battleWarMachineQueueQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	battleWarMachineQueueType                 = reflect.TypeOf(&BattleWarMachineQueue{})
	battleWarMachineQueueMapping              = queries.MakeStructMapping(battleWarMachineQueueType)
	battleWarMachineQueuePrimaryKeyMapping, _ = queries.BindMapping(battleWarMachineQueueType, battleWarMachineQueueMapping, battleWarMachineQueuePrimaryKeyColumns)
	battleWarMachineQueueInsertCacheMut       sync.RWMutex
	battleWarMachineQueueInsertCache          = make(map[string]insertCache)
	battleWarMachineQueueUpdateCacheMut       sync.RWMutex
	battleWarMachineQueueUpdateCache          = make(map[string]updateCache)
	battleWarMachineQueueUpsertCacheMut       sync.RWMutex
	battleWarMachineQueueUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var battleWarMachineQueueBeforeInsertHooks []BattleWarMachineQueueHook
var battleWarMachineQueueBeforeUpdateHooks []BattleWarMachineQueueHook
var battleWarMachineQueueBeforeDeleteHooks []BattleWarMachineQueueHook
var battleWarMachineQueueBeforeUpsertHooks []BattleWarMachineQueueHook

var battleWarMachineQueueAfterInsertHooks []BattleWarMachineQueueHook
var battleWarMachineQueueAfterSelectHooks []BattleWarMachineQueueHook
var battleWarMachineQueueAfterUpdateHooks []BattleWarMachineQueueHook
var battleWarMachineQueueAfterDeleteHooks []BattleWarMachineQueueHook
var battleWarMachineQueueAfterUpsertHooks []BattleWarMachineQueueHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *BattleWarMachineQueue) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *BattleWarMachineQueue) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *BattleWarMachineQueue) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *BattleWarMachineQueue) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *BattleWarMachineQueue) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *BattleWarMachineQueue) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *BattleWarMachineQueue) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *BattleWarMachineQueue) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *BattleWarMachineQueue) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleWarMachineQueueAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddBattleWarMachineQueueHook registers your hook function for all future operations.
func AddBattleWarMachineQueueHook(hookPoint boil.HookPoint, battleWarMachineQueueHook BattleWarMachineQueueHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		battleWarMachineQueueBeforeInsertHooks = append(battleWarMachineQueueBeforeInsertHooks, battleWarMachineQueueHook)
	case boil.BeforeUpdateHook:
		battleWarMachineQueueBeforeUpdateHooks = append(battleWarMachineQueueBeforeUpdateHooks, battleWarMachineQueueHook)
	case boil.BeforeDeleteHook:
		battleWarMachineQueueBeforeDeleteHooks = append(battleWarMachineQueueBeforeDeleteHooks, battleWarMachineQueueHook)
	case boil.BeforeUpsertHook:
		battleWarMachineQueueBeforeUpsertHooks = append(battleWarMachineQueueBeforeUpsertHooks, battleWarMachineQueueHook)
	case boil.AfterInsertHook:
		battleWarMachineQueueAfterInsertHooks = append(battleWarMachineQueueAfterInsertHooks, battleWarMachineQueueHook)
	case boil.AfterSelectHook:
		battleWarMachineQueueAfterSelectHooks = append(battleWarMachineQueueAfterSelectHooks, battleWarMachineQueueHook)
	case boil.AfterUpdateHook:
		battleWarMachineQueueAfterUpdateHooks = append(battleWarMachineQueueAfterUpdateHooks, battleWarMachineQueueHook)
	case boil.AfterDeleteHook:
		battleWarMachineQueueAfterDeleteHooks = append(battleWarMachineQueueAfterDeleteHooks, battleWarMachineQueueHook)
	case boil.AfterUpsertHook:
		battleWarMachineQueueAfterUpsertHooks = append(battleWarMachineQueueAfterUpsertHooks, battleWarMachineQueueHook)
	}
}

// One returns a single battleWarMachineQueue record from the query.
func (q battleWarMachineQueueQuery) One(exec boil.Executor) (*BattleWarMachineQueue, error) {
	o := &BattleWarMachineQueue{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for battle_war_machine_queues")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all BattleWarMachineQueue records from the query.
func (q battleWarMachineQueueQuery) All(exec boil.Executor) (BattleWarMachineQueueSlice, error) {
	var o []*BattleWarMachineQueue

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to BattleWarMachineQueue slice")
	}

	if len(battleWarMachineQueueAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all BattleWarMachineQueue records in the query.
func (q battleWarMachineQueueQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count battle_war_machine_queues rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q battleWarMachineQueueQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if battle_war_machine_queues exists")
	}

	return count > 0, nil
}

// BattleWarMachineQueues retrieves all the records using an executor.
func BattleWarMachineQueues(mods ...qm.QueryMod) battleWarMachineQueueQuery {
	mods = append(mods, qm.From("\"battle_war_machine_queues\""))
	return battleWarMachineQueueQuery{NewQuery(mods...)}
}

// FindBattleWarMachineQueue retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBattleWarMachineQueue(exec boil.Executor, iD string, selectCols ...string) (*BattleWarMachineQueue, error) {
	battleWarMachineQueueObj := &BattleWarMachineQueue{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"battle_war_machine_queues\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, battleWarMachineQueueObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from battle_war_machine_queues")
	}

	if err = battleWarMachineQueueObj.doAfterSelectHooks(exec); err != nil {
		return battleWarMachineQueueObj, err
	}

	return battleWarMachineQueueObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BattleWarMachineQueue) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_war_machine_queues provided for insertion")
	}

	var err error
	currTime := time.Now().In(boil.GetLocation())

	if queries.MustTime(o.CreatedAt).IsZero() {
		queries.SetScanner(&o.CreatedAt, currTime)
	}

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleWarMachineQueueColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	battleWarMachineQueueInsertCacheMut.RLock()
	cache, cached := battleWarMachineQueueInsertCache[key]
	battleWarMachineQueueInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			battleWarMachineQueueAllColumns,
			battleWarMachineQueueColumnsWithDefault,
			battleWarMachineQueueColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(battleWarMachineQueueType, battleWarMachineQueueMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(battleWarMachineQueueType, battleWarMachineQueueMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"battle_war_machine_queues\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"battle_war_machine_queues\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into battle_war_machine_queues")
	}

	if !cached {
		battleWarMachineQueueInsertCacheMut.Lock()
		battleWarMachineQueueInsertCache[key] = cache
		battleWarMachineQueueInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the BattleWarMachineQueue.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BattleWarMachineQueue) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	battleWarMachineQueueUpdateCacheMut.RLock()
	cache, cached := battleWarMachineQueueUpdateCache[key]
	battleWarMachineQueueUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			battleWarMachineQueueAllColumns,
			battleWarMachineQueuePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update battle_war_machine_queues, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"battle_war_machine_queues\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, battleWarMachineQueuePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(battleWarMachineQueueType, battleWarMachineQueueMapping, append(wl, battleWarMachineQueuePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update battle_war_machine_queues row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for battle_war_machine_queues")
	}

	if !cached {
		battleWarMachineQueueUpdateCacheMut.Lock()
		battleWarMachineQueueUpdateCache[key] = cache
		battleWarMachineQueueUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q battleWarMachineQueueQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for battle_war_machine_queues")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for battle_war_machine_queues")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BattleWarMachineQueueSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleWarMachineQueuePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"battle_war_machine_queues\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, battleWarMachineQueuePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in battleWarMachineQueue slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all battleWarMachineQueue")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BattleWarMachineQueue) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_war_machine_queues provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	if queries.MustTime(o.CreatedAt).IsZero() {
		queries.SetScanner(&o.CreatedAt, currTime)
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleWarMachineQueueColumnsWithDefault, o)

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

	battleWarMachineQueueUpsertCacheMut.RLock()
	cache, cached := battleWarMachineQueueUpsertCache[key]
	battleWarMachineQueueUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			battleWarMachineQueueAllColumns,
			battleWarMachineQueueColumnsWithDefault,
			battleWarMachineQueueColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			battleWarMachineQueueAllColumns,
			battleWarMachineQueuePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert battle_war_machine_queues, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(battleWarMachineQueuePrimaryKeyColumns))
			copy(conflict, battleWarMachineQueuePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"battle_war_machine_queues\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(battleWarMachineQueueType, battleWarMachineQueueMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(battleWarMachineQueueType, battleWarMachineQueueMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert battle_war_machine_queues")
	}

	if !cached {
		battleWarMachineQueueUpsertCacheMut.Lock()
		battleWarMachineQueueUpsertCache[key] = cache
		battleWarMachineQueueUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single BattleWarMachineQueue record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BattleWarMachineQueue) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no BattleWarMachineQueue provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), battleWarMachineQueuePrimaryKeyMapping)
	sql := "DELETE FROM \"battle_war_machine_queues\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from battle_war_machine_queues")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for battle_war_machine_queues")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q battleWarMachineQueueQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no battleWarMachineQueueQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battle_war_machine_queues")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_war_machine_queues")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BattleWarMachineQueueSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(battleWarMachineQueueBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleWarMachineQueuePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"battle_war_machine_queues\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleWarMachineQueuePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battleWarMachineQueue slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_war_machine_queues")
	}

	if len(battleWarMachineQueueAfterDeleteHooks) != 0 {
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
func (o *BattleWarMachineQueue) Reload(exec boil.Executor) error {
	ret, err := FindBattleWarMachineQueue(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BattleWarMachineQueueSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BattleWarMachineQueueSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleWarMachineQueuePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"battle_war_machine_queues\".* FROM \"battle_war_machine_queues\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleWarMachineQueuePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in BattleWarMachineQueueSlice")
	}

	*o = slice

	return nil
}

// BattleWarMachineQueueExists checks if the BattleWarMachineQueue row exists.
func BattleWarMachineQueueExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"battle_war_machine_queues\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if battle_war_machine_queues exists")
	}

	return exists, nil
}
