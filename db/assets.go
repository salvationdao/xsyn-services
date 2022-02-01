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
	AssetColumnID           AssetColumn = "id"
	AssetColumnTokenID      AssetColumn = "token_id"
	AssetColumnUserID       AssetColumn = "user_id"
	AssetColumnCollectionID AssetColumn = "collection_id"
	AssetColumnName         AssetColumn = "name"
	AssetColumnDescription  AssetColumn = "description"
	AssetColumnExternalUrl  AssetColumn = "external_url"
	AssetColumnImage        AssetColumn = "image"
	AssetColumnAttributes   AssetColumn = "attributes"

	AssetColumnDeletedAt AssetColumn = "deleted_at"
	AssetColumnUpdatedAt AssetColumn = "updated_at"
	AssetColumnCreatedAt AssetColumn = "created_at"
)

func (ic AssetColumn) IsValid() error {
	switch ic {
	case AssetColumnID,
		AssetColumnTokenID,
		AssetColumnUserID,
		AssetColumnCollectionID,
		AssetColumnName,
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
xsyn_nft_metadata.token_id,
xsyn_nft_metadata.name,
xsyn_nft_metadata.description,
xsyn_nft_metadata.external_url,
xsyn_nft_metadata.image,
xsyn_nft_metadata.attributes,
xsyn_nft_metadata.deleted_at,
xsyn_nft_metadata.updated_at,
xsyn_nft_metadata.created_at,
xsyn_assets.user_id,
xsyn_assets.frozen_at
` + AessetGetQueryFrom

const AessetGetQueryFrom = `
FROM xsyn_nft_metadata 
LEFT OUTER JOIN xsyn_assets ON xsyn_nft_metadata.token_id = xsyn_assets.token_id
`

// AssetList gets a list of assets depending on the filters
func AssetList(
	ctx context.Context,
	conn Conn,
	result *[]*passport.XsynNftMetadata,
	search string,
	archived bool,
	includedTokenIDs []int,
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

	// select specific assets via tokenIDs
	if includedTokenIDs != nil {
		cond := "("
		for i, nftTokenID := range includedTokenIDs {
			cond += fmt.Sprintf("%d", nftTokenID)
			if i < len(includedTokenIDs)-1 {
				cond += ","
				continue
			}

			cond += ")"
		}
		filterConditionsString += fmt.Sprintf(" AND xsyn_nft_metadata.token_id  IN %v", cond)
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
		SELECT COUNT(DISTINCT xsyn_nft_metadata.token_id)
		%s
		WHERE xsyn_nft_metadata.deleted_at %s
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
		WHERE xsyn_nft_metadata.deleted_at %s
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

// AssetGet returns a asset by given ID
func AssetGet(ctx context.Context, conn Conn, tokenID uint64) (*passport.XsynNftMetadata, error) {
	asset := &passport.XsynNftMetadata{}
	q := AssetGetQuery + ` WHERE xsyn_nft_metadata.token_id = $1`

	err := pgxscan.Get(ctx, conn, asset, q, tokenID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting asset from token ID.")
	}
	return asset, nil
}
