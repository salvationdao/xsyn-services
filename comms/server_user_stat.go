package comms

import (
	"context"
	"fmt"
	"passport"
	"passport/api"
	"passport/db"
	"passport/passlog"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/shopspring/decimal"
)

type UsersGetReq struct {
	UserIDs []passport.UserID `json:"userIDs"`
}

type UsersGetResp struct {
	Users []*passport.User `json:"users"`
}

func (c *S) SupremacyUsersGetHandler(req UsersGetReq, resp *UsersGetResp) error {
	ctx := context.Background()

	var err error
	resp.Users, err = db.UserGetByIDs(ctx, c.Conn, req.UserIDs)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

type UserStatSendReq struct {
	UserStatSends []*UserStatSend `json:"userStatSends"`
}

type UserStatSend struct {
	ToUserSessionID *hub.SessionID     `json:"toUserSessionID,omitempty"`
	Stat            *passport.UserStat `json:"stat"`
}

type UserStatSendResp struct{}

func (c *S) SupremacyUserStatSendHandler(req UserStatSendReq, resp *UserStatSendResp) error {
	ctx := context.Background()
	for _, userStatSend := range req.UserStatSends {

		if userStatSend.ToUserSessionID == nil {
			// broadcast to all faction stat subscribers
			go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserStatSubscribe, userStatSend.Stat.ID)), userStatSend.Stat)
			continue
		}

		// broadcast to specific subscribers
		filterOption := messagebus.BusSendFilterOption{}
		if userStatSend.ToUserSessionID != nil {
			filterOption.SessionID = *userStatSend.ToUserSessionID
		}

		go c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyUserStatSubscribe, userStatSend.Stat.ID)), userStatSend.Stat, filterOption)
	}

	return nil
}

type UserBalanceGetReq struct {
	UserID uuid.UUID `json:"userID"`
}

type UserBalanceGetResp struct {
	Balance decimal.Decimal `json:"balance"`
}

func (c *S) SupremacyUserBalanceGetHandler(req UserBalanceGetReq, resp *UserBalanceGetResp) error {
	balance, err := c.UserCacheMap.Get(req.UserID.String())
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
