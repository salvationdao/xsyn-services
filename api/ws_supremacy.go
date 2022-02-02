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
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-software/tickle"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

// SupremacyControllerWS holds handlers for supremacy and the supremacy held transactions
type SupremacyControllerWS struct {
	Conn               *pgxpool.Pool
	Log                *zerolog.Logger
	API                *API
	SupremacyUserID    passport.UserID
	XsynTreasuryUserID passport.UserID
	BattleUserID       passport.UserID
}

// NewSupremacyController creates the supremacy hub
func NewSupremacyController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *SupremacyControllerWS {
	supremacyHub := &SupremacyControllerWS{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "supremacy"),
		API:  api,
	}

	supremacyHub.SupremacyUserID = passport.SupremacyGameUserID
	supremacyHub.XsynTreasuryUserID = passport.XsynTreasuryUserID
	supremacyHub.BattleUserID = passport.SupremacyBattleUserID

	// start nft repair ticker
	tickle.New("NFT Repair Ticker", 60, func() (int, error) {
		err := db.XsynNftMetadataDurabilityBulkIncrement(context.Background(), conn)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err)
		}
		return http.StatusOK, nil
	}).Start()

	// sup control
	api.SupremacyCommand(HubKeySupremacyHoldSups, supremacyHub.SupremacyHoldSupsHandler)
	api.SupremacyCommand(HubKeySupremacyCommitTransactions, supremacyHub.SupremacyCommitTransactionsHandler)
	api.SupremacyCommand(HubKeySupremacyReleaseTransactions, supremacyHub.SupremacyReleaseTransactionsHandler)
	api.SupremacyCommand(HubKeySupremacyTickerTick, supremacyHub.SupremacyTickerTickHandler)
	api.SupremacyCommand(HubKeySupremacyDistributeBattleReward, supremacyHub.SupremacyDistributeBattleRewardHandler)

	// user connection upgrade
	api.SupremacyCommand(HubKeySupremacyUserConnectionUpgrade, supremacyHub.SupremacyUserConnectionUpgradeHandler)

	// battle queue
	api.SupremacyCommand(HubKeySupremacyWarMachineQueuePositionClear, supremacyHub.SupremacyWarMachineQueuePositionClearHandler)

	// asset control
	api.SupremacyCommand(HubKeySupremacyAssetFreeze, supremacyHub.SupremacyAssetFreezeHandler)
	api.SupremacyCommand(HubKeySupremacyAssetLock, supremacyHub.SupremacyAssetLockHandler)
	api.SupremacyCommand(HubKeySupremacyAssetRelease, supremacyHub.SupremacyAssetReleaseHandler)
	api.SupremacyCommand(HubKeySupremacyWarMachineQueuePosition, supremacyHub.SupremacyWarMachineQueuePositionHandler)

	// other?
	api.SupremacyCommand(HubKeySupremacyDefaultWarMachines, supremacyHub.SupremacyDefaultWarMachinesHandler)

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
		Amount               passport.BigInt      `json:"amount"`
		FromUserID           passport.UserID      `json:"userID"`
		TransactionReference TransactionReference `json:"transactionReference"`
		IsBattleVote         bool                 `json:"isBattleVote"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyHoldSupsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyHoldSupsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	tx := &NewTransaction{
		From:                 req.Payload.FromUserID,
		To:                   sc.SupremacyUserID,
		TransactionReference: req.Payload.TransactionReference,
		Amount:               req.Payload.Amount.Int,
	}

	if req.Payload.IsBattleVote {
		tx.To = sc.BattleUserID
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

	reference := fmt.Sprintf("supremacy|ticker|%s", time.Now())

	var transactions []*NewTransaction

	// 50 sups per 60 second
	// supremacy ticker tick every 3 second, so grab 2.5 sups on every tick
	supPool := big.NewInt(0)
	supPool, ok := supPool.SetString("2500000000000000000", 10)
	if !ok {
		return terror.Error(fmt.Errorf("failed to convert 2500000000000000000 to big int"))
	}

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
				From:                 sc.SupremacyUserID,
				To:                   *user,
				Amount:               *usersSups,
				TransactionReference: TransactionReference(reference),
			})

			supPool = supPool.Sub(supPool, usersSups)
		}
	}

	// send through transactions
	for _, tx := range transactions {
		tx.ResultChan = make(chan *passport.Transaction, 1)
		sc.API.transaction <- tx
		result := <-tx.ResultChan
		// if result is success, update the cache map
		if result.Status == passport.TransactionSuccess {
			errChan := make(chan error, 10)
			sc.API.UpdateUserCacheRemoveSups(tx.From, tx.Amount, errChan)
			err := <-errChan
			if err != nil {
				sc.API.Log.Err(err).Msg(err.Error())
				continue
			}
			sc.API.UpdateUserCacheAddSups(tx.To, tx.Amount)
		}
	}

	reply(true)
	return nil
}

const HubKeySupremacyDistributeBattleReward = hub.HubCommandKey("SUPREMACY:DISTRIBUTE_BATTLE_REWARD")

type SupremacyDistributeBattleRewardRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		WinnerFactionID               passport.FactionID `json:"winnerFactionID"`
		WinningFactionViewerIDs       []passport.UserID  `json:"winningFactionViewerIDs"`
		WinningWarMachineOwnerIDs     []passport.UserID  `json:"winningWarMachineOwnerIDs"`
		ExecuteKillWarMachineOwnerIDs []passport.UserID  `json:"executeKillWarMachineOwnerIDs"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyDistributeBattleRewardHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyDistributeBattleRewardRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get sups from battle user
	battleUser, err := db.UserGet(ctx, sc.Conn, passport.SupremacyBattleUserID, "")
	if err != nil {
		return terror.Error(err, "Failed to get battle arena user")
	}

	// skip calculation, if the is not sups in the pool
	if battleUser.Sups.Int.Cmp(big.NewInt(0)) <= 0 {
		reply(true)
		return nil
	}

	// portion sups

	// 25% sups for winner war machine owners and execute kill war machine owners
	supsPortionPercentage25 := big.NewInt(0)
	supsPortionPercentage25.Div(&battleUser.Sups.Int, big.NewInt(4))

	// 50% sups for winner faction viewers
	supsPortionPercentage50 := big.NewInt(0)
	supsPortionPercentage50.Sub(&battleUser.Sups.Int, supsPortionPercentage25.Mul(supsPortionPercentage25, big.NewInt(2)))

	// start distributing sups
	var transactions []*NewTransaction

	// get winning faction user
	winningFactionUserID, err := db.FactionUserIDGetByFactionID(ctx, sc.Conn, req.Payload.WinnerFactionID)
	if err != nil {
		return terror.Error(err)
	}

	/*************************
	* Winner Faction Viewers *
	*************************/
	viewerIDs := req.Payload.WinningFactionViewerIDs
	// set faction user as viewer if there is no viewer online
	if len(viewerIDs) == 0 {
		viewerIDs = []passport.UserID{winningFactionUserID}
	}

	reference := fmt.Sprintf("supremacy|battle_reward|winning_faction_viewer|%s", time.Now())

	supsPerUser := big.NewInt(0)
	supsPerUser.Div(supsPortionPercentage50, big.NewInt(int64(len(viewerIDs))))

	for _, viewerID := range viewerIDs {
		amount := supsPerUser

		transactions = append(transactions, &NewTransaction{
			From:                 sc.BattleUserID,
			To:                   viewerID,
			Amount:               *amount,
			TransactionReference: TransactionReference(reference),
		})
	}

	/****************************
	* Winning War Machine Owner *
	****************************/

	ownerIDs := req.Payload.WinningWarMachineOwnerIDs
	if len(ownerIDs) == 0 {
		ownerIDs = []passport.UserID{winningFactionUserID}
	}

	reference = fmt.Sprintf("supremacy|battle_reward|winning_war_machine_owner|%s", time.Now())

	supsPerUser = big.NewInt(0)
	supsPerUser.Div(supsPortionPercentage25, big.NewInt(int64(len(ownerIDs))))

	for _, ownerID := range ownerIDs {
		amount := supsPerUser

		transactions = append(transactions, &NewTransaction{
			From:                 sc.BattleUserID,
			To:                   ownerID,
			Amount:               *amount,
			TransactionReference: TransactionReference(reference),
		})
	}

	/*********************************
	* Execute Kill War Machine Owner *
	*********************************/

	killOwnerIDs := req.Payload.ExecuteKillWarMachineOwnerIDs
	if len(killOwnerIDs) == 0 {
		killOwnerIDs = []passport.UserID{winningFactionUserID}
	}

	reference = fmt.Sprintf("supremacy|battle_reward|execute_kill_war_machine_owner|%s", time.Now())

	supsPerUser = big.NewInt(0)
	supsPerUser.Div(supsPortionPercentage25, big.NewInt(int64(len(killOwnerIDs))))

	for _, killOwnerID := range killOwnerIDs {
		amount := supsPerUser

		transactions = append(transactions, &NewTransaction{
			From:                 sc.BattleUserID,
			To:                   killOwnerID,
			Amount:               *amount,
			TransactionReference: TransactionReference(reference),
		})
	}

	// send through transactions
	for _, tx := range transactions {
		tx.ResultChan = make(chan *passport.Transaction, 1)
		sc.API.transaction <- tx
		result := <-tx.ResultChan
		// if result is success, update the cache map
		if result.Status == passport.TransactionSuccess {
			errChan := make(chan error, 10)
			sc.API.UpdateUserCacheRemoveSups(tx.From, tx.Amount, errChan)
			err := <-errChan
			if err != nil {
				sc.API.Log.Err(err).Msg(err.Error())
				continue
			}
			sc.API.UpdateUserCacheAddSups(tx.To, tx.Amount)
		}
	}

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

	asset, err := db.AssetGet(ctx, sc.Conn, int(req.Payload.AssetTokenID))
	if err != nil {
		reply(false)
		return terror.Error(err)
	}

	// TODO: In the future, charge user's sups for joining the queue

	sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, req.Payload.AssetTokenID)), asset)

	sc.API.SendToAllServerClient(&ServerClientMessage{
		Key: AssetUpdated,
		Payload: struct {
			Asset *passport.XsynNftMetadata `json:"asset"`
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

type SupremacyWarMachineQueuePositionClearRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"factionID"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeySupremacyWarMachineQueuePositionClear, AssetController.RegisterHandler)
const HubKeySupremacyWarMachineQueuePositionClear hub.HubCommandKey = "SUPREMACY:WAR:MACHINE:QUEUE:POSITION:CLEAR"

// SupremacyWarMachineQueuePositionClearHandler broadcast user to clear the war machine queue
func (sc *SupremacyControllerWS) SupremacyWarMachineQueuePositionClearHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyWarMachineQueuePositionClearRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// get faction users
	userIDs, err := db.UserIDsGetByFactionID(ctx, sc.Conn, req.Payload.FactionID)
	if err != nil {
		return terror.Error(err, "Failed to get user id from faction")
	}

	if len(userIDs) == 0 {
		return nil
	}

	// broadcast war machine position to all user client
	for _, userID := range userIDs {
		go sc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueuePositionSubscribe, userID)), []*WarMachineQueuePosition{})
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
		TransactionReferences []TransactionReference `json:"transactions"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyReleaseTransactionsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyReleaseTransactionsRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	resultChan := make(chan []*passport.Transaction)

	sc.API.ReleaseHeldTransaction(req.Payload.TransactionReferences...)

	results := <-resultChan

	reply(results)
	return nil
}

const HubKeySupremacyDefaultWarMachines = hub.HubCommandKey("SUPREMACY:GET_DEFAULT_WAR_MACHINES")

type SupremacyDefaultWarMachinesRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FactionID passport.FactionID `json:"factionID"`
		Amount    int                `json:"amount"`
	} `json:"payload"`
}

func (sc *SupremacyControllerWS) SupremacyDefaultWarMachinesHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupremacyDefaultWarMachinesRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	var warMachines []*passport.WarMachineNFT
	// check user own this asset, and it has not joined the queue yet
	switch req.Payload.FactionID {
	case passport.RedMountainFactionID:
		faction, err := db.FactionGet(ctx, sc.Conn, passport.RedMountainFactionID)

		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, sc.Conn, passport.SupremacyRedMountainUserID, req.Payload.Amount)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineNFT := &passport.WarMachineNFT{}
			// parse nft
			passport.ParseWarMachineNFT(wmmd, warMachineNFT)
			warMachineNFT.OwnedByID = passport.SupremacyRedMountainUserID
			warMachineNFT.FactionID = passport.RedMountainFactionID
			warMachineNFT.Faction = faction

			warMachines = append(warMachines, warMachineNFT)
		}

	case passport.BostonCyberneticsFactionID:
		faction, err := db.FactionGet(ctx, sc.Conn, passport.BostonCyberneticsFactionID)
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, sc.Conn, passport.SupremacyBostonCyberneticsUserID, req.Payload.Amount)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineNFT := &passport.WarMachineNFT{}
			// parse nft
			passport.ParseWarMachineNFT(wmmd, warMachineNFT)
			warMachineNFT.OwnedByID = passport.SupremacyBostonCyberneticsUserID
			warMachineNFT.FactionID = passport.BostonCyberneticsFactionID
			warMachineNFT.Faction = faction
			warMachines = append(warMachines, warMachineNFT)
		}
	case passport.ZaibatsuFactionID:
		faction, err := db.FactionGet(ctx, sc.Conn, passport.ZaibatsuFactionID)
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, sc.Conn, passport.SupremacyZaibatsuUserID, req.Payload.Amount)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineNFT := &passport.WarMachineNFT{}
			// parse nft
			passport.ParseWarMachineNFT(wmmd, warMachineNFT)
			warMachineNFT.OwnedByID = passport.SupremacyZaibatsuUserID
			warMachineNFT.FactionID = passport.ZaibatsuFactionID
			warMachineNFT.Faction = faction
			warMachines = append(warMachines, warMachineNFT)
		}
	}

	reply(warMachines)
	return nil
}
