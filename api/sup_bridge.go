package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"passport/db"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v4"
	"github.com/jpillora/backoff"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"

	"github.com/ethereum/go-ethereum/common"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/ethclient"
	client "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

type ChainClients struct {
	EthClient       *ethclient.Client
	BscClient       *ethclient.Client
	Params          *passport.BridgeParams
	API             *API
	Log             *zerolog.Logger
	updateStateFunc func(chainID int64, newBlock uint64)
}

func RunChainListeners(log *zerolog.Logger, api *API, p *passport.BridgeParams) *ChainClients {
	ctx := context.Background()
	cc := &ChainClients{
		Params: p,
		API:    api,
		Log:    log,
	}

	// func to update state
	cc.updateStateFunc = func(chainID int64, newBlock uint64) {
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

	go cc.runETHBridgeListener(ctx)
	go cc.runBSCBridgeListener(ctx)

	return cc
}

func (cc *ChainClients) handleTransfer() func(xfer *bridge.Transfer) {
	fn := func(xfer *bridge.Transfer) {
		chainID := xfer.ChainID
		switch chainID {
		case cc.Params.BSCChainID:
			switch xfer.Symbol {
			case "BUSD":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with BUSD
					ctx := context.Background()

					amountTimes100 := xfer.Amount.Mul(xfer.Amount, big.NewInt(1000))
					supUSDPriceTimes100 := cc.Params.SUPToUSD.Mul(decimal.New(1000, 0)).BigInt()
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
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
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

					resultChan := make(chan *passport.TransactionResult)

					cc.API.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID)
					if err != nil {
						cc.API.transaction <- &passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
						}
						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
			case "WBNB":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with WBNB
					ctx := context.Background()
					// TODO: probably do a * 1000 here? currently no decimals in conversion but possibly in future?
					supAmount := cc.Params.WBNBToSUPS.BigInt()
					supAmount = supAmount.Mul(supAmount, xfer.Amount)

					cc.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", decimal.NewFromBigInt(supAmount, 0).Div(decimal.New(1, int32(18))).String()).
						Str("WBNB", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("Buyer", xfer.From.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
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

					resultChan := make(chan *passport.TransactionResult)

					cc.API.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on BSC with WBNB %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID)
					if err != nil {
						cc.API.transaction <- &passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with WBNB %s", xfer.TxID.Hex()),
						}
						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
			case "SUPS":
				if xfer.To == cc.Params.PurchaseAddr {
					ctx := context.Background()

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
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
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

					resultChan := make(chan *passport.TransactionResult)

					cc.API.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *xfer.Amount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup deposit on BSC %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID)
					if err != nil {
						cc.API.transaction <- &passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup deposit on BSC %s", xfer.TxID.Hex()),
						}
						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
				if xfer.To == cc.Params.RedemptionAddr {
					// UNTESTED
					ctx := context.Background()
					//busdAmount := d.Div(p.BUSDToSUPS)

					// make sup cost 1000 * bigger to not deal with decimals
					supUSDPriceTimes1000 := cc.Params.SUPToUSD.Mul(decimal.New(1000, 0)).BigInt()
					// amount * sup to usd price
					amountTimesSupsPrice := xfer.Amount.Mul(xfer.Amount, supUSDPriceTimes1000)
					// divide by 1000 to bring it back down
					amountTimesSupsPriceNormalized := amountTimesSupsPrice.Div(amountTimesSupsPrice, big.NewInt(1000))
					// so now we have it at 18 decimals because that is what sups are, we need to reduce it to match the given token decimal
					// TODO: get decimals for from chain for BUSD
					busdDecimals := 6
					decimalDifference := xfer.Decimals - busdDecimals
					toDivideBy := big.NewInt(10)
					toDivideBy = toDivideBy.Exp(toDivideBy, big.NewInt(int64(decimalDifference)), nil)
					amountOfBUSD := amountTimesSupsPriceNormalized.Div(amountTimesSupsPriceNormalized, toDivideBy)

					cc.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("BUSD", decimal.NewFromBigInt(amountOfBUSD, 0).Div(decimal.New(1, int32(6))).String()). // TODO: get decimals from chain for this
						Str("User", xfer.From.Hex()).
						Msg("redeem")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
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

					resultChan := make(chan *passport.TransactionResult)

					cc.API.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   passport.XsynTreasuryUserID,
						From:                 user.ID,
						Amount:               *xfer.Amount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup redeem on BSC to BUSD %s", xfer.TxID.Hex()),
					}

					result := <-resultChan
					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}
					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID)
					if err != nil {
						cc.API.transaction <- &passport.NewTransaction{
							To:                   user.ID,
							From:                 passport.XsynTreasuryUserID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup redeem on BSC to BUSD %s", xfer.TxID.Hex()),
						}
						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
				if xfer.From == cc.Params.WithdrawAddr {
					ctx := context.Background()

					// UNTESTED
					// if withdrawing sups
					cc.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("User", xfer.To.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("withdraw")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
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
					resultChan := make(chan *passport.TransactionResult)

					cc.API.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   passport.XsynTreasuryUserID,
						From:                 user.ID,
						Amount:               *xfer.Amount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup withdraw on BSC to %s", xfer.TxID.Hex()),
					}

					result := <-resultChan
					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}
					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID)
					if err != nil {
						cc.API.transaction <- &passport.NewTransaction{
							To:                   user.ID,
							From:                 passport.XsynTreasuryUserID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - ssup withdraw on BSC to %s", xfer.TxID.Hex()),
						}
						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)

				}
			}

		case cc.Params.ETHChainID:
			switch xfer.Symbol {
			case "USDC":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with USDC
					ctx := context.Background()
					amountTimes100 := xfer.Amount.Mul(xfer.Amount, big.NewInt(1000))
					supUSDPriceTimes100 := cc.Params.SUPToUSD.Mul(decimal.New(1000, 0)).BigInt()
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
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
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
					resultChan := make(chan *passport.TransactionResult)

					cc.API.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
					}

					result := <-resultChan
					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}
					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID)
					if err != nil {
						cc.API.transaction <- &passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
						}
						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)
				}
			case "WETH":
				if xfer.To == cc.Params.PurchaseAddr {
					// if buying sups with WETH
					ctx := context.Background()
					// TODO: probably do a * 1000 here? currently no decimals in conversion but possibly in future?
					supAmount := cc.Params.WETHToSUPS.BigInt()
					supAmount = supAmount.Mul(supAmount, xfer.Amount)

					cc.Log.Info().
						Str("Chain", "Ethereum").
						Str("SUPS", decimal.NewFromBigInt(supAmount, 0).Div(decimal.New(1, int32(18))).String()).
						Str("WETH", decimal.NewFromBigInt(xfer.Amount, 0).Div(decimal.New(1, int32(xfer.Decimals))).String()).
						Str("Buyer", xfer.From.Hex()).
						Str("TxID", xfer.TxID.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, cc.API.Conn, xfer.From.Hex())
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
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
					resultChan := make(chan *passport.TransactionResult)

					cc.API.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.XsynSaleUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on Ethereum with WETH %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						cc.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					conf, err := db.CreateChainConfirmationEntry(ctx, cc.API.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID)
					if err != nil {
						cc.API.transaction <- &passport.NewTransaction{
							To:                   passport.XsynSaleUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with WETH %s", xfer.TxID.Hex()),
						}
						cc.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
					cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), conf)

				}
			}
		}
	}
	return fn
}

func (cc *ChainClients) handleBlock(client *ethclient.Client, chainID int64) func(header *bridge.Header) {
	fn := func(header *bridge.Header) {
		ctx := context.Background()
		cc.Log.Info().Str("Block", header.Number.String())

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
				cc.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, conf.UserID.String())), conf)
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
			bscClient, err := client.DialContext(ctx, cc.Params.BscNodeAddr)
			if err != nil {
				cc.Log.Err(err).Msg("failed to connected to bsc node")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
			cc.BscClient = bscClient
			cc.Log.Info().Msg("Successfully to connect to BSC node")

			busdListener, err := bridge.NewERC20Listener(cc.Params.BusdAddr, cc.BscClient, cc.handleTransfer())
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for busd")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
			wbnbListener, err := bridge.NewERC20Listener(cc.Params.WbnbAddr, cc.BscClient, cc.handleTransfer())
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for wbnb")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
			supListener, err := bridge.NewERC20Listener(cc.Params.SupAddr, cc.BscClient, cc.handleTransfer())
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for sups")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}
			blockBSC := bridge.NewHeadListener(cc.BscClient, cc.handleBlock(cc.BscClient, cc.Params.BSCChainID))

			// get current BSC block
			currentBSCBlock, err := cc.BscClient.BlockNumber(ctx)
			if err != nil {
				cc.Log.Err(err).Msg("failed to get bsc block number")
				cancel()
				time.Sleep(b.Duration())
				continue bscClientLoop
			}

			errChan := make(chan error)

			// create ping pong using the block number function
			go func() {
				for {
					ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
					_, err := cc.BscClient.BlockNumber(ctxTimeout)
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

			/*****************
			This second replays any blocks we've missed.
			The cc.state object is retrieved in the cc.Run func
			If we have difference in blocks we get all the transaction records and then rerun them on our transaction handler
			We update the state block in the block listeners (cc.handleBlock())
			 *****************/

			// if diff is greater than 1 we need to replay some blocks
			if currentBSCBlock-cc.API.State.LatestBscBlock > 0 {
				cc.Log.Info().Uint64("amount", currentBSCBlock-cc.API.State.LatestBscBlock).Msg("Replaying BSC blocks")
				go func() {
					BUSDrecords, err := busdListener.Replay(int(cc.API.State.LatestBscBlock), int(currentBSCBlock))
					if err != nil {
						cc.Log.Err(err).Msg("failed to replay transactions for BUSD")
						errChan <- err
						return
					}
					for _, record := range BUSDrecords {
						fn := cc.handleTransfer()
						fn(record)
					}
				}()
				go func() {
					WBNBrecords, err := wbnbListener.Replay(int(cc.API.State.LatestBscBlock), int(currentBSCBlock))
					if err != nil {
						cc.Log.Err(err).Msg("failed to replay transactions for WBNB")
						errChan <- err
						return
					}
					for _, record := range WBNBrecords {
						fn := cc.handleTransfer()
						fn(record)
					}
				}()
				go func() {
					SUPRecords, err := supListener.Replay(int(cc.API.State.LatestBscBlock), int(currentBSCBlock))
					if err != nil {
						cc.Log.Err(err).Msg("failed to replay transactions for SUPS")
						errChan <- err
						return
					}
					for _, record := range SUPRecords {
						fn := cc.handleTransfer()
						fn(record)
					}
				}()
			}

			/*************
			start block/header listeners
			************/

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("chain", "BSC").Msg("start header listener")
						err := blockBSC.Listen()
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
						cc.Log.Info().Str("sym", "WBNB").Msg("start listener")
						err := wbnbListener.Listen()
						if err != nil {
							cc.Log.Err(err).Msg("error listening to wbnb")
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
						cc.Log.Info().Str("sym", "SUP").Msg("start listener")
						err := supListener.Listen()
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
							cc.Log.Info().Str("sym", "BUSD").Msg("start listener")
							err := busdListener.Listen()
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

			ethClient, err := client.DialContext(ctx, cc.Params.EthNodeAddr)
			if err != nil {
				cc.Log.Err(err).Msg("failed to connected to eth node")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}
			cc.EthClient = ethClient
			cc.Log.Info().Msg("Successfully to connect to ETH node")

			usdcListener, err := bridge.NewERC20Listener(cc.Params.UsdcAddr, cc.EthClient, cc.handleTransfer())
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for usdc")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}
			wethListener, err := bridge.NewERC20Listener(cc.Params.WethAddr, cc.EthClient, cc.handleTransfer())
			if err != nil {
				cc.Log.Err(err).Msg("failed create listener for weth")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}
			blockEth := bridge.NewHeadListener(cc.EthClient, cc.handleBlock(ethClient, cc.Params.ETHChainID))
			// get current ETH block
			currentETHBlock, err := cc.EthClient.BlockNumber(ctx)
			if err != nil {
				cc.Log.Err(err).Msg("failed to get eth block number")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}

			errChan := make(chan error)

			// start out ping pong
			go func() {
				for {
					ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
					_, err := ethClient.BlockNumber(ctxTimeout)
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

			// if diff is greater than 1 we need to replay some blocks
			if currentETHBlock-cc.API.State.LatestEthBlock > 0 {
				cc.Log.Info().Uint64("amount", currentETHBlock-cc.API.State.LatestEthBlock).Msg("Replaying ETH blocks")
				go func() {
					USDCrecords, err := usdcListener.Replay(int(cc.API.State.LatestEthBlock), int(currentETHBlock))
					if err != nil {
						cc.Log.Err(err).Msg("failed to replay transactions for USDC")
						errChan <- err
					}
					for _, record := range USDCrecords {
						fn := cc.handleTransfer()
						fn(record)
					}
				}()
				go func() {
					WETHrecords, err := wethListener.Replay(int(cc.API.State.LatestEthBlock), int(currentETHBlock))
					if err != nil {
						cc.Log.Err(err).Msg("failed to replay transactions for WETH")
						errChan <- err
					}
					for _, record := range WETHrecords {
						fn := cc.handleTransfer()
						fn(record)
					}
				}()
			}

			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					default:
						cc.Log.Info().Str("chain", "ETH").Msg("start header listener")
						err := blockEth.Listen()
						if err != nil {
							cc.Log.Err(err).Msg("error listening to eth blocks")
							errChan <- err
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
						cc.Log.Info().Str("sym", "WETH").Msg("start listener")
						err := wethListener.Listen()
						if err != nil {
							cc.Log.Err(err).Msg("error listening to weth")
							errChan <- err
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
						cc.Log.Info().Str("sym", "USDC").Msg("start listener")
						err := usdcListener.Listen()
						if err != nil {
							cc.Log.Err(err).Msg("error listening to usdc")
							errChan <- err
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
	ctx := context.Background()
	// Get token id
	txID := chi.URLParam(r, "tx_id")
	if txID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tx id"), "Missing Tx.")
	}

	if cc.EthClient == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("eth client is nil"), "Issue accessing ETH node, please try again or contact support.")
	}

	record, _, err := bridge.GetTransfer(ctx, cc.EthClient, cc.Params.ETHChainID, common.HexToHash(txID))
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
	fn := cc.handleTransfer()
	fn(record)

	return http.StatusOK, nil
}

func (cc *ChainClients) CheckBscTx(w http.ResponseWriter, r *http.Request) (int, error) {
	ctx := context.Background()
	// Get token id
	txID := chi.URLParam(r, "tx_id")
	if txID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tx id"), "Missing Tx.")
	}

	if cc.BscClient == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("bsc client is nil"), "Issue accessing BSC node, please try again or contact support.")
	}

	record, pending, err := bridge.GetTransfer(ctx, cc.BscClient, cc.Params.BSCChainID, common.HexToHash(txID))
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
	fn := cc.handleTransfer()
	fn(record)

	return http.StatusOK, nil
}
