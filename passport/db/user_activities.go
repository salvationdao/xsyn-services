package db

import (
	"fmt"
	"strings"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
)

type UserActivityColumn string

const (
	UserActivityColumnID         UserActivityColumn = "id"
	UserActivityColumnUserID     UserActivityColumn = "user_id"
	UserActivityColumnAction     UserActivityColumn = "action"
	UserActivityColumnObjectID   UserActivityColumn = "object_id"
	UserActivityColumnObjectSlug UserActivityColumn = "object_slug"
	UserActivityColumnObjectName UserActivityColumn = "object_name"
	UserActivityColumnObjectType UserActivityColumn = "object_type"
	UserActivityColumnOldData    UserActivityColumn = "old_data"
	UserActivityColumnNewData    UserActivityColumn = "new_data"
	UserActivityColumnCreatedAt  UserActivityColumn = "created_at"
)

func (c UserActivityColumn) IsValid() error {
	switch c {
	case UserActivityColumnID,
		UserActivityColumnUserID,
		UserActivityColumnAction,
		UserActivityColumnObjectID,
		UserActivityColumnObjectSlug,
		UserActivityColumnObjectName,
		UserActivityColumnObjectType,
		UserActivityColumnOldData,
		UserActivityColumnNewData,
		UserActivityColumnCreatedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid intake column type"))
}

// UserActivityCreate will create a brand new User Activity with name and filters
func UserActivityCreate(
	userID types.UserID,
	action string,
	objectType types.ObjectType,
	objectID *string,
	objectSlug *string,
	objectName *string,
	oldData null.JSON,
	newData null.JSON,
) error {
	q := `INSERT INTO user_activities (user_id, action, object_type, object_id, object_slug, object_name, old_data, new_data) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := passdb.StdConn.Exec(q, userID, action, objectType, objectID, objectSlug, objectName, oldData, newData)
	if err != nil {
		return err
	}
	return nil
}

// UserActivityGet will grab an existing User Activity by its id
func UserActivityGet(result *types.UserActivity, id types.UserActivityID) error {
	q := `
		SELECT a.id, a.user_id, a.action, a.object_type, a.object_id, a.object_slug, a.object_name, a.created_at, a.old_data, a.new_data,
			u.id as "user.id", u.email as "user.email", u.username as "user.username", u.avatar_id as "user.avatar_id", u.role as "user.role"
		FROM user_activities a
		JOIN users u ON u.id = a.user_id
		WHERE a.id = $1`
	err := passdb.StdConn.QueryRow(q, id).Scan(
		&result.ID,
		&result.UserID,
		&result.Action,
		&result.ObjectType,
		&result.ObjectID,
		&result.ObjectSlug,
		&result.ObjectName,
		&result.CreatedAt,
		&result.OldData,
		&result.NewData,
		&result.User.ID,
		&result.User.Email,
		&result.User.Username,
		&result.User.AvatarID,
		&result.User.RoleID,
	)
	if err != nil {
		return err
	}
	return nil
}

// UserActivityList will grab a list of User Activity templates in offset pagination format
func UserActivityList(result []*types.UserActivity, search string, filter *ListFilterRequest, page int, pageSize int, sortBy UserActivityColumn, sortDir SortByDir) (int, error) {
	// Prepare Filters
	var args []interface{}
	filterConditionsString := ""

	whereStr := ""

	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			condition, value := GenerateListFilterSQL(f.Column, f.Value, f.Operator, i+1)
			if condition != "" {
				filterConditions = append(filterConditions, condition)
				args = append(args, value)
			}
		}

		if len(filterConditions) > 0 {
			filterConditionsString = " (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
			whereStr = " WHERE" + filterConditionsString
		}
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			searchCondition = fmt.Sprintf(" a.keywords @@ to_tsquery($%d)", len(args))
			if whereStr == "" {
				whereStr = " WHERE" + searchCondition
			} else {
				whereStr += " AND" + searchCondition
			}
		}
	}

	// Get Total Found
	q := `
		SELECT COUNT(*)
		FROM user_activities a
		JOIN (SELECT id, email as "user.email" from users) u ON u.id = a.user_id` + whereStr
	var totalRows int
	err := passdb.StdConn.QueryRow(q, args...).Scan(&totalRows)
	if err != nil {
		return 0, err
	}
	if totalRows == 0 {
		return totalRows, nil
	}

	// Get Paginated Result
	q = fmt.Sprintf(
		`--sql
		SELECT a.id, a.user_id, a.action, a.object_type, a.object_id, a.object_slug, a.object_name, a.created_at, a.old_data, a.new_data,
			"user.id", "user.email", "user.username", "user.avatar_id"
		FROM user_activities a
		JOIN (
			SELECT users.id as "user.id", email as "user.email", username as "user.username", avatar_id as "user.avatar_id"
			FROM users
		) u ON "user.id" = a.user_id
		%s
		ORDER BY %s %s
		LIMIT %d
		OFFSET %d`,
		whereStr,
		sortBy,
		sortDir,
		pageSize,
		page,
	)
	r, err := passdb.StdConn.Query(q, args...)
	if err != nil {
		return 0, err
	}

	for r.Next() {
		act := &types.UserActivity{}
		err = r.Scan(
			&act.ID,
			&act.UserID,
			&act.Action,
			&act.ObjectID,
			&act.ObjectSlug,
			&act.ObjectName,
			&act.CreatedAt,
			&act.OldData,
			&act.NewData,
			&act.User.ID,
			&act.User.Email,
			&act.User.Username,
			&act.User.AvatarID,
		)
		if err != nil {
			return 0, err
		}

		result = append(result, act)
	}
	return totalRows, nil
}
