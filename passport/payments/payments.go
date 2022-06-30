package payments

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type AvantDataResp struct {
	Time int    `json:"time"`
	USD  string `json:"usd"`
}

type PriceExchangeRates struct {
	SUPtoUSD   decimal.Decimal `json:"sup_to_usd"`
	ETHtoUSD   decimal.Decimal `json:"eth_to_usd"`
	BNBtoUSD   decimal.Decimal `json:"bnb_to_usd"`
	EnableSale bool            `json:"enable_sale"`
}
type UserCacheMap interface {
	Transact(nt *types.NewTransaction) (string, error)
}

const SUPDecimals = 18

func CreateOrGetUser(userAddr common.Address) (*types.User, error) {
	var user *types.User
	var err error
	user, err = users.PublicAddress(userAddr)
	if errors.Is(err, sql.ErrNoRows) {
		username := helpers.TrimUsername(userAddr.Hex())
		runes := []rune(username)
		username = string(runes[0:10])
		user, err = users.UserCreator("", "", username, "", "", "", "", "", "", "", userAddr, "")
		if err != nil {
			return nil, err
		}
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return user, nil
}

func ProcessValues(sups string, inputValue string, inputTokenDecimals int) (decimal.Decimal, decimal.Decimal, error) {
	outputAmt, err := decimal.NewFromString(sups)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	bigOutputAmt := outputAmt.Shift(1 * types.SUPSDecimals)
	inputAmt, err := decimal.NewFromString(inputValue)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return inputAmt, bigOutputAmt, nil
}

func StoreRecord(ctx context.Context, fromUserID types.UserID, toUserID types.UserID, ucm UserCacheMap, record *PurchaseRecord, passportExchangeRatesEnabled bool) error {

	tokenValue, supsValue, err := ProcessValues(record.Sups, record.ValueInt, record.ValueDecimals)

	if err != nil {
		return err
	}

	if passportExchangeRatesEnabled {
		// From Record
		usdRate, err := decimal.NewFromString(record.UsdRate)
		if err != nil {
			return err
		}
		supsAmt, err := decimal.NewFromString(record.Sups)
		if err != nil {
			return err
		}

		supToUsd := tokenValue.Shift(-1 * int32(record.ValueDecimals)).Mul(usdRate).Div(supsAmt)

		supPrice, err := fetchPrice("sups", passportExchangeRatesEnabled)
		if err != nil {
			return err
		}
		rateDifference := (supToUsd).Div(supPrice)

		record.Sups = supsAmt.Mul(rateDifference).String()

		tokenValue, supsValue, err = ProcessValues(record.Sups, record.ValueInt, record.ValueDecimals)
		if err != nil {
			return err
		}

	}

	msg := fmt.Sprintf("purchased %s SUPS for %s [%s]", supsValue.Shift(-1*types.SUPSDecimals).StringFixed(4), tokenValue.Shift(-1*int32(record.ValueDecimals)).StringFixed(4), strings.ToUpper(record.Symbol))
	trans := &types.NewTransaction{
		To:                   toUserID,
		From:                 fromUserID,
		Amount:               supsValue,
		TransactionReference: types.TransactionReference(record.TxHash),
		Description:          msg,
		Group:                types.TransactionGroupStore,
	}

	_, err = ucm.Transact(trans)
	if err != nil {
		return fmt.Errorf("create tx entry for tx %s: %w", record.TxHash, err)
	}
	return nil
}

func CheckIsCurrentBlockAfter() bool {
	latestBNBBlock := db.GetIntWithDefault(db.KeyLatestBNBBlock, 0)
	latestBUSDBlock := db.GetIntWithDefault(db.KeyLatestBUSDBlock, 0)
	afterBSCBlock := db.GetIntWithDefault(db.KeyEnablePassportExchangeRateAfterBSCBlock, 0)

	latestETHBlock := db.GetIntWithDefault(db.KeyLatestETHBlock, 0)
	latestUSDCBlock := db.GetIntWithDefault(db.KeyLatestUSDCBlock, 0)
	afterETHBlock := db.GetIntWithDefault(db.KeyEnablePassportExchangeRateAfterETHBlock, 0)

	if afterBSCBlock == 0 && afterETHBlock == 0 {
		return false
	}

	return latestBNBBlock > afterBSCBlock &&
		latestBUSDBlock > afterBSCBlock && latestETHBlock > afterETHBlock &&
		latestUSDCBlock > afterETHBlock
}

func BUSD(isTestnet bool) ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestBUSDBlock)
	records, latestBlock, err := getPurchaseRecords(BUSDPurchasePath, currentBlock, isTestnet)
	if err != nil {
		return nil, err
	}
	// Avant data testnet BUSD doesnt work
	if latestBlock == 0 {
		db.PutInt(db.KeyLatestBUSDBlock, db.GetIntWithDefault(db.KeyLatestBNBBlock, 0))
		return records, nil
	}
	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestBUSDBlock, latestBlock)
	}

	return records, nil
}

func USDC(isTestnet bool) ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestUSDCBlock)
	records, latestBlock, err := getPurchaseRecords(USDCPurchasePath, currentBlock, isTestnet)
	if err != nil {
		return nil, err
	}

	// Avant data testnet USDC doesnt work
	if latestBlock == 0 {
		db.PutInt(db.KeyLatestUSDCBlock, db.GetIntWithDefault(db.KeyLatestETHBlock, 0))
		return records, nil
	}

	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestUSDCBlock, latestBlock)
	}
	return records, nil

}

func ETH(isTestnet bool) ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestETHBlock)
	records, latestBlock, err := getPurchaseRecords(ETHPurchasePath, currentBlock, isTestnet)
	if err != nil {
		return nil, err
	}

	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestETHBlock, latestBlock)
	}

	return records, nil
}

func BNB(isTestnet bool) ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestBNBBlock)
	records, latestBlock, err := getPurchaseRecords(BNBPurchasePath, currentBlock, isTestnet)
	if err != nil {
		return nil, err
	}
	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestBNBBlock, latestBlock)
	}

	return records, nil
}
func fetchPrice(symbol string, passportExchangeRateEnabled bool) (decimal.Decimal, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf(`%s/api/%s_price`, baseURL, symbol), nil)

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
	result := &AvantDataResp{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return decimal.Zero, err
	}

	dec, err := decimal.NewFromString(result.USD)
	if err != nil {
		return decimal.Zero, err
	}

	if symbol == "sups" {
		defaultFloorPrice, err := decimal.NewFromString("0.02")
		if err != nil {
			return decimal.Zero, err
		}
		defaultMarketMultiplier, err := decimal.NewFromString("1.1")
		if err != nil {
			return decimal.Zero, err
		}

		priceFloor := db.GetDecimalWithDefault(db.KeyPurchaseSupsFloorPrice, defaultFloorPrice)
		marketPriceMultiplier := db.GetDecimalWithDefault(db.KeyPurchaseSupsMarketPriceMultiplier, defaultMarketMultiplier)

		// Increase market price
		dec = dec.Mul(marketPriceMultiplier)
		// Check if less than floor price
		if dec.LessThan(priceFloor) {
			dec = priceFloor
		}

		if dec.LessThanOrEqual(decimal.Zero) {
			return decimal.Zero, fmt.Errorf("0 price returned")
		}

		if !passportExchangeRateEnabled {
			dec, err = decimal.NewFromString("0.12")
			if err != nil {
				return decimal.Zero, err
			}
		}
	}
	return dec, nil
}
func catchPriceFetchError(symbol string, dbKey db.KVKey, passportExchangeRatesEnabled bool) (decimal.Decimal, error) {
	passlog.L.Warn().Msg(fmt.Sprintf("could not fetch %s price", symbol))
	dec, err := decimal.NewFromString(db.GetStr(dbKey))

	if !passportExchangeRatesEnabled {
		dec, err = decimal.NewFromString("0.12")
		if err != nil {
			return decimal.Zero, err
		}
	}
	if err != nil {
		return decimal.Zero, err
	}
	if dec.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("0 price returned")
	}
	return dec, nil
}

func FetchExchangeRates(passportExchangeRatesEnabled bool) (*PriceExchangeRates, error) {
	enableSale := db.GetBoolWithDefault(db.KeyEnableSyncSale, true)

	supsPrice, err := fetchPrice("sups", passportExchangeRatesEnabled)
	if err != nil {
		supsPrice, err = catchPriceFetchError("sups", db.KeySupsToUSD, passportExchangeRatesEnabled)
		if err != nil {
			return nil, err
		}
	}
	ethPrice, err := fetchPrice("eth", passportExchangeRatesEnabled)
	if err != nil {
		ethPrice, err = catchPriceFetchError("eth", db.KeyEthToUSD, passportExchangeRatesEnabled)
		if err != nil {
			return nil, err
		}
	}
	bnbPrice, err := fetchPrice("bnb", passportExchangeRatesEnabled)
	if err != nil {
		if err != nil {
			bnbPrice, err = catchPriceFetchError("bnb", db.KeyBNBToUSD, passportExchangeRatesEnabled)
			return nil, err
		}
	}

	priceExchangeRates := &PriceExchangeRates{SUPtoUSD: supsPrice, ETHtoUSD: ethPrice, BNBtoUSD: bnbPrice, EnableSale: enableSale}

	db.PutDecimal(db.KeySupsToUSD, priceExchangeRates.SUPtoUSD)
	db.PutDecimal(db.KeyEthToUSD, priceExchangeRates.ETHtoUSD)
	db.PutDecimal(db.KeyBNBToUSD, priceExchangeRates.BNBtoUSD)
	return priceExchangeRates, nil
}
