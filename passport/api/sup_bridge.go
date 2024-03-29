package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/jpillora/backoff"
	"github.com/rs/zerolog"

	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
	"github.com/sasha-s/go-deadlock"
)

type ChainClients struct {
	isTestnetBlockchain bool
	runBlockchainBridge bool
	SUPS                *bridge.SUPS
	EthClient           *ethclient.Client
	BscClient           *ethclient.Client
	Params              *types.Web3Params
	API                 *API
	Log                 *zerolog.Logger

	updatePriceFuncMu deadlock.Mutex
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
		return decimal.Zero, fmt.Errorf("0 price returned")
	}
	return dec, nil
}

func FetchETHPrice() (decimal.Decimal, error) {
	return fetchPrice("ETH")
}

func FetchBNBPrice() (decimal.Decimal, error) {
	return fetchPrice("BNB")
}

func NewChainClients(log *zerolog.Logger, api *API, p *types.Web3Params, isTestnetBlockchain bool, runBlockchainBridge bool, enablePurchaseSubscription bool) *ChainClients {
	cc := &ChainClients{
		Params:              p,
		API:                 api,
		Log:                 log,
		updatePriceFuncMu:   deadlock.Mutex{},
		isTestnetBlockchain: isTestnetBlockchain,
		runBlockchainBridge: runBlockchainBridge,
	}
	ctx := context.Background()

	cc.updatePriceFunc = func(symbol string, amount decimal.Decimal) {
		if !enablePurchaseSubscription {
			return
		}
		if cc.API == nil {
			passlog.L.Warn().Msg("API pointer is nil, likely a race condition in spin up")
			return
		}
		if cc.API.State == nil {
			passlog.L.Warn().Msg("API.State pointer is nil, likely a race condition in spin up")
			return
		}
		switch symbol {
		case types.SUPSSymbol:
			cc.API.State.SUPtoUSD = amount
		case types.ETHSymbol:
			cc.API.State.ETHtoUSD = amount
		case types.BNBSymbol:
			cc.API.State.BNBtoUSD = amount
		}

		_, err := db.UpdateExchangeRates(isTestnetBlockchain, cc.API.State)
		if err != nil {
			api.Log.Err(err).Msg("failed to update exchange rates")
		}
		cc.Log.Info().
			Str(symbol, amount.String()).
			Msg("update rate")

	}

	if runBlockchainBridge {
		go cc.runGoETHPriceListener(ctx)
		go cc.runGoBNBPriceListener(ctx)
	}

	return cc
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
				cc.updatePriceFunc(types.ETHSymbol, result)
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
				cc.updatePriceFunc(types.BNBSymbol, result)
				cc.updatePriceFuncMu.Unlock()

				time.Sleep(10 * time.Second)
			}
		}
	}()
}

func pingFunc(ctx context.Context, client *ethclient.Client) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_, err := client.BlockNumber(ctxTimeout)
	if err != nil {
		return err
	}
	return nil
}
