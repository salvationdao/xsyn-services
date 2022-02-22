package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"passport/db"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jpillora/backoff"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"

	"github.com/ethereum/go-ethereum/common"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

const ETHSymbol = "ETH"
const BNBSymbol = "BNB"
const BUSDSymbol = "BUSD"
const USDCSymbol = "USDC"

type ChainClients struct {
	SUPS            *bridge.SUPS
	EthClient       *ethclient.Client
	BscClient       *ethclient.Client
	Params          *passport.BridgeParams
	API             *API
	Log             *zerolog.Logger
	updateStateFunc func(chainID int64, newBlock uint64)

	updatePriceFuncMu sync.Mutex
	updatePriceFunc   func(symbol string, amount decimal.Decimal)
}

func RunChainListeners(log *zerolog.Logger, api *API, p *passport.BridgeParams) *ChainClients {
	log.Debug().Str("purchase_addr", p.PurchaseAddr.Hex()).Str("deposit_addr", p.DepositAddr.Hex()).Str("busd_addr", p.BusdAddr.Hex()).Str("usdc_addr", p.UsdcAddr.Hex()).Msg("addresses")
	ctx := context.Background()
	cc := &ChainClients{
		Params:            p,
		API:               api,
		Log:               log,
		updatePriceFuncMu: sync.Mutex{},
	}

	// func to update state
	cc.updateStateFunc = func(chainID int64, newBlock uint64) {
		cc.Log.Debug().Int64("ChainID", chainID).Uint64("Block", newBlock).Msg("updating state")

		if chainID == p.ETHChainID {
			_, err := db.UpdateLatestETHBlock(ctx, cc.API.Conn, newBlock)
			if err != nil {
				api.Log.Err(err).Msgf("failed to update latest eth block to %d", newBlock)
			}
		}

		if chainID == p.BSCChainID {
			_, err := db.UpdateLatestBSCBlock(ctx, cc.API.Conn, newBlock)
			if err != nil {
				api.Log.Err(err).Msgf("failed to update latest bsc block to %d", newBlock)
			}
		}
	}

	cc.updatePriceFunc = func(symbol string, amount decimal.Decimal) {
		switch symbol {
		case "SUPS":
			cc.API.State.SUPtoUSD = amount
		case "ETH":
			cc.API.State.ETHtoUSD = amount
		case "BNB":
			cc.API.State.BNBtoUSD = amount
		}

		_, err := db.UpdateExchangeRates(ctx, cc.API.Conn, cc.API.State)
		if err != nil {
			api.Log.Err(err).Msg("failed to update exchange rates")
		}
		cc.Log.Debug().
			Str(symbol, amount.String()).
			Msg("update rate")

		go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeySUPSExchangeRates), cc.API.State)
	}

	go cc.runETHBridgeListener(ctx)
	go cc.runBSCBridgeListener(ctx)

	return cc
}

func (cc *ChainClients) handleTransfer(ctx context.Context) func(xfer *bridge.Transfer) {
	fn := func(xfer *bridge.Transfer) {
		if xfer.From.Hex() == cc.Params.OperatorAddr.Hex() || xfer.To.Hex() == cc.Params.OperatorAddr.Hex() {
			amt := decimal.NewFromBigInt(xfer.Amount, -18)
			cc.Log.Debug().
				Str("txid", xfer.TxID.Hex()).
				Str("from", xfer.From.Hex()).
				Str("to", xfer.To.Hex()).
				Str("sym", xfer.Symbol).
				Str("amt", amt.Round(2).String()).
				Msg("operator tx detected. Skipping...")
			return
		}
		chainID := xfer.ChainID
		switch chainID {
		case cc.Params.BSCChainID:
			switch xfer.Symbol {
			case "BUSD":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with BUSD

					amountTimes100 := xfer.Amount.Mul(xfer.Amount, big.NewInt(1000))
					supUSDPriceTimes100 := cc.API.State.SUPtoUSD.Mul(decimal.New(1000, 0)).BigInt()
					supAmount := amountTimes100.Div(amountTimes100, supUSDPriceTimes100)

					cc.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", decimal.NewFromBigInt(supAmount, 0).Div(decimal.New(1, int32(18))).String()).
						Str("BUSD", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("Buyer", xfer.From.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex()), RoleID: passport.UserRoleMemberID}
							err = db.UserCreate(ctx, cc.API.Conn, user)
							if err != nil {
								cc.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							cc.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}

					tx := passport.NewTransaction{
						// ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
					}

					// process user cache map
					err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
					if err != nil {
						cc.Log.Err(err).Msg("insufficient fund")
						return
					}
					// resultChan := make(chan *passport.TransactionResult)

					// process transaction
					transactonID := cc.API.transactionCache.Process(tx)

					// TODO: manage user cache

					// select {
					// case cc.API.transaction <- &passport.NewTransaction{
					// 	ResultChan:           resultChan,
					// 	To:                   user.ID,
					// 	From:                 passport.XsynSaleUserID,
					// 	Amount:               *supAmount,
					// 	TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
					// 	Description:          fmt.Sprintf("sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
					// }:

					// case <-time.After(10 * time.Second):
					// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
					// 	panic("sup purchase on BSC with BUSD ")
					// }

					// result := <-resultChan

					// if result.Error != nil {
					// 	return // believe error logs already
					// }

					// if result.Transaction.Status != passport.TransactionSuccess {
					// 	cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
					// 	return
					// }

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), transactonID, xfer.Block, xfer.ChainID)
					if err != nil {
						tx := passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
						}

						err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
						if err != nil {
							cc.Log.Err(err).Msg("insufficient fund")
							return
						}

						cc.API.transactionCache.Process(tx)

						// TODO: manage user cache

						// select {
						// case cc.API.transaction <- &passport.NewTransaction{
						// 	To:                   passport.XsynSaleUserID,
						// 	From:                 user.ID,
						// 	Amount:               *supAmount,
						// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
						// 	Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
						// }:

						// case <-time.After(10 * time.Second):
						// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
						// 	panic("failed insert chan")
						// }

						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
			case "BNB":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with BNB

					// TODO: probably do a * 1000 here? currently no decimals in conversion but possibly in future?
					supAmount := cc.API.State.BNBtoUSD.Div(cc.API.State.SUPtoUSD).BigInt()
					supAmount = supAmount.Mul(supAmount, xfer.Amount)

					cc.Log.Info().
						Str("BNBtoUSD", cc.API.State.BNBtoUSD.String()).
						Str("SUPtoUSD", cc.API.State.SUPtoUSD.String()).
						Str("Chain", "BSC").
						Str("SUPS", supAmount.String()).
						Str("BNB", xfer.Amount.String()).
						Str("Buyer", xfer.From.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex()), RoleID: passport.UserRoleMemberID}
							err = db.UserCreate(ctx, cc.API.Conn, user)
							if err != nil {
								cc.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							cc.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}

					// resultChan := make(chan *passport.TransactionResult)
					tx := passport.NewTransaction{
						// ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on BSC with BNB %s", xfer.TxID.Hex()),
					}

					// process user cache map
					err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
					if err != nil {
						cc.Log.Err(err).Msg("insufficient fund")
						return
					}

					txID := cc.API.transactionCache.Process(tx)

					// TODO: manage user cache

					// select {
					// case cc.API.transaction <- &passport.NewTransaction{
					// 	ResultChan:           resultChan,
					// 	To:                   user.ID,
					// 	From:                 passport.XsynSaleUserID,
					// 	Amount:               *supAmount,
					// 	TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
					// 	Description:          fmt.Sprintf("sup purchase on BSC with BNB %s", xfer.TxID.Hex()),
					// }:

					// case <-time.After(10 * time.Second):
					// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
					// 	panic("sup purchase on BSC with BNB")
					// }

					// result := <-resultChan

					// if result.Error != nil {
					// 	return // believe error logs already
					// }

					// if result.Transaction.Status != passport.TransactionSuccess {
					// 	cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
					// 	return
					// }

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), txID, xfer.Block, xfer.ChainID)
					if err != nil {
						tx := passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with BNB %s", xfer.TxID.Hex()),
						}

						// process user cache map
						err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
						if err != nil {
							cc.Log.Err(err).Msg("insufficient fund")
							return
						}

						cc.API.transactionCache.Process(tx)

						// TODO: manage user cache
						// select {
						// case cc.API.transaction <- &passport.NewTransaction{
						// 	To:                   passport.XsynSaleUserID,
						// 	From:                 user.ID,
						// 	Amount:               *supAmount,
						// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
						// 	Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with BNB %s", xfer.TxID.Hex()),
						// }:

						// case <-time.After(10 * time.Second):
						// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
						// 	panic("FAILED TO INSERT CHAIN CONFIRM ENTRY ")
						// }

						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
			case "SUPS":
				if xfer.To == cc.Params.PurchaseAddr {

					// if deposit sups
					cc.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("User", xfer.From.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("deposit")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex()), RoleID: passport.UserRoleMemberID}
							err = db.UserCreate(ctx, cc.API.Conn, user)
							if err != nil {
								cc.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							cc.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}

					// resultChan := make(chan *passport.TransactionResult)
					tx := passport.NewTransaction{
						// ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *xfer.Amount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("[DEPOSIT] SUPS on BSC %s", xfer.TxID.Hex()),
					}

					// process user cache map
					err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
					if err != nil {
						cc.Log.Err(err).Msg("insufficient fund")
						return
					}

					txID := cc.API.transactionCache.Process(tx)

					// TODO: manage user cache
					// select {
					// case cc.API.transaction <- &passport.NewTransaction{
					// 	ResultChan:           resultChan,
					// 	To:                   user.ID,
					// 	From:                 passport.XsynSaleUserID,
					// 	Amount:               *xfer.Amount,
					// 	TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
					// 	Description:          fmt.Sprintf("[DEPOSIT] SUPS on BSC %s", xfer.TxID.Hex()),
					// }:

					// case <-time.After(10 * time.Second):
					// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
					// 	panic("[DEPOSIT] SUPS on BSC")
					// }

					// result := <-resultChan

					// if result.Error != nil {
					// 	return // believe error logs already
					// }

					// if result.Transaction.Status != passport.TransactionSuccess {
					// 	cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
					// 	return
					// }

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), txID, xfer.Block, xfer.ChainID)
					if err != nil {
						tx := passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup deposit on BSC %s", xfer.TxID.Hex()),
						}

						// process user cache map
						err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
						if err != nil {
							cc.Log.Err(err).Msg("insufficient fund")
							return
						}

						cc.API.transactionCache.Process(tx)

						// TODO: manage user cache
						// select {
						// case cc.API.transaction <- &passport.NewTransaction{
						// 	To:                   passport.XsynSaleUserID,
						// 	From:                 user.ID,
						// 	Amount:               *xfer.Amount,
						// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
						// 	Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup deposit on BSC %s", xfer.TxID.Hex()),
						// }:

						// case <-time.After(10 * time.Second):
						// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
						// 	panic("FAILED TO INSERT CHAIN CONFIRM ENTRY")
						// }

						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}

				if xfer.From == cc.Params.WithdrawAddr {

					// UNTESTED
					// if withdrawing sups
					cc.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("User", xfer.To.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("withdraw")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.To.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex()), RoleID: passport.UserRoleMemberID}
							err = db.UserCreate(ctx, cc.API.Conn, user)
							if err != nil {
								cc.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							cc.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}
					// resultChan := make(chan *passport.TransactionResult)
					tx := passport.NewTransaction{
						// ResultChan:           resultChan,
						From:                 user.ID,
						To:                   passport.XsynTreasuryUserID,
						Amount:               *xfer.Amount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("[SUPS] Withdraw on BSC to %s", xfer.To.Hex()),
					}

					// process user cache map
					err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
					if err != nil {
						cc.Log.Err(err).Msg("insufficient fund")
						return
					}

					txID := cc.API.transactionCache.Process(tx)

					// TODO: manage user cache
					// select {
					// case cc.API.transaction <- &passport.NewTransaction{
					// 	ResultChan:           resultChan,
					// 	From:                 user.ID,
					// 	To:                   passport.XsynTreasuryUserID,
					// 	Amount:               *xfer.Amount,
					// 	TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
					// 	Description:          fmt.Sprintf("[SUPS] Withdraw on BSC to %s", xfer.To.Hex()),
					// }:

					// case <-time.After(10 * time.Second):
					// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
					// 	panic("[SUPS] Withdraw on BSC to")
					// }

					// result := <-resultChan
					// if result.Error != nil {
					// 	return // believe error logs already
					// }

					// if result.Transaction.Status != passport.TransactionSuccess {
					// 	cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
					// 	return
					// }
					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), txID, xfer.Block, xfer.ChainID)
					if err != nil {
						tx := passport.NewTransaction{
							To:                   user.ID,
							From:                 passport.XsynTreasuryUserID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - ssup withdraw on BSC to %s", xfer.TxID.Hex()),
						}

						// process user cache map
						err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
						if err != nil {
							cc.Log.Err(err).Msg("insufficient fund")
							return
						}

						cc.API.transactionCache.Process(tx)

						// TODO: manage user cache
						// select {
						// case cc.API.transaction <- &passport.NewTransaction{
						// 	To:                   user.ID,
						// 	From:                 passport.XsynTreasuryUserID,
						// 	Amount:               *xfer.Amount,
						// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
						// 	Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - ssup withdraw on BSC to %s", xfer.TxID.Hex()),
						// }:

						// case <-time.After(10 * time.Second):
						// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
						// 	panic(fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - ssup withdraw on BSC to %s", xfer.TxID.Hex()))
						// }

						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)

				}
			}

		case cc.Params.ETHChainID:
			switch xfer.Symbol {
			case "USDC":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with USDC
					amountTimes100 := xfer.Amount.Mul(xfer.Amount, big.NewInt(1000))
					supUSDPriceTimes100 := cc.API.State.SUPtoUSD.Mul(decimal.New(1000, 0)).BigInt()
					supAmount := amountTimes100.Div(amountTimes100, supUSDPriceTimes100)

					cc.Log.Info().
						Str("Chain", "Ethereum").
						Str("SUPS", decimal.NewFromBigInt(supAmount, 0).Div(decimal.New(1, int32(18))).String()).
						Str("USDC", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("Buyer", xfer.From.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex()), RoleID: passport.UserRoleMemberID}
							err = db.UserCreate(ctx, cc.API.Conn, user)
							if err != nil {
								cc.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							cc.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}
					// resultChan := make(chan *passport.TransactionResult)
					tx := passport.NewTransaction{
						// ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
					}

					// process user cache map
					err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
					if err != nil {
						cc.Log.Err(err).Msg("insufficient fund")
						return
					}

					txID := cc.API.transactionCache.Process(tx)

					// TODO: manage user cache
					// select {
					// case cc.API.transaction <- &passport.NewTransaction{
					// 	ResultChan:           resultChan,
					// 	To:                   user.ID,
					// 	From:                 passport.XsynSaleUserID,
					// 	Amount:               *supAmount,
					// 	TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
					// 	Description:          fmt.Sprintf("sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
					// }:

					// case <-time.After(10 * time.Second):
					// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
					// 	panic(fmt.Sprintf("sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()))
					// }

					// result := <-resultChan
					// if result.Error != nil {
					// 	return // believe error logs already
					// }

					// if result.Transaction.Status != passport.TransactionSuccess {
					// 	cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
					// 	return
					// }
					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), txID, xfer.Block, xfer.ChainID)
					if err != nil {
						tx := passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
						}

						// process user cache map
						err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
						if err != nil {
							cc.Log.Err(err).Msg("insufficient fund")
							return
						}

						cc.API.transactionCache.Process(tx)

						// TODO: manage user cache
						// select {
						// case cc.API.transaction <- &passport.NewTransaction{
						// 	To:                   passport.XsynSaleUserID,
						// 	From:                 user.ID,
						// 	Amount:               *supAmount,
						// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
						// 	Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
						// }:

						// case <-time.After(10 * time.Second):
						// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
						// 	panic(fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()))
						// }

						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
			case "ETH":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with ETH
					// TODO: probably do a * 1000 here? currently no decimals in conversion but possibly in future?
					supAmount := cc.API.State.ETHtoUSD.Div(cc.API.State.SUPtoUSD).BigInt()
					supAmount = supAmount.Mul(supAmount, xfer.Amount)

					cc.Log.Info().
						Str("Chain", "Ethereum").
						Str("SUPS", decimal.NewFromBigInt(supAmount, 0).Div(decimal.New(1, int32(18))).String()).
						Str("ETH", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("Buyer", xfer.From.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex()), RoleID: passport.UserRoleMemberID}
							err = db.UserCreate(ctx, cc.API.Conn, user)
							if err != nil {
								cc.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							cc.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}
					// resultChan := make(chan *passport.TransactionResult)
					tx := passport.NewTransaction{
						// ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("[SUPS] purchase on Ethereum with ETH %s", xfer.TxID.Hex()),
					}

					// process user cache map
					err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
					if err != nil {
						cc.Log.Err(err).Msg("insufficient fund")
						return
					}

					txID := cc.API.transactionCache.Process(tx)

					// TODO: mange user cache
					// select {
					// case cc.API.transaction <- &passport.NewTransaction{
					// 	ResultChan:           resultChan,
					// 	To:                   user.ID,
					// 	From:                 passport.XsynSaleUserID,
					// 	Amount:               *supAmount,
					// 	TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
					// 	Description:          fmt.Sprintf("[SUPS] purchase on Ethereum with ETH %s", xfer.TxID.Hex()),
					// }:

					// case <-time.After(10 * time.Second):
					// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
					// 	panic(fmt.Sprintf("[SUPS] purchase on Ethereum with ETH %s", xfer.TxID.Hex()))
					// }

					// result := <-resultChan

					// if result.Error != nil {
					// 	return // believe error logs already
					// }

					// if result.Transaction.Status != passport.TransactionSuccess {
					// 	cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
					// 	return
					// }

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), txID, xfer.Block, xfer.ChainID)
					if err != nil {

						tx := passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with ETH %s", xfer.TxID.Hex()),
						}

						// process user cache map
						err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
						if err != nil {
							cc.Log.Err(err).Msg("insufficient fund")
							return
						}

						cc.API.transactionCache.Process(tx)

						// TODO: manage user cache
						// select {
						// case cc.API.transaction <- &passport.NewTransaction{
						// 	To:                   passport.XsynSaleUserID,
						// 	From:                 user.ID,
						// 	Amount:               *supAmount,
						// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
						// 	Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with ETH %s", xfer.TxID.Hex()),
						// }:

						// case <-time.After(10 * time.Second):
						// 	cc.API.Log.Err(errors.New("timeout on channel send exceeded"))
						// 	panic(fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with ETH %s", xfer.TxID.Hex()))
						// }

						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)

				}
			}
		}
	}
	return fn
}

func (cc *ChainClients) handleBlock(ctx context.Context, client *ethclient.Client, chainID int64) func(header *bridge.Header) {
	fn := func(header *bridge.Header) {
		cc.Log.Trace().Str("Block", header.Number.String()).Msg("")

		cc.updateStateFunc(chainID, header.Number.Uint64())

		// get all transaction confirmations that are not finished
		confirmations, err := db.PendingChainConfirmationsByChainID(ctx, cc.API.Conn, chainID)
		if err != nil {
			cc.Log.Err(err).Msg("issue getting pending chain confirmations in block listener")
			return
		}

		// loop over the confirmations, check if block difference is >= 6, then validate and update
		for _, conf := range confirmations {
			confirmedBlocks, err := bridge.TransactionConfirmations(ctx, client, common.HexToHash(conf.Tx))
			if err != nil {
				cc.Log.Err(err).Msg("issue confirming transaction")
				return
			}

			// updated confirmation amount on db object
			if confirmedBlocks <= 6 {
				conf, err = db.UpdateConfirmationAmount(ctx, cc.API.Conn, conf.Tx, confirmedBlocks)
				if err != nil {
					cc.Log.Err(err).Msg("issue updating confirmed amount")
					return
				}
			}

			// if confirmed blocks greater than 6, finalize it
			if confirmedBlocks >= 6 {
				conf, err = db.ConfirmChainConfirmation(ctx, cc.API.Conn, conf.Tx)
				if err != nil {
					cc.Log.Err(err).Msg("issue setting as confirmed")
					return
				}

				cc.Log.Info().Msg("chain transaction finalized")
			}

			if confirmedBlocks > 0 {
				go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, conf.UserID.String())), conf)
			}
		}
	}
	return fn
}

func (cc *ChainClients) runBSCBridgeListener(ctx context.Context) {
	// stuff on bsc chain
	go func() {
		b := &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    30 * time.Second,
			Factor: 2,
		}

	bscClientLoop:
		for {
			ctx, cancel := context.WithCancel(ctx)

			cc.Log.Info().Msg("Attempting to connect to BSC node")
			bscClient, err := ethclient.DialContext(ctx, cc.Params.BscNodeAddr)
			if err != nil {
				cc.Log.Err(err).Msg("failed to connected to bsc node")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
			cc.BscClient = bscClient
			cc.Log.Info().Str("url", cc.Params.BscNodeAddr).Msg("Successfully connect to BSC node")

			// call the first ping outside the loop
			err = pingFunc(ctx, bscClient)
			if err != nil {
				cc.Log.Err(err).Msg("failed to connected to bsc node")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}

			errChan := make(chan error)

			// create ping pong using the block number function
			go func() {
				for {
					err := pingFunc(ctx, bscClient)
					if err != nil {
						cc.Log.Err(err).Msg("failed our ping pong with bsc")
						cancel()
						time.Sleep(b.Duration())
						errChan <- err
						return
					}
					time.Sleep(10 * time.Second)
				}
			}()

			busdListener, err := bridge.NewERC20Listener(cc.Params.BusdAddr, cc.Params.BSCChainID, cc.BscClient, cc.handleTransfer(ctx))
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for busd")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
			bnbListener := bridge.NewNativeListener(cc.BscClient, cc.Params.PurchaseAddr, "BNB", 18, cc.Params.BSCChainID, cc.handleTransfer(ctx))
			supListener, err := bridge.NewERC20Listener(cc.Params.SupAddr, cc.Params.BSCChainID, cc.BscClient, cc.handleTransfer(ctx))
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for sups")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}

			// Create the sups' controller for transferring sups
			supsController, err := bridge.NewSUPS(cc.Params.SupAddr, cc.Params.SignerPrivateKey, cc.BscClient, big.NewInt(cc.Params.BSCChainID))
			if err != nil {
				cc.Log.Err(err).Msg("failed create sups controller")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
			cc.Log.Info().Msg("Started sup controller")
			cc.SUPS = supsController

			//Path address gets the value of [0] in terms of [1]
			SUPSPathAddr := []common.Address{cc.Params.SupAddr, cc.Params.BusdAddr}

			// creates a struct that then can be used to get sup price in USD
			supGetter, err := bridge.NewPriceGetter(cc.BscClient, cc.Params.BSCRouterAddr, SUPSPathAddr) // pathAddrs are an array of contract addresses, from one token to the other
			if err != nil {
				cc.Log.Err(err).Msg("failed to get sup to busd price getter struct")
				cancel()
				return
			}

			o := bridge.NewOracle(cc.Params.MoralisKey)
			go func() {

				exchangeRateBackoff := &backoff.Backoff{
					Min:    1 * time.Second,
					Max:    30 * time.Second,
					Factor: 2,
				}

				for {
					select {
					case <-ctx.Done():
						return
					default:
						// gets how many sups for 1 busd
						supBigPrice, err := supGetter.Price(decimal.New(1, int32(18)).BigInt())
						if err != nil {
							cc.Log.Err(err).Msg("failed to get sup to busd price")
							time.Sleep(exchangeRateBackoff.Duration())
							continue
						}

						supPrice := decimal.NewFromBigInt(supBigPrice, -18)
						if supPrice == decimal.NewFromInt(0) {
							cc.Log.Warn().Msg("new supPrice was 0, exiting loop")
							continue
						}
						exchangeRateBackoff.Reset()
						cc.updatePriceFuncMu.Lock()
						cc.updatePriceFunc(supListener.TokenSymbol, supPrice)
						cc.updatePriceFuncMu.Unlock()

						time.Sleep(10 * time.Minute)
					}
				}

			}()
			go func() {
				exchangeRateBackoff := &backoff.Backoff{
					Min:    1 * time.Second,
					Max:    30 * time.Second,
					Factor: 2,
				}

				for {
					select {
					case <-ctx.Done():
						return
					default:

						//gets how many bnb for 1 busd
						bnbPrice, err := o.BNBUSDPrice()
						if err != nil {
							cc.Log.Err(err).Msg("failed to get bnb price")
							time.Sleep(exchangeRateBackoff.Duration())
							continue
						}
						if bnbPrice == decimal.NewFromInt(0) {
							cc.Log.Warn().Msg("new bnbPrice was 0, exiting loop")
							continue
						}

						exchangeRateBackoff.Reset()

						cc.updatePriceFuncMu.Lock()
						cc.updatePriceFunc(bnbListener.Symbol, bnbPrice)
						cc.updatePriceFuncMu.Unlock()

						time.Sleep(10 * time.Minute)
					}

				}
			}()

			/*****************
			This second replays any blocks we've missed.
			The cc.state object is retrieved in the cc.Run func
			If we have difference in blocks we get all the transaction records and then rerun them on our transaction handler
			We update the state block in the block listeners (cc.handleBlock())
			 *****************/

			go func() {
				// get current BSC block
				currentBSCBlock, err := cc.BscClient.BlockNumber(ctx)
				if err != nil {
					cc.Log.Err(err).Msg("failed to get bsc block number")
					errChan <- err
					return
				}

				// if diff is greater than 1 we need to replay some blocks
				if currentBSCBlock-cc.API.State.LatestBscBlock > 0 {
					cc.Log.Info().Uint64("amount", currentBSCBlock-cc.API.State.LatestBscBlock).Msg("Replaying BSC blocks")
					go func() {
						BUSDrecords, err := busdListener.Replay(ctx, int(cc.API.State.LatestBscBlock), int(currentBSCBlock))
						if err != nil {
							cc.Log.Err(err).Msg("failed to replay transactions for BUSD")
							errChan <- err
							return
						}
						for _, record := range BUSDrecords {
							fn := cc.handleTransfer(ctx)
							fn(record)
						}
					}()
					go func() {
						err := bnbListener.Replay(ctx, int(cc.API.State.LatestBscBlock), int(currentBSCBlock))
						if err != nil {
							cc.Log.Err(err).Msg("failed to replay transactions for BNB")
							errChan <- err
							return
						}
					}()
					go func() {
						SUPRecords, err := supListener.Replay(ctx, int(cc.API.State.LatestBscBlock), int(currentBSCBlock))
						if err != nil {
							cc.Log.Err(err).Msg("failed to replay transactions for SUPS")
							errChan <- err
							return
						}
						for _, record := range SUPRecords {
							fn := cc.handleTransfer(ctx)
							fn(record)
						}
					}()
					cc.updateStateFunc(cc.Params.BSCChainID, currentBSCBlock)
				}
			}()

			/*************
			start block/header listeners
			************/

			// listen for withdraw contract

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						listener := bridge.NewWithdrawerListener(cc.Params.WithdrawAddr, cc.BscClient, cc.Params.BSCChainID, func(withdraw *bridge.Withdraw) {
							cc.Log.Info().
								Str("User", withdraw.To.Hex()).
								Str("Amount", decimal.NewFromBigInt(withdraw.Amount, 0).Div(decimal.New(1, int32(18))).String()).
								Msg("withdraw")
							user, err := db.UserByPublicAddress(ctx, cc.API.Conn, withdraw.To.Hex())
							if err != nil {
								cc.Log.Err(err).Msg("issue finding user for withdraw")
								return
							}

							txID := uuid.Must(uuid.NewV4())

							// resultChan := make(chan *passport.TransactionResult)

							tx := passport.NewTransaction{
								// ResultChan:           resultChan,
								To:                   passport.OnChainUserID,
								From:                 user.ID,
								Amount:               *withdraw.Amount,
								TransactionReference: passport.TransactionReference(fmt.Sprintf("%s:%s:%d:%s", withdraw.To, withdraw.Amount.String(), withdraw.Block, txID.String())),
								Description:          "sup withdraw on bsc",
							}

							// process user cache map
							err = cc.API.userCacheMap.Process(tx.From, tx.To, tx.Amount)
							if err != nil {
								cc.Log.Err(err).Msg("insufficient fund")
								return
							}

							cc.API.transactionCache.Process(tx)

							// TODO: manage user cache

							// cc.API.transaction <- &passport.NewTransaction{
							// 	ResultChan:           resultChan,
							// 	To:                   passport.OnChainUserID,
							// 	From:                 user.ID,
							// 	Amount:               *withdraw.Amount,
							// 	TransactionReference: passport.TransactionReference(fmt.Sprintf("%s:%s:%d:%s", withdraw.To, withdraw.Amount.String(), withdraw.Block, txID.String())),
							// 	Description:          "sup withdraw on bsc",
							// }

							// result := <-resultChan

							// if result.Error != nil {
							// 	return // believe error logs already
							// }

							// if result.Transaction.Status != passport.TransactionSuccess {
							// 	cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
							// 	return
							// }
						})
						err = listener.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to bsc blocks")
							errChan <- err
							return
						}

					}
				}
			}()

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("chain", "BSC").Msg("Start header listener")
						blockBSC := bridge.NewHeadListener(cc.BscClient, cc.Params.BSCChainID, cc.handleBlock(ctx, cc.BscClient, cc.Params.BSCChainID))
						err := blockBSC.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to bsc blocks")
							errChan <- err
							return
						}
					}
				}
			}()

			/*************
			start wallet and contract listeners
			************/
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("sym", "WNB").Msg("Start listener")
						err := bnbListener.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to bnb")
							errChan <- err
							return
						}
					}
				}

			}()
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("sym", "SUP").Msg("Start listener")
						err := supListener.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to sups")
							errChan <- err
							return
						}
					}

				}
			}()
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						for {
							cc.Log.Info().Str("sym", "BUSD").Msg("Start listener")
							err := busdListener.Listen(ctx)
							if err != nil {
								cc.Log.Err(err).Msg("error listening to busd")
								errChan <- err
								return
							}
						}
					}
				}
			}()

			b.Reset()
			// listen for err chan
			err = <-errChan
			if err != nil {
				cc.Log.Err(err).Msg("error listening to busd, attempting to reconnect...")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
		}
	}()
}

func (cc *ChainClients) runETHBridgeListener(ctx context.Context) {

	// stuff on eth chain
	go func() {
		b := &backoff.Backoff{
			Min:    1 * time.Second,
			Max:    30 * time.Second,
			Factor: 2,
		}

	ethClientLoop:
		for {
			ctx, cancel := context.WithCancel(ctx)

			cc.Log.Info().Msg("Attempting to connect to ETH node")

			ethClient, err := ethclient.DialContext(ctx, cc.Params.EthNodeAddr)
			if err != nil {
				cc.Log.Err(err).Msg("failed to connected to eth node")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}
			cc.EthClient = ethClient
			cc.Log.Info().Str("url", cc.Params.EthNodeAddr).Msg("Successfully connect to ETH node")

			// call the first ping outside the loop
			err = pingFunc(ctx, ethClient)
			if err != nil {
				cc.Log.Err(err).Msg("failed to ping the eth node")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}

			errChan := make(chan error)

			// create ping pong using the block number function
			go func() {
				for {
					err := pingFunc(ctx, ethClient)
					if err != nil {
						cc.Log.Err(err).Msg("failed our ping pong with eth")
						cancel()
						time.Sleep(b.Duration())
						errChan <- err
						return
					}
					time.Sleep(10 * time.Second)
				}
			}()

			usdcListener, err := bridge.NewERC20Listener(cc.Params.UsdcAddr, cc.Params.ETHChainID, cc.EthClient, cc.handleTransfer(ctx))
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for usdc")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}

			ethListener := bridge.NewNativeListener(cc.EthClient, cc.Params.PurchaseAddr, "ETH", 18, cc.Params.ETHChainID, cc.handleTransfer(ctx))

			nftListener, err := bridge.NewERC721Listener(
				cc.Params.EthNftAddr,
				cc.EthClient,
				cc.Params.ETHChainID,
				cc.handleNFTTransfer(ctx),
				func(event *bridge.NFTStakeEvent) {},
				func(event *bridge.NFTUnstakeEvent) {},
				func(*bridge.NFTLockEvent) {},
				func(*bridge.NFTUnlockEvent) {},
				func(*bridge.NFTRemapEvent) {},
			)
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for eth nft")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}

			nftStakingListener, err := bridge.NewERC721Listener(
				cc.Params.EthNftStakingAddr,
				cc.EthClient,
				cc.Params.ETHChainID,
				func(event *bridge.NFTTransferEvent) {},
				func(event *bridge.NFTStakeEvent) {
					cc.Log.Info().Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msg("nft stake")

					// user from wallet address
					var user *passport.User
					var err error
					user, err = db.UserByPublicAddress(ctx, cc.API.Conn, event.Owner.Hex())
					if err != nil {
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: event.Owner.Hex(), PublicAddress: passport.NewString(event.Owner.Hex()), RoleID: passport.UserRoleMemberID}
							err = db.UserCreate(ctx, cc.API.Conn, user)
							if err != nil {
								cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("issue creating new user and unable to find a user")
								return
							}
						}
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("issue finding user to assign token")
						return
					}

					// get asset
					asset, err := db.AssetGet(ctx, cc.API.Conn, event.TokenID.Uint64())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("unable to find asset to transfer")
						return
					}

					// check if asset has handled this tx already
					if asset == nil {
						cc.Log.Err(err).Msgf("failed to find asset, asset was nil: %s", event.TokenID.String())
						return
					}
					for _, tx := range asset.TxHistory {
						if tx == event.TxID.Hex() {
							return
						}
					}

					// check asset is owned by on chain user
					if asset.UserID == nil || asset.UserID.IsNil() ||
						*asset.UserID != passport.OnChainUserID {
						cc.Log.Err(fmt.Errorf("not owned by on chain user")).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("asset is not owned by the on chain user")
						return
					}

					//func AssetTransfer(ctx context.Context, conn Conn, tokenID uint64, oldUserID, newUserID passport.UserID, txHash string) error {
					err = db.AssetTransfer(ctx, cc.API.Conn, asset.TokenID, passport.OnChainUserID, user.ID, event.TxID.Hex())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("failed to transfer asset")
						return
					}

					// TODO: remove this and update asset transfer to return updated asset instead
					asset, err = db.AssetGet(ctx, cc.API.Conn, event.TokenID.Uint64())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("unable to find asset to transfer")
						return
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.TokenID)), asset)
				},
				func(event *bridge.NFTUnstakeEvent) {
					cc.Log.Info().Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msg("nft unstake")
					// get asset
					asset, err := db.AssetGet(ctx, cc.API.Conn, event.TokenID.Uint64())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("unable to find asset to remove")
						return
					}

					// check if asset has handled this tx already
					if asset == nil {
						cc.Log.Err(err).Msgf("failed to find asset, asset was nil: %s", event.TokenID.String())
						return
					}
					for _, tx := range asset.TxHistory {
						if tx == event.TxID.Hex() {
							return
						}
					}

					// remove the asset from user
					err = db.AssetTransfer(ctx, cc.API.Conn, asset.TokenID, *asset.UserID, passport.OnChainUserID, event.TxID.Hex())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("failed to transfer asset")
						return
					}

					// TODO: remove this and update asset transfer to return updated asset instead
					asset, err = db.AssetGet(ctx, cc.API.Conn, event.TokenID.Uint64())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("unable to find asset to transfer")
						return
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.TokenID)), asset)
				},
				func(*bridge.NFTLockEvent) {},
				func(*bridge.NFTUnlockEvent) {},
				func(*bridge.NFTRemapEvent) {},
			)
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for eth staking nft")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}

			// TODO use real deposit address
			blockEth := bridge.NewHeadListener(cc.EthClient, cc.Params.ETHChainID, cc.handleBlock(ctx, ethClient, cc.Params.ETHChainID))

			o := bridge.NewOracle(cc.Params.MoralisKey)

			go func() {
				exchangeRateBackoff := &backoff.Backoff{
					Min:    1 * time.Second,
					Max:    30 * time.Second,
					Factor: 2,
				}
				select {
				case <-ctx.Done():
					return
				default:

					for {
						// //gets how many eth for 1 busd
						ethPrice, err := o.ETHUSDPrice()
						if err != nil {
							cc.Log.Err(err).Msg("Could not get ETH price")
							time.Sleep(exchangeRateBackoff.Duration())
							continue
						}

						if ethPrice == decimal.NewFromInt(0) {
							cc.Log.Warn().Msg("new ethPrice was 0, exiting loop")
							continue
						}

						exchangeRateBackoff.Reset()

						cc.updatePriceFuncMu.Lock()
						cc.updatePriceFunc(ethListener.Symbol, ethPrice)
						cc.updatePriceFuncMu.Unlock()

						time.Sleep(10 * time.Second)
					}
				}
			}()

			// replay
			go func() {
				// get current ETH block
				currentETHBlock, err := cc.EthClient.BlockNumber(ctx)
				if err != nil {
					cc.Log.Err(err).Msg("failed to get eth block number")
					errChan <- err
					return
				}
				// if diff is greater than 1 we need to replay some blocks
				if currentETHBlock-cc.API.State.LatestEthBlock > 0 {
					cc.Log.Info().Uint64("amount", currentETHBlock-cc.API.State.LatestEthBlock).Msg("Replaying ETH blocks")
					go func() {
						USDCrecords, err := usdcListener.Replay(ctx, int(cc.API.State.LatestEthBlock), int(currentETHBlock))
						if err != nil {
							cc.Log.Err(err).Msg("failed to replay transactions for USDC")
							errChan <- err
							return
						}
						for _, record := range USDCrecords {
							fn := cc.handleTransfer(ctx)
							fn(record)
						}

					}()
					go func() {
						err := ethListener.Replay(ctx, int(cc.API.State.LatestEthBlock), int(currentETHBlock))
						if err != nil {
							cc.Log.Err(err).Msg("failed to replay transactions for ETH")
							errChan <- err
							return
						}
					}()
					go func() {
						err := nftListener.Replay(ctx, int(cc.API.State.LatestEthBlock), int(currentETHBlock))
						if err != nil {
							cc.Log.Err(err).Msg("failed to replay transactions for NFT")
							errChan <- err
							return
						}
					}()
					go func() {
						err := nftStakingListener.Replay(ctx, int(cc.API.State.LatestEthBlock), int(currentETHBlock))
						if err != nil {
							cc.Log.Err(err).Msg("failed to replay staking for NFT")
							errChan <- err
							return
						}
					}()
					cc.updateStateFunc(cc.Params.ETHChainID, currentETHBlock)
				}
			}()

			// block listener
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("chain", "ETH").Msg("Start header listener")
						err := blockEth.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to eth blocks")
							errChan <- err
							return
						}
					}
				}
			}()

			// nft mint listener
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("sym", "NFT").Str("type", "Mint/Transfer").Msg("Start listener")
						err := nftListener.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to eth nft")
							errChan <- err
							return
						}
					}
				}
			}()
			// nft staking listener
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("sym", "NFT").Str("type", "Staking").Msg("Start listener")
						err := nftStakingListener.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to eth nft staking")
							errChan <- err
							return
						}
					}
				}
			}()

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("sym", "ETH").Msg("Start listener")
						err := ethListener.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to eth")
							errChan <- err
							return
						}
					}
				}
			}()
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("sym", "USDC").Msg("Start listener")
						err := usdcListener.Listen(ctx)
						if err != nil {
							cc.Log.Err(err).Msg("error listening to usdc")
							errChan <- err
							return
						}
					}
				}
			}()

			err = <-errChan
			if err != nil {
				cc.Log.Err(err).Msg("error listening to busd, attempting to reconnect...")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}
		}
	}()

}

func (cc *ChainClients) CheckEthTx(w http.ResponseWriter, r *http.Request) (int, error) {
	// Get token id
	txID := chi.URLParam(r, "tx_id")
	if txID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tx id"), "Missing Tx.")
	}

	if cc.EthClient == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("eth client is nil"), "Issue accessing ETH node, please try again or contact support.")
	}

	record, _, err := bridge.GetTransfer(r.Context(), cc.EthClient, cc.Params.ETHChainID, common.HexToHash(txID))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, fmt.Sprintf("Issue finding transaction: %s on chain: %d", txID, cc.Params.ETHChainID))
	}

	cc.API.Log.Info().
		Str("Symbol", record.Symbol).
		Str("Amount", decimal.NewFromBigInt(record.Amount, 0).Div(decimal.New(1, int32(record.Decimals))).String()).
		Str("TxID", record.TxID.String()).
		Str("From", record.From.String()).
		Str("To", record.To.String()).
		Msg("running eth tx checker")
	fn := cc.handleTransfer(r.Context())
	fn(record)

	return http.StatusOK, nil
}

func (cc *ChainClients) CheckBscTx(w http.ResponseWriter, r *http.Request) (int, error) {
	// Get token id
	txID := chi.URLParam(r, "tx_id")
	if txID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tx id"), "Missing Tx.")
	}

	if cc.BscClient == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("bsc client is nil"), "Issue accessing BSC node, please try again or contact support.")
	}

	record, pending, err := bridge.GetTransfer(r.Context(), cc.BscClient, cc.Params.BSCChainID, common.HexToHash(txID))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, fmt.Sprintf("Issue finding transaction: %s on chain: %d", txID, cc.Params.BSCChainID))
	}

	if pending {
		_, err := w.Write([]byte("pending"))
		if err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	}

	cc.API.Log.Info().
		Str("Symbol", record.Symbol).
		Str("Amount", decimal.NewFromBigInt(record.Amount, 0).Div(decimal.New(1, int32(record.Decimals))).String()).
		Str("TxID", record.TxID.String()).
		Str("From", record.From.String()).
		Str("To", record.To.String()).
		Msg("running bsc tx checker")
	fn := cc.handleTransfer(r.Context())
	fn(record)

	return http.StatusOK, nil
}

func (cc *ChainClients) handleNFTTransfer(ctx context.Context) func(xfer *bridge.NFTTransferEvent) {
	fn := func(ev *bridge.NFTTransferEvent) {
		func() {
			if ev.From.Hex() == cc.Params.EthNftStakingAddr.Hex() || ev.To.Hex() == cc.Params.EthNftStakingAddr.Hex() {
				return
			}

			asset, err := db.AssetGet(ctx, cc.API.Conn, ev.TokenID.Uint64())
			if err != nil {
				cc.Log.Err(err).Msgf("issue getting asset: %s", ev.TokenID.String())
				return
			}
			if asset == nil {
				cc.Log.Err(err).Msgf("failed to find asset, asset was nil: %s", ev.TokenID.String())
				return
			}

			// check if asset has handled this tx already
			for _, tx := range asset.TxHistory {
				if tx == ev.TxID.Hex() {
					return
				}
			}

			// if asset owner is passport.onchainuser then it is an external transfer, so just store the tx hash
			if asset.UserID != nil && *asset.UserID == passport.OnChainUserID {
				err := db.AssetTransferOnChain(ctx, cc.API.Conn, ev.TokenID.Uint64(), ev.TxID.Hex())
				if err != nil {
					cc.Log.Err(err).Msgf("failed to add tx hash to array asset: %s, tx: %s", ev.TokenID.String(), ev.TxID.Hex())
					return
				}
				return
			}

			cc.Log.Info().
				Str("Chain", "ETH").
				Str("From", ev.From.Hex()).
				Str("To", ev.To.Hex()).
				Str("Token ID", ev.TokenID.String()).
				Msg("nft mint")

			// get user
			user, err := db.UserByPublicAddress(ctx, cc.API.Conn, ev.To.Hex())
			if err != nil {
				cc.Log.Err(err).Msgf("issue finding user from public address: %s, locking and freezing asset token id %s", ev.To.Hex(), ev.TokenID.String())
				// if issue transferring asset, LOCK IT!
				err := db.XsynAssetLock(ctx, cc.API.Conn, ev.TokenID.Uint64(), passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO LOCK ASSET token id: %s", ev.TokenID.String())
				}
				err = db.XsynAssetFreeze(ctx, cc.API.Conn, ev.TokenID.Uint64(), passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO FREEZE ASSET token id: %s", ev.TokenID.String())
				}
				return
			}

			// remove all the stores sigs for their other assets
			err = db.XsynAssetMintUnLock(ctx, cc.API.Conn, user.ID)
			if err != nil {
				cc.Log.Err(err).Msgf("failed to clear users asset mint signatures, user: %s", user.ID)
			}

			// check user owns asset
			if asset.UserID == nil || *asset.UserID != user.ID {
				cc.Log.Err(err).Msgf("this wallet address doesn't own this asset, locking and freezing asset token id %s", ev.TokenID.String())
				// if issue transferring asset, LOCK IT!
				err := db.XsynAssetLock(ctx, cc.API.Conn, ev.TokenID.Uint64(), passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO LOCK ASSET token id: %s", ev.TokenID.String())
				}
				err = db.XsynAssetFreeze(ctx, cc.API.Conn, ev.TokenID.Uint64(), passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO FREEZE ASSET token id: %s", ev.TokenID.String())
				}
				return
			}

			err = db.AssetTransfer(ctx, cc.API.Conn, ev.TokenID.Uint64(), user.ID, passport.OnChainUserID, ev.TxID.Hex())
			if err != nil {
				cc.Log.Err(err).Msgf("issue transferring asset token id: %s, locking and freezing it", ev.TokenID.String())
				// if issue transferring asset, LOCK IT!
				err := db.XsynAssetLock(ctx, cc.API.Conn, ev.TokenID.Uint64(), passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO LOCK ASSET token id: %s", ev.TokenID.String())
				}
				err = db.XsynAssetFreeze(ctx, cc.API.Conn, ev.TokenID.Uint64(), passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO FREEZE ASSET token id: %s", ev.TokenID.String())
				}
				return
			}
		}()

		// mark as minted
		err := db.XsynAssetMinted(ctx, cc.API.Conn, ev.TokenID.Uint64())
		if err != nil {
			cc.Log.Err(err).Msgf("failed to find asset to mark minted: %s", ev.TokenID.String())
			return
		}

		// get updated asset
		asset, err := db.AssetGet(ctx, cc.API.Conn, ev.TokenID.Uint64())
		if err != nil {
			cc.Log.Err(err).Msgf("failed to find asset to reply: %s", ev.TokenID.String())
			return
		}
		if asset == nil {
			cc.Log.Err(err).Msgf("failed to find asset its nil?: %s", ev.TokenID.String())
			return
		}
		go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, ev.TokenID.String())), asset)

	}
	return fn
}

func pingFunc(ctx context.Context, client *ethclient.Client) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := client.BlockNumber(ctxTimeout)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}
