package comms

import (
	"fmt"
	"html"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"

	"github.com/microcosm-cc/bluemonday"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

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

	return nil
}

type UserGetResp struct {
	User *UserResp `json:"user"`
}

type UserResp struct {
	ID            string
	Username      string
	FactionID     null.String
	PublicAddress null.String
}

func (s *S) UserBalanceGetHandler(req UserBalanceGetReq, resp *UserBalanceGetResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	balance, err := s.UserCacheMap.Get(req.UserID.String())
	if err != nil {
		passlog.L.Error().Str("user_id", req.UserID.String()).Err(err).Msg("Failed to get user balance")
		return err
	}

	// convert balance from big int to decimal
	b, err := decimal.NewFromString(balance.String())
	if err != nil {
		passlog.L.Error().Str("big int amount", balance.String()).Err(err).Msg("Failed to get convert big int to decimal")
		return terror.Error(err, "failed to convert big int to decimal")
	}

	resp.Balance = b
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

func (s *S) UserUpdateUsername(req UsernameUpdateReq, resp *UserResp) error {
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

	if req.NewUsername != "" {
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

		resp.ID = user.ID
		resp.Username = sanitizedUsername
		resp.FactionID = user.FactionID
		resp.PublicAddress = user.PublicAddress

	}

	// add to user activity

	return nil
}
