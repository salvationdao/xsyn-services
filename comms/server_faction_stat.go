package comms

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"math/big"
	"passport"
	"passport/api"
	"passport/db"
	"passport/passlog"

	"github.com/jackc/pgx/v4"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

func (c *S) SupremacyFactionAllHandler(req FactionAllReq, resp *FactionAllResp) error {
	passlog.L.Trace().Str("fn", "SupremacyFactionAllHandler").Msg("rpc handler")
	var err error
	resp.Factions, err = db.FactionAll(context.Background(), c.Conn)
	if err != nil {
		return terror.Error(err, "Failed to query faction from db")
	}

	return nil
}

type FactionStatSendReq struct {
	FactionStatSends []*FactionStatSend `json:"faction_stat_sends"`
}

type FactionStatSend struct {
	FactionStat *passport.FactionStat `json:"faction_stat"`
}

type FactionStatSendResp struct{}

func (c *S) SupremacyFactionStatSendHandler(req FactionStatSendReq, resp *FactionStatSendResp) error {
	passlog.L.Trace().Str("fn", "SupremacyFactionStatSendHandler").Msg("rpc handler")
	ctx := context.Background()
	for _, factionStatSend := range req.FactionStatSends {
		// get recruit number
		recruitNumber, err := db.FactionGetRecruitNumber(ctx, c.Conn, factionStatSend.FactionStat.ID)
		if err != nil {
			c.Log.Err(err).Msgf("Failed to get recruit number from faction %s", factionStatSend.FactionStat.ID)
			continue
		}
		factionStatSend.FactionStat.RecruitNumber = recruitNumber

		// get velocity number
		// TODO: figure out what velocity is
		factionStatSend.FactionStat.Velocity = 0

		// get mvp

		mvp, err := db.FactionMvpGet(ctx, c.Conn, factionStatSend.FactionStat.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			c.Log.Err(err).Msgf("failed to get mvp from faction %s", factionStatSend.FactionStat.ID)
			continue
		}
		factionStatSend.FactionStat.MVP = mvp

		supsVoted, err := db.FactionSupsVotedGet(ctx, c.Conn, factionStatSend.FactionStat.ID)
		if err != nil {
			c.Log.Err(err).Msgf("failed to get sups voted from faction %s", factionStatSend.FactionStat.ID)
			continue
		}

		factionStatSend.FactionStat.SupsVoted = supsVoted.String()

		// broadcast to all faction stat subscribers
		c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyFactionStatUpdatedSubscribe, factionStatSend.FactionStat.ID)), factionStatSend.FactionStat)
		continue
	}

	return nil
}

//****************************************
//  CONTRACT REWARD
//****************************************
type FactionContractRewardUpdateReq struct {
	FactionContractRewards []*FactionContractReward `json:"faction_contract_rewards"`
}

type FactionContractReward struct {
	FactionID      passport.FactionID `json:"faction_id"`
	ContractReward string             `json:"contract_reward"`
}

type FactionContractRewardUpdateResp struct {
}

func (c *S) SupremacyFactionContractRewardUpdateHandler(req FactionContractRewardUpdateReq, resp *FactionContractRewardUpdateResp) error {
	passlog.L.Trace().Str("fn", "SupremacyFactionContractRewardUpdateHandler").Msg("rpc handler")
	ctx := context.Background()
	for _, fcr := range req.FactionContractRewards {
		c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyFactionContractRewardSubscribe, fcr.FactionID)), fcr.ContractReward)
	}

	return nil
}

type RedeemFactionContractRewardReq struct {
	UserID               passport.UserID               `json:"user_id"`
	FactionID            passport.FactionID            `json:"faction_id"`
	BattleID             string                        `json:"battle_id"`
	Amount               string                        `json:"amount"`
	TransactionReference passport.TransactionReference `json:"transaction_reference"`
}

type RedeemFactionContractRewardResp struct{}

func (c *S) SupremacyRedeemFactionContractRewardHandler(req RedeemFactionContractRewardReq, resp *RedeemFactionContractRewardResp) error {
	passlog.L.Trace().Str("fn", "SupremacyRedeemFactionContractRewardHandler").Msg("rpc handler")

	amount := big.NewInt(0)
	amount, ok := amount.SetString(req.Amount, 10)
	if !ok {
		return terror.Error(fmt.Errorf("failed to parse amount into big int"), "Could not parse reward amount")
	}

	if amount.Cmp(big.NewInt(0)) <= 0 {
		return terror.Error(terror.ErrInvalidInput, "Sups amount can not be negative")
	}

	tx := &passport.NewTransaction{
		To:                   req.UserID,
		TransactionReference: req.TransactionReference,
		Amount:               decimal.NewFromBigInt(amount, 0),
		Description:          "Contract Reward",
		Group:                passport.TransactionGroupBattle,
		SubGroup:             req.BattleID,
	}

	switch req.FactionID {
	case passport.RedMountainFactionID:
		tx.From = passport.SupremacyRedMountainUserID
	case passport.BostonCyberneticsFactionID:
		tx.From = passport.SupremacyBostonCyberneticsUserID
	case passport.ZaibatsuFactionID:
		tx.From = passport.SupremacyZaibatsuUserID
	default:
		return terror.Error(terror.ErrInvalidInput, "Provided faction does not exist")
	}

	// process user cache map
	_, _, _, err := c.UserCacheMap.Process(tx)
	if err != nil {
		return terror.Error(err, "failed to process fund")
	}
	return nil
}

type FactionQueuingCostReq struct {
	FactionID     passport.FactionID `json:"faction_id"`
	QueuingLength int                `json:"queuing_length"`
}

type FactionQueuingCostResp struct{}

func (c *S) SupremacyFactionQueuingCostHandler(req FactionQueuingCostReq, resp *FactionQueuingCostResp) error {
	passlog.L.Trace().Str("fn", "SupremacyFactionQueuingCostHandler").Msg("rpc handler")
	cost := big.NewInt(1000000000000000000)
	cost.Mul(cost, big.NewInt(int64(req.QueuingLength)+1))

	c.MessageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyAssetRepairStatUpdate, req.FactionID)), cost.String())
	return nil
}
