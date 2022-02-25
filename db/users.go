package db

import (
	"context"
	"errors"
	"fmt"
	"passport"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/ninja-software/terror/v2"
)

type UserColumn string

const (
	UserColumnID                  UserColumn = "id"
	UserColumnUsername            UserColumn = "username"
	UserColumnRoleID              UserColumn = "role_id"
	UserColumnAvatarID            UserColumn = "avatar_id"
	UserColumnEmail               UserColumn = "email"
	UserColumnFirstName           UserColumn = "first_name"
	UserColumnLastName            UserColumn = "last_name"
	UserColumnVerified            UserColumn = "verified"
	UserColumnOldPasswordRequired UserColumn = "old_password_required"

	UserColumnDeletedAt UserColumn = "deleted_at"
	UserColumnUpdatedAt UserColumn = "updated_at"
	UserColumnCreatedAt UserColumn = "created_at"

	// Columns in the user list page
	UserColumnRoleName         UserColumn = "role.name"
	UserColumnOrganisationName UserColumn = "organisation.name"
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
		UserColumnOrganisationName:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid user column type"))
}

const UserGetQuery string = `--sql
SELECT 
	users.id, users.role_id, users.two_factor_authentication_activated, users.two_factor_authentication_is_set, users.first_name, users.last_name, users.email, users.username, users.avatar_id, users.verified, users.old_password_required,
	users.created_at, sups, users.updated_at, users.deleted_at, users.facebook_id, users.google_id, users.twitch_id, users.twitter_id, users.discord_id, users.public_address, users.nonce, users.faction_id,
	(SELECT COUNT(id) FROM user_recovery_codes urc WHERE urc.user_id = users.id) > 0 as has_recovery_code,
	row_to_json(role) as role,
	row_to_json(faction) as faction,
	row_to_json(organisation) as organisation
` + UserGetQueryFrom
const UserGetQueryFrom string = `--sql
FROM users
LEFT JOIN (SELECT id, name, permissions, tier FROM roles) role ON role.id = users.role_id
LEFT JOIN (
	SELECT id, user_id, organisation_id, slug, name
	FROM user_organisations
	INNER JOIN organisations o ON o.id = organisation_id
) organisation ON organisation.user_id = users.id
LEFT JOIN (
    SELECT id, label, theme, logo_blob_id as logoBlobID
    FROM factions
) faction ON faction.id = users.faction_id
`

// UserByPublicAddress returns a user by given public wallet address
func UserByPublicAddress(ctx context.Context, conn Conn, publicAddress string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.public_address = LOWER($1)`
	err := pgxscan.Get(ctx, conn, user, q, publicAddress)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from Public Address.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserByGoogleID returns a user by google id
func UserByGoogleID(ctx context.Context, conn Conn, googleID string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.google_id = $1`
	err := pgxscan.Get(ctx, conn, user, q, googleID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from google id.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserByTwitchID returns a user by twitch id
func UserByTwitchID(ctx context.Context, conn Conn, twitchID string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.twitch_id = $1`
	err := pgxscan.Get(ctx, conn, user, q, twitchID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from twitch id.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserByTwitterID returns a user by Twitter id
func UserByTwitterID(ctx context.Context, conn Conn, twitterID string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.twitter_id = $1`
	err := pgxscan.Get(ctx, conn, user, q, twitterID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from twitter id.")
	}
	return user, nil
}

// UserByDiscordID returns a user by Discord id
func UserByDiscordID(ctx context.Context, conn Conn, discordID string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.discord_id = $1`
	err := pgxscan.Get(ctx, conn, user, q, discordID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from discord id.")
	}
	return user, nil
}

// UserByFacebookID returns a user by google id
func UserByFacebookID(ctx context.Context, conn Conn, facebookID string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.facebook_id = $1`
	err := pgxscan.Get(ctx, conn, user, q, facebookID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from facebook id.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserGet returns a user by given ID
func UserGet(ctx context.Context, conn Conn, userID passport.UserID) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.id = $1`

	err := pgxscan.Get(ctx, conn, user, q, userID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from ID.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserGet returns a user by given ID
func UserGetByUsername(ctx context.Context, conn Conn, username string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE users.username = $1`

	err := pgxscan.Get(ctx, conn, user, q, username)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from ID.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserGetByIDs returns a user by given ID
func UserGetByIDs(ctx context.Context, conn Conn, userIDs []passport.UserID) ([]*passport.User, error) {
	users := []*passport.User{}
	q := UserGetQuery + ` WHERE users.id IN (`
	for i, userID := range userIDs {
		q += fmt.Sprintf("'%s'", userID)
		if i < len(userIDs)-1 {
			q += ","
			continue
		}
		q += ")"
	}

	err := pgxscan.Select(ctx, conn, &users, q)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from ID.")
	}

	return users, nil
}

// UserByUsername returns a user by given username
func UserByUsername(ctx context.Context, conn Conn, username string) (*passport.User, error) {
	user := &passport.User{}
	q := UserGetQuery + ` WHERE username = $1`

	err := pgxscan.Get(ctx, conn, user, q, username)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from Username.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserByEmail returns a user by given email address
func UserByEmail(ctx context.Context, conn Conn, email string) (*passport.User, error) {
	user := &passport.User{}

	q := UserGetQuery + ` WHERE email = $1`
	err := pgxscan.Get(ctx, conn, user, q, email)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user from Email.")
	}

	// calc sups
	user.Sups.Init()

	return user, nil
}

// UserIDFromUsername takes a username and returns the user id
func UserIDFromUsername(ctx context.Context, conn Conn, username string) (*passport.UserID, error) {
	q := `SELECT id FROM users WHERE username = $1`
	var id passport.UserID
	err := pgxscan.Get(ctx, conn, &id, q, username)
	if err != nil {
		return nil, terror.Error(err)
	}
	return &id, nil
}

// User2FASecretGet returns a user 2FA secret by given ID
func User2FASecretGet(ctx context.Context, conn Conn, userID passport.UserID) (string, error) {
	secret := ""
	q := `
		SELECT 
			two_factor_authentication_secret
		FROM users 
		WHERE users.id = $1
	`
	err := pgxscan.Get(ctx, conn, &secret, q, userID)
	if err != nil {
		return "", terror.Error(err)
	}
	return secret, nil
}

// User2FASecretGet set users' 2fa secret
func User2FASecretSet(ctx context.Context, conn Conn, userID passport.UserID, secret string) error {
	q := `
		UPDATE 
			users
		SET
			two_factor_authentication_secret = $2
		WHERE users.id = $1
	`
	_, err := conn.Exec(ctx, q, userID, secret)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// User2FAIsActivated check whether users' 2fa is activated
func User2FAIsActivated(ctx context.Context, conn Conn, userID passport.UserID) (bool, error) {
	isActivated := false
	q := `
		SELECT 
			two_factor_authentication_activated
		FROM users 
		WHERE users.id = $1
	`
	err := pgxscan.Get(ctx, conn, &isActivated, q, userID)
	if err != nil {
		return false, terror.Error(err)
	}
	return isActivated, nil
}

// User2FAIsSet check a user has set set yet
func User2FAIsSet(ctx context.Context, conn Conn, userID passport.UserID) (bool, error) {
	isSet := false
	q := `
		SELECT 
			two_factor_authentication_is_set
		FROM users 
		WHERE users.id = $1
	`
	err := pgxscan.Get(ctx, conn, &isSet, q, userID)
	if err != nil {
		return false, terror.Error(err)
	}
	return isSet, nil
}

// UserUpdate2FAIsSet update users' 2fa flag
func UserUpdate2FAIsSet(ctx context.Context, conn Conn, userID passport.UserID, isSet bool) error {
	q := `
		UPDATE users
		SET	two_factor_authentication_is_set = $2
		WHERE id = $1;
	`
	_, err := conn.Exec(ctx, q, userID, isSet)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserCreate will create a new user
func UserCreate(ctx context.Context, conn Conn, user *passport.User) error {
	usernameOK, err := UsernameAvailable(ctx, conn, user.Username, nil)
	if err != nil {
		return terror.Error(err)
	}
	if !usernameOK {
		return terror.Error(fmt.Errorf("username is taken: %s", user.Username))
	}

	q := `--sql
		INSERT INTO users (first_name, last_name, email, username, public_address, avatar_id, role_id, verified, facebook_id, google_id, twitch_id, twitter_id, discord_id)
		VALUES ($1, $2, $3, $4, LOWER($5), $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING
			id, role_id, first_name, last_name, email, username, avatar_id, created_at, updated_at, deleted_at, facebook_id, google_id, twitch_id, twitter_id, discord_id`
	err = pgxscan.Get(ctx,
		conn,
		user,
		q,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Username,
		user.PublicAddress,
		user.AvatarID,
		user.RoleID,
		user.Verified,
		user.FacebookID,
		user.GoogleID,
		user.TwitchID,
		user.TwitterID,
		user.DiscordID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserFactionEnlist will assign the faction to the user
func UserFactionEnlist(ctx context.Context, conn Conn, user *passport.User) error {
	q := `--sql
		UPDATE users
		SET faction_id = $2
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		user.ID,
		user.FactionID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserUpdate will update a user
func UserUpdate(ctx context.Context, conn Conn, user *passport.User) error {
	usernameOK, err := UsernameAvailable(ctx, conn, user.Username, &user.ID)
	if err != nil {
		return terror.Error(err)
	}
	if !usernameOK {
		return terror.Error(fmt.Errorf("username is taken: %s", user.Username))
	}

	if user.Email.String != "" {
		emailOK, err := EmailAvailable(ctx, conn, user.Email.String, &user.ID)
		if err != nil {
			return terror.Error(err)
		}
		if !emailOK {
			return terror.Error(fmt.Errorf("email is taken: %s", user.Email.String), "Email is already in use, please use another one.")
		}
	}

	q := `--sql
		UPDATE users
		SET first_name = $2, last_name = $3, email = $4, username = $5, avatar_id = $6, role_id = $7, two_factor_authentication_activated = $8
		WHERE id = $1`
	_, err = conn.Exec(ctx,
		q,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Username,
		user.AvatarID,
		user.RoleID,
		user.TwoFactorAuthenticationActivated,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserRemoveWallet will remove a users wallet
func UserRemoveWallet(ctx context.Context, conn Conn, user *passport.User) error {
	q := `--sql
		UPDATE users
		SET public_address = null
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		user.ID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserAddWallet will add a users wallet
func UserAddWallet(ctx context.Context, conn Conn, user *passport.User, publicAddress string) error {
	count := 0

	q := `--sql
		SELECT count(*)
		FROM users
		WHERE public_address = LOWER($1)`

	err := pgxscan.Get(ctx, conn, &count, q, publicAddress)
	if err != nil {
		return terror.Error(err)
	}

	if count != 0 {
		return terror.Error(fmt.Errorf("wallet already assigned to a user"), "This wallet is already assigned to a user.")
	}

	q = `--sql
		UPDATE users
		SET public_address = LOWER($2)
		WHERE id = $1`
	_, err = conn.Exec(ctx,
		q,
		user.ID,
		publicAddress,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserRemoveFacebook will remove a users associated Facebook account
func UserRemoveFacebook(ctx context.Context, conn Conn, user *passport.User) error {
	q := `--sql
		UPDATE users
		SET facebook_id = null
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		user.ID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserAddFacebook will associate a user with a Facebook account
func UserAddFacebook(ctx context.Context, conn Conn, user *passport.User, facebookID string) error {
	count := 0

	q := `--sql
		SELECT count(*)
		FROM users
		WHERE facebook_id = $1`

	err := pgxscan.Get(ctx, conn, &count, q, facebookID)
	if err != nil {
		return terror.Error(err)
	}

	if count != 0 {
		return terror.Error(fmt.Errorf("facebook already assigned to a user"), "This Facebook account is already associated with a user.")
	}

	q = `--sql
		UPDATE users
		SET facebook_id = $2
		WHERE id = $1`
	_, err = conn.Exec(ctx,
		q,
		user.ID,
		facebookID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserRemoveGoogle will remove a users associated Google account
func UserRemoveGoogle(ctx context.Context, conn Conn, user *passport.User) error {
	q := `--sql
		UPDATE users
		SET google_id = null
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		user.ID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserAddGoogle wil associate a user with a Google account
func UserAddGoogle(ctx context.Context, conn Conn, user *passport.User, googleID string) error {
	count := 0

	q := `--sql
		SELECT count(*)
		FROM users
		WHERE google_id = $1`

	err := pgxscan.Get(ctx, conn, &count, q, googleID)
	if err != nil {
		return terror.Error(err)
	}

	if count != 0 {
		return terror.Error(fmt.Errorf("google already assigned to a user"), "This Google account is already associated with a user.")
	}

	q = `--sql
		UPDATE users
		SET google_id = $2
		WHERE id = $1`
	_, err = conn.Exec(ctx,
		q,
		user.ID,
		googleID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserRemoveTwitch will remove a users associated Twitch account
func UserRemoveTwitch(ctx context.Context, conn Conn, user *passport.User) error {
	q := `--sql
		UPDATE users
		SET twitch_id = null
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		user.ID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserAddTwitch wil associate a user with a Twitch account
func UserAddTwitch(ctx context.Context, conn Conn, user *passport.User, twitchID string) error {
	count := 0

	q := `--sql
		SELECT count(*)
		FROM users
		WHERE twitch_id = $1`

	err := pgxscan.Get(ctx, conn, &count, q, twitchID)
	if err != nil {
		return terror.Error(err)
	}

	if count != 0 {
		return terror.Error(fmt.Errorf("twitch already assigned to a user"), "This Twitch account is already associated with a user.")
	}

	q = `--sql
		UPDATE users
		SET twitch_id = $2
		WHERE id = $1`
	_, err = conn.Exec(ctx,
		q,
		user.ID,
		twitchID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserRemoveTwitter will remove a users associated Twitter account
func UserRemoveTwitter(ctx context.Context, conn Conn, user *passport.User) error {
	q := `--sql
		UPDATE users
		SET twitter_id = null
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		user.ID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserAddTwitter wil associate a user with a Twitter account
func UserAddTwitter(ctx context.Context, conn Conn, user *passport.User, twitterID string) error {
	count := 0

	q := `--sql
		SELECT count(*)
		FROM users
		WHERE twitter_id = $1`

	err := pgxscan.Get(ctx, conn, &count, q, twitterID)
	if err != nil {
		return terror.Error(err)
	}

	if count != 0 {
		return terror.Error(fmt.Errorf("twitter already assigned to a user"), "This Twitter account is already associated with a user.")
	}

	q = `--sql
		UPDATE users
		SET twitter_id = $2
		WHERE id = $1`
	_, err = conn.Exec(ctx,
		q,
		user.ID,
		twitterID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserRemoveDiscord will remove a users associated Discord account
func UserRemoveDiscord(ctx context.Context, conn Conn, user *passport.User) error {
	q := `--sql
		UPDATE users
		SET discord_id = null
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		user.ID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserAddDiscord wil associate a user with a Discord account
func UserAddDiscord(ctx context.Context, conn Conn, user *passport.User, discordID string) error {
	count := 0

	q := `--sql
		SELECT count(*)
		FROM users
		WHERE discord_id = $1`

	err := pgxscan.Get(ctx, conn, &count, q, discordID)
	if err != nil {
		return terror.Error(err)
	}

	if count != 0 {
		return terror.Error(fmt.Errorf("discord already assigned to a user"), "This Discord account is already associated with a user.")
	}

	q = `--sql
		UPDATE users
		SET discord_id = $2
		WHERE id = $1`
	_, err = conn.Exec(ctx,
		q,
		user.ID,
		discordID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserVerify will mark a user as verified
func UserVerify(ctx context.Context, conn Conn, id passport.UserID) error {
	q := `
		UPDATE users
		SET verified = true
		WHERE id = $1`
	_, err := conn.Exec(ctx, q, id.String())
	if err != nil {
		return terror.Error(err, "")
	}
	return nil
}

// UserUpdatePasswordSetting will change whether a user needs an old password to change password
func UserUpdatePasswordSetting(ctx context.Context, conn Conn, id passport.UserID, oldPasswordRequired bool) error {
	q := `
		UPDATE users
		SET old_password_required = $2
		WHERE id = $1`
	_, err := conn.Exec(ctx, q, id.String(), oldPasswordRequired)
	if err != nil {
		return terror.Error(err, "")
	}
	return nil
}

// UserList gets a list of patients depending on the filters
func UserList(
	ctx context.Context,
	conn Conn,
	result *[]*passport.User,
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
			column := UserColumn(f.ColumnField)
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
	err = pgxscan.Select(ctx, conn, result, q, args...)
	if err != nil {
		return 0, terror.Error(err)
	}

	for _, user := range *result {
		// calc sups
		user.Sups.Init()

	}
	return totalRows, nil
}

// UserArchiveUpdate will update a user archive status
func UserArchiveUpdate(ctx context.Context, conn Conn, id passport.UserID, archived bool) error {
	var deletedAt *time.Time
	if archived {
		now := time.Now()
		deletedAt = &now
	}
	q := `
		UPDATE users
		SET deleted_at = $2
		WHERE id = $1 `
	_, err := conn.Exec(ctx, q, id, deletedAt)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

//// UserGenerateUsername generates a user slug in the format "JohnSmith3".
//func UserGenerateUsername(ctx context.Context, conn Conn, firstName string, lastName string, oldUsername string) (string, error) {
//	seperator := "_" // use underscore to prevent loosing hyphened names ie John Brown-Smith
//	username := slug.Make(fmt.Sprintf("%s%s%s", firstName, seperator, lastName))
//
//	if username == oldUsername {
//		return oldUsername, nil
//	}
//
//	// check if slug exists
//	count := 0
//	countQ := `
//	SELECT
//		count(*)
//	FROM
//		users
//	WHERE
//		username ~ $1
//	`
//	// Match the
//	// %s[.]?[0-9]*$
//	// `%s`: the username
//	// `[.]?`: zero or one hyphen
//	// `[0-9]*`: zero or more digits
//	// `$`: on the end of the string
//	clause := fmt.Sprintf("%s[%s]?[0-9]*$", username, seperator)
//	err := pgxscan.Get(ctx, conn, &count, countQ, clause)
//	if err != nil {
//		return "", terror.Error(err)
//	}
//	if count == 0 {
//		return username, nil
//	}
//
//	return username + fmt.Sprintf("%s%d", seperator, count), nil
//}

// UserExistsByEmail checks if a user is found through their email address
func UserExistsByEmail(ctx context.Context, conn Conn, email string) (bool, error) {
	var count int
	q := "SELECT COUNT(*) FROM users WHERE email = $1"
	err := pgxscan.Get(ctx, conn, &count, q, email)
	if err != nil {
		return false, terror.Error(err)
	}
	return count > 0, nil
}

// UserSetOrganisations will set a user's organisations
func UserSetOrganisations(ctx context.Context, conn Conn, userID passport.UserID, organisations ...passport.OrganisationID) error {
	args := []interface{}{userID}
	values := []string{}
	removeValues := []string{}

	for i, orgID := range organisations {
		args = append(args, orgID)
		values = append(values, fmt.Sprintf("($1, $%d)", i+2))
		removeValues = append(removeValues, fmt.Sprintf("$%d", i+2))
	}

	// Add new organisations to user
	insertQuery := fmt.Sprintf(`--sql
		INSERT INTO user_organisations (user_id, organisation_id)
		VALUES %s
		ON CONFLICT (user_id, organisation_id) DO NOTHING`,
		strings.Join(values, ", "),
	)
	_, err := conn.Exec(ctx, insertQuery, args...)
	if err != nil {
		return terror.Error(err)
	}

	// Remove old organisations from user
	removeQuery := fmt.Sprintf(`--sql
		DELETE FROM user_organisations
		WHERE user_id = $1 AND organisation_id NOT IN (%s)`,
		strings.Join(removeValues, ", "),
	)
	_, err = conn.Exec(ctx, removeQuery, args...)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// UserSetRecoveryCodes set users' recovery codes
func UserSetRecoveryCodes(ctx context.Context, conn Conn, userID passport.UserID, recoveryCodes []string) error {
	q := `
		INSERT INTO 
			user_recovery_codes (user_id, recovery_code)
		VALUES	
	`

	for i, recoveryCode := range recoveryCodes {
		q += fmt.Sprintf("('%s','%s')", userID, recoveryCode)

		if i < len(recoveryCodes)-1 {
			q += ","
			continue
		}

		q += ";"
	}

	_, err := conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// UserCheckRecoveryCode return recovery code by given user id and recovery code
func UserCheckRecoveryCode(ctx context.Context, conn Conn, userID passport.UserID, recoveryCode string) error {
	q := `
		SELECT id FROM user_recovery_codes
		WHERE 	user_id = $1 AND 
				recovery_code = $2 AND 
				used_at ISNULL;
	`

	_, err := conn.Exec(ctx, q, userID, recoveryCode)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// UserUseRecoveryCode use users' recovery code
func UserUseRecoveryCode(ctx context.Context, conn Conn, userID passport.UserID, recoveryCode string) error {
	q := `
		UPDATE
			user_recovery_codes
		SET
			used_at = NOW()
		WHERE 	user_id = $1 AND 
				recovery_code = $2 AND
				used_at ISNULL;
	`

	_, err := conn.Exec(ctx, q, userID, recoveryCode)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// UserDeleteRecoveryCode delete users' recovery code
func UserDeleteRecoveryCode(ctx context.Context, conn Conn, userID passport.UserID) error {
	q := `
		DELETE FROM
			user_recovery_codes
		WHERE user_id = $1;
	`

	_, err := conn.Exec(ctx, q, userID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// UserUpdateNonce updates a user's nonce, used for wallet auth
func UserUpdateNonce(ctx context.Context, conn Conn, userID passport.UserID, newNonce string) error {
	q := `
		UPDATE users
		SET	nonce = $2
		WHERE id = $1;
	`
	_, err := conn.Exec(ctx, q, userID, newNonce)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// EmailAvailable returns true if an email is free
func EmailAvailable(ctx context.Context, conn Conn, emailToCheck string, userID *passport.UserID) (bool, error) {
	count := 0

	if userID != nil && !userID.IsNil() {
		q := `
        		SELECT count(*) FROM users
        		WHERE email = $1 and id != $2
        	`
		err := pgxscan.Get(ctx, conn, &count, q, emailToCheck, userID)
		if err != nil {
			return false, terror.Error(err)
		}
		return count == 0, nil
	}

	q := `
		SELECT count(*) FROM users
		WHERE email = $1
	`
	err := pgxscan.Get(ctx, conn, &count, q, emailToCheck)
	if err != nil {
		return false, terror.Error(err)
	}
	return count == 0, nil
}

// UsernameAvailable returns true if a username is free
func UsernameAvailable(ctx context.Context, conn Conn, nameToCheck string, userID *passport.UserID) (bool, error) {
	if nameToCheck == "" {
		return false, terror.Error(fmt.Errorf("username cannot be empty"), "Username cannot be empty.")
	}
	count := 0

	if userID != nil && !userID.IsNil() {
		q := `
        		SELECT count(*) FROM users
        		WHERE 	username = $1 and id != $2
        	`
		err := pgxscan.Get(ctx, conn, &count, q, nameToCheck, userID)
		if err != nil {
			return false, terror.Error(err)
		}
		return count == 0, nil
	}

	q := `
		SELECT count(*) FROM users
		WHERE 	username = $1
	`
	err := pgxscan.Get(ctx, conn, &count, q, nameToCheck)
	if err != nil {
		return false, terror.Error(err)
	}
	return count == 0, nil
}

// InsertSystemUser this allows manually giving a user a ID
func InsertSystemUser(ctx context.Context, conn Conn, user *passport.User) error {

	q := `--sql
			INSERT INTO users (id, username, role_id, verified, faction_id)
			VALUES ($1, $2, $3, $4, $5)`

	_, err := conn.Exec(ctx,
		q,
		user.ID,
		user.Username,
		user.RoleID,
		user.Verified,
		user.FactionID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// UserIDsGetByFactionID return a list of user id from the given faction id
func UserIDsGetByFactionID(ctx context.Context, conn Conn, factionID passport.FactionID) ([]passport.UserID, error) {
	userIDs := []passport.UserID{}

	q := `
		SELECT
			id
		FROM
			users
		WHERE
			faction_id = $1
	`

	err := pgxscan.Select(ctx, conn, &userIDs, q, factionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return userIDs, nil
}

// FactionUserIDGetByFactionID return the faction game account user by faction id
func FactionUserIDGetByFactionID(ctx context.Context, conn Conn, factionID passport.FactionID) (passport.UserID, error) {
	userID := passport.UserID{}

	q := `
		SELECT
			id
		FROM
			users
		WHERE
			faction_id = $1 AND role_id = $2
	`

	err := pgxscan.Get(ctx, conn, &userID, q, factionID, passport.UserRoleGameAccount)
	if err != nil {
		return userID, terror.Error(err)
	}

	return userID, nil
}

type Address struct {
	WalletAddress string `json:"walletAddress" db:"wallet_address"`
}

// IsUserWhitelisted check if user is whitelisted
func IsUserWhitelisted(ctx context.Context, conn Conn, walletAddress string) (bool, error) {
	user := &Address{}
	q := "SELECT * FROM whitelisted_addresses WHERE wallet_address = $1"
	err := pgxscan.Get(ctx, conn, user, q, walletAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, terror.Error(err, "Issue getting user whitelisted user")
	}

	return true, nil
}

// IsUserWhitelisted check if user is whitelisted
func IsUserDeathlisted(ctx context.Context, conn Conn, walletAddress string) (bool, error) {
	user := &Address{}
	q := "SELECT * FROM death_addresses WHERE wallet_address = $1"
	err := pgxscan.Get(ctx, conn, user, q, walletAddress)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, terror.Error(err, "Issue getting user death listed user")
	}

	return true, nil
}

func UpdateUserMetadata(ctx context.Context, conn Conn, userID passport.UserID, metadata *passport.UserMetadata) error {
	q := "UPDATE users SET metadata = $1 WHERE id = $2"

	_, err := conn.Exec(ctx, q, metadata, userID)
	if err != nil {
		return terror.Error(err, "Issue inserting user data.")
	}

	return nil
}

func GetUserMetadata(ctx context.Context, conn Conn, userID passport.UserID) (*passport.UserMetadata, error) {
	md := &passport.User{}
	q := "SELECT metadata FROM users WHERE id = $1"

	err := pgxscan.Get(ctx, conn, md, q, userID)
	if err != nil {
		return nil, terror.Error(err, "Issue getting user metadata.")
	}

	return &md.Metadata, nil
}
