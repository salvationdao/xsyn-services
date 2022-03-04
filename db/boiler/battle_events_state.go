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

// BattleEventsState is an object representing the database table.
type BattleEventsState struct {
	ID      string      `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	EventID string      `boiler:"event_id" boil:"event_id" json:"eventID" toml:"eventID" yaml:"eventID"`
	State   null.String `boiler:"state" boil:"state" json:"state,omitempty" toml:"state" yaml:"state,omitempty"`
	Detail  types.JSON  `boiler:"detail" boil:"detail" json:"detail" toml:"detail" yaml:"detail"`

	R *battleEventsStateR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L battleEventsStateL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var BattleEventsStateColumns = struct {
	ID      string
	EventID string
	State   string
	Detail  string
}{
	ID:      "id",
	EventID: "event_id",
	State:   "state",
	Detail:  "detail",
}

var BattleEventsStateTableColumns = struct {
	ID      string
	EventID string
	State   string
	Detail  string
}{
	ID:      "battle_events_state.id",
	EventID: "battle_events_state.event_id",
	State:   "battle_events_state.state",
	Detail:  "battle_events_state.detail",
}

// Generated where

type whereHelpertypes_JSON struct{ field string }

func (w whereHelpertypes_JSON) EQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertypes_JSON) NEQ(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertypes_JSON) LT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertypes_JSON) LTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertypes_JSON) GT(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertypes_JSON) GTE(x types.JSON) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var BattleEventsStateWhere = struct {
	ID      whereHelperstring
	EventID whereHelperstring
	State   whereHelpernull_String
	Detail  whereHelpertypes_JSON
}{
	ID:      whereHelperstring{field: "\"battle_events_state\".\"id\""},
	EventID: whereHelperstring{field: "\"battle_events_state\".\"event_id\""},
	State:   whereHelpernull_String{field: "\"battle_events_state\".\"state\""},
	Detail:  whereHelpertypes_JSON{field: "\"battle_events_state\".\"detail\""},
}

// BattleEventsStateRels is where relationship names are stored.
var BattleEventsStateRels = struct {
	Event string
}{
	Event: "Event",
}

// battleEventsStateR is where relationships are stored.
type battleEventsStateR struct {
	Event *BattleEvent `boiler:"Event" boil:"Event" json:"Event" toml:"Event" yaml:"Event"`
}

// NewStruct creates a new relationship struct
func (*battleEventsStateR) NewStruct() *battleEventsStateR {
	return &battleEventsStateR{}
}

// battleEventsStateL is where Load methods for each relationship are stored.
type battleEventsStateL struct{}

var (
	battleEventsStateAllColumns            = []string{"id", "event_id", "state", "detail"}
	battleEventsStateColumnsWithoutDefault = []string{"event_id", "state", "detail"}
	battleEventsStateColumnsWithDefault    = []string{"id"}
	battleEventsStatePrimaryKeyColumns     = []string{"id"}
)

type (
	// BattleEventsStateSlice is an alias for a slice of pointers to BattleEventsState.
	// This should almost always be used instead of []BattleEventsState.
	BattleEventsStateSlice []*BattleEventsState
	// BattleEventsStateHook is the signature for custom BattleEventsState hook methods
	BattleEventsStateHook func(boil.Executor, *BattleEventsState) error

	battleEventsStateQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	battleEventsStateType                 = reflect.TypeOf(&BattleEventsState{})
	battleEventsStateMapping              = queries.MakeStructMapping(battleEventsStateType)
	battleEventsStatePrimaryKeyMapping, _ = queries.BindMapping(battleEventsStateType, battleEventsStateMapping, battleEventsStatePrimaryKeyColumns)
	battleEventsStateInsertCacheMut       sync.RWMutex
	battleEventsStateInsertCache          = make(map[string]insertCache)
	battleEventsStateUpdateCacheMut       sync.RWMutex
	battleEventsStateUpdateCache          = make(map[string]updateCache)
	battleEventsStateUpsertCacheMut       sync.RWMutex
	battleEventsStateUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var battleEventsStateBeforeInsertHooks []BattleEventsStateHook
var battleEventsStateBeforeUpdateHooks []BattleEventsStateHook
var battleEventsStateBeforeDeleteHooks []BattleEventsStateHook
var battleEventsStateBeforeUpsertHooks []BattleEventsStateHook

var battleEventsStateAfterInsertHooks []BattleEventsStateHook
var battleEventsStateAfterSelectHooks []BattleEventsStateHook
var battleEventsStateAfterUpdateHooks []BattleEventsStateHook
var battleEventsStateAfterDeleteHooks []BattleEventsStateHook
var battleEventsStateAfterUpsertHooks []BattleEventsStateHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *BattleEventsState) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *BattleEventsState) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *BattleEventsState) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *BattleEventsState) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *BattleEventsState) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *BattleEventsState) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *BattleEventsState) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *BattleEventsState) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *BattleEventsState) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range battleEventsStateAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddBattleEventsStateHook registers your hook function for all future operations.
func AddBattleEventsStateHook(hookPoint boil.HookPoint, battleEventsStateHook BattleEventsStateHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		battleEventsStateBeforeInsertHooks = append(battleEventsStateBeforeInsertHooks, battleEventsStateHook)
	case boil.BeforeUpdateHook:
		battleEventsStateBeforeUpdateHooks = append(battleEventsStateBeforeUpdateHooks, battleEventsStateHook)
	case boil.BeforeDeleteHook:
		battleEventsStateBeforeDeleteHooks = append(battleEventsStateBeforeDeleteHooks, battleEventsStateHook)
	case boil.BeforeUpsertHook:
		battleEventsStateBeforeUpsertHooks = append(battleEventsStateBeforeUpsertHooks, battleEventsStateHook)
	case boil.AfterInsertHook:
		battleEventsStateAfterInsertHooks = append(battleEventsStateAfterInsertHooks, battleEventsStateHook)
	case boil.AfterSelectHook:
		battleEventsStateAfterSelectHooks = append(battleEventsStateAfterSelectHooks, battleEventsStateHook)
	case boil.AfterUpdateHook:
		battleEventsStateAfterUpdateHooks = append(battleEventsStateAfterUpdateHooks, battleEventsStateHook)
	case boil.AfterDeleteHook:
		battleEventsStateAfterDeleteHooks = append(battleEventsStateAfterDeleteHooks, battleEventsStateHook)
	case boil.AfterUpsertHook:
		battleEventsStateAfterUpsertHooks = append(battleEventsStateAfterUpsertHooks, battleEventsStateHook)
	}
}

// One returns a single battleEventsState record from the query.
func (q battleEventsStateQuery) One(exec boil.Executor) (*BattleEventsState, error) {
	o := &BattleEventsState{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for battle_events_state")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all BattleEventsState records from the query.
func (q battleEventsStateQuery) All(exec boil.Executor) (BattleEventsStateSlice, error) {
	var o []*BattleEventsState

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to BattleEventsState slice")
	}

	if len(battleEventsStateAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all BattleEventsState records in the query.
func (q battleEventsStateQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count battle_events_state rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q battleEventsStateQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if battle_events_state exists")
	}

	return count > 0, nil
}

// Event pointed to by the foreign key.
func (o *BattleEventsState) Event(mods ...qm.QueryMod) battleEventQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.EventID),
	}

	queryMods = append(queryMods, mods...)

	query := BattleEvents(queryMods...)
	queries.SetFrom(query.Query, "\"battle_events\"")

	return query
}

// LoadEvent allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (battleEventsStateL) LoadEvent(e boil.Executor, singular bool, maybeBattleEventsState interface{}, mods queries.Applicator) error {
	var slice []*BattleEventsState
	var object *BattleEventsState

	if singular {
		object = maybeBattleEventsState.(*BattleEventsState)
	} else {
		slice = *maybeBattleEventsState.(*[]*BattleEventsState)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &battleEventsStateR{}
		}
		args = append(args, object.EventID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &battleEventsStateR{}
			}

			for _, a := range args {
				if a == obj.EventID {
					continue Outer
				}
			}

			args = append(args, obj.EventID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`battle_events`),
		qm.WhereIn(`battle_events.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load BattleEvent")
	}

	var resultSlice []*BattleEvent
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice BattleEvent")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for battle_events")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for battle_events")
	}

	if len(battleEventsStateAfterSelectHooks) != 0 {
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
		object.R.Event = foreign
		if foreign.R == nil {
			foreign.R = &battleEventR{}
		}
		foreign.R.EventBattleEventsStates = append(foreign.R.EventBattleEventsStates, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.EventID == foreign.ID {
				local.R.Event = foreign
				if foreign.R == nil {
					foreign.R = &battleEventR{}
				}
				foreign.R.EventBattleEventsStates = append(foreign.R.EventBattleEventsStates, local)
				break
			}
		}
	}

	return nil
}

// SetEvent of the battleEventsState to the related item.
// Sets o.R.Event to related.
// Adds o to related.R.EventBattleEventsStates.
func (o *BattleEventsState) SetEvent(exec boil.Executor, insert bool, related *BattleEvent) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"battle_events_state\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"event_id"}),
		strmangle.WhereClause("\"", "\"", 2, battleEventsStatePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.EventID = related.ID
	if o.R == nil {
		o.R = &battleEventsStateR{
			Event: related,
		}
	} else {
		o.R.Event = related
	}

	if related.R == nil {
		related.R = &battleEventR{
			EventBattleEventsStates: BattleEventsStateSlice{o},
		}
	} else {
		related.R.EventBattleEventsStates = append(related.R.EventBattleEventsStates, o)
	}

	return nil
}

// BattleEventsStates retrieves all the records using an executor.
func BattleEventsStates(mods ...qm.QueryMod) battleEventsStateQuery {
	mods = append(mods, qm.From("\"battle_events_state\""))
	return battleEventsStateQuery{NewQuery(mods...)}
}

// FindBattleEventsState retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindBattleEventsState(exec boil.Executor, iD string, selectCols ...string) (*BattleEventsState, error) {
	battleEventsStateObj := &BattleEventsState{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"battle_events_state\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, battleEventsStateObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from battle_events_state")
	}

	if err = battleEventsStateObj.doAfterSelectHooks(exec); err != nil {
		return battleEventsStateObj, err
	}

	return battleEventsStateObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *BattleEventsState) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_events_state provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleEventsStateColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	battleEventsStateInsertCacheMut.RLock()
	cache, cached := battleEventsStateInsertCache[key]
	battleEventsStateInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			battleEventsStateAllColumns,
			battleEventsStateColumnsWithDefault,
			battleEventsStateColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(battleEventsStateType, battleEventsStateMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(battleEventsStateType, battleEventsStateMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"battle_events_state\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"battle_events_state\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into battle_events_state")
	}

	if !cached {
		battleEventsStateInsertCacheMut.Lock()
		battleEventsStateInsertCache[key] = cache
		battleEventsStateInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the BattleEventsState.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *BattleEventsState) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	battleEventsStateUpdateCacheMut.RLock()
	cache, cached := battleEventsStateUpdateCache[key]
	battleEventsStateUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			battleEventsStateAllColumns,
			battleEventsStatePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update battle_events_state, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"battle_events_state\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, battleEventsStatePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(battleEventsStateType, battleEventsStateMapping, append(wl, battleEventsStatePrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update battle_events_state row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for battle_events_state")
	}

	if !cached {
		battleEventsStateUpdateCacheMut.Lock()
		battleEventsStateUpdateCache[key] = cache
		battleEventsStateUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q battleEventsStateQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for battle_events_state")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for battle_events_state")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o BattleEventsStateSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleEventsStatePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"battle_events_state\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, battleEventsStatePrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in battleEventsState slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all battleEventsState")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *BattleEventsState) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no battle_events_state provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(battleEventsStateColumnsWithDefault, o)

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

	battleEventsStateUpsertCacheMut.RLock()
	cache, cached := battleEventsStateUpsertCache[key]
	battleEventsStateUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			battleEventsStateAllColumns,
			battleEventsStateColumnsWithDefault,
			battleEventsStateColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			battleEventsStateAllColumns,
			battleEventsStatePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert battle_events_state, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(battleEventsStatePrimaryKeyColumns))
			copy(conflict, battleEventsStatePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"battle_events_state\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(battleEventsStateType, battleEventsStateMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(battleEventsStateType, battleEventsStateMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert battle_events_state")
	}

	if !cached {
		battleEventsStateUpsertCacheMut.Lock()
		battleEventsStateUpsertCache[key] = cache
		battleEventsStateUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single BattleEventsState record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *BattleEventsState) Delete(exec boil.Executor) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no BattleEventsState provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), battleEventsStatePrimaryKeyMapping)
	sql := "DELETE FROM \"battle_events_state\" WHERE \"id\"=$1"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete from battle_events_state")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for battle_events_state")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q battleEventsStateQuery) DeleteAll(exec boil.Executor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no battleEventsStateQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battle_events_state")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_events_state")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o BattleEventsStateSlice) DeleteAll(exec boil.Executor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(battleEventsStateBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleEventsStatePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"battle_events_state\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleEventsStatePrimaryKeyColumns, len(o))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from battleEventsState slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for battle_events_state")
	}

	if len(battleEventsStateAfterDeleteHooks) != 0 {
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
func (o *BattleEventsState) Reload(exec boil.Executor) error {
	ret, err := FindBattleEventsState(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *BattleEventsStateSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := BattleEventsStateSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), battleEventsStatePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"battle_events_state\".* FROM \"battle_events_state\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, battleEventsStatePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in BattleEventsStateSlice")
	}

	*o = slice

	return nil
}

// BattleEventsStateExists checks if the BattleEventsState row exists.
func BattleEventsStateExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"battle_events_state\" where \"id\"=$1 limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if battle_events_state exists")
	}

	return exists, nil
}
