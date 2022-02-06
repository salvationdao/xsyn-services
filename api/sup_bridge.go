package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"passport"
	"passport/db"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/ethereum/go-ethereum/common"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/ethclient"
	client "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

func (api *API) handleTransfer(p *passport.BridgeParams) func(xfer *bridge.Transfer) {
	fn := func(xfer *bridge.Transfer) {
		chainID := int(xfer.ChainID.Int64())
		switch chainID {
		case p.BSCChainID:
			switch xfer.Symbol {
			case "BUSD":
				if xfer.To == p.PurchaseAddr {
					// if buying sups with BUSD
					ctx := context.Background()

					amountTimes100 := xfer.Amount.Mul(xfer.Amount, big.NewInt(1000))
					supUSDPriceTimes100 := p.SUPToUSD.Mul(decimal.New(1000, 0)).BigInt()
					supAmount := amountTimes100.Div(amountTimes100, supUSDPriceTimes100)

					api.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", supAmount.String()).
						Str("BUSD", xfer.Amount.String()).
						Str("Buyer", xfer.From.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, api.Conn, xfer.From.Hex(), api.HostUrl)
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
							err = db.UserCreate(ctx, api.Conn, user)
							if err != nil {
								api.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							api.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}

					resultChan := make(chan *passport.TransactionResult)

					api.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.OnChainUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						api.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					err = db.CreateChainConfirmationEntry(ctx, api.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID.Uint64())
					if err != nil {
						api.transaction <- &passport.NewTransaction{
							To:                   passport.OnChainUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with BUSD %s", xfer.TxID.Hex()),
						}
						api.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
				}
			case "WBNB":
				if xfer.To == p.PurchaseAddr {
					// if buying sups with WBNB
					ctx := context.Background()
					// TODO: probably do a * 1000 here? currently no decimals in conversion but possibly in future?
					supAmount := p.WBNBToSUPS.BigInt()
					supAmount = supAmount.Mul(supAmount, xfer.Amount)

					api.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", supAmount.String()).
						Str("WBNB", xfer.Amount.String()).
						Str("Buyer", xfer.From.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, api.Conn, xfer.From.Hex(), api.HostUrl)
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
							err = db.UserCreate(ctx, api.Conn, user)
							if err != nil {
								api.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							api.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}

					resultChan := make(chan *passport.TransactionResult)

					api.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.OnChainUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on BSC with WBNB %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						api.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					err = db.CreateChainConfirmationEntry(ctx, api.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID.Uint64())
					if err != nil {
						api.transaction <- &passport.NewTransaction{
							To:                   passport.OnChainUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on BSC with WBNB %s", xfer.TxID.Hex()),
						}
						api.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
				}
			case "SUPS":
				if xfer.To == p.PurchaseAddr {
					ctx := context.Background()
					// if deposit sups
					api.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", xfer.Amount.String()).
						Str("User", xfer.From.Hex()).
						Msg("deposit")

					user, err := db.UserByPublicAddress(ctx, api.Conn, xfer.From.Hex(), api.HostUrl)
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
							err = db.UserCreate(ctx, api.Conn, user)
							if err != nil {
								api.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							api.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}

					resultChan := make(chan *passport.TransactionResult)

					api.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.OnChainUserID,
						Amount:               *xfer.Amount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup deposit on BSC %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						api.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					err = db.CreateChainConfirmationEntry(ctx, api.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID.Uint64())
					if err != nil {
						api.transaction <- &passport.NewTransaction{
							To:                   passport.OnChainUserID,
							From:                 user.ID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup deposit on BSC %s", xfer.TxID.Hex()),
						}
						api.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}

				}
				if xfer.To == p.RedemptionAddr {
					// UNTESTED
					ctx := context.Background()
					//busdAmount := d.Div(p.BUSDToSUPS)

					// make sup cost 1000 * bigger to not deal with decimals
					supUSDPriceTimes1000 := p.SUPToUSD.Mul(decimal.New(1000, 0)).BigInt()
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

					api.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", xfer.Amount.String()).
						Str("BUSD", amountOfBUSD.String()).
						Str("User", xfer.From.Hex()).
						Msg("redeem")

					user, err := db.UserByPublicAddress(ctx, api.Conn, xfer.From.Hex(), api.HostUrl)
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
							err = db.UserCreate(ctx, api.Conn, user)
							if err != nil {
								api.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							api.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}

					resultChan := make(chan *passport.TransactionResult)

					api.transaction <- &passport.NewTransaction{
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
						api.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}
					err = db.CreateChainConfirmationEntry(ctx, api.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID.Uint64())
					if err != nil {
						api.transaction <- &passport.NewTransaction{
							To:                   user.ID,
							From:                 passport.XsynTreasuryUserID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup redeem on BSC to BUSD %s", xfer.TxID.Hex()),
						}
						api.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}

				}
				if xfer.From == p.WithdrawAddr {
					ctx := context.Background()
					// UNTESTED
					// if withdrawing sups
					api.Log.Info().
						Str("Chain", "BSC").
						Str("SUPS", xfer.Amount.String()).
						Str("User", xfer.To.Hex()).
						Msg("withdraw")

					user, err := db.UserByPublicAddress(ctx, api.Conn, xfer.From.Hex(), api.HostUrl)
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
							err = db.UserCreate(ctx, api.Conn, user)
							if err != nil {
								api.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							api.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}
					resultChan := make(chan *passport.TransactionResult)

					api.transaction <- &passport.NewTransaction{
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
						api.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}
					err = db.CreateChainConfirmationEntry(ctx, api.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID.Uint64())
					if err != nil {
						api.transaction <- &passport.NewTransaction{
							To:                   user.ID,
							From:                 passport.XsynTreasuryUserID,
							Amount:               *xfer.Amount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - ssup withdraw on BSC to %s", xfer.TxID.Hex()),
						}
						api.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}

				}
			}

		case p.ETHChainID:
			switch xfer.Symbol {
			case "USDC":
				if xfer.To == p.PurchaseAddr {
					// if buying sups with USDC
					ctx := context.Background()
					amountTimes100 := xfer.Amount.Mul(xfer.Amount, big.NewInt(1000))
					supUSDPriceTimes100 := p.SUPToUSD.Mul(decimal.New(1000, 0)).BigInt()
					supAmount := amountTimes100.Div(amountTimes100, supUSDPriceTimes100)

					api.Log.Info().
						Str("Chain", "Ethereum").
						Str("SUPS", supAmount.String()).
						Str("USDC", xfer.Amount.String()).
						Str("Buyer", xfer.From.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, api.Conn, xfer.From.Hex(), api.HostUrl)
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
							err = db.UserCreate(ctx, api.Conn, user)
							if err != nil {
								api.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							api.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}
					resultChan := make(chan *passport.TransactionResult)

					api.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.OnChainUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
					}

					result := <-resultChan
					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						api.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}
					err = db.CreateChainConfirmationEntry(ctx, api.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID.Uint64())
					if err != nil {
						api.transaction <- &passport.NewTransaction{
							To:                   passport.OnChainUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with USDC %s", xfer.TxID.Hex()),
						}
						api.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}

				}
			case "WETH":
				if xfer.To == p.PurchaseAddr {
					// if buying sups with WETH
					ctx := context.Background()
					// TODO: probably do a * 1000 here? currently no decimals in conversion but possibly in future?
					supAmount := p.WETHToSUPS.BigInt()
					supAmount = supAmount.Mul(supAmount, xfer.Amount)

					api.Log.Info().
						Str("Chain", "Ethereum").
						Str("SUPS", supAmount.String()).
						Str("WETH", xfer.Amount.String()).
						Str("Buyer", xfer.From.Hex()).
						Msg("purchase")

					user, err := db.UserByPublicAddress(ctx, api.Conn, xfer.From.Hex(), api.HostUrl)
					if err != nil {
						// if error is no rows, create user!
						if errors.Is(err, pgx.ErrNoRows) {
							user = &passport.User{Username: xfer.From.Hex(), PublicAddress: passport.NewString(xfer.From.Hex())}
							err = db.UserCreate(ctx, api.Conn, user)
							if err != nil {
								api.Log.Err(err).Msg("issue creating new user")
								return
							}
						} else {
							api.Log.Err(err).Msg("issue finding users public address")
							return
						}
					}
					resultChan := make(chan *passport.TransactionResult)

					api.transaction <- &passport.NewTransaction{
						ResultChan:           resultChan,
						To:                   user.ID,
						From:                 passport.OnChainUserID,
						Amount:               *supAmount,
						TransactionReference: passport.TransactionReference(xfer.TxID.Hex()),
						Description:          fmt.Sprintf("sup purchase on Ethereum with WETH %s", xfer.TxID.Hex()),
					}

					result := <-resultChan

					if result.Error != nil {
						return // believe error logs already
					}

					if result.Transaction.Status != passport.TransactionSuccess {
						api.Log.Err(fmt.Errorf("transaction unsuccessful reason: %s", result.Transaction.Reason))
						return
					}

					err = db.CreateChainConfirmationEntry(ctx, api.Conn, xfer.TxID.Hex(), result.Transaction.ID, xfer.Block, xfer.ChainID.Uint64())
					if err != nil {
						api.transaction <- &passport.NewTransaction{
							To:                   passport.OnChainUserID,
							From:                 user.ID,
							Amount:               *supAmount,
							TransactionReference: passport.TransactionReference(fmt.Sprintf("%s %s", xfer.TxID.Hex(), "FAILED TO INSERT CHAIN CONFIRM ENTRY")),
							Description:          fmt.Sprintf("FAILED TO INSERT CHAIN CONFIRM ENTRY - Revert - sup purchase on Ethereum with WETH %s", xfer.TxID.Hex()),
						}
						api.Log.Err(err).Msg("failed to insert chain confirmation entry")
					}
				}
			}
		}
	}
	return fn
}

func (api *API) handleBlock(client *ethclient.Client, chainID uint64) func(header *bridge.Header) {
	fn := func(header *bridge.Header) {
		ctx := context.Background()
		api.Log.Info().Str("Block", header.Number.String())

		// get all transaction confirmations that are not finished
		confirmations, err := db.PendingChainConfirmationsByChainID(ctx, api.Conn, chainID)
		if err != nil {
			api.Log.Err(err).Msg("issue getting pending chain confirmations in block listener")
			return
		}

		// loop over the confirmations, check if block difference is >= 6, then validate and update
		for _, conf := range confirmations {
			confirmedBlocks, err := bridge.TransactionConfirmations(ctx, client, common.HexToHash(conf.Tx))
			if err != nil {
				api.Log.Err(err).Msg("issue confirming transaction")
				return
			}

			// if confirmed blocks greater than 6, finalize it
			if confirmedBlocks >= 6 {

				_, err = db.ConfirmChainConfirmation(ctx, api.Conn, conf.Tx)
				if err != nil {
					api.Log.Err(err).Msg("issue setting as confirmed")
					return
				}

				api.Log.Info().Msg("chain transaction finalized")
			}

			// TODO: setup subscription
			//api.MessageBus.Sub("qweqw", conf)
		}
	}
	return fn
}

func (api *API) RunBridgeListener(p *passport.BridgeParams) {
	infuraAddr := p.EthNodeAddr
	ethclient, err := client.Dial(infuraAddr)
	if err != nil {
		api.Log.Fatal().Err(err).Msg("failed to connected to eth node")
	}

	bscAddr := p.BscNodeAddr
	bscclient, err := client.Dial(bscAddr)
	if err != nil {
		api.Log.Fatal().Err(err).Msg("failed to connected to bsc node")
	}

	usdcListener := bridge.NewERC20Listener(p.UsdcAddr, "USDC", 18, ethclient, api.handleTransfer(p))
	wethListener := bridge.NewERC20Listener(p.WethAddr, "WETH", 18, ethclient, api.handleTransfer(p))
	busdListener := bridge.NewERC20Listener(p.BusdAddr, "BUSD", 18, bscclient, api.handleTransfer(p))
	wbnbListener := bridge.NewERC20Listener(p.WbnbAddr, "WBNB", 18, bscclient, api.handleTransfer(p))

	blockEth := bridge.NewHeadListener(ethclient, api.handleBlock(ethclient, uint64(p.ETHChainID)))
	blockBSC := bridge.NewHeadListener(bscclient, api.handleBlock(bscclient, uint64(p.BSCChainID)))

	// start block listeners
	go func() {
		for {
			api.Log.Info().Str("chain", "ETH").Msg("start header listener")
			err := blockEth.Listen()
			if err != nil {
				api.Log.Err(err).Msg("error listening to eth blocks")
			}
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for {
			api.Log.Info().Str("chain", "BSC").Msg("start header listener")
			err := blockBSC.Listen()
			if err != nil {
				api.Log.Err(err).Msg("error listening to bsc blocks")
			}
			time.Sleep(5 * time.Second)
		}
	}()

	// start wallet and contract listeners
	go func() {
		for {
			api.Log.Info().Str("sym", "WBNB").Msg("start listener")
			err := wbnbListener.Listen()
			if err != nil {
				api.Log.Err(err).Msg("error listening to wbnb")
			}
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for {
			api.Log.Info().Str("sym", "WETH").Msg("start listener")
			err := wethListener.Listen()
			if err != nil {
				api.Log.Err(err).Msg("error listening to weth")
			}
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for {
			api.Log.Info().Str("sym", "BUSD").Msg("start listener")
			err := busdListener.Listen()
			if err != nil {
				api.Log.Err(err).Msg("error listening to busd")
			}
			time.Sleep(5 * time.Second)
		}
	}()
	go func() {
		for {
			api.Log.Info().Str("sym", "USDC").Msg("start listener")
			err := usdcListener.Listen()
			if err != nil {
				api.Log.Err(err).Msg("error listening to usdc")
			}
			time.Sleep(5 * time.Second)
		}
	}()

	select {}
}
