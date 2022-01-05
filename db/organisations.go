package db

import (
	"passport"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gosimple/slug"
	"github.com/ninja-software/terror/v2"
)

type OrganisationColumn string

const (
	OrganisationColumnID   OrganisationColumn = "id"
	OrganisationColumnSlug OrganisationColumn = "slug"
	OrganisationColumnName OrganisationColumn = "name"

	OrganisationColumnDeletedAt OrganisationColumn = "deleted_at"
	OrganisationColumnUpdatedAt OrganisationColumn = "updated_at"
	OrganisationColumnCreatedAt OrganisationColumn = "created_at"
)

func (ic OrganisationColumn) IsValid() error {
	switch ic {
	case OrganisationColumnID,
		OrganisationColumnSlug,
		OrganisationColumnName,
		OrganisationColumnDeletedAt,
		OrganisationColumnUpdatedAt,
		OrganisationColumnCreatedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid organisation column type"))
}

// OrganisationGet returns a organisation by given ID
func OrganisationGet(ctx context.Context, conn Conn, organisationID passport.OrganisationID) (*passport.Organisation, error) {
	organisation := &passport.Organisation{}
	q := `--sql
		SELECT id, slug, name, deleted_at, updated_at, created_at
		FROM organisations
		WHERE id = $1`
	err := pgxscan.Get(ctx, conn, organisation, q, organisationID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return organisation, nil
}

// OrganisationGetBySlug returns a organisation by given slug
func OrganisationGetBySlug(ctx context.Context, conn Conn, slug string) (*passport.Organisation, error) {
	organisation := &passport.Organisation{}
	q := `--sql
		SELECT id, slug, name, deleted_at, updated_at, created_at
		FROM organisations
		WHERE slug = $1`
	err := pgxscan.Get(ctx, conn, organisation, q, slug)
	if err != nil {
		return nil, terror.Error(err)
	}
	return organisation, nil
}

// OrganisationCreate will create a new organisation
func OrganisationCreate(ctx context.Context, conn Conn, organisation *passport.Organisation) error {
	slug := slug.Make(organisation.Name)
	q := `--sql
		INSERT INTO organisations (slug, name)
		VALUES ($1, $2)
		RETURNING
			id, slug, name, deleted_at, updated_at, created_at`
	err := pgxscan.Get(ctx,
		conn,
		organisation,
		q,
		slug,
		organisation.Name,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// OrganisationUpdate will update an existing organisation
func OrganisationUpdate(ctx context.Context, conn Conn, organisation *passport.Organisation) error {
	slug := slug.Make(organisation.Name)
	q := `--sql
		UPDATE organisations
		SET 
			slug = $2, name = $3
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		organisation.ID,
		slug,
		organisation.Name,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// OrganisationArchiveUpdate will update an existing organisation as archived or unarchived
func OrganisationArchiveUpdate(ctx context.Context, conn Conn, id passport.OrganisationID, archived bool) error {
	var deletedAt *time.Time
	if archived {
		now := time.Now()
		deletedAt = &now
	}
	q := `--sql
		UPDATE organisations
		SET deleted_at = $2
		WHERE id = $1`
	_, err := conn.Exec(ctx, q, id, deletedAt)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// OrganisationList will grab a list of organisations in offset pagination format
func OrganisationList(ctx context.Context,
	conn Conn,
	result *[]*passport.Organisation,
	search string,
	archived bool,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy OrganisationColumn,
	sortDir SortByDir,
) (int, error) {
	// Prepare Filters
	var args []interface{}
	filterConditionsString := ""

	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			column := OrganisationColumn(f.ColumnField)
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
			filterConditionsString = " AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
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
			searchCondition = fmt.Sprintf(" AND keywords @@ to_tsquery($%d)", len(args))
		}
	}

	// Get Total Found
	q := `--sql
		SELECT COUNT(*)
		FROM organisations
		WHERE deleted_at ` + archiveCondition + filterConditionsString + searchCondition
	var totalRows int
	err := pgxscan.Get(ctx, conn, &totalRows, q, args...)
	if err != nil {
		return 0, terror.Error(err)
	}
	if totalRows == 0 {
		return totalRows, nil
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
	q = fmt.Sprintf(
		`--sql
		SELECT id, slug, name, deleted_at, updated_at, created_at
		FROM organisations
		WHERE deleted_at %s
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
