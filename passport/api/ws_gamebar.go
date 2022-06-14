package api

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/shopspring/decimal"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/log_helpers"
	"github.com/rs/zerolog"
)

// GamebarController holds handlers for authentication
type GamebarController struct {
	Log *zerolog.Logger
	API *API
}

func NewGamebarController(log *zerolog.Logger, api *API) *GamebarController {
	gamebarHub := &GamebarController{
		Log: log_helpers.NamedLogger(log, "user_hub"),
		API: api,
	}
	if os.Getenv("PASSPORT_ENVIRONMENT") == "development" || os.Getenv("PASSPORT_ENVIRONMENT") == "staging" {

		api.SecureCommand(HubKeyGamebarGetFreeSups, gamebarHub.GetFreeSups)
	}
	//api.SecureCommand(HubKeyGamebarUserSubscribe, gamebarHub.UserUpdatedSubscribeHandler)

	return gamebarHub
}

var timeMap = sync.Map{}

const HubKeyGamebarGetFreeSups = "GAMEBAR:GET:SUPS"

func (gc *GamebarController) GetFreeSups(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	if os.Getenv("PASSPORT_ENVIRONMENT") != "development" && os.Getenv("PASSPORT_ENVIRONMENT") != "staging" {
		// If not development or staging
		passlog.L.
			Error().
			Str("env", os.Getenv("PASSPORT_ENVIRONMENT")).
			Msg("NO SUPS FOR YOU :p (not staging or development)")
		reply(false)
		return nil
	}

	cooldown := time.Hour
	if os.Getenv("PASSPORT_ENVIRONMENT") == "development" {
		cooldown = time.Second * 5
	}

	allowed := false
	t, found := timeMap.Load(fmt.Sprintf("%s:GET_SUPS", user.ID))
	// Get time til next claim
	tm, ok := t.(time.Time)
	if found {
		// If user has claimed sups before
		if ok && time.Now().After(tm) {
			// If the current time is after time til next claim
			allowed = true
		}
	} else {
		// If user has not claimed sups before
		allowed = true
	}

	if !allowed {
		passlog.L.
			Error().
			Interface("timeUntilClaim", tm).
			Msg(fmt.Sprintf("NO SUPS FOR YOU :p (on cooldown)"))
		reply(tm)
		return nil
	}

	// If user can claim free sups
	// Give them 100 sups
	oneSups := big.NewInt(1000000000000000000)
	oneSups.Mul(oneSups, big.NewInt(100))
	tx := &types.NewTransaction{
		To:                   types.UserID(uuid.Must(uuid.FromString(user.ID))),
		From:                 types.XsynSaleUserID,
		Amount:               decimal.NewFromBigInt(oneSups, 0),
		NotSafe:              true,
		TransactionReference: types.TransactionReference(fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())),
		Description:          "100 SUPS giveaway for testing",
		Group:                types.TransactionGroupTesting,
	}
	_, err := gc.API.userCacheMap.Transact(tx)
	if err != nil {
		passlog.L.
			Err(err).
			Str("to", tx.To.String()).
			Str("from", tx.From.String()).
			Str("amount", tx.Amount.String()).
			Str("description", tx.Description).
			Str("transaction_reference", string(tx.TransactionReference)).
			Msg("NO SUPS FOR YOU :p (transaction failed)")
		reply(false)
		return nil
	}
	// Set their timer to one hour from now (or whatever the cooldown is)
	nextClaimTime := time.Now().Add(cooldown)
	timeMap.Store(fmt.Sprintf("%s:GET_SUPS", user.ID), nextClaimTime)

	reply(nextClaimTime)
	return nil
}
