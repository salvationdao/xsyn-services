package db

import (
	"context"
	"fmt"
	"passport"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

const CollectionGetQuery string = `
SELECT 
DISTINCT collections.name,
collections.id,
collections.deleted_at,
collections.updated_at,
collections.created_at
` + CollectionGetQueryFrom

const CollectionGetQueryFrom = `
FROM collections
LEFT OUTER JOIN xsyn_metadata ON collections.id = xsyn_metadata.collection_id 
LEFT OUTER JOIN xsyn_assets ON xsyn_assets.token_id = xsyn_metadata.token_id 
`

type CollectionColumn string

const (
	CollectionColumnID     CollectionColumn = "id"
	CollectionColumnName   CollectionColumn = "name"
	CollectionColumnUserID CollectionColumn = "user_id"

	CollectionColumnDeletedAt CollectionColumn = "deleted_at"
	CollectionColumnUpdatedAt CollectionColumn = "updated_at"
	CollectionColumnCreatedAt CollectionColumn = "created_at"
)

func (cc CollectionColumn) IsValid() error {
	switch cc {
	case CollectionColumnID,
		CollectionColumnName,
		CollectionColumnUserID,
		CollectionColumnDeletedAt,
		CollectionColumnUpdatedAt,
		CollectionColumnCreatedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid collection column type"))
}

// CollectionGet returns a collection by name
func CollectionGet(ctx context.Context, conn Conn, name string) (*passport.Collection, error) {
	collection := &passport.Collection{}
	q := CollectionGetQuery + `WHERE collections.name = $1`

	err := pgxscan.Get(ctx, conn, collection, q, name)
	if err != nil {
		return nil, terror.Error(err, "Issue getting collection.")
	}
	return collection, nil
}

// CollectionsList gets a list of collections depending on the filters
func CollectionsList(
	ctx context.Context,
	conn Conn,
	result *[]*passport.Collection,
	search string,
	archived bool,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy CollectionColumn,
	sortDir SortByDir,
) (int, error) {
	// Prepare Filters
	var args []interface{}

	filterConditionsString := ""
	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			column := CollectionColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, terror.Error(err)
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

	archiveCondition := "IS NULL"
	if archived {
		archiveCondition = "IS NOT NULL"
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			searchCondition = fmt.Sprintf(" AND collections.keywords @@ to_tsquery($%d)", len(args))
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT collections.id)
		%s
		WHERE collections.deleted_at %s
			%s
			%s
		`,
		CollectionGetQueryFrom,
		archiveCondition,
		filterConditionsString,
		searchCondition,
	)

	var totalRows int
	err := pgxscan.Get(ctx, conn, &totalRows, countQ, args...)
	if err != nil {
		return 0, terror.Error(err)
	}
	if totalRows == 0 {
		return 0, nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, terror.Error(err)
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(
		CollectionGetQuery+`--sql
		WHERE collections.deleted_at %s
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
	err = pgxscan.Select(ctx, conn, result, q, args...)
	if err != nil {
		return 0, terror.Error(err)
	}
	return totalRows, nil
}
