package db

import (
	"context"
	"fmt"
	"strings"
	"time"
	"xsyn-services/types"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

type RoleColumn string

const (
	RoleColumnID          RoleColumn = "id"
	RoleColumnName        RoleColumn = "name"
	RoleColumnPermissions RoleColumn = "permissions"
	RoleColumnTier        RoleColumn = "tier"
	RoleColumnReserved    RoleColumn = "reserved"

	RoleColumnDeletedAt RoleColumn = "deleted_at"
	RoleColumnUpdatedAt RoleColumn = "updated_at"
	RoleColumnCreatedAt RoleColumn = "created_at"
)

func (ic RoleColumn) IsValid() error {
	switch ic {
	case RoleColumnID,
		RoleColumnName,
		RoleColumnPermissions,
		RoleColumnTier,
		RoleColumnReserved,
		RoleColumnDeletedAt,
		RoleColumnUpdatedAt,
		RoleColumnCreatedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid role column type"))
}

// RoleGet returns a role by given ID
func RoleGet(ctx context.Context, conn Conn, roleID types.RoleID) (*types.Role, error) {
	role := &types.Role{}
	q := `--sql
		SELECT id, name, permissions, tier, reserved, deleted_at, updated_at, created_at
		FROM roles
		WHERE id = $1`
	err := pgxscan.Get(ctx, conn, role, q, roleID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return role, nil
}

// RoleGetByName returns a role by name
func RoleByName(ctx context.Context, conn Conn, name string) (*types.Role, error) {
	role := &types.Role{}
	q := `--sql
		SELECT id, name, permissions, tier, reserved, deleted_at, updated_at, created_at
		FROM roles
		WHERE name = $1`
	err := pgxscan.Get(ctx, conn, role, q, name)
	if err != nil {
		return nil, terror.Error(err)
	}
	return role, nil
}

// RoleCreate will create a new role
func RoleCreate(ctx context.Context, conn Conn, role *types.Role) error {
	q := `--sql
		INSERT INTO roles (name, permissions)
		VALUES ($1, $2)
		RETURNING
			id, name, permissions, tier, reserved, deleted_at, updated_at, created_at`
	err := pgxscan.Get(ctx,
		conn,
		role,
		q,
		role.Name,
		role.Permissions,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// RoleCreateReserved will create a reserved new role
func RoleCreateReserved(ctx context.Context, conn Conn, role *types.Role) error {
	q := `--sql
		INSERT INTO roles (id, name, permissions, tier, reserved)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id, name, permissions, tier, reserved, deleted_at, updated_at, created_at`
	err := pgxscan.Get(ctx,
		conn,
		role,
		q,
		role.ID,
		role.Name,
		role.Permissions,
		role.Tier,
		true,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// RoleUpdate will update an existing role
func RoleUpdate(ctx context.Context, conn Conn, role *types.Role) error {
	q := `--sql
		UPDATE roles
		SET 
			name = $2, permissions = $3, tier = $4, reserved = $5
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		role.ID,
		role.Name,
		role.Permissions,
		role.Tier,
		role.Reserved,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// RoleArchiveUpdate will update an existing role as archived or unarchived
func RoleArchiveUpdate(ctx context.Context, conn Conn, id types.RoleID, archived bool) error {
	var deletedAt *time.Time
	if archived {
		now := time.Now()
		deletedAt = &now
	}
	q := `--sql
		UPDATE roles
		SET deleted_at = $2
		WHERE id = $1`
	_, err := conn.Exec(ctx, q, id, deletedAt)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// RoleList will grab a list of roles in offset pagination format
func RoleList(ctx context.Context,
	conn Conn,
	result *[]*types.Role,
	search string,
	archived bool,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy RoleColumn,
	sortDir SortByDir,
) (int, error) {
	// Prepare Filters
	var args []interface{}
	filterConditionsString := ""

	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			column := RoleColumn(f.ColumnField)
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
		FROM roles
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
		SELECT id, name, permissions, tier, reserved, deleted_at, updated_at, created_at
		FROM roles
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
