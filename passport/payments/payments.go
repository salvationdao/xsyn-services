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
	Transact(nt *types.NewTransaction) (decimal.Decimal, decimal.Decimal, string, error)
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

func StoreRecord(ctx context.Context, fromUserID types.UserID, toUserID types.UserID, ucm UserCacheMap, record *PurchaseRecord) error {
	input, output, err := ProcessValues(record.Sups, record.ValueInt, record.ValueDecimals)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("purchased %s SUPS for %s [%s]", output.Shift(-1*types.SUPSDecimals).StringFixed(4), input.Shift(-1*int32(record.ValueDecimals)).StringFixed(4), strings.ToUpper(record.Symbol))
	trans := &types.NewTransaction{
		To:                   toUserID,
		From:                 fromUserID,
		Amount:               output,
		TransactionReference: types.TransactionReference(record.TxHash),
		Description:          msg,
		Group:                types.TransactionGroupStore,
	}

	_, _, _, err = ucm.Transact(trans)
	if err != nil {
		return fmt.Errorf("create tx entry for tx %s: %w", record.TxHash, err)
	}
	return nil
}

func BUSD() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestBUSDBlock)
	records, latestBlock, err := getPurchaseRecords(BUSDPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}
	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestBUSDBlock, latestBlock)
	}
	return records, nil
}

func USDC() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestUSDCBlock)
	records, latestBlock, err := getPurchaseRecords(USDCPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}

	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestUSDCBlock, latestBlock)
	}

	return records, nil
}

func ETH() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestETHBlock)
	records, latestBlock, err := getPurchaseRecords(ETHPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}

	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestETHBlock, latestBlock)
	}
	return records, nil
}

func BNB() ([]*PurchaseRecord, error) {
	currentBlock := db.GetInt(db.KeyLatestBNBBlock)
	records, latestBlock, err := getPurchaseRecords(BNBPurchasePath, currentBlock, false)
	if err != nil {
		return nil, err
	}
	if latestBlock > currentBlock {
		db.PutInt(db.KeyLatestBNBBlock, latestBlock)
	}

	return records, nil
}
func fetchPrice(symbol string) (decimal.Decimal, error) {
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
		priceFloor := db.GetDecimal(db.KeyPurchaseSupsFloorPrice)
		marketPriceMultiplier := db.GetDecimal(db.KeyPurchaseSupsMarketPriceMultiplier)
		// Increase market price
		dec = dec.Mul(marketPriceMultiplier)
		// Check if less than floor price
		if dec.LessThan(priceFloor) {
			dec = priceFloor
		}
	}

	if dec.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("0 price returned")
	}
	return dec, nil
}
func catchPriceFetchError(symbol string, dbKey db.KVKey) (decimal.Decimal, error) {
	passlog.L.Warn().Msg(fmt.Sprintf("could not fetch %s price", symbol))
	dec, err := decimal.NewFromString(db.GetStr(dbKey))
	if err != nil {
		return decimal.Zero, err
	}
	if dec.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("0 price returned")
	}
	return dec, nil
}

func FetchExchangeRates() (*PriceExchangeRates, error) {
	supsPrice, err := fetchPrice("sups")
	if err != nil {
		supsPrice, err = catchPriceFetchError("sups", db.KeySupsToUSD)
		if err != nil {
			return nil, err
		}
	}
	ethPrice, err := fetchPrice("eth")
	if err != nil {
		ethPrice, err = catchPriceFetchError("eth", db.KeyEthToUSD)
		if err != nil {
			return nil, err
		}
	}
	bnbPrice, err := fetchPrice("bnb")
	if err != nil {
		bnbPrice, err = catchPriceFetchError("bnb", db.KeyBNBToUSD)
		if err != nil {
			return nil, err
		}
	}
	priceExchangeRates := &PriceExchangeRates{SUPtoUSD: supsPrice, ETHtoUSD: ethPrice, BNBtoUSD: bnbPrice, EnableSale: db.GetBool(db.KeyEnableSyncSale)}

	db.PutDecimal(db.KeySupsToUSD, priceExchangeRates.SUPtoUSD)
	db.PutDecimal(db.KeyEthToUSD, priceExchangeRates.ETHtoUSD)
	db.PutDecimal(db.KeyBNBToUSD, priceExchangeRates.BNBtoUSD)
	return priceExchangeRates, nil
}
