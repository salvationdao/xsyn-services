package db

import (
	"context"
	"fmt"
	"passport"
	"strings"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/ninja-software/terror/v2"
)

type StoreColumn string

const (
	StoreColumnID                 StoreColumn = "id"
	StoreColumnFactionID          StoreColumn = "faction_id"
	StoreColumnName               StoreColumn = "name"
	StoreColumnCollectionID       StoreColumn = "collection_id"
	StoreColumnDescription        StoreColumn = "description"
	StoreColumnImage              StoreColumn = "image"
	StoreColumnAttributes         StoreColumn = "attributes"
	StoreColumnAdditionalMetadata StoreColumn = "additional_metadata"
	StoreColumnUsdCentCost        StoreColumn = "usd_cent_cost"
	StoreColumnAmountSold         StoreColumn = "amount_sold"
	StoreColumnAmountAvailable    StoreColumn = "amount_available"
	StoreColumnSoldAfter          StoreColumn = "sold_after"
	StoreColumnSoldBefore         StoreColumn = "sold_before"

	StoreColumnDeletedAt StoreColumn = "deleted_at"
	StoreColumnUpdatedAt StoreColumn = "updated_at"
	StoreColumnCreatedAt StoreColumn = "created_at"
)

func (ic StoreColumn) IsValid() error {
	switch ic {
	case
		StoreColumnID,
		StoreColumnFactionID,
		StoreColumnName,
		StoreColumnCollectionID,
		StoreColumnDescription,
		StoreColumnImage,
		StoreColumnAttributes,
		StoreColumnAdditionalMetadata,
		StoreColumnUsdCentCost,
		StoreColumnAmountSold,
		StoreColumnAmountAvailable,
		StoreColumnSoldAfter,
		StoreColumnSoldBefore,
		StoreColumnDeletedAt,
		StoreColumnUpdatedAt,
		StoreColumnCreatedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid store item column type"))
}

const StoreGetQuery string = `
SELECT 
row_to_json(c) as collection,
xsyn_store.id,
xsyn_store.faction_id,
xsyn_store.name,
xsyn_store.collection_id,
xsyn_store.description,
xsyn_store.image,
xsyn_store.attributes,
xsyn_store.additional_metadata,
xsyn_store.usd_cent_cost,
xsyn_store.amount_sold,
xsyn_store.amount_available,
xsyn_store.sold_after,
xsyn_store.sold_before,
xsyn_store.deleted_at,
xsyn_store.updated_at,
xsyn_store.created_at
` + StoreGetQueryFrom

const StoreGetQueryFrom = `
FROM xsyn_store 
INNER JOIN collections c ON xsyn_store.collection_id = c.id
`

// AddItemToStore added the object to the xsyn nft store table
func AddItemToStore(ctx context.Context, conn Conn, storeObject *passport.StoreItem) error {
	q := `INSERT INTO xsyn_store (faction_id,
                                      name,
                                      collection_id,
                                      description,
                                      image,
                                      attributes,
                                      usd_cent_cost,
                                      amount_sold,
                                      amount_available,
                                      sold_after,
                                      sold_before)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id`

	_, err := conn.Exec(ctx, q,
		storeObject.FactionID,
		storeObject.Name,
		storeObject.CollectionID,
		storeObject.Description,
		storeObject.Image,
		storeObject.Attributes,
		storeObject.UsdCentCost,
		storeObject.AmountSold,
		storeObject.AmountAvailable,
		storeObject.SoldAfter,
		storeObject.SoldAfter,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// StoreItemByID get store item by id
func StoreItemByID(ctx context.Context, conn Conn, storeItemID passport.StoreItemID) (*passport.StoreItem, error) {
	storeItem := &passport.StoreItem{}
	q := `SELECT 	xsyn_store.id, 
					xsyn_store.faction_id,
					xsyn_store.name,
					xsyn_store.collection_id,
					xsyn_store.description,
					xsyn_store.image,
					xsyn_store.attributes,
					xsyn_store.usd_cent_cost,
					xsyn_store.amount_sold,
					xsyn_store.amount_available,
					xsyn_store.sold_after,
					xsyn_store.sold_before,
					row_to_json(faction) as faction
			FROM xsyn_store
			LEFT JOIN (
				SELECT id, label, theme, logo_blob_id as logoBlobID
				FROM factions
			) faction ON faction.id = xsyn_store.faction_id
			WHERE xsyn_store.id = $1`

	err := pgxscan.Get(ctx, conn, storeItem, q, storeItemID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return storeItem, nil
}

// StoreItemPurchased bumps a store items amount sold up
func StoreItemPurchased(ctx context.Context, conn Conn, storeItem *passport.StoreItem) error {
	q := `UPDATE xsyn_store SET amount_sold = amount_sold + 1 WHERE id = $1 RETURNING amount_sold`

	err := pgxscan.Get(ctx, conn, storeItem, q, storeItem.ID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// StoreItemListByCollectionAndFaction returns a list of store item IDs
func StoreItemListByCollectionAndFaction(ctx context.Context, conn Conn,
	collectionID passport.CollectionID,
	factionID passport.FactionID,
	page int,
	pageSize int,
) ([]*passport.StoreItemID, error) {
	storeItems := []*passport.StoreItemID{}

	q := `	SELECT 	id
			FROM xsyn_store
			WHERE collection_id = $1
			AND faction_id = $2
			LIMIT $3 OFFSET $4`

	err := pgxscan.Select(ctx, conn, &storeItems, q, collectionID, factionID, pageSize, page*pageSize)
	if err != nil {
		return nil, terror.Error(err)
	}

	return storeItems, nil
}

// StoreList gets a list of store items depending on the filters
func StoreList(
	ctx context.Context,
	conn Conn,
	search string,
	archived bool,
	includedStoreItemIDs []passport.StoreItemID,
	filter *ListFilterRequest,
	assetType string,
	offset int,
	pageSize int,
	sortBy StoreColumn,
	sortDir SortByDir,
) (int, []*passport.StoreItem, error) {
	// Prepare Filters
	var args []interface{}

	filterConditionsString := ""
	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			column := StoreColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			condition, value := GenerateListFilterSQL(f.ColumnField, f.Value, f.OperatorValue, i+1)
			if condition != "" {
				filterConditions = append(filterConditions, condition)
				args = append(args, value)
			}
		}
		if len(filterConditions) > 0 {
			filterConditionsString = "AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	// asset type filter
	if assetType != "" {
		filterConditionsString += fmt.Sprintf(`
		AND xsyn_store.attributes @> '[{"trait_type": "Asset Type"}]' 
        AND xsyn_store.attributes @> '[{"value": "%s"}]' `, assetType)
	}

	// select specific assets via tokenIDs
	if includedStoreItemIDs != nil {
		cond := "("
		for i, storeItemID := range includedStoreItemIDs {
			cond += fmt.Sprintf("%s", storeItemID.String())
			if i < len(includedStoreItemIDs)-1 {
				cond += ","
				continue
			}

			cond += ")"
		}
		filterConditionsString += fmt.Sprintf(" AND xsyn_store.id  IN %v", cond)
	}

	archiveCondition := "IS NULL"
	if archived {
		archiveCondition = "IS NOT NULL"
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			searchCondition = fmt.Sprintf(" AND xsyn_store.keywords @@ to_tsquery($%d)", len(args))
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT xsyn_store.id)
		%s
		WHERE xsyn_store.deleted_at %s
			%s
			%s
		`,
		StoreGetQueryFrom,
		archiveCondition,
		filterConditionsString,
		searchCondition,
	)

	var totalRows int

	err := pgxscan.Get(ctx, conn, &totalRows, countQ, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	if totalRows == 0 {
		return 0, make([]*passport.StoreItem, 0), nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, nil, terror.Error(err)
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(
		StoreGetQuery+`--sql
		WHERE xsyn_store.deleted_at %s
			%s
			%s
		%s
		%s`,
		archiveCondition,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := make([]*passport.StoreItem, 0)
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	return totalRows, result, nil
}
