package db

import (
	"errors"
	"fmt"
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type UserColumn string

const (
	UserColumnID                  UserColumn = "id"
	UserColumnUsername            UserColumn = "username"
	UserColumnRoleID              UserColumn = "role_id"
	UserColumnAvatarID            UserColumn = "avatar_id"
	UserColumnEmail               UserColumn = "email"
	UserColumnMobileNumber        UserColumn = "mobile_number"
	UserColumnFirstName           UserColumn = "first_name"
	UserColumnLastName            UserColumn = "last_name"
	UserColumnVerified            UserColumn = "verified"
	UserColumnOldPasswordRequired UserColumn = "old_password_required"

	UserColumnDeletedAt UserColumn = "deleted_at"
	UserColumnUpdatedAt UserColumn = "updated_at"
	UserColumnCreatedAt UserColumn = "created_at"

	// Columns in the user list page
	UserColumnRoleName UserColumn = "role.name"
)

func (ic UserColumn) IsValid() error {
	switch ic {
	case UserColumnID,
		UserColumnUsername,
		UserColumnRoleID,
		UserColumnAvatarID,
		UserColumnEmail,
		UserColumnFirstName,
		UserColumnLastName,
		UserColumnVerified,
		UserColumnOldPasswordRequired,
		UserColumnDeletedAt,
		UserColumnUpdatedAt,
		UserColumnCreatedAt,
		UserColumnRoleName,
		UserColumnMobileNumber:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid user column type"))
}

const UserGetQuery string = `--sql
SELECT 
	users.id, users.role_id, users.two_factor_authentication_activated, users.two_factor_authentication_is_set, users.first_name, users.last_name, users.email, users.username, users.avatar_id, users.verified, users.old_password_required,
	users.created_at, sups, users.updated_at, users.deleted_at, users.facebook_id, users.google_id, users.twitch_id, users.twitter_id, users.discord_id, users.public_address, users.nonce, users.faction_id, users.withdraw_lock, users.mint_lock, users.total_lock,
	(SELECT COUNT(id) FROM user_recovery_codes urc WHERE urc.user_id = users.id) > 0 as has_recovery_code, users.mobile_number,
	row_to_json(role) as role,
	row_to_json(faction) as faction
` + UserGetQueryFrom
const UserGetQueryFrom string = `--sql
FROM users
LEFT JOIN (SELECT id, name, permissions, tier FROM roles) role ON role.id = users.role_id
LEFT JOIN (
    SELECT id, label, theme, logo_blob_id
    FROM factions
) faction ON faction.id = users.faction_id
`

// UserList gets a list of users depending on the filters
func UserList(
	result []*types.User,
	search string,
	archived bool,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy UserColumn,
	sortDir SortByDir,
) (int, error) {
	// Prepare Filters
	var args []interface{}

	filterConditionsString := ""
	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			column := UserColumn(f.Column)
			err := column.IsValid()
			if err != nil {
				return 0, err
			}

			condition, value := GenerateListFilterSQL(f.Column, f.Value, f.Operator, i+1)
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
			searchCondition = fmt.Sprintf(" AND users.keywords @@ to_tsquery($%d)", len(args))
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT users.id)
		%s
		WHERE users.deleted_at %s
			%s
			%s
		`,
		UserGetQueryFrom,
		archiveCondition,
		filterConditionsString,
		searchCondition,
	)

	var totalRows int
	err := passdb.StdConn.QueryRow(countQ, args...).Scan(&totalRows)
	if err != nil {
		return 0, err
	}
	if totalRows == 0 {
		return 0, nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, err
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
		WHERE users.deleted_at %s
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
	rows, err := passdb.StdConn.Query(q, args...)
	if err != nil {
		return 0, err
	}

	for rows.Next() {
		u := &types.User{}
		err := rows.Scan(
			&u.ID,
			&u.RoleID,
			&u.TwoFactorAuthenticationActivated,
			&u.TwoFactorAuthenticationIsSet,
			&u.FirstName,
			&u.LastName,
			&u.Email,
			&u.Username,
			&u.AvatarID,
			&u.Verified,
			&u.OldPasswordRequired,
			&u.CreatedAt,
			&u.Sups,
			&u.UpdatedAt,
			&u.DeletedAt,
			&u.FacebookID,
			&u.GoogleID,
			&u.TwitchID,
			&u.TwitterID,
			&u.DiscordID,
			&u.PublicAddress,
			&u.Nonce,
			&u.FactionID,
			&u.MobileNumber,
		)
		if err != nil {
			return 0, err
		}

		result = append(result, u)
	}

	return totalRows, nil
}

// UsernameAvailable returns true if a username is free
func UsernameAvailable(nameToCheck string, userID string) (bool, error) {
	if nameToCheck == "" {
		return false, terror.Error(fmt.Errorf("username cannot be empty"), "Username cannot be empty.")
	}
	nameToCheck = strings.ToLower(nameToCheck)

	count := 0

	if userID != "" {
		q := `
        		SELECT count(*) FROM users
        		WHERE 	username = $1 and id != $2
        	`
		err := passdb.StdConn.QueryRow(q, nameToCheck, userID).Scan(&count)
		if err != nil {
			return false, err
		}
		return count == 0, nil
	}

	q := `
		SELECT count(*) FROM users
		WHERE 	username = $1
	`
	err := passdb.StdConn.QueryRow(q, nameToCheck).Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

type Address struct {
	WalletAddress string `json:"walletAddress" db:"wallet_address"`
}

// IsUserWhitelisted check if user is whitelisted
func IsUserWhitelisted(walletAddress string) (bool, error) {

	addr := common.HexToAddress(walletAddress).Hex()
	_, err := boiler.WhitelistedAddresses(
		boiler.WhitelistedAddressWhere.WalletAddress.EQ(addr),
	).One(passdb.StdConn)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, terror.Error(err, "Issue getting user whitelisted user")
	}

	return true, nil
}

// IsUserWhitelisted check if user is whitelisted
func IsUserDeathlisted(walletAddress string) (bool, error) {
	addr := common.HexToAddress(walletAddress).Hex()
	_, err := boiler.WhitelistedAddresses(
		boiler.WhitelistedAddressWhere.WalletAddress.EQ(addr),
	).One(passdb.StdConn)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, terror.Error(err, "Issue getting user death listed user")
	}

	return true, nil
}

// UserTransactionGetList returns list of transactions based on userid == credit/ debit
func UserTransactionGetList(userID string, limit int) ([]*boiler.Transaction, error) {
	transactions, err := boiler.Transactions(
		qm.Where(
			fmt.Sprintf(
				"%s = ? OR %s = ?",
				qm.Rels(boiler.TableNames.Transactions, boiler.TransactionColumns.Credit),
				qm.Rels(boiler.TableNames.Transactions, boiler.TransactionColumns.Debit),
			),
			userID,
			userID,
		),
		qm.OrderBy(
			fmt.Sprintf(
				"%s desc",
				qm.Rels(boiler.TableNames.Transactions, boiler.TransactionColumns.CreatedAt),
			),
		),
		qm.Limit(limit),
	).All(passdb.StdConn)
	if err != nil {
		return nil, err
	}

	for _, t := range transactions {
		_, err := t.Amount.Value()
		if err != nil {
			return nil, err
		}
	}

	return transactions, nil
}

func UserMixedCaseUpdateAll() error {
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	users, err := boiler.Users().All(tx)
	if err != nil {
		return err
	}
	for _, u := range users {
		if !u.PublicAddress.Valid || u.PublicAddress.String == "" {
			continue
		}
		if strings.Contains(u.PublicAddress.String, "bnb") {
			continue
		}
		passlog.L.Info().Str("user_id", u.ID).Msg("updating user to mixed case")
		u.PublicAddress = null.StringFrom(common.HexToAddress(u.PublicAddress.String).Hex())
		_, err = u.Update(tx, boil.Whitelist(boiler.UserColumns.PublicAddress))
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
