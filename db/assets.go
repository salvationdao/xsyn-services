package db

import (
	"context"
	"fmt"
	"passport"
	"strings"

	goaway "github.com/TwiN/go-away"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

type AssetColumn string

const (
	AssetColumnID           AssetColumn = "id"
	AssetColumnTokenID      AssetColumn = "token_id"
	AssetColumnUserID       AssetColumn = "user_id"
	AssetColumnUsername     AssetColumn = "username"
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
		AssetColumnUsername,
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
row_to_json(c) as collection,
xsyn_metadata.token_id,
xsyn_metadata.name,
xsyn_metadata.description,
xsyn_metadata.external_url,
xsyn_metadata.image,
xsyn_metadata.durability,
xsyn_metadata.attributes,
xsyn_metadata.deleted_at,
xsyn_metadata.updated_at,
xsyn_metadata.created_at,
xsyn_assets.user_id,
xsyn_assets.frozen_at,
xsyn_assets.locked_by_id,
xsyn_assets.tx_history,
COALESCE(xsyn_assets.minting_signature, '') as minting_signature,
u.username
` + AssetGetQueryFrom

const AssetGetQueryFrom = `
FROM xsyn_metadata 
LEFT OUTER JOIN xsyn_assets ON xsyn_metadata.token_id = xsyn_assets.token_id
INNER JOIN collections c ON xsyn_metadata.collection_id = c.id
INNER JOIN users u ON xsyn_assets.user_id = u.id
`

// AssetList gets a list of assets depending on the filters
func AssetList(
	ctx context.Context,
	conn Conn,
	search string,
	archived bool,
	includedTokenIDs []uint64,
	filter *ListFilterRequest,
	attributeFilter *AttributeFilterRequest,
	offset int,
	pageSize int,
	sortBy AssetColumn,
	sortDir SortByDir,
) (int, []*passport.XsynMetadata, error) {
	// Prepare Filters
	var args []interface{}

	filterConditionsString := ""
	argIndex := 0
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := AssetColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			argIndex += 1
			condition, value := GenerateListFilterSQL(f.ColumnField, f.Value, f.OperatorValue, argIndex)
			if condition != "" {
				filterConditions = append(filterConditions, condition)
				args = append(args, value)
			}
		}
		if len(filterConditions) > 0 {
			filterConditionsString = "AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	if attributeFilter != nil {
		filterConditions := []string{}
		for _, f := range attributeFilter.Items {
			column := TraitType(f.Trait)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			// argIndex += 1
			condition, err := GenerateAttributeFilterSQL(f.Trait, f.Value, f.OperatorValue, argIndex, "xsyn_metadata")
			if err != nil {
				return 0, nil, terror.Error(err)
			}
			filterConditions = append(filterConditions, *condition)
			// args = append(args, f.Value)
		}
		if len(filterConditions) > 0 {
			filterConditionsString += "AND (" + strings.Join(filterConditions, " "+string(attributeFilter.LinkOperator)+" ") + ")"
		}
	}

	// select specific assets via tokenIDs
	if len(includedTokenIDs) > 0 {
		cond := "("
		for i, nftTokenID := range includedTokenIDs {
			cond += fmt.Sprintf("%d", nftTokenID)
			if i < len(includedTokenIDs)-1 {
				cond += ","
				continue
			}

			cond += ")"
		}
		filterConditionsString += fmt.Sprintf(" AND xsyn_metadata.token_id  IN %v", cond)
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
			searchCondition = fmt.Sprintf(" AND xsyn_metadata.keywords @@ to_tsquery($%d)", len(args))
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT xsyn_metadata.token_id)
		%s
		WHERE xsyn_metadata.deleted_at %s
			%s
			%s
		`,
		AssetGetQueryFrom,
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
		return 0, make([]*passport.XsynMetadata, 0), nil
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
		AssetGetQuery+`--sql
		WHERE xsyn_metadata.deleted_at %s
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

	result := make([]*passport.XsynMetadata, 0)
	err = pgxscan.Select(ctx, conn, &result, q, args...)
	if err != nil {
		return 0, nil, terror.Error(err)
	}
	return totalRows, result, nil
}

// AssetGet returns a asset by given ID
func AssetGet(ctx context.Context, conn Conn, tokenID uint64) (*passport.XsynMetadata, error) {
	asset := &passport.XsynMetadata{}
	q := AssetGetQuery + `WHERE xsyn_metadata.token_id = $1`

	err := pgxscan.Get(ctx, conn, asset, q, tokenID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting asset from token ID.")
	}
	return asset, nil
}

// AssetGet returns a asset by given ID
func AssetGetByName(ctx context.Context, conn Conn, name string) (*passport.XsynMetadata, error) {
	asset := &passport.XsynMetadata{}
	q := AssetGetQuery + `WHERE xsyn_metadata.name = $1`
	err := pgxscan.Get(ctx, conn, asset, q, name)
	if err != nil {
		return nil, terror.Error(err, "Issue getting asset from name.")
	}
	return asset, nil
}

// AssetUpdate will update an asset name entry in attribute
func AssetUpdate(ctx context.Context, conn Conn, tokenID uint64, newName string) error {

	// profanity check
	if goaway.IsProfane(newName) {
		return terror.Error(fmt.Errorf("invalid asset name: cannot contain profanity"), "invalid asset name: cannot contain profanity")
	}

	nameAvailable, err := AssetNameAvailable(ctx, conn, newName, tokenID)
	if err != nil {
		return terror.Error(err)
	}
	if !nameAvailable {
		return terror.Error(fmt.Errorf("name is taken: %s", newName), fmt.Sprintf("Name is taken: %s", newName))
	}

	// sql to update a 'Name' entry in the attributes column
	// reference: https://stackoverflow.com/a/38996799
	q := `--sql
	UPDATE 
    xsyn_metadata 
	SET
	    -- updates attributes with new name entry
	    attributes = JSONB_SET(attributes, ARRAY[elem_index::TEXT, 'value'], TO_JSON($1::TEXT)::JSONB, FALSE)
	FROM (
	    SELECT 
		    -- selects the indexes of attributes entries
	        pos- 1 AS elem_index
	    FROM 
	        xsyn_metadata, 
			-- gets indexes of attribute's entries
	        JSONB_ARRAY_ELEMENTS(attributes) WITH ORDINALITY arr(elem, pos)
	    WHERE
	        token_id = $2 and
			-- gets the name entry
	        elem->>'trait_type' = 'Name'
	    ) sub
	WHERE
	    token_id = $2;    
	`
	_, err = conn.Exec(ctx,
		q,
		newName,
		tokenID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// AssetNameAvailable returns true if an asset name is free
func AssetNameAvailable(ctx context.Context, conn Conn, nameToCheck string, tokenID uint64) (bool, error) {

	if nameToCheck == "" {
		return false, terror.Error(fmt.Errorf("name cannot be empty"), "Name cannot be empty.")
	}
	count := 0

	q := `
	SELECT 
		count(token_id) 
	FROM 
		xsyn_metadata, 
		JSONB_ARRAY_ELEMENTS(attributes) WITH ORDINALITY arr(elem, pos)
	WHERE 
	    elem ->>'trait_type' = 'Name'
		AND elem->>'value' = $2
		AND xsyn_metadata.token_id != $1
		`
	err := pgxscan.Get(ctx, conn, &count, q, tokenID, nameToCheck)
	if err != nil {
		return false, terror.Error(err)
	}

	return count == 0, nil
}

// AssetTransfer changes ownership of an asset to another user // TODO: add internal transaciton details tx jsonb array
func AssetTransfer(ctx context.Context, conn Conn, tokenID uint64, oldUserID, newUserID passport.UserID, txHash string) error {
	args := []interface{}{
		newUserID, oldUserID, tokenID,
	}

	hashUpdate := ""
	if txHash != "" {
		hashUpdate = fmt.Sprintf(`, tx_history = tx_history || '["%s"]'::jsonb`, txHash)
	}

	q := fmt.Sprintf(`
	UPDATE xsyn_assets
	SET user_id = $1 %s
	WHERE user_id = $2 AND token_id = $3`, hashUpdate)

	_, err := conn.Exec(ctx, q, args...)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// AssetTransferOnChain adds a tx hash to the history
func AssetTransferOnChain(ctx context.Context, conn Conn, tokenID uint64, txHash string) error {
	q := fmt.Sprintf(`
	UPDATE xsyn_assets
	SET  tx_history = tx_history || '["%s"]'::jsonb
	WHERE token_id = $1 AND NOT (tx_history @> '["%s"]'::jsonb) `, txHash, txHash)
	_, err := conn.Exec(ctx, q, tokenID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}
