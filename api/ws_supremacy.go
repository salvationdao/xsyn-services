package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"passport"
	"passport/db"
	"sync"
	"time"

	"github.com/ninja-software/log_helpers"

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
	Conn            *pgxpool.Pool
	Log             *zerolog.Logger
	API             *API
	TickerPoolCache struct {
		outerMx            sync.Mutex
		nextAccessMx       sync.Mutex
		dataMx             sync.Mutex
		TricklingAmountMap map[string]*big.Int
	}

	txs *transactions
}

type transactions struct {
	Txes []*passport.NewTransaction
	txMx sync.Mutex
}

// NewSupremacyController creates the supremacy hub
func NewSupremacyController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *SupremacyControllerWS {
	supremacyHub := &SupremacyControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "supremacy"),
		API:  api,
		TickerPoolCache: struct {
			outerMx            sync.Mutex
			nextAccessMx       sync.Mutex
			dataMx             sync.Mutex
			TricklingAmountMap map[string]*big.Int
		}{
			outerMx:            sync.Mutex{},
			nextAccessMx:       sync.Mutex{},
			dataMx:             sync.Mutex{},
			TricklingAmountMap: make(map[string]*big.Int),
		},
		txs: &transactions{
			Txes: []*passport.NewTransaction{},
		},
	}

	// sup control
	api.SupremacyCommand(HubKeySupremacySpendSups, supremacyHub.SupremacySpendSupsHandler)
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
	api.SupremacyCommand(HubKeySupremacyTopSupsContruteUser, supremacyHub.SupremacyTopSupsContributeUser)
	api.SupremacyCommand(HubKeySupremacyUsersGet, supremacyHub.SupremacyUsersGet)

	// faction stat
	api.SupremacyCommand(HubKeySupremacyFactionStatSend, supremacyHub.SupremacyFactionStatSend)
	// user stat
	api.SupremacyCommand(HubKeySupremacyUserStatSend, supremacyHub.SupremacyUserStatSend)

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

const HubKeySupremacySpendSups = hub.HubCommandKey("SUPREMACY:HOLD_SUPS")

type SupremacyHoldSupsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount               passport.BigInt               `json:"amount"`
		FromUserID           passport.UserID               `json:"userID"`
		TransactionReference passport.TransactionReference `json:"transactionReference"`
		GroupID              string                        `json:"groupID"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacySpendSupsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

	if req.Payload.GroupID != "" {
		tx.To = passport.SupremacyBattleUserID
		tx.GroupID = &req.Payload.GroupID
	}

	nfb, ntb, txID, err := sc.API.userCacheMap.Process(tx)
	if err != nil {
		return terror.Error(err, "failed to process sups")
	}

	if !tx.From.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.From)), nfb.String())
	}

	if !tx.To.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.To)), ntb.String())
	}

	tx.ID = txID

	// for refund
	sc.txs.txMx.Lock()
	sc.txs.Txes = append(sc.txs.Txes, &passport.NewTransaction{
		ID:                   txID,
		From:                 tx.To,
		To:                   tx.From,
		Amount:               tx.Amount,
		TransactionReference: passport.TransactionReference(fmt.Sprintf("refund|sups vote|%s", txID)),
	})
	sc.txs.txMx.Unlock()

	reply(txID)
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

func (sc *SupremacyControllerWS) SupremacyFeed() {
	fund := big.NewInt(0)
	fund, ok := fund.SetString("4000000000000000000", 10)
	if !ok {
		sc.Log.Err(errors.New("setting string not ok on fund big int")).Msg("too many strings")
		return
	}

	tx := &passport.NewTransaction{
		From:                 passport.XsynTreasuryUserID,
		To:                   passport.SupremacySupPoolUserID,
		Amount:               *fund,
		TransactionReference: passport.TransactionReference(fmt.Sprintf("treasury|ticker|%s", time.Now())),
	}

	// process user cache map
	fromBalance, toBalance, _, err := sc.API.userCacheMap.Process(tx)
	if err != nil {
		sc.Log.Err(err).Msg(err.Error())
		return
	}

	ctx := context.Background()

	if !tx.From.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.From)), fromBalance.String())
	}

	if !tx.To.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.To)), toBalance.String())
	}
}

func (sc *SupremacyControllerWS) SupremacyTickerTickHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	// make treasury send game server user moneys
	sc.SupremacyFeed()

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

	// var transactions []*passport.NewTransaction

	// we take the whole balance of supremacy sup pool and give it to the users watching
	// amounts depend on their multiplier
	// the supremacy sup pool user gets sups trickled into it from the last battle and 4 every 5 seconds
	supsForTick, err := sc.API.userCacheMap.Get(passport.SupremacySupPoolUserID.String())
	if err != nil {
		return terror.Error(err)
	}

	supPool := &supsForTick
	onePointWorth := big.NewInt(0)
	onePointWorth = onePointWorth.Div(supPool, big.NewInt(int64(totalPoints)))
	// loop again to create all transactions
	for multiplier, users := range req.Payload.UserMap {
		for _, user := range users {
			usersSups := big.NewInt(0)
			usersSups = usersSups.Mul(onePointWorth, big.NewInt(int64(multiplier)))

			tx := &passport.NewTransaction{
				From:                 passport.SupremacySupPoolUserID,
				To:                   *user,
				Amount:               *usersSups,
				TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|ticker|%s|%s", *user, time.Now())),
			}

			nfb, ntb, _, err := sc.API.userCacheMap.Process(tx)
			if err != nil {
				return terror.Error(err, "failed to process user fund")
			}

			if !tx.From.IsSystemUser() {
				go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.From)), nfb.String())
			}

			if !tx.To.IsSystemUser() {
				go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.To)), ntb.String())
			}

			supPool = supPool.Sub(supPool, usersSups)
		}
	}

	reply(true)
	return nil
}

const HubKeySupremacyTransferBattleFundToSupPool = hub.HubCommandKey("SUPREMACY:TRANSFER_BATTLE_FUND_TO_SUP_POOL")

func (sc *SupremacyControllerWS) SupremacyTransferBattleFundToSupPoolHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	// recalculate faction mvp user
	err := db.FactionMvpMaterialisedViewRefresh(ctx, sc.Conn)
	if err != nil {
		return terror.Error(err, "Failed to refresh faction mvp list")
	}

	// generate new go routine to trickle sups
	sc.poolHighPriorityLock()

	// get current battle user sups
	battleUser, err := sc.API.userCacheMap.Get(passport.SupremacyBattleUserID.String())
	if err != nil {
		sc.poolHighPriorityUnlock()
		return terror.Error(err, "failed to get battle user balance from db")
	}

	// calc trickling sups for current round
	supsForTrickle := big.NewInt(0)
	supsForTrickle.Add(supsForTrickle, &battleUser)

	// subtrack the sups that is trickling at the moment
	for _, tricklingSups := range sc.TickerPoolCache.TricklingAmountMap {
		supsForTrickle.Sub(supsForTrickle, tricklingSups)
	}

	// so here we want to trickle the battle pool out over 5 minutes, so we create a ticker that ticks every 5 seconds with a max ticks of 300 / 5
	ticksInFiveMinutes := 300 / 5
	supsPerTick := big.NewInt(0)
	supsPerTick.Div(supsForTrickle, big.NewInt(int64(ticksInFiveMinutes)))

	// skip, if trickle amount is empty
	if supsPerTick.BitLen() == 0 {
		sc.poolHighPriorityUnlock()
		reply(true)
		return nil
	}

	// append the amount set to the list
	key := uuid.Must(uuid.NewV4()).String()
	sc.TickerPoolCache.TricklingAmountMap[key] = big.NewInt(0)
	sc.TickerPoolCache.TricklingAmountMap[key].Add(sc.TickerPoolCache.TricklingAmountMap[key], supsForTrickle)

	// start a new go routine for current round
	go sc.trickleFactory(key, ticksInFiveMinutes, supsPerTick)
	sc.poolHighPriorityUnlock()

	reply(true)
	return nil
}

// priority locks

// poolHighPriorityLock
func (sc *SupremacyControllerWS) poolHighPriorityLock() {
	sc.TickerPoolCache.nextAccessMx.Lock()
	sc.TickerPoolCache.dataMx.Lock()
	sc.TickerPoolCache.nextAccessMx.Unlock()
}

// poolHighPriorityUnlock
func (sc *SupremacyControllerWS) poolHighPriorityUnlock() {
	sc.TickerPoolCache.dataMx.Unlock()
}

// poolLowPriorityLock
func (sc *SupremacyControllerWS) poolLowPriorityLock() {
	sc.TickerPoolCache.outerMx.Lock()
	sc.TickerPoolCache.nextAccessMx.Lock()
	sc.TickerPoolCache.dataMx.Lock()
	sc.TickerPoolCache.nextAccessMx.Unlock()
}

// poolLowPriorityUnlock
func (sc *SupremacyControllerWS) poolLowPriorityUnlock() {
	sc.TickerPoolCache.dataMx.Unlock()
	sc.TickerPoolCache.outerMx.Unlock()
}

// trickle factory
func (sc *SupremacyControllerWS) trickleFactory(key string, totalTick int, supsPerTick *big.Int) {
	i := 0
	for {
		i++
		// resultChan := make(chan *passport.TransactionResult)

		// TODO: manage user cache
		// transaction := &passport.NewTransaction{
		// 	ResultChan:           resultChan,
		// 	From:                 passport.SupremacyBattleUserID,
		// 	To:                   passport.SupremacySupPoolUserID,
		// 	Amount:               *supsPerTick,
		// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|battle_sups_spend_transfer|%s", time.Now())),
		// }

		sc.poolLowPriorityLock()
		tx := &passport.NewTransaction{
			From:                 passport.SupremacyBattleUserID,
			To:                   passport.SupremacySupPoolUserID,
			Amount:               *supsPerTick,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("supremacy|battle_sups_spend_transfer|%s", time.Now())),
		}

		// process user cache map
		nfb, ntb, _, err := sc.API.userCacheMap.Process(tx)
		if err != nil {
			sc.Log.Err(err).Msg("insufficient fund")
			return
		}

		ctx := context.Background()

		if !tx.From.IsSystemUser() {
			go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.From)), nfb.String())
		}

		if !tx.To.IsSystemUser() {
			go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.To)), ntb.String())
		}

		// if the routine is not finished
		if i < totalTick {
			// update current trickling amount
			sc.TickerPoolCache.TricklingAmountMap[key].Sub(sc.TickerPoolCache.TricklingAmountMap[key], supsPerTick)
			sc.poolLowPriorityUnlock()

			time.Sleep(5 * time.Second)
			continue
		}

		// otherwise, delete the trickle amount from the map
		delete(sc.TickerPoolCache.TricklingAmountMap, key)
		sc.poolLowPriorityUnlock()
		break
	}
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

	asset, err := db.AssetGet(ctx, sc.Conn, req.Payload.AssetTokenID)
	if err != nil {
		reply(false)
		return terror.Error(err)
	}
	if asset == nil {
		return terror.Error(fmt.Errorf("asset doesn't exist"), "Failed to get asset.")
	}

	frozenAt := time.Now()

	err = db.XsynAssetFreeze(ctx, sc.Conn, req.Payload.AssetTokenID, userID)
	if err != nil {
		reply(false)
		return terror.Error(err)
	}

	asset.FrozenAt = &frozenAt

	go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, req.Payload.AssetTokenID)), asset)

	sc.API.SendToAllServerClient(ctx, &ServerClientMessage{
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
		"", false, req.Payload.AssetTokenIDs, nil, nil, 0, len(req.Payload.AssetTokenIDs), "", "",
	)
	if err != nil {
		return terror.Error(err)
	}

	for _, asset := range assets {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.ExternalTokenID)), asset)
	}

	// auto release it after an hour
	for _, asset := range assets {
		go func(asset *passport.XsynMetadata) {
			frozenAt := asset.FrozenAt
			time.Sleep(3 * time.Hour)
			// get asset from db
			a, err := db.AssetGet(ctx, sc.Conn, asset.ExternalTokenID)
			if err != nil {
				sc.API.Log.Err(err).Msgf("Failed to get asset, token id: %d", asset.ExternalTokenID)
				return
			}

			// check whether the frozen at has changed
			if a.FrozenAt == nil || !frozenAt.Equal(*a.FrozenAt) {
				// skip, if frozen at is changed
				return
			}

			// otherwise release it
			err = db.XsynAssetBulkRelease(ctx, sc.Conn, []*passport.WarMachineMetadata{
				{
					TokenID: a.ExternalTokenID,
				},
			}, userID)
			if err != nil {
				sc.API.Log.Err(err).Msgf("Failed to auto released, token id: %d", asset.ExternalTokenID)
				return
			}

		}(asset)
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

	//err = db.XsynAsseetDurabilityBulkUpdate(ctx, tx, req.Payload.ReleasedAssets)
	//if err != nil {
	//	return terror.Error(err)
	//}

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
		"", false, tokenIDs, nil, nil, 0, len(tokenIDs), "", "",
	)
	if err != nil {
		return terror.Error(err)
	}

	for _, asset := range assets {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.ExternalTokenID)), asset)
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
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueuePositionSubscribe, uwm.UserID)), uwm.WarMachineQueuePositions)
	}

	return nil
}

const HubKeySupremacyReleaseTransactions = hub.HubCommandKey("SUPREMACY:RELEASE_TRANSACTIONS")

type SupremacyReleaseTransactionsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TxIDs []string `json:"txIDs"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyReleaseTransactionsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyReleaseTransactionsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	sc.txs.txMx.Lock()
	defer sc.txs.txMx.Unlock()
	for _, txID := range req.Payload.TxIDs {
		for _, tx := range sc.txs.Txes {
			if txID != tx.ID {
				continue
			}
			nfb, ntb, _, err := sc.API.userCacheMap.Process(tx)
			if err != nil {
				sc.API.Log.Err(err).Msg("failed to process user sups fund")
				continue
			}

			if !tx.From.IsSystemUser() {
				go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.From)), nfb.String())
			}

			if !tx.To.IsSystemUser() {
				go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.To)), ntb.String())
			}
		}
	}

	sc.txs.Txes = []*passport.NewTransaction{}

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
	supsPoolUser, err := sc.API.userCacheMap.Get(passport.SupremacySupPoolUserID.String())
	if err != nil {
		return terror.Error(err)
	}

	battleUser, err := sc.API.userCacheMap.Get(passport.SupremacyBattleUserID.String())
	if err != nil {
		return terror.Error(err)
	}

	result := big.NewInt(0)
	result.Add(result, &supsPoolUser)
	result.Add(result, &battleUser)

	reply(result.String())
	return nil
}

type SupremacyAbilityTargetPriceUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		AbilityTokenID    uint64 `json:"abilityTokenID"`
		WarMachineTokenID uint64 `json:"warMachineTokenID"`
		SupsCost          string `json:"supsCost"`
	} `json:"payload"`
}

const HubKeySupremacyAbilityTargetPriceUpdate = hub.HubCommandKey("SUPREMACY:ABILITY:TARGET:PRICE:UPDATE")

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

type SupremacyTopSupsContributorRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		StartTime time.Time `json:"startTime"`
		EndTime   time.Time `json:"endTime"`
	} `json:"payload"`
}

type SupremacyTopSupsContributorResponse struct {
	TopSupsContributors       []*passport.User    `json:"topSupsContributors"`
	TopSupsContributeFactions []*passport.Faction `json:"topSupsContributeFactions"`
}

const HubKeySupremacyTopSupsContruteUser = hub.HubCommandKey("SUPREMACY:TOP_SUPS_CONTRIBUTORS")

func (sc *SupremacyControllerWS) SupremacyTopSupsContributeUser(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyTopSupsContributorRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get top contribute users
	topSupsContributors, err := db.BattleArenaSupsTopContributors(ctx, sc.Conn, req.Payload.StartTime, req.Payload.EndTime)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return terror.Error(err)
	}

	// get top contribute faction
	topSupsContributeFactions, err := db.BattleArenaSupsTopContributeFaction(ctx, sc.Conn, req.Payload.StartTime, req.Payload.EndTime)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return terror.Error(err)
	}

	reply(&SupremacyTopSupsContributorResponse{
		TopSupsContributors:       topSupsContributors,
		TopSupsContributeFactions: topSupsContributeFactions,
	})

	return nil
}

type SupremacyUserGetRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserIDs []passport.UserID `json:"userIDs"`
	} `json:"payload"`
}

const HubKeySupremacyUsersGet = hub.HubCommandKey("SUPREMACY:GET_USERS")

func (sc *SupremacyControllerWS) SupremacyUsersGet(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyUserGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	users, err := db.UserGetByIDs(ctx, sc.Conn, req.Payload.UserIDs)
	if err != nil {
		return terror.Error(err)
	}

	reply(users)
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
			go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers, messagebus.BusSendFilterOption{
				SessionID: *usm.ToUserSessionID,
			})
			continue
		}

		// otherwise, broadcast to the target user
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsMultiplierSubscribe, usm.ToUserID)), usm.SupsMultipliers)
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
		mvp, err := db.FactionMvpGet(ctx, sc.Conn, factionStatSend.FactionStat.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			sc.Log.Err(err).Msgf("failed to get mvp from faction %s", factionStatSend.FactionStat.ID)
			continue
		}
		factionStatSend.FactionStat.MVP = mvp

		supsVoted, err := db.FactionSupsVotedGet(ctx, sc.Conn, factionStatSend.FactionStat.ID)
		if err != nil {
			sc.Log.Err(err).Msgf("failed to get sups voted from faction %s", factionStatSend.FactionStat.ID)
			continue
		}

		factionStatSend.FactionStat.SupsVoted = supsVoted.String()

		if factionStatSend.ToUserID == nil && factionStatSend.ToUserSessionID == nil {
			// broadcast to all faction stat subscribers
			go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionStatUpdatedSubscribe, factionStatSend.FactionStat.ID)), factionStatSend.FactionStat)
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
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyFactionStatUpdatedSubscribe, factionStatSend.FactionStat.ID)), factionStatSend.FactionStat, filterOption)
	}

	return nil
}

type SupremacyUserStatSendRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserStatSends []*UserStatSend `json:"userStatSends"`
	} `json:"payload"`
}

type UserStatSend struct {
	ToUserSessionID *hub.SessionID     `json:"toUserSessionID,omitempty"`
	Stat            *passport.UserStat `json:"stat"`
}

const HubKeySupremacyUserStatSend = hub.HubCommandKey("SUPREMACY:USER_STAT_SEND")

func (sc *SupremacyControllerWS) SupremacyUserStatSend(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyUserStatSendRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	for _, userStatSend := range req.Payload.UserStatSends {

		if userStatSend.ToUserSessionID == nil {
			// broadcast to all faction stat subscribers
			go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, userStatSend.Stat.ID)), userStatSend.Stat)
			continue
		}

		// broadcast to specific subscribers
		filterOption := messagebus.BusSendFilterOption{}
		if userStatSend.ToUserSessionID != nil {
			filterOption.SessionID = *userStatSend.ToUserSessionID
		}

		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, userStatSend.Stat.ID)), userStatSend.Stat, filterOption)
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
		go sc.API.recalculateContractReward(ctx, fwq.FactionID, fwq.QueueTotal)
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

	// resultChan := make(chan *passport.TransactionResult)

	tx := &passport.NewTransaction{
		// ResultChan:           resultChan,
		From:                 req.Payload.UserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount.Int,
	}

	// TODO: validate the insurance is 10% of current reward price

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

	nfb, ntb, _, err := sc.API.userCacheMap.Process(tx)
	if err != nil {
		return terror.Error(err, "failed to process user fund")
	}

	if !tx.From.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.From)), nfb.String())
	}

	if !tx.To.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.To)), ntb.String())
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

	if req.Payload.Amount.Cmp(big.NewInt(0)) <= 0 {
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

	// process user cache map
	nfb, ntb, _, err := sc.API.userCacheMap.Process(tx)
	if err != nil {
		return terror.Error(err, "failed to process fund")
	}

	if !tx.From.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.From)), nfb.String())
	}

	if !tx.To.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, tx.To)), ntb.String())
	}

	reply(true)
	return nil
}
