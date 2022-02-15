package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"passport"
	"passport/db"
	"passport/log_helpers"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

// SupremacyControllerWS holds handlers for supremacy and the supremacy held transactions
type SupremacyControllerWS struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewSupremacyController creates the supremacy hub
func NewSupremacyController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *SupremacyControllerWS {
	supremacyHub := &SupremacyControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "supremacy"),
		API:  api,
	}

	// sup control
	api.SupremacyCommand(HubKeySupremacyHoldSups, supremacyHub.SupremacyHoldSupsHandler)
	api.SupremacyCommand(HubKeySupremacyCommitTransactions, supremacyHub.SupremacyCommitTransactionsHandler)
	api.SupremacyCommand(HubKeySupremacyReleaseTransactions, supremacyHub.SupremacyReleaseTransactionsHandler)
	api.SupremacyCommand(HubKeySupremacyTickerTick, supremacyHub.SupremacyTickerTickHandler)
	api.SupremacyCommand(HubKeySupremacyGetSpoilOfWar, supremacyHub.SupremacyGetSpoilOfWarHandler)
	api.SupremacyCommand(HubKeySupremacyTransferBattleFundToSupPool, supremacyHub.SupremacyTransferBattleFundToSupPoolHandler)
	api.SupremacyCommand(HubKeySupremacyUserSupsMultiplierSend, supremacyHub.SupremacyUserSupsMultiplierSendHandler)

	// user connection upgrade
	api.SupremacyCommand(HubKeySupremacyUserConnectionUpgrade, supremacyHub.SupremacyUserConnectionUpgradeHandler)

	// asset control
	api.SupremacyCommand(HubKeySupremacyAssetFreeze, supremacyHub.SupremacyAssetFreezeHandler)
	api.SupremacyCommand(HubKeySupremacyAssetLock, supremacyHub.SupremacyAssetLockHandler)
	api.SupremacyCommand(HubKeySupremacyAssetRelease, supremacyHub.SupremacyAssetReleaseHandler)
	api.SupremacyCommand(HubKeySupremacyWarMachineQueuePosition, supremacyHub.SupremacyWarMachineQueuePositionHandler)
	api.SupremacyCommand(HubKeySupremacyPayAssetInsurance, supremacyHub.SupremacyPayAssetInsuranceHandler)

	// battle queue
	api.SupremacyCommand(HubKeySupremacyDefaultWarMachines, supremacyHub.SupremacyDefaultWarMachinesHandler)
	api.SupremacyCommand(HubKeySupremacyWarMachineQueueContractUpdate, supremacyHub.SupremacyWarMachineQueueContractUpdateHandler)
	api.SupremacyCommand(HubKeySupremacyRedeemFactionContractReward, supremacyHub.SupremacyRedeemFactionContractRewardHandler)

	// sups contribute
	api.SupremacyCommand(HubKeySupremacyAbilityTargetPriceUpdate, supremacyHub.SupremacyAbilityTargetPriceUpdate)

	// faction stat
	api.SupremacyCommand(HubKeySupremacyFactionStatSend, supremacyHub.SupremacyFactionStatSend)

	return supremacyHub
}

type SupremacyUserConnectionUpgradeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SessionID hub.SessionID `json:"sessionID"`
	} `json:"payload"`
}

const HubKeySupremacyUserConnectionUpgrade = hub.HubCommandKey("SUPREMACY:USER_CONNECTION_UPGRADE")

func (sc *SupremacyControllerWS) SupremacyUserConnectionUpgradeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyUserConnectionUpgradeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	sc.API.Hub.Clients(func(clients hub.ClientsList) {
		for cl := range clients {
			if cl.SessionID == req.Payload.SessionID {
				cl.SetLevel(2)

				sc.API.Log.Info().Msgf("Hub client %s has been upgraded to level 2 client", cl.SessionID)
				break
			}
		}
	})

	reply(true)

	return nil
}

const HubKeySupremacyHoldSups = hub.HubCommandKey("SUPREMACY:HOLD_SUPS")

type SupremacyHoldSupsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount               passport.BigInt               `json:"amount"`
		FromUserID           passport.UserID               `json:"userID"`
		TransactionReference passport.TransactionReference `json:"transactionReference"`
		IsBattleVote         bool                          `json:"isBattleVote"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyHoldSupsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyHoldSupsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Amount.Cmp(big.NewInt(0)) < 0 {
		return terror.Error(terror.ErrInvalidInput, "Sups amount can not be negative")
	}

	tx := &passport.NewTransaction{
		From:                 req.Payload.FromUserID,
		To:                   passport.SupremacyGameUserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount.Int,
	}

	if req.Payload.IsBattleVote {
		tx.To = passport.SupremacyBattleUserID
	}

	errChan := make(chan error, 10)
	sc.API.HoldTransaction(errChan, tx)

	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	reply(true)
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

func (sc *SupremacyControllerWS) SupremacyTickerTickHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyTickerTickRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	//  to avoid working in floats, a 100% multiplier is 100 points, a 25% is 25 points
	// This will give us what we need to divide the pool by and then times by to give the user the correct share of the pool

	totalPoints := 0
	// loop once to get total point count
	for multiplier, users := range req.Payload.UserMap {
		totalPoints = totalPoints + (multiplier * len(users))
	}

	if totalPoints == 0 {
		return nil
	}

	var transactions []*passport.NewTransaction

	/////////////////////////////////////////
	//  Standard Sups From Supremacy User  //
	/////////////////////////////////////////

	// 50 sups per 60 second
	// supremacy ticker tick every 3 second, so grab 2.5 sups on every tick
	supPool := big.NewInt(0)
	supPool, ok := supPool.SetString("2500000000000000000", 10)
	if !ok {
		return terror.Error(fmt.Errorf("failed to convert 2500000000000000000 to big int"))
	}

	onePointWorth := big.NewInt(0)
	onePointWorth.Div(supPool, big.NewInt(int64(totalPoints)))

	// loop again to create all transactions
	for multiplier, users := range req.Payload.UserMap {
		for _, user := range users {
			usersSups := big.NewInt(0)
			usersSups = usersSups.Mul(onePointWorth, big.NewInt(int64(multiplier)))

			transactions = append(transactions, &passport.NewTransaction{
				From:                 passport.SupremacyGameUserID,
				To:                   *user,
				Amount:               *usersSups,
				TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|ticker|%s|%s", *user, time.Now())),
			})

			supPool = supPool.Sub(supPool, usersSups)
		}
	}

	///////////////////////////////////
	//  Sups From  Battle Sups Pool  //
	///////////////////////////////////

	// get trickle amount
	trickleAmount := sc.API.SupremacySupPoolGetTrickleAmount()
	if trickleAmount.Cmp(big.NewInt(0)) > 0 {
		// cal
		onePointWorth := big.NewInt(0)
		onePointWorth.Div(&trickleAmount.Int, big.NewInt(int64(totalPoints)))

		// loop again to create all transactions
		for multiplier, users := range req.Payload.UserMap {
			for _, user := range users {
				usersSups := big.NewInt(0)
				usersSups = usersSups.Mul(onePointWorth, big.NewInt(int64(multiplier)))

				transactions = append(transactions, &passport.NewTransaction{
					From:                 passport.SupremacySupPoolUserID,
					To:                   *user,
					Amount:               *usersSups,
					TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|spoil_of_war|%s|%s", *user, time.Now())),
				})

				supPool = supPool.Sub(supPool, usersSups)
			}
		}
	}

	///////////////////////////
	//  Insert Transactions  //
	///////////////////////////

	// send through transactions
	for _, tx := range transactions {
		tx.ResultChan = make(chan *passport.TransactionResult, 1)
		sc.API.transaction <- tx
		result := <-tx.ResultChan
		// if result is success, update the cache map

		if result.Error != nil {
			continue // believe error logs already
		}

		if result.Transaction.Status != passport.TransactionSuccess {
			sc.API.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
			continue
		}

		errChan := make(chan error, 10)
		sc.API.UpdateUserCacheRemoveSups(tx.From, tx.Amount, errChan)
		err := <-errChan
		if err != nil {
			sc.API.Log.Err(err).Msg(err.Error())
			continue
		}
		sc.API.UpdateUserCacheAddSups(tx.To, tx.Amount)
	}

	reply(true)
	return nil
}

const HubKeySupremacyTransferBattleFundToSupPool = hub.HubCommandKey("SUPREMACY:TRANSFER_BATTLE_FUND_TO_SUP_POOL")

func (sc *SupremacyControllerWS) SupremacyTransferBattleFundToSupPoolHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	// get sups from battle user
	battleUser, err := db.UserGet(ctx, sc.Conn, passport.SupremacyBattleUserID)
	if err != nil {
		return terror.Error(err, "Failed to get battle arena user")
	}

	// skip calculation, if the is not sups in the pool
	if battleUser.Sups.Int.Cmp(big.NewInt(0)) <= 0 {
		reply(true)
		return nil
	}

	transaction := &passport.NewTransaction{
		From:                 passport.SupremacyBattleUserID,
		To:                   passport.SupremacySupPoolUserID,
		Amount:               battleUser.Sups.Int,
		TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|battle_sups_spend_transfer|%s", time.Now())),
	}

	// transfer all the sups to sups pool user
	transaction.ResultChan = make(chan *passport.TransactionResult, 1)
	sc.API.transaction <- transaction
	result := <-transaction.ResultChan
	// if result is success, update the cache map
	if result.Error != nil {
		return terror.Error(result.Error)
	}
	if result.Transaction.Status == passport.TransactionFailed {
		return terror.Error(fmt.Errorf("Failed to transfer sups spend over"))
	}

	// get sups pool user
	supsPoolUser, err := db.UserGet(context.Background(), sc.Conn, passport.SupremacySupPoolUserID)
	if err != nil {
		return terror.Error(err)
	}

	// set total sups pool
	sc.API.SupremacySupPoolSet(supsPoolUser.Sups)

	reply(true)

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

func (sc *SupremacyControllerWS) SupremacyAssetFreezeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyAssetFreezeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	err = db.XsynAssetFreeze(ctx, sc.Conn, req.Payload.AssetTokenID, userID)
	if err != nil {
		reply(false)
		return terror.Error(err)
	}

	asset, err := db.AssetGet(ctx, sc.Conn, req.Payload.AssetTokenID)
	if err != nil {
		reply(false)
		return terror.Error(err)
	}

	// TODO: In the future, charge user's sups for joining the queue

	sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, req.Payload.AssetTokenID)), asset)

	sc.API.SendToAllServerClient(&ServerClientMessage{
		Key: AssetUpdated,
		Payload: struct {
			Asset *passport.XsynMetadata `json:"asset"`
		}{
			Asset: asset,
		},
	})

	reply(true)
	return nil
}

type SupremacyAssetLockRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AssetTokenIDs []uint64 `json:"assetTokenIDs"`
	} `json:"payload"`
}

const HubKeySupremacyAssetLock hub.HubCommandKey = "SUPREMACY:ASSET:LOCK"

func (sc *SupremacyControllerWS) SupremacyAssetLockHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyAssetLockRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	err = db.XsynAssetBulkLock(ctx, sc.Conn, req.Payload.AssetTokenIDs, userID)
	if err != nil {
		return terror.Error(err)
	}

	_, assets, err := db.AssetList(
		ctx, sc.Conn,
		"", false, req.Payload.AssetTokenIDs, nil, "", 0, len(req.Payload.AssetTokenIDs), "", "",
	)
	if err != nil {
		return terror.Error(err)
	}

	for _, asset := range assets {
		sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.TokenID)), asset)
	}

	reply(true)
	return nil
}

type SupremacyAssetReleaseRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ReleasedAssets []*passport.WarMachineMetadata `json:"releasedAssets"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeySupremacyAssetFreeze, AssetController.RegisterHandler)
const HubKeySupremacyAssetRelease hub.HubCommandKey = "SUPREMACY:ASSET:RELEASE"

func (sc *SupremacyControllerWS) SupremacyAssetReleaseHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyAssetReleaseRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}

	tx, err := sc.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			sc.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	err = db.XsynAssetBulkRelease(ctx, tx, req.Payload.ReleasedAssets, userID)
	if err != nil {
		return terror.Error(err)
	}

	err = db.XsynAsseetDurabilityBulkUpdate(ctx, tx, req.Payload.ReleasedAssets)
	if err != nil {
		return terror.Error(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	tokenIDs := []uint64{}
	for _, ra := range req.Payload.ReleasedAssets {
		tokenIDs = append(tokenIDs, ra.TokenID)
		if ra.Durability < 100 {
			if ra.IsInsured {
				sc.API.RegisterRepairCenter(RepairTypeFast, ra.TokenID)
			} else {
				sc.API.RegisterRepairCenter(RepairTypeStandard, ra.TokenID)
			}
		}
	}

	_, assets, err := db.AssetList(
		ctx, sc.Conn,
		"", false, tokenIDs, nil, "", 0, len(tokenIDs), "", "",
	)
	if err != nil {
		return terror.Error(err)
	}

	for _, asset := range assets {
		sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.TokenID)), asset)
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
	WarMachineMetadata *passport.WarMachineMetadata `json:"warMachineMetadata"`
	Position           int                          `json:"position"`
}

// SupremacyWarMachineQueuePositionHandler broadcast the updated battle queue position detail
func (sc *SupremacyControllerWS) SupremacyWarMachineQueuePositionHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyWarMachineQueuePositionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// broadcast war machine position to all user client
	for _, uwm := range req.Payload.UserWarMachineQueuePosition {
		go sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueuePositionSubscribe, uwm.UserID)), uwm.WarMachineQueuePositions)
	}

	return nil
}

// 	api.SupremacyCommand(HubKeySupremacyCommitTransactions, supremacyHub.SupremacyCommitTransactions)
const HubKeySupremacyCommitTransactions = hub.HubCommandKey("SUPREMACY:COMMIT_TRANSACTIONS")

type SupremacyCommitTransactionsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TransactionReferences []passport.TransactionReference `json:"transactionReferences"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyCommitTransactionsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyCommitTransactionsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	resultChan := make(chan []*passport.Transaction, len(req.Payload.TransactionReferences)+5)
	sc.API.CommitTransactions(resultChan, req.Payload.TransactionReferences...)

	results := <-resultChan
	reply(results)
	return nil
}

const HubKeySupremacyReleaseTransactions = hub.HubCommandKey("SUPREMACY:RELEASE_TRANSACTIONS")

type SupremacyReleaseTransactionsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TransactionReferences []passport.TransactionReference `json:"transactionReferences"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyReleaseTransactionsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyReleaseTransactionsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	sc.API.ReleaseHeldTransaction(req.Payload.TransactionReferences...)

	return nil
}

// 	api.SupremacyCommand(HubKeySupremacyGetSpoilOfWar, supremacyHub.SupremacyGetSpoilOfWarHandler)
const HubKeySupremacyGetSpoilOfWar = hub.HubCommandKey("SUPREMACY:SUPS_POOL_AMOUNT")

func (sc *SupremacyControllerWS) SupremacyGetSpoilOfWarHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	// get current sup pool user sups
	supsPoolUser, err := db.UserGet(ctx, sc.Conn, passport.SupremacySupPoolUserID)
	if err != nil {
		return terror.Error(err)
	}

	battleUser, err := db.UserGet(ctx, sc.Conn, passport.SupremacyBattleUserID)
	if err != nil {
		return terror.Error(err)
	}

	result := big.NewInt(0)
	result.Add(&supsPoolUser.Sups.Int, &battleUser.Sups.Int)

	reply(result.String())
	return nil
}

const HubKeySupremacyAbilityTargetPriceUpdate = hub.HubCommandKey("SUPREMACY:ABILITY:TARGET:PRICE:UPDATE")

type SupremacyAbilityTargetPriceUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityTokenID    uint64 `json:"abilityTokenID"`
		WarMachineTokenID uint64 `json:"warMachineTokenID"`
		SupsCost          string `json:"supsCost"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyAbilityTargetPriceUpdate(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyAbilityTargetPriceUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// store new sups cost
	err = db.WarMachineAbilityCostUpsert(ctx, sc.Conn, req.Payload.WarMachineTokenID, req.Payload.AbilityTokenID, req.Payload.SupsCost)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

type SupremacyUserSupsMultiplierSendRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserSupsMultiplierSends []*UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
	} `json:"payload"`
}

type UserSupsMultiplierSend struct {
	ToUserID        passport.UserID   `json:"toUserID"`
	ToUserSessionID *hub.SessionID    `json:"toUserSessionID,omitempty"`
	SupsMultipliers []*SupsMultiplier `json:"supsMultiplier"`
}

type SupsMultiplier struct {
	Key       string    `json:"key"`
	Value     int       `json:"value"`
	ExpiredAt time.Time `json:"expiredAt"`
}

const HubKeySupremacyUserSupsMultiplierSend = hub.HubCommandKey("SUPREMACY:USER_SUPS_MULTIPLIER_SEND")

func (sc *SupremacyControllerWS) SupremacyUserSupsMultiplierSendHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyUserSupsMultiplierSendRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	for _, usm := range req.Payload.UserSupsMultiplierSends {
		// broadcast to specific hub client if session id is provided
		if usm.ToUserSessionID != nil && *usm.ToUserSessionID != "" {
			sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers, messagebus.BusSendFilterOption{
				SessionID: *usm.ToUserSessionID,
			})
			continue
		}

		// otherwise, broadcast to the target user
		sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers)
	}

	reply(true)
	return nil
}

type SupremacyFactionStatSendRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionStatSends []*FactionStatSend `json:"factionStatSends"`
	} `json:"payload"`
}

type FactionStatSend struct {
	FactionStat     *passport.FactionStat `json:"factionStat"`
	ToUserID        *passport.UserID      `json:"toUserID,omitempty"`
	ToUserSessionID *hub.SessionID        `json:"toUserSessionID,omitempty"`
}

const HubKeySupremacyFactionStatSend = hub.HubCommandKey("SUPREMACY:FACTION_STAT_SEND")

func (sc *SupremacyControllerWS) SupremacyFactionStatSend(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyFactionStatSendRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	for _, factionStatSend := range req.Payload.FactionStatSends {
		// get recruit number
		recruitNumber, err := db.FactionGetRecruitNumber(ctx, sc.Conn, factionStatSend.FactionStat.ID)
		if err != nil {
			sc.Log.Err(err).Msgf("Failed to get recruit number from faction %s", factionStatSend.FactionStat.ID)
			continue
		}
		factionStatSend.FactionStat.RecruitNumber = recruitNumber

		// get velocity number
		// TODO: figure out what velocity is
		factionStatSend.FactionStat.Velocity = 0

		// get mvp
		if factionStatSend.FactionStat.MvpTokenID > 0 {
			assetMetadata, err := db.XsynMetadataGet(ctx, sc.Conn, factionStatSend.FactionStat.MvpTokenID)
			if err != nil {
				sc.Log.Err(err).Msgf("Failed to get mvp asset %v", factionStatSend.FactionStat.MvpTokenID)
				continue
			}

			// NOTE: currently just return the of asset name,
			//       can pass back more data back as we want in the future
			factionStatSend.FactionStat.MVP = assetMetadata.Name
		}

		if factionStatSend.ToUserID == nil && factionStatSend.ToUserSessionID == nil {
			// broadcast to all faction stat subscribers
			sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionStatUpdatedSubscribe, factionStatSend.FactionStat.ID)), factionStatSend.FactionStat)
			continue
		}

		// broadcast to specific subscribers
		filterOption := messagebus.BusSendFilterOption{}
		if factionStatSend.ToUserID != nil {
			filterOption.Ident = factionStatSend.ToUserID.String()
		}
		if factionStatSend.ToUserSessionID != nil {
			filterOption.SessionID = *factionStatSend.ToUserSessionID
		}

		// broadcast to the target user
		sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionStatUpdatedSubscribe, factionStatSend.FactionStat.ID)), factionStatSend.FactionStat, filterOption)
	}

	return nil
}

type SupremacyDefaultWarMachinesRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"factionID"`
		Amount    int                `json:"amount"`
	} `json:"payload"`
}

const HubKeySupremacyDefaultWarMachines = hub.HubCommandKey("SUPREMACY:GET_DEFAULT_WAR_MACHINES")

func (sc *SupremacyControllerWS) SupremacyDefaultWarMachinesHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyDefaultWarMachinesRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	var warMachines []*passport.WarMachineMetadata
	// check user own this asset, and it has not joined the queue yet
	switch req.Payload.FactionID {
	case passport.RedMountainFactionID:
		faction, err := db.FactionGet(ctx, sc.Conn, passport.RedMountainFactionID)
		if err != nil {
			return terror.Error(err)
		}
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, sc.Conn, passport.SupremacyRedMountainUserID, req.Payload.Amount)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineMetadata := &passport.WarMachineMetadata{}
			// parse metadata
			passport.ParseWarMachineMetadata(wmmd, warMachineMetadata)
			warMachineMetadata.OwnedByID = passport.SupremacyRedMountainUserID
			warMachineMetadata.FactionID = passport.RedMountainFactionID
			warMachineMetadata.Faction = faction

			// parse war machine abilities
			if len(warMachineMetadata.Abilities) > 0 {
				for _, abilityMetadata := range warMachineMetadata.Abilities {
					err := db.AbilityAssetGet(ctx, sc.Conn, abilityMetadata)
					if err != nil {
						return terror.Error(err)
					}

					supsCost, err := db.WarMachineAbilityCostGet(ctx, sc.Conn, warMachineMetadata.TokenID, abilityMetadata.TokenID)
					if err != nil {
						return terror.Error(err)
					}

					abilityMetadata.SupsCost = supsCost
				}
			}

			warMachines = append(warMachines, warMachineMetadata)
		}

	case passport.BostonCyberneticsFactionID:
		faction, err := db.FactionGet(ctx, sc.Conn, passport.BostonCyberneticsFactionID)
		if err != nil {
			return terror.Error(err)
		}
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, sc.Conn, passport.SupremacyBostonCyberneticsUserID, req.Payload.Amount)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineMetadata := &passport.WarMachineMetadata{}
			// parse metadata
			passport.ParseWarMachineMetadata(wmmd, warMachineMetadata)
			warMachineMetadata.OwnedByID = passport.SupremacyBostonCyberneticsUserID
			warMachineMetadata.FactionID = passport.BostonCyberneticsFactionID
			warMachineMetadata.Faction = faction

			// parse war machine abilities
			if len(warMachineMetadata.Abilities) > 0 {
				for _, abilityMetadata := range warMachineMetadata.Abilities {
					err := db.AbilityAssetGet(ctx, sc.Conn, abilityMetadata)
					if err != nil {
						return terror.Error(err)
					}

					supsCost, err := db.WarMachineAbilityCostGet(ctx, sc.Conn, warMachineMetadata.TokenID, abilityMetadata.TokenID)
					if err != nil {
						return terror.Error(err)
					}

					abilityMetadata.SupsCost = supsCost
				}
			}

			warMachines = append(warMachines, warMachineMetadata)
		}
	case passport.ZaibatsuFactionID:
		faction, err := db.FactionGet(ctx, sc.Conn, passport.ZaibatsuFactionID)
		if err != nil {
			return terror.Error(err)
		}
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, sc.Conn, passport.SupremacyZaibatsuUserID, req.Payload.Amount)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineMetadata := &passport.WarMachineMetadata{}
			// parse metadata
			passport.ParseWarMachineMetadata(wmmd, warMachineMetadata)
			warMachineMetadata.OwnedByID = passport.SupremacyZaibatsuUserID
			warMachineMetadata.FactionID = passport.ZaibatsuFactionID
			warMachineMetadata.Faction = faction

			// parse war machine abilities
			if len(warMachineMetadata.Abilities) > 0 {
				for _, abilityMetadata := range warMachineMetadata.Abilities {
					err := db.AbilityAssetGet(ctx, sc.Conn, abilityMetadata)
					if err != nil {
						return terror.Error(err)
					}

					supsCost, err := db.WarMachineAbilityCostGet(ctx, sc.Conn, warMachineMetadata.TokenID, abilityMetadata.TokenID)
					if err != nil {
						return terror.Error(err)
					}

					abilityMetadata.SupsCost = supsCost
				}
			}
			warMachines = append(warMachines, warMachineMetadata)
		}
	}

	reply(warMachines)
	return nil
}

const HubKeySupremacyWarMachineQueueContractUpdate = hub.HubCommandKey("SUPREMACY:WAR_MACHINE_QUEUE_CONTRACT_UPDATE")

type SupremacyWarMachineQueueContractUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionWarMachineQueues []*FactionWarMachineQueue `json:"factionWarMachineQueues"`
	} `json:"payload"`
}

type FactionWarMachineQueue struct {
	FactionID  passport.FactionID `json:"factionID"`
	QueueTotal int                `json:"queueTotal"`
}

func (sc *SupremacyControllerWS) SupremacyWarMachineQueueContractUpdateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyWarMachineQueueContractUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	for _, fwq := range req.Payload.FactionWarMachineQueues {
		go sc.API.recalculateContractReward(fwq.FactionID, fwq.QueueTotal)
	}

	return nil
}

const HubKeySupremacyPayAssetInsurance = hub.HubCommandKey("SUPREMACY:PAY_ASSET_INSURANCE")

type SupremacyPayAssetInsuranceRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID               passport.UserID               `json:"userID"`
		FactionID            passport.FactionID            `json:"factionID"`
		Amount               passport.BigInt               `json:"amount"`
		TransactionReference passport.TransactionReference `json:"transactionReference"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyPayAssetInsuranceHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyPayAssetInsuranceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Amount.Cmp(big.NewInt(0)) < 0 {
		return terror.Error(terror.ErrInvalidInput, "Sups amount can not be negative")
	}

	tx := &passport.NewTransaction{
		From:                 req.Payload.UserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount.Int,
	}

	switch req.Payload.FactionID {
	case passport.RedMountainFactionID:
		tx.To = passport.SupremacyRedMountainUserID
	case passport.BostonCyberneticsFactionID:
		tx.To = passport.SupremacyBostonCyberneticsUserID
	case passport.ZaibatsuFactionID:
		tx.To = passport.SupremacyZaibatsuUserID
	default:
		return terror.Error(terror.ErrInvalidInput, "Provided faction does not exist")
	}

	errChan := make(chan error, 10)
	sc.API.HoldTransaction(errChan, tx)
	err = <-errChan
	if err != nil {
		return terror.Error(err)
	}

	resultChan := make(chan []*passport.Transaction, 1)
	sc.API.CommitTransactions(resultChan, tx.TransactionReference)
	results := <-resultChan
	for _, result := range results {
		if result.Status != passport.TransactionSuccess {
			return terror.Error(fmt.Errorf("Transaction Failed"))
		}
	}

	reply(true)
	return nil
}

const HubKeySupremacyRedeemFactionContractReward = hub.HubCommandKey("SUPREMACY:REDEEM_FACTION_CONTRACT_REWARD")

type SupremacyRedeemFactionContractRewardRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID               passport.UserID               `json:"userID"`
		FactionID            passport.FactionID            `json:"factionID"`
		Amount               passport.BigInt               `json:"amount"`
		TransactionReference passport.TransactionReference `json:"transactionReference"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyRedeemFactionContractRewardHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyPayAssetInsuranceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Amount.Cmp(big.NewInt(0)) < 0 {
		return terror.Error(terror.ErrInvalidInput, "Sups amount can not be negative")
	}

	tx := &passport.NewTransaction{
		To:                   req.Payload.UserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount.Int,
	}

	switch req.Payload.FactionID {
	case passport.RedMountainFactionID:
		tx.From = passport.SupremacyRedMountainUserID
	case passport.BostonCyberneticsFactionID:
		tx.From = passport.SupremacyBostonCyberneticsUserID
	case passport.ZaibatsuFactionID:
		tx.From = passport.SupremacyZaibatsuUserID
	default:
		return terror.Error(terror.ErrInvalidInput, "Provided faction does not exist")
	}

	sc.API.transaction <- tx

	//errChan := make(chan error, 10)
	//sc.API.HoldTransaction(errChan, tx)
	//err = <-errChan
	//if err != nil {
	//	return terror.Error(err)
	//}
	//
	//resultChan := make(chan []*passport.Transaction, 1)
	//sc.API.CommitTransactions(resultChan, tx.TransactionReference)
	//results := <-resultChan
	//for _, result := range results {
	//	if result == nil || result.Status == passport.TransactionFailed {
	//		return terror.Error(fmt.Errorf("Transaction Failed"))
	//	}
	//}

	reply(true)
	return nil
}
