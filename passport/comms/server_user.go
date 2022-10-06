package comms

import (
	"fmt"
	"html"
	"net/mail"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/microcosm-cc/bluemonday"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type UserResp struct {
	ID               string
	Username         string
	FactionID        null.String
	PublicAddress    null.String
	AcceptsMarketing null.Bool
}

func (s *S) UserGetHandler(req UserGetReq, resp *UserResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	user, err := boiler.Users(
		qm.Select(
			boiler.UserColumns.ID,
			boiler.UserColumns.Username,
			boiler.UserColumns.FactionID,
			boiler.UserColumns.PublicAddress,
		),
		boiler.UserWhere.ID.EQ(req.UserID.String()),
	).One(passdb.StdConn)
	if err != nil {
		return err
	}

	resp.ID = user.ID
	resp.Username = user.Username
	resp.FactionID = user.FactionID
	resp.PublicAddress = user.PublicAddress
	resp.AcceptsMarketing = user.AcceptsMarketing

	return nil
}

type UserMarketingUpdateRequest struct {
	ApiKey           string
	UserID           string `json:"userID"`
	AcceptsMarketing bool   `json:"acceptsMarketing"`
	NewEmail         string `json:"newEmail"`
}

func (s *S) UserMarketingUpdateHandler(req UserMarketingUpdateRequest, resp *struct{}) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	// check user is enlisted
	user, err := boiler.Users(
		boiler.UserWhere.ID.EQ(req.UserID),
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user from db")
	}

	user.AcceptsMarketing = null.BoolFrom(req.AcceptsMarketing)

	if !user.Email.Valid {
		if req.NewEmail == "" {
			return terror.Error(fmt.Errorf("user email was not provided"), "User email is null, but no new email was provided when updating marketing preferences")
		}

		lowerEmail := strings.ToLower(req.NewEmail)
		_, err := mail.ParseAddress(lowerEmail)
		if err != nil {
			return terror.Error(err, "Invalid email address.")
		}

		// Check if email address is already taken
		u, _ := users.Email(lowerEmail)
		if u != nil {
			err = fmt.Errorf("email address is already taken by another user")
			return terror.Error(err, "Email address is already taken by another user.")
		}
		user.Email = null.StringFrom(lowerEmail)
	}

	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.AcceptsMarketing, boiler.UserColumns.Email))
	if err != nil {
		return terror.Error(err, "Failed to update user's marketing preferences.")
	}

	return nil
}

func (s *S) UserBalanceGetHandler(req UserBalanceGetReq, resp *UserBalanceGetResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	sups, err := s.UserCacheMap.Get(req.UserID.String())
	if err != nil {
		passlog.L.Error().Str("user_id", req.UserID.String()).Err(err).Msg("Failed to get user balance")
		return err
	}

	resp.Balance = sups
	return nil
}

func (s *S) UserFactionEnlistHandler(req UserFactionEnlistReq, resp *UserFactionEnlistResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	// check user is enlisted
	user, err := boiler.Users(
		boiler.UserWhere.ID.EQ(req.UserID),
	).One(passdb.StdConn)
	if err != nil {
		return terror.Error(err, "Failed to get user from db")
	}

	if user.FactionID.Valid {
		return terror.Error(fmt.Errorf("user already has faction"), "User already has faction")
	}

	user.FactionID = null.StringFrom(req.FactionID)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.FactionID))
	if err != nil {
		return terror.Error(err, "Failed to update user faction")
	}

	return nil
}

type UsernameUpdateReq struct {
	UserID      string `json:"user_id"`
	NewUsername string `json:"new_username"`
	ApiKey      string
}

type UsernameUpdateResp struct {
	Username string
}

func (s *S) UserUpdateUsername(req UsernameUpdateReq, resp *UsernameUpdateResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	// get user
	user, err := boiler.FindUser(passdb.StdConn, req.UserID)
	if err != nil {
		passlog.L.Error().Msg("Failed to get user")
		return terror.Error(err, "unable to get user")
	}

	// for activity record
	oldUser := user

	if req.NewUsername == "" {
		passlog.L.Error().Msg("Username cannot be empty")
		return terror.Error(err, "Username cannot be empty")
	}
	// Validate username
	err = helpers.IsValidUsername(req.NewUsername)
	if err != nil {
		passlog.L.Error().Msg("username invalid")
		return terror.Error(err, "username invalid")
	}

	bm := bluemonday.StrictPolicy()
	sanitizedUsername := html.UnescapeString(bm.Sanitize(strings.TrimSpace(req.NewUsername)))

	user.Username = sanitizedUsername
	// update
	user.UpdatedAt = time.Now()
	_, err = user.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Msg("unable to update username")
		return terror.Error(err, "unable to update username, try again or contact support")
	}

	resp.Username = sanitizedUsername

	// add to user activity
	s.API.RecordUserActivity(nil,
		user.ID,
		"Updated User",
		types.ObjectTypeUser,
		helpers.StringPointer(user.ID),
		&user.Username,
		helpers.StringPointer(user.FirstName.String+" "+user.LastName.String),
		&types.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	return nil
}
