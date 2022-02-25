package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"passport"
	"passport/db"
	"sync"
	"time"

	"github.com/ninja-software/terror/v2"

	"github.com/jackc/pgx/v4"
	"github.com/jpillora/backoff"
	"github.com/rs/zerolog"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

const ETHSymbol = "ETH"
const BNBSymbol = "BNB"
const BUSDSymbol = "BUSD"
const USDCSymbol = "USDC"

const ETHDecimals = 18
const BNBDecimals = 18
const SUPSDecimals = 18

type ChainClients struct {
	isTestnetBlockchain bool
	runBlockchainBridge bool
	SUPS                *bridge.SUPS
	EthClient           *ethclient.Client
	BscClient           *ethclient.Client
	Params              *passport.BridgeParams
	API                 *API
	Log                 *zerolog.Logger

	updatePriceFuncMu sync.Mutex
	updatePriceFunc   func(symbol string, amount decimal.Decimal)
}

type Prices struct {
	ETH float64
	BTC float64
}

type BNBPriceResp struct {
	Binancecoin struct {
		Usd float64 `json:"usd"`
	} `json:"binancecoin"`
}

type ETHPriceResp struct {
	Ethereum struct {
		Usd float64 `json:"usd"`
	} `json:"ethereum"`
}
type CoinbaseResp struct {
	Data struct {
		Currency string `json:"currency"`
		Rates    struct {
			Usd string `json:"USD"`
		} `json:"rates"`
	} `json:"data"`
}

func fetchPrice(symbol string) (decimal.Decimal, error) {
	// use ETH or BNB for symbol
	req, err := http.NewRequest("GET", fmt.Sprintf(`https://api.coinbase.com/v2/exchange-rates?currency=%s`, symbol), nil)
	if err != nil {
		return decimal.Zero, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return decimal.Zero, fmt.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	result := &CoinbaseResp{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return decimal.Zero, err
	}

	dec, err := decimal.NewFromString(result.Data.Rates.Usd)
	if err != nil {
		return decimal.Zero, err
	}
	if dec.Equal(decimal.Zero) {
		return decimal.Zero, errors.New("0 price returned")
	}
	return dec, nil
}

func FetchETHPrice() (decimal.Decimal, error) {
	return fetchPrice("ETH")
}

func FetchBNBPrice() (decimal.Decimal, error) {
	return fetchPrice("BNB")
}

func NewChainClients(log *zerolog.Logger, api *API, p *passport.BridgeParams, isTestnetBlockchain bool, runBlockchainBridge bool, enablePurchaseSubscription bool) *ChainClients {
	cc := &ChainClients{
		Params:              p,
		API:                 api,
		Log:                 log,
		updatePriceFuncMu:   sync.Mutex{},
		isTestnetBlockchain: isTestnetBlockchain,
		runBlockchainBridge: runBlockchainBridge,
	}
	ctx := context.Background()

	cc.updatePriceFunc = func(symbol string, amount decimal.Decimal) {
		if !enablePurchaseSubscription {
			return
		}
		switch symbol {
		//case "SUPS":
		//	cc.API.State.SUPtoUSD = amount
		case ETHSymbol:
			cc.API.State.ETHtoUSD = amount
		case BNBSymbol:
			cc.API.State.BNBtoUSD = amount
		}

		_, err := db.UpdateExchangeRates(ctx, isTestnetBlockchain, cc.API.Conn, cc.API.State)
		if err != nil {
			api.Log.Err(err).Msg("failed to update exchange rates")
		}
		cc.Log.Debug().
			Str(symbol, amount.String()).
			Msg("update rate")

		go api.MessageBus.Send(ctx, messagebus.BusKey(HubKeySUPSExchangeRates), cc.API.State)
	}

	if runBlockchainBridge {
		go cc.runChainListeners()
		go cc.runGoETHPriceListener(ctx)
		go cc.runGoBNBPriceListener(ctx)
	}

	return cc
}

func (cc *ChainClients) runChainListeners() *ChainClients {
	cc.Log.Debug().Bool("is_testnet", cc.isTestnetBlockchain).Str("purchase_addr", cc.Params.PurchaseAddr.Hex()).Str("deposit_addr", cc.Params.DepositAddr.Hex()).Str("busd_addr", cc.Params.BusdAddr.Hex()).Str("usdc_addr", cc.Params.UsdcAddr.Hex()).Msg("addresses")
	ctx := context.Background()

	go cc.runETHBridgeListener(ctx)

	return cc
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
					asset, err := db.AssetGetFromContractAndID(ctx, cc.API.Conn, event.Contract.Hex(), event.TokenID.Uint64())
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

					//func AssetTransfer(ctx context.Context, conn Conn, tokenID int, oldUserID, newUserID passport.UserID, txHash string) error {
					err = db.AssetTransfer(ctx, cc.API.Conn, asset.ExternalTokenID, passport.OnChainUserID, user.ID, event.TxID.Hex())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("failed to transfer asset")
						return
					}

					// TODO: remove this and update asset transfer to return updated asset instead
					asset, err = db.AssetGetFromContractAndID(ctx, cc.API.Conn, event.Contract.Hex(), event.TokenID.Uint64())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("unable to find asset to transfer")
						return
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.ExternalTokenID)), asset)
				},
				func(event *bridge.NFTUnstakeEvent) {
					cc.Log.Info().Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msg("nft unstake")
					// get asset
					asset, err := db.AssetGetFromContractAndID(ctx, cc.API.Conn, event.Contract.Hex(), event.TokenID.Uint64())
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
					err = db.AssetTransfer(ctx, cc.API.Conn, asset.ExternalTokenID, *asset.UserID, passport.OnChainUserID, event.TxID.Hex())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("failed to transfer asset")
						return
					}

					// TODO: remove this and update asset transfer to return updated asset instead (optimization)
					asset, err = db.AssetGetFromContractAndID(ctx, cc.API.Conn, event.Contract.Hex(), event.TokenID.Uint64())
					if err != nil {
						cc.Log.Err(err).Str("Owner", event.Owner.Hex()).Uint64("Token", event.TokenID.Uint64()).Str("tx", event.TxID.Hex()).Msgf("unable to find asset to transfer")
						return
					}
					go cc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, asset.ExternalTokenID)), asset)
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

			err = <-errChan
			if err != nil {
				cc.Log.Err(err).Msg("error with eth client, attempting to reconnect...")
				cancel()
				time.Sleep(b.Duration())
				continue ethClientLoop
			}
		}
	}()
}

func (cc *ChainClients) runGoETHPriceListener(ctx context.Context) {

	// ETH price listener
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

				result, err := FetchETHPrice()
				if err != nil {
					cc.Log.Err(err).Msg("failed to get ETH price")
					time.Sleep(exchangeRateBackoff.Duration())
					continue
				}
				exchangeRateBackoff.Reset()

				cc.updatePriceFuncMu.Lock()
				cc.updatePriceFunc(ETHSymbol, result)
				cc.updatePriceFuncMu.Unlock()

				time.Sleep(10 * time.Second)
			}
		}
	}()
}

func (cc *ChainClients) runGoBNBPriceListener(ctx context.Context) {
	// BNB price listener
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

				result, err := FetchBNBPrice()
				if err != nil {
					cc.Log.Err(err).Msg("failed to get BNB price")
					time.Sleep(exchangeRateBackoff.Duration())
					continue
				}
				exchangeRateBackoff.Reset()

				cc.updatePriceFuncMu.Lock()
				cc.updatePriceFunc(BNBSymbol, result)
				cc.updatePriceFuncMu.Unlock()

				time.Sleep(10 * time.Second)
			}
		}
	}()
}

func (cc *ChainClients) handleNFTTransfer(ctx context.Context) func(xfer *bridge.NFTTransferEvent) {
	fn := func(ev *bridge.NFTTransferEvent) {
		func() {
			if ev.From.Hex() == cc.Params.EthNftStakingAddr.Hex() || ev.To.Hex() == cc.Params.EthNftStakingAddr.Hex() {
				return
			}

			asset, err := db.AssetGetFromContractAndID(ctx, cc.API.Conn, ev.Contract.Hex(), ev.TokenID.Uint64())
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
				err := db.XsynAssetLock(ctx, cc.API.Conn, asset.Hash, passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO LOCK ASSET token id: %s", ev.TokenID.String())
					return
				}

				err = db.XsynAssetFreeze(ctx, cc.API.Conn, asset.Hash, passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO FREEZE ASSET token id: %s", ev.TokenID.String())
					return
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
				err := db.XsynAssetLock(ctx, cc.API.Conn, asset.Hash, passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO LOCK ASSET token id: %s", ev.TokenID.String())
				}
				err = db.XsynAssetFreeze(ctx, cc.API.Conn, asset.Hash, passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO FREEZE ASSET token id: %s", ev.TokenID.String())
				}
				return
			}

			err = db.AssetTransfer(ctx, cc.API.Conn, ev.TokenID.Uint64(), user.ID, passport.OnChainUserID, ev.TxID.Hex())
			if err != nil {
				cc.Log.Err(err).Msgf("issue transferring asset token id: %s, locking and freezing it", ev.TokenID.String())
				// if issue transferring asset, LOCK IT!
				err := db.XsynAssetLock(ctx, cc.API.Conn, asset.Hash, passport.XsynTreasuryUserID)
				if err != nil {
					cc.Log.Err(err).Msgf("FAILED TO LOCK ASSET token id: %s", ev.TokenID.String())
				}
				err = db.XsynAssetFreeze(ctx, cc.API.Conn, asset.Hash, passport.XsynTreasuryUserID)
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
		asset, err := db.AssetGetFromContractAndID(ctx, cc.API.Conn, ev.Contract.Hex(), ev.TokenID.Uint64())
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
