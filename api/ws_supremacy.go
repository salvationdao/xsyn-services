package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"passport"
	"passport/db"
	"passport/log_helpers"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// SupremacyControllerWS holds handlers for supremacying supremacy status
type SupremacyControllerWS struct {
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	SupremacyUserID passport.UserID
}

// NewSupremacyController creates the supremacy hub
func NewSupremacyController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *SupremacyControllerWS {
	supremacyHub := &SupremacyControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "supremacy"),
		API:  api,
	}

	supID, err := db.UserIDFromUsername(context.Background(), conn, passport.SupremacyGameUsername)
	if err != nil {
		log.Panic().Err(err).Msgf("unable to find supremacy user")
	}

	supremacyHub.SupremacyUserID = *supID

	api.SupremacyCommand(HubKeySupremacyTakeSups, supremacyHub.SupremacyTakeSupsHandler)
	api.SupremacyCommand(HubKeySupremacyTickerTick, supremacyHub.SupremacyTickerTickHandler)

	return supremacyHub
}

const HubKeySupremacyTakeSups = hub.HubCommandKey("SUPREMACY:TAKE_SUPS")

type SupremacyTakeSupsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount               passport.BigInt `json:"amount"`
		FromUserID           passport.UserID `json:"userId"`
		TransactionReference string          `json:"transactionReference"`
	} `json:"payload"`
}

func (ctrlr *SupremacyControllerWS) SupremacyTakeSupsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyTakeSupsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	tx := &Transaction{
		From:                 req.Payload.FromUserID,
		To:                   ctrlr.SupremacyUserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount,
	}

	ctrlr.API.transaction <- tx

	reply(struct {
		IsSuccess bool `json:"isSuccess"`
	}{
		IsSuccess: true,
	})
	return nil
}

const HubKeySupremacyTickerTick = hub.HubCommandKey("SUPREMACY:TICKER_TICK")

type SupremacyTickerTickRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		// this is a map of multipliers with a slice of users per multiplier
		UserMap map[int][]*passport.UserID `json:"userMap"`
	} `json:"payload"`
}

func (ctrlr *SupremacyControllerWS) SupremacyTickerTickHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyTickerTickRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	reference := fmt.Sprintf("supremacy|ticker|%s", time.Now())

	// TODO: get token pool from somewhere
	supPool := big.NewInt(10000000000000000) // just setting the pool at 1000
	var transactions []*Transaction
	totalPoints := 0
	//  to avoid working in floats, a 100% multiplier is 100 points, a 25% is 25 points
	// This will give us what we need to divide the pool by and then times by to give the user the correct share of the pool

	// loop once to get total point count
	for multiplier, users := range req.Payload.UserMap {
		totalPoints = totalPoints + (multiplier * len(users))
	}

	if totalPoints == 0 {
		return nil
	}

	onePointWorth := big.NewInt(0)
	onePointWorth.Div(supPool, big.NewInt(int64(totalPoints)))

	// loop again to create all transactions
	for multiplier, users := range req.Payload.UserMap {
		for _, user := range users {
			usersSups := big.NewInt(0)
			usersSups = usersSups.Mul(onePointWorth, big.NewInt(int64(multiplier)))

			transactions = append(transactions, &Transaction{
				From:                 ctrlr.SupremacyUserID,
				To:                   *user,
				Amount:               passport.BigInt{Int: *usersSups},
				TransactionReference: reference,
			})

			supPool = supPool.Sub(supPool, usersSups)
		}
	}

	// send through transactions
	for _, tx := range transactions {
		ctrlr.API.transaction <- tx
	}

	reply(struct {
		IsSuccess bool `json:"isSuccess"`
	}{
		IsSuccess: true,
	})
	return nil
}
