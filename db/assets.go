package db

import (
	"context"
	"fmt"
	"passport"
	"strings"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

type AssetColumn string

const (
	AssetColumnID          AssetColumn = "id"
	AssetColumnTokenID     AssetColumn = "token_id"
	AssetColumnUserID      AssetColumn = "user_id"
	AssetColumnName        AssetColumn = "name"
	AssetColumnCollection  AssetColumn = "collection"
	AssetColumnDescription AssetColumn = "description"
	AssetColumnExternalUrl AssetColumn = "external_url"
	AssetColumnImage       AssetColumn = "image"
	AssetColumnAttributes  AssetColumn = "attributes"

	AssetColumnDeletedAt AssetColumn = "deleted_at"
	AssetColumnUpdatedAt AssetColumn = "updated_at"
	AssetColumnCreatedAt AssetColumn = "created_at"
)

func (ic AssetColumn) IsValid() error {
	switch ic {
	case AssetColumnID,
		AssetColumnTokenID,
		AssetColumnUserID,
		AssetColumnName,
		AssetColumnCollection,
		AssetColumnDescription,
		AssetColumnExternalUrl,
		AssetColumnImage,
		AssetColumnAttributes,
		AssetColumnDeletedAt,
		AssetColumnUpdatedAt,
		AssetColumnCreatedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid asset column type"))
}

const AssetGetQuery string = `
SELECT 
xnm.token_id,
xnm.name,
xnm.collection,
xnm.description,
xnm.external_url,
xnm.image,
xnm.attributes,
xnm.deleted_at,
xnm.updated_at,
xnm.created_at
` + AessetGetQueryFrom
const AessetGetQueryFrom = `
FROM xsyn_assets xa
INNER JOIN xsyn_nft_metadata xnm ON xnm.token_id = xa.token_id
`

// AssetList gets a list of assets depending on the filters
func AssetList(
	ctx context.Context,
	conn Conn,
	result *[]*passport.Asset,
	search string,
	archived bool,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy AssetColumn,
	sortDir SortByDir,
) (int, error) {
	// Prepare Filters
	var args []interface{}

	filterConditionsString := ""
	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			column := AssetColumn(f.ColumnField)
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
			searchCondition = fmt.Sprintf(" AND assets.keywords @@ to_tsquery($%d)", len(args))
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT xa.token_id)
		%s
		WHERE xnm.deleted_at %s
			%s
			%s
		`,
		AessetGetQueryFrom,
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
		AssetGetQuery+`--sql
		WHERE xnm.deleted_at %s
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
