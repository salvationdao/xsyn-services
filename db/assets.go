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

const AssetGetQuery string = `--sql
SELECT 
	users.id, users.role_id, users.two_factor_authentication_activated, users.two_factor_authentication_is_set, users.first_name, users.last_name, users.email, users.username, users.avatar_id, users.verified,
	users.created_at, sups, users.updated_at, users.deleted_at, users.facebook_id, users.google_id, users.twitch_id, users.public_address, users.nonce, users.faction_id,
	(SELECT COUNT(id) FROM user_recovery_codes urc WHERE urc.user_id = users.id) > 0 as has_recovery_code,
	row_to_json(role) as role,
	row_to_json(faction) as faction,
	row_to_json(organisation) as organisation
` + UserGetQueryFrom

const AssetGetQueryFrom string = `--sql
FROM users
LEFT JOIN (SELECT id, name, permissions, tier FROM roles) role ON role.id = users.role_id
LEFT JOIN (
	SELECT id, user_id, organisation_id, slug, name
	FROM user_organisations
	INNER JOIN organisations o ON o.id = organisation_id
) organisation ON organisation.user_id = users.id
LEFT JOIN (
    SELECT id, label, theme
    FROM factions
) faction ON faction.id = users.faction_id
`

// AssetList gets a list of patients depending on the filters
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
		SELECT COUNT(DISTINCT assets.id)
		%s
		WHERE assets.deleted_at %s
			%s
			%s
		`,
		UserGetQueryFrom,
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
		UserGetQuery+`--sql
		WHERE assets.deleted_at %s
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
