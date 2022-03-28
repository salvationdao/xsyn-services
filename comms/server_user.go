package comms

import (
	"context"
	"passport/db"
	"passport/passlog"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

func (s *S) UserGetHandler(req UserGetReq, resp *UserGetResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return terror.Error(err)
	}

	ctx := context.Background()

	resp.User, err = db.UserGet(ctx, s.Conn, req.UserID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func (s *S) UserBalanceGetHandler(req UserBalanceGetReq, resp *UserBalanceGetResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return terror.Error(err)
	}

	balance, err := s.UserCacheMap.Get(req.UserID.String())
	if err != nil {
		passlog.L.Error().Str("user_id", req.UserID.String()).Err(err).Msg("Failed to get user balance")
		return terror.Error(err)
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
