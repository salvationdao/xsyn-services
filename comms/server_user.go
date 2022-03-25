package comms

import (
	"context"
	"passport"
	"passport/db"
	"passport/passdb"
	"passport/passlog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

type UserReq struct {
	ID uuid.UUID
}
type UserResp struct {
	ID            uuid.UUID
	Username      string
	FactionID     null.String
	PublicAddress common.Address
}

func (c *S) User(req UserReq, resp *UserResp) error {
	ctx := context.Background()
	u, err := db.UserGet(ctx, passdb.Conn, passport.UserID(req.ID))
	if err != nil {
		return err
	}
	resp.FactionID = null.NewString(u.FactionID.String(), u.FactionID == nil)
	resp.ID = uuid.UUID(u.ID)
	resp.PublicAddress = common.HexToAddress(u.PublicAddress.String)
	resp.Username = u.Username

	return nil
}

type UserGetReq struct {
	ApiKey string
	UserID passport.UserID `json:"userID"`
}

type UserGetResp struct {
	User *passport.User `json:"user"`
}

func (s *S) UserGetHandler(req UserGetReq, resp *UserGetResp) error {
	err := IsServerClient(req.ApiKey)
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

type UserBalanceGetReq struct {
	ApiKey string
	UserID uuid.UUID `json:"userID"`
}

type UserBalanceGetResp struct {
	Balance decimal.Decimal `json:"balance"`
}

func (s *S) UserBalanceGetHandler(req UserBalanceGetReq, resp *UserBalanceGetResp) error {
	err := IsServerClient(req.ApiKey)
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
