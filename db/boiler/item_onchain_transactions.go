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

// ItemOnchainTransaction is an object representing the database table.
type ItemOnchainTransaction struct {
	ID              string    `boiler:"id" boil:"id" json:"id" toml:"id" yaml:"id"`
	PurchasedItemID string    `boiler:"purchased_item_id" boil:"purchased_item_id" json:"purchasedItemID" toml:"purchasedItemID" yaml:"purchasedItemID"`
	TXID            string    `boiler:"tx_id" boil:"tx_id" json:"txID" toml:"txID" yaml:"txID"`
	ContractAddr    string    `boiler:"contract_addr" boil:"contract_addr" json:"contractAddr" toml:"contractAddr" yaml:"contractAddr"`
	FromAddr        string    `boiler:"from_addr" boil:"from_addr" json:"fromAddr" toml:"fromAddr" yaml:"fromAddr"`
	ToAddr          string    `boiler:"to_addr" boil:"to_addr" json:"toAddr" toml:"toAddr" yaml:"toAddr"`
	DeletedAt       null.Time `boiler:"deleted_at" boil:"deleted_at" json:"deletedAt,omitempty" toml:"deletedAt" yaml:"deletedAt,omitempty"`
	UpdatedAt       time.Time `boiler:"updated_at" boil:"updated_at" json:"updatedAt" toml:"updatedAt" yaml:"updatedAt"`
	CreatedAt       time.Time `boiler:"created_at" boil:"created_at" json:"createdAt" toml:"createdAt" yaml:"createdAt"`

	R *itemOnchainTransactionR `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
	L itemOnchainTransactionL  `boiler:"-" boil:"-" json:"-" toml:"-" yaml:"-"`
}

var ItemOnchainTransactionColumns = struct {
	ID              string
	PurchasedItemID string
	TXID            string
	ContractAddr    string
	FromAddr        string
	ToAddr          string
	DeletedAt       string
	UpdatedAt       string
	CreatedAt       string
}{
	ID:              "id",
	PurchasedItemID: "purchased_item_id",
	TXID:            "tx_id",
	ContractAddr:    "contract_addr",
	FromAddr:        "from_addr",
	ToAddr:          "to_addr",
	DeletedAt:       "deleted_at",
	UpdatedAt:       "updated_at",
	CreatedAt:       "created_at",
}

var ItemOnchainTransactionTableColumns = struct {
	ID              string
	PurchasedItemID string
	TXID            string
	ContractAddr    string
	FromAddr        string
	ToAddr          string
	DeletedAt       string
	UpdatedAt       string
	CreatedAt       string
}{
	ID:              "item_onchain_transactions.id",
	PurchasedItemID: "item_onchain_transactions.purchased_item_id",
	TXID:            "item_onchain_transactions.tx_id",
	ContractAddr:    "item_onchain_transactions.contract_addr",
	FromAddr:        "item_onchain_transactions.from_addr",
	ToAddr:          "item_onchain_transactions.to_addr",
	DeletedAt:       "item_onchain_transactions.deleted_at",
	UpdatedAt:       "item_onchain_transactions.updated_at",
	CreatedAt:       "item_onchain_transactions.created_at",
}

// Generated where

var ItemOnchainTransactionWhere = struct {
	ID              whereHelperstring
	PurchasedItemID whereHelperstring
	TXID            whereHelperstring
	ContractAddr    whereHelperstring
	FromAddr        whereHelperstring
	ToAddr          whereHelperstring
	DeletedAt       whereHelpernull_Time
	UpdatedAt       whereHelpertime_Time
	CreatedAt       whereHelpertime_Time
}{
	ID:              whereHelperstring{field: "\"item_onchain_transactions\".\"id\""},
	PurchasedItemID: whereHelperstring{field: "\"item_onchain_transactions\".\"purchased_item_id\""},
	TXID:            whereHelperstring{field: "\"item_onchain_transactions\".\"tx_id\""},
	ContractAddr:    whereHelperstring{field: "\"item_onchain_transactions\".\"contract_addr\""},
	FromAddr:        whereHelperstring{field: "\"item_onchain_transactions\".\"from_addr\""},
	ToAddr:          whereHelperstring{field: "\"item_onchain_transactions\".\"to_addr\""},
	DeletedAt:       whereHelpernull_Time{field: "\"item_onchain_transactions\".\"deleted_at\""},
	UpdatedAt:       whereHelpertime_Time{field: "\"item_onchain_transactions\".\"updated_at\""},
	CreatedAt:       whereHelpertime_Time{field: "\"item_onchain_transactions\".\"created_at\""},
}

// ItemOnchainTransactionRels is where relationship names are stored.
var ItemOnchainTransactionRels = struct {
	PurchasedItem string
}{
	PurchasedItem: "PurchasedItem",
}

// itemOnchainTransactionR is where relationships are stored.
type itemOnchainTransactionR struct {
	PurchasedItem *PurchasedItem `boiler:"PurchasedItem" boil:"PurchasedItem" json:"PurchasedItem" toml:"PurchasedItem" yaml:"PurchasedItem"`
}

// NewStruct creates a new relationship struct
func (*itemOnchainTransactionR) NewStruct() *itemOnchainTransactionR {
	return &itemOnchainTransactionR{}
}

// itemOnchainTransactionL is where Load methods for each relationship are stored.
type itemOnchainTransactionL struct{}

var (
	itemOnchainTransactionAllColumns            = []string{"id", "purchased_item_id", "tx_id", "contract_addr", "from_addr", "to_addr", "deleted_at", "updated_at", "created_at"}
	itemOnchainTransactionColumnsWithoutDefault = []string{"purchased_item_id", "tx_id", "contract_addr", "from_addr", "to_addr", "deleted_at"}
	itemOnchainTransactionColumnsWithDefault    = []string{"id", "updated_at", "created_at"}
	itemOnchainTransactionPrimaryKeyColumns     = []string{"id"}
)

type (
	// ItemOnchainTransactionSlice is an alias for a slice of pointers to ItemOnchainTransaction.
	// This should almost always be used instead of []ItemOnchainTransaction.
	ItemOnchainTransactionSlice []*ItemOnchainTransaction
	// ItemOnchainTransactionHook is the signature for custom ItemOnchainTransaction hook methods
	ItemOnchainTransactionHook func(boil.Executor, *ItemOnchainTransaction) error

	itemOnchainTransactionQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	itemOnchainTransactionType                 = reflect.TypeOf(&ItemOnchainTransaction{})
	itemOnchainTransactionMapping              = queries.MakeStructMapping(itemOnchainTransactionType)
	itemOnchainTransactionPrimaryKeyMapping, _ = queries.BindMapping(itemOnchainTransactionType, itemOnchainTransactionMapping, itemOnchainTransactionPrimaryKeyColumns)
	itemOnchainTransactionInsertCacheMut       sync.RWMutex
	itemOnchainTransactionInsertCache          = make(map[string]insertCache)
	itemOnchainTransactionUpdateCacheMut       sync.RWMutex
	itemOnchainTransactionUpdateCache          = make(map[string]updateCache)
	itemOnchainTransactionUpsertCacheMut       sync.RWMutex
	itemOnchainTransactionUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var itemOnchainTransactionBeforeInsertHooks []ItemOnchainTransactionHook
var itemOnchainTransactionBeforeUpdateHooks []ItemOnchainTransactionHook
var itemOnchainTransactionBeforeDeleteHooks []ItemOnchainTransactionHook
var itemOnchainTransactionBeforeUpsertHooks []ItemOnchainTransactionHook

var itemOnchainTransactionAfterInsertHooks []ItemOnchainTransactionHook
var itemOnchainTransactionAfterSelectHooks []ItemOnchainTransactionHook
var itemOnchainTransactionAfterUpdateHooks []ItemOnchainTransactionHook
var itemOnchainTransactionAfterDeleteHooks []ItemOnchainTransactionHook
var itemOnchainTransactionAfterUpsertHooks []ItemOnchainTransactionHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *ItemOnchainTransaction) doBeforeInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionBeforeInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *ItemOnchainTransaction) doBeforeUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionBeforeUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *ItemOnchainTransaction) doBeforeDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionBeforeDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *ItemOnchainTransaction) doBeforeUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionBeforeUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *ItemOnchainTransaction) doAfterInsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionAfterInsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *ItemOnchainTransaction) doAfterSelectHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionAfterSelectHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *ItemOnchainTransaction) doAfterUpdateHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionAfterUpdateHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *ItemOnchainTransaction) doAfterDeleteHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionAfterDeleteHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *ItemOnchainTransaction) doAfterUpsertHooks(exec boil.Executor) (err error) {
	for _, hook := range itemOnchainTransactionAfterUpsertHooks {
		if err := hook(exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddItemOnchainTransactionHook registers your hook function for all future operations.
func AddItemOnchainTransactionHook(hookPoint boil.HookPoint, itemOnchainTransactionHook ItemOnchainTransactionHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		itemOnchainTransactionBeforeInsertHooks = append(itemOnchainTransactionBeforeInsertHooks, itemOnchainTransactionHook)
	case boil.BeforeUpdateHook:
		itemOnchainTransactionBeforeUpdateHooks = append(itemOnchainTransactionBeforeUpdateHooks, itemOnchainTransactionHook)
	case boil.BeforeDeleteHook:
		itemOnchainTransactionBeforeDeleteHooks = append(itemOnchainTransactionBeforeDeleteHooks, itemOnchainTransactionHook)
	case boil.BeforeUpsertHook:
		itemOnchainTransactionBeforeUpsertHooks = append(itemOnchainTransactionBeforeUpsertHooks, itemOnchainTransactionHook)
	case boil.AfterInsertHook:
		itemOnchainTransactionAfterInsertHooks = append(itemOnchainTransactionAfterInsertHooks, itemOnchainTransactionHook)
	case boil.AfterSelectHook:
		itemOnchainTransactionAfterSelectHooks = append(itemOnchainTransactionAfterSelectHooks, itemOnchainTransactionHook)
	case boil.AfterUpdateHook:
		itemOnchainTransactionAfterUpdateHooks = append(itemOnchainTransactionAfterUpdateHooks, itemOnchainTransactionHook)
	case boil.AfterDeleteHook:
		itemOnchainTransactionAfterDeleteHooks = append(itemOnchainTransactionAfterDeleteHooks, itemOnchainTransactionHook)
	case boil.AfterUpsertHook:
		itemOnchainTransactionAfterUpsertHooks = append(itemOnchainTransactionAfterUpsertHooks, itemOnchainTransactionHook)
	}
}

// One returns a single itemOnchainTransaction record from the query.
func (q itemOnchainTransactionQuery) One(exec boil.Executor) (*ItemOnchainTransaction, error) {
	o := &ItemOnchainTransaction{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(nil, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: failed to execute a one query for item_onchain_transactions")
	}

	if err := o.doAfterSelectHooks(exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all ItemOnchainTransaction records from the query.
func (q itemOnchainTransactionQuery) All(exec boil.Executor) (ItemOnchainTransactionSlice, error) {
	var o []*ItemOnchainTransaction

	err := q.Bind(nil, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "boiler: failed to assign all query results to ItemOnchainTransaction slice")
	}

	if len(itemOnchainTransactionAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all ItemOnchainTransaction records in the query.
func (q itemOnchainTransactionQuery) Count(exec boil.Executor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to count item_onchain_transactions rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q itemOnchainTransactionQuery) Exists(exec boil.Executor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRow(exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "boiler: failed to check if item_onchain_transactions exists")
	}

	return count > 0, nil
}

// PurchasedItem pointed to by the foreign key.
func (o *ItemOnchainTransaction) PurchasedItem(mods ...qm.QueryMod) purchasedItemQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.PurchasedItemID),
		qmhelper.WhereIsNull("deleted_at"),
	}

	queryMods = append(queryMods, mods...)

	query := PurchasedItems(queryMods...)
	queries.SetFrom(query.Query, "\"purchased_items\"")

	return query
}

// LoadPurchasedItem allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (itemOnchainTransactionL) LoadPurchasedItem(e boil.Executor, singular bool, maybeItemOnchainTransaction interface{}, mods queries.Applicator) error {
	var slice []*ItemOnchainTransaction
	var object *ItemOnchainTransaction

	if singular {
		object = maybeItemOnchainTransaction.(*ItemOnchainTransaction)
	} else {
		slice = *maybeItemOnchainTransaction.(*[]*ItemOnchainTransaction)
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &itemOnchainTransactionR{}
		}
		args = append(args, object.PurchasedItemID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &itemOnchainTransactionR{}
			}

			for _, a := range args {
				if a == obj.PurchasedItemID {
					continue Outer
				}
			}

			args = append(args, obj.PurchasedItemID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`purchased_items`),
		qm.WhereIn(`purchased_items.id in ?`, args...),
		qmhelper.WhereIsNull(`purchased_items.deleted_at`),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.Query(e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load PurchasedItem")
	}

	var resultSlice []*PurchasedItem
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice PurchasedItem")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for purchased_items")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for purchased_items")
	}

	if len(itemOnchainTransactionAfterSelectHooks) != 0 {
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
		object.R.PurchasedItem = foreign
		if foreign.R == nil {
			foreign.R = &purchasedItemR{}
		}
		foreign.R.ItemOnchainTransactions = append(foreign.R.ItemOnchainTransactions, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.PurchasedItemID == foreign.ID {
				local.R.PurchasedItem = foreign
				if foreign.R == nil {
					foreign.R = &purchasedItemR{}
				}
				foreign.R.ItemOnchainTransactions = append(foreign.R.ItemOnchainTransactions, local)
				break
			}
		}
	}

	return nil
}

// SetPurchasedItem of the itemOnchainTransaction to the related item.
// Sets o.R.PurchasedItem to related.
// Adds o to related.R.ItemOnchainTransactions.
func (o *ItemOnchainTransaction) SetPurchasedItem(exec boil.Executor, insert bool, related *PurchasedItem) error {
	var err error
	if insert {
		if err = related.Insert(exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"item_onchain_transactions\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"purchased_item_id"}),
		strmangle.WhereClause("\"", "\"", 2, itemOnchainTransactionPrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, updateQuery)
		fmt.Fprintln(boil.DebugWriter, values)
	}
	if _, err = exec.Exec(updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.PurchasedItemID = related.ID
	if o.R == nil {
		o.R = &itemOnchainTransactionR{
			PurchasedItem: related,
		}
	} else {
		o.R.PurchasedItem = related
	}

	if related.R == nil {
		related.R = &purchasedItemR{
			ItemOnchainTransactions: ItemOnchainTransactionSlice{o},
		}
	} else {
		related.R.ItemOnchainTransactions = append(related.R.ItemOnchainTransactions, o)
	}

	return nil
}

// ItemOnchainTransactions retrieves all the records using an executor.
func ItemOnchainTransactions(mods ...qm.QueryMod) itemOnchainTransactionQuery {
	mods = append(mods, qm.From("\"item_onchain_transactions\""), qmhelper.WhereIsNull("\"item_onchain_transactions\".\"deleted_at\""))
	return itemOnchainTransactionQuery{NewQuery(mods...)}
}

// FindItemOnchainTransaction retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindItemOnchainTransaction(exec boil.Executor, iD string, selectCols ...string) (*ItemOnchainTransaction, error) {
	itemOnchainTransactionObj := &ItemOnchainTransaction{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"item_onchain_transactions\" where \"id\"=$1 and \"deleted_at\" is null", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(nil, exec, itemOnchainTransactionObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "boiler: unable to select from item_onchain_transactions")
	}

	if err = itemOnchainTransactionObj.doAfterSelectHooks(exec); err != nil {
		return itemOnchainTransactionObj, err
	}

	return itemOnchainTransactionObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *ItemOnchainTransaction) Insert(exec boil.Executor, columns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no item_onchain_transactions provided for insertion")
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

	nzDefaults := queries.NonZeroDefaultSet(itemOnchainTransactionColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	itemOnchainTransactionInsertCacheMut.RLock()
	cache, cached := itemOnchainTransactionInsertCache[key]
	itemOnchainTransactionInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			itemOnchainTransactionAllColumns,
			itemOnchainTransactionColumnsWithDefault,
			itemOnchainTransactionColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(itemOnchainTransactionType, itemOnchainTransactionMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(itemOnchainTransactionType, itemOnchainTransactionMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"item_onchain_transactions\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"item_onchain_transactions\" %sDEFAULT VALUES%s"
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
		return errors.Wrap(err, "boiler: unable to insert into item_onchain_transactions")
	}

	if !cached {
		itemOnchainTransactionInsertCacheMut.Lock()
		itemOnchainTransactionInsertCache[key] = cache
		itemOnchainTransactionInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(exec)
}

// Update uses an executor to update the ItemOnchainTransaction.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *ItemOnchainTransaction) Update(exec boil.Executor, columns boil.Columns) (int64, error) {
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime

	var err error
	if err = o.doBeforeUpdateHooks(exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	itemOnchainTransactionUpdateCacheMut.RLock()
	cache, cached := itemOnchainTransactionUpdateCache[key]
	itemOnchainTransactionUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			itemOnchainTransactionAllColumns,
			itemOnchainTransactionPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("boiler: unable to update item_onchain_transactions, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"item_onchain_transactions\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, itemOnchainTransactionPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(itemOnchainTransactionType, itemOnchainTransactionMapping, append(wl, itemOnchainTransactionPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to update item_onchain_transactions row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by update for item_onchain_transactions")
	}

	if !cached {
		itemOnchainTransactionUpdateCacheMut.Lock()
		itemOnchainTransactionUpdateCache[key] = cache
		itemOnchainTransactionUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(exec)
}

// UpdateAll updates all rows with the specified column values.
func (q itemOnchainTransactionQuery) UpdateAll(exec boil.Executor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all for item_onchain_transactions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected for item_onchain_transactions")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o ItemOnchainTransactionSlice) UpdateAll(exec boil.Executor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), itemOnchainTransactionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"item_onchain_transactions\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, itemOnchainTransactionPrimaryKeyColumns, len(o)))

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, args...)
	}
	result, err := exec.Exec(sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to update all in itemOnchainTransaction slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to retrieve rows affected all in update all itemOnchainTransaction")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *ItemOnchainTransaction) Upsert(exec boil.Executor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("boiler: no item_onchain_transactions provided for upsert")
	}
	currTime := time.Now().In(boil.GetLocation())

	o.UpdatedAt = currTime
	if o.CreatedAt.IsZero() {
		o.CreatedAt = currTime
	}

	if err := o.doBeforeUpsertHooks(exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(itemOnchainTransactionColumnsWithDefault, o)

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

	itemOnchainTransactionUpsertCacheMut.RLock()
	cache, cached := itemOnchainTransactionUpsertCache[key]
	itemOnchainTransactionUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			itemOnchainTransactionAllColumns,
			itemOnchainTransactionColumnsWithDefault,
			itemOnchainTransactionColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			itemOnchainTransactionAllColumns,
			itemOnchainTransactionPrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("boiler: unable to upsert item_onchain_transactions, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(itemOnchainTransactionPrimaryKeyColumns))
			copy(conflict, itemOnchainTransactionPrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"item_onchain_transactions\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(itemOnchainTransactionType, itemOnchainTransactionMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(itemOnchainTransactionType, itemOnchainTransactionMapping, ret)
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
		return errors.Wrap(err, "boiler: unable to upsert item_onchain_transactions")
	}

	if !cached {
		itemOnchainTransactionUpsertCacheMut.Lock()
		itemOnchainTransactionUpsertCache[key] = cache
		itemOnchainTransactionUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(exec)
}

// Delete deletes a single ItemOnchainTransaction record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *ItemOnchainTransaction) Delete(exec boil.Executor, hardDelete bool) (int64, error) {
	if o == nil {
		return 0, errors.New("boiler: no ItemOnchainTransaction provided for delete")
	}

	if err := o.doBeforeDeleteHooks(exec); err != nil {
		return 0, err
	}

	var (
		sql  string
		args []interface{}
	)
	if hardDelete {
		args = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), itemOnchainTransactionPrimaryKeyMapping)
		sql = "DELETE FROM \"item_onchain_transactions\" WHERE \"id\"=$1"
	} else {
		currTime := time.Now().In(boil.GetLocation())
		o.DeletedAt = null.TimeFrom(currTime)
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"item_onchain_transactions\" SET %s WHERE \"id\"=$2",
			strmangle.SetParamNames("\"", "\"", 1, wl),
		)
		valueMapping, err := queries.BindMapping(itemOnchainTransactionType, itemOnchainTransactionMapping, append(wl, itemOnchainTransactionPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "boiler: unable to delete from item_onchain_transactions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by delete for item_onchain_transactions")
	}

	if err := o.doAfterDeleteHooks(exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q itemOnchainTransactionQuery) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("boiler: no itemOnchainTransactionQuery provided for delete all")
	}

	if hardDelete {
		queries.SetDelete(q.Query)
	} else {
		currTime := time.Now().In(boil.GetLocation())
		queries.SetUpdate(q.Query, M{"deleted_at": currTime})
	}

	result, err := q.Query.Exec(exec)
	if err != nil {
		return 0, errors.Wrap(err, "boiler: unable to delete all from item_onchain_transactions")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for item_onchain_transactions")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o ItemOnchainTransactionSlice) DeleteAll(exec boil.Executor, hardDelete bool) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(itemOnchainTransactionBeforeDeleteHooks) != 0 {
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
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), itemOnchainTransactionPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
		}
		sql = "DELETE FROM \"item_onchain_transactions\" WHERE " +
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, itemOnchainTransactionPrimaryKeyColumns, len(o))
	} else {
		currTime := time.Now().In(boil.GetLocation())
		for _, obj := range o {
			pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), itemOnchainTransactionPrimaryKeyMapping)
			args = append(args, pkeyArgs...)
			obj.DeletedAt = null.TimeFrom(currTime)
		}
		wl := []string{"deleted_at"}
		sql = fmt.Sprintf("UPDATE \"item_onchain_transactions\" SET %s WHERE "+
			strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 2, itemOnchainTransactionPrimaryKeyColumns, len(o)),
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
		return 0, errors.Wrap(err, "boiler: unable to delete all from itemOnchainTransaction slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "boiler: failed to get rows affected by deleteall for item_onchain_transactions")
	}

	if len(itemOnchainTransactionAfterDeleteHooks) != 0 {
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
func (o *ItemOnchainTransaction) Reload(exec boil.Executor) error {
	ret, err := FindItemOnchainTransaction(exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *ItemOnchainTransactionSlice) ReloadAll(exec boil.Executor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := ItemOnchainTransactionSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), itemOnchainTransactionPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"item_onchain_transactions\".* FROM \"item_onchain_transactions\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, itemOnchainTransactionPrimaryKeyColumns, len(*o)) +
		"and \"deleted_at\" is null"

	q := queries.Raw(sql, args...)

	err := q.Bind(nil, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "boiler: unable to reload all in ItemOnchainTransactionSlice")
	}

	*o = slice

	return nil
}

// ItemOnchainTransactionExists checks if the ItemOnchainTransaction row exists.
func ItemOnchainTransactionExists(exec boil.Executor, iD string) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"item_onchain_transactions\" where \"id\"=$1 and \"deleted_at\" is null limit 1)"

	if boil.DebugMode {
		fmt.Fprintln(boil.DebugWriter, sql)
		fmt.Fprintln(boil.DebugWriter, iD)
	}
	row := exec.QueryRow(sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "boiler: unable to check if item_onchain_transactions exists")
	}

	return exists, nil
}
