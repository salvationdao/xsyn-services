package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"passport/db"
	"passport/log_helpers"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v3"
	"github.com/ninja-software/hub/v3/ext/messagebus"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/rs/zerolog"
)

// SupremacyControllerWS holds handlers for supremacy and the supremacy held transactions
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

	supremacyHub.SupremacyUserID = passport.SupremacyGameUserID

	// start nft repair ticker
	tickle.New("NFT Repair Ticker", 60, func() (int, error) {
		err := db.XsynNftMetadataDurabilityBulkIncrement(context.Background(), conn)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err)
		}
		return http.StatusOK, nil
	}).Start()

	// hold sups
	api.SupremacyCommand(HubKeySupremacyHoldSups, supremacyHub.SupremacyHoldSupsHandler)
	// commit holds
	api.SupremacyCommand(HubKeySupremacyCommitTransactions, supremacyHub.SupremacyCommitTransactionsHandler)
	// release holds
	api.SupremacyCommand(HubKeySupremacyReleaseTransactions, supremacyHub.SupremacyReleaseTransactionsHandler)

	api.SupremacyCommand(HubKeySupremacyTickerTick, supremacyHub.SupremacyTickerTickHandler)

	api.SupremacyCommand(HubKeySupremacyAssetFreeze, supremacyHub.SupremacyAssetFreezeHandler)
	api.SupremacyCommand(HubKeySupremacyAssetLock, supremacyHub.SupremacyAssetLockHandler)
	api.SupremacyCommand(HubKeySupremacyAssetRelease, supremacyHub.SupremacyAssetReleaseHandler)
	api.SupremacyCommand(HubKeySupremacyWarMachineQueuePosition, supremacyHub.SupremacyWarMachineQueuePositionHandler)

	return supremacyHub
}

const HubKeySupremacyHoldSups = hub.HubCommandKey("SUPREMACY:HOLD_SUPS")

type SupremacyHoldSupsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount               passport.BigInt      `json:"amount"`
		FromUserID           passport.UserID      `json:"userID"`
		TransactionReference TransactionReference `json:"transactionReference"`
	} `json:"payload"`
}

func (ctrlr *SupremacyControllerWS) SupremacyHoldSupsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyHoldSupsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	tx := &NewTransaction{
		From:                 req.Payload.FromUserID,
		To:                   ctrlr.SupremacyUserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount.Int,
	}

	ctrlr.API.HoldTransaction(tx)

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
	supPool := big.NewInt(0) // just setting the pool at 1000
	supPool, ok := supPool.SetString("1000000000000000000000", 10)
	if !ok {
		return terror.Error(fmt.Errorf("failed to convert 1000000000000000000000 to big int"))
	}
	var transactions []*NewTransaction
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

			transactions = append(transactions, &NewTransaction{
				From:                 ctrlr.SupremacyUserID,
				To:                   *user,
				Amount:               *usersSups,
				TransactionReference: TransactionReference(reference),
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

type SupremacyAssetFreezeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetTokenID uint64 `json:"assetTokenID"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeySupremacyAssetFreeze, AssetController.RegisterHandler)
const HubKeySupremacyAssetFreeze hub.HubCommandKey = "SUPREMACY:ASSET:FREEZE"

func (ctrlr *SupremacyControllerWS) SupremacyAssetFreezeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyAssetFreezeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	err = db.XsynAssetFreeze(ctx, ctrlr.Conn, req.Payload.AssetTokenID, userID)
	if err != nil {
		reply(false)
		return terror.Error(err)
	}

	// TODO: In the future, charge user's sups for joining the queue

	reply(true)
	return nil
}

type SupremacyAssetLockRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetTokenIDs []uint64 `json:"assetTokenIDs"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeySupremacyAssetFreeze, AssetController.RegisterHandler)
const HubKeySupremacyAssetLock hub.HubCommandKey = "SUPREMACY:ASSET:LOCK"

func (ctrlr *SupremacyControllerWS) SupremacyAssetLockHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyAssetLockRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	err = db.XsynAssetBulkLock(ctx, ctrlr.Conn, req.Payload.AssetTokenIDs, userID)
	if err != nil {
		reply(false)
		return terror.Error(err)
	}

	reply(true)
	return nil
}

type SupremacyAssetReleaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ReleasedAssets []*passport.WarMachineNFT `json:"releasedAssets"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeySupremacyAssetFreeze, AssetController.RegisterHandler)
const HubKeySupremacyAssetRelease hub.HubCommandKey = "SUPREMACY:ASSET:RELEASE"

func (ctrlr *SupremacyControllerWS) SupremacyAssetReleaseHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyAssetReleaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	tx, err := ctrlr.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ctrlr.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	err = db.XsynAssetBulkRelease(ctx, tx, req.Payload.ReleasedAssets, userID)
	if err != nil {
		return terror.Error(err)
	}

	err = db.XsynNftMetadataDurabilityBulkUpdate(ctx, tx, req.Payload.ReleasedAssets)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// 	rootHub.SecureCommand(HubKeySupremacyAssetFreeze, AssetController.RegisterHandler)
const HubKeySupremacyWarMachineQueuePosition hub.HubCommandKey = "SUPREMACY:WAR:MACHINE:QUEUE:POSITION"

type SupremacyWarMachineQueuePositionRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserWarMachineQueuePosition []*UserWarMachineQueuePosition `json:"userWarMachineQueuePosition"`
	} `json:"payload"`
}
type UserWarMachineQueuePosition struct {
	UserID                   passport.UserID            `json:"userID"`
	WarMachineQueuePositions []*WarMachineQueuePosition `json:"warMachineQueuePositions"`
}

type WarMachineQueuePosition struct {
	WarMachineNFT *passport.WarMachineNFT `json:"warMachineNFT"`
	Position      int                     `json:"position"`
}

// SupremacyWarMachineQueuePositionHandler broadcast the updated battle queue position detail
func (ctrlr *SupremacyControllerWS) SupremacyWarMachineQueuePositionHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyWarMachineQueuePositionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// broadcast war machine position to all user client
	for _, uwm := range req.Payload.UserWarMachineQueuePosition {
		go ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueuePositionSubscribe, uwm.UserID)), uwm.WarMachineQueuePositions)
	}

	return nil
}

type SupremacyWarMachineQueuePositionClearRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"factionID"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeySupremacyWarMachineQueuePositionClear, AssetController.RegisterHandler)
const HubKeySupremacyWarMachineQueuePositionClear hub.HubCommandKey = "SUPREMACY:WAR:MACHINE:QUEUE:POSITION:CLEAR"

// SupremacyWarMachineQueuePositionClearHandler broadcast user to clear the war machine queue
func (ctrlr *SupremacyControllerWS) SupremacyWarMachineQueuePositionClearHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyWarMachineQueuePositionClearRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get faction users
	userIDs, err := db.UserIDsGetByFactionID(ctx, ctrlr.Conn, req.Payload.FactionID)
	if err != nil {
		return terror.Error(err, "Failed to get user id from faction")
	}

	if len(userIDs) == 0 {
		return nil
	}

	// broadcast war machine position to all user client
	for _, userID := range userIDs {
		go ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueuePositionSubscribe, userID)), []*WarMachineQueuePosition{})
	}

	return nil
}

// 	api.SupremacyCommand(HubKeySupremacyCommitTransactions, supremacyHub.SupremacyCommitTransactions)
const HubKeySupremacyCommitTransactions = hub.HubCommandKey("SUPREMACY:COMMIT_TRANSACTIONS")

type SupremacyCommitTransactionsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TransactionReferences []TransactionReference `json:"transactionReferences"`
	} `json:"payload"`
}

func (ctrlr *SupremacyControllerWS) SupremacyCommitTransactionsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyCommitTransactionsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	resultChan := make(chan []*passport.Transaction, len(req.Payload.TransactionReferences)+5)

	ctrlr.API.CommitTransactions(resultChan, req.Payload.TransactionReferences...)

	results := <-resultChan

	reply(results)
	return nil
}

const HubKeySupremacyReleaseTransactions = hub.HubCommandKey("SUPREMACY:RELEASE_TRANSACTIONS")

type SupremacyReleaseTransactionsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TransactionReferences []TransactionReference `json:"transactions"`
	} `json:"payload"`
}

func (ctrlr *SupremacyControllerWS) SupremacyReleaseTransactionsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyReleaseTransactionsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// TODO: this is totally untested btw.

	resultChan := make(chan []*passport.Transaction)

	ctrlr.API.ReleaseHeldTransaction(req.Payload.TransactionReferences...)

	results := <-resultChan

	reply(results)
	return nil
}
