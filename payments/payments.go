package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"passport"
	"passport/api"
	"passport/db"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

type Record struct {
	Symbol      string    `json:"symbol"`
	Usd         float64   `json:"usd"`
	Value       string    `json:"value"`
	Sups        string    `json:"sups"`
	Bucket      time.Time `json:"bucket"`
	TxHash      string    `json:"tx_hash"`
	Time        float64   `json:"time"`
	ToAddress   string    `json:"to_address"`
	FromAddress string    `json:"from_address"`
	JSON        struct {
		To                string `json:"to"`
		Gas               string `json:"gas"`
		From              string `json:"from"`
		Hash              string `json:"hash"`
		Input             string `json:"input"`
		Nonce             string `json:"nonce"`
		Value             string `json:"value"`
		GasUsed           string `json:"gasUsed"`
		GasPrice          string `json:"gasPrice"`
		BlockHash         string `json:"blockHash"`
		TimeStamp         string `json:"timeStamp"`
		TokenName         string `json:"tokenName"`
		BlockNumber       string `json:"blockNumber"`
		TokenSymbol       string `json:"tokenSymbol"`
		TokenDecimal      string `json:"tokenDecimal"`
		Confirmations     string `json:"confirmations"`
		ContractAddress   string `json:"contractAddress"`
		TransactionIndex  string `json:"transactionIndex"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"`
	} `json:"json"`
}
type Symbol string

const BNBSymbol Symbol = "bnb_txs"
const BUSDSymbol Symbol = "busd_txs"
const ETHSymbol Symbol = "eth_txs"
const USDCSymbol Symbol = "usdc_txs"

const SUPDecimals = 18

func CreateOrGetUser(ctx context.Context, conn *pgxpool.Pool, userAddr string) (*passport.User, error) {
	from := common.HexToAddress(userAddr).Hex()
	var user *passport.User
	var err error
	user, err = db.UserByPublicAddress(ctx, conn, from)
	if errors.Is(err, pgx.ErrNoRows) {
		user = &passport.User{}
		user.Username = from
		user.PublicAddress = null.NewString(from, true)
		user.RoleID = passport.UserRoleMemberID
		err := db.UserCreate(ctx, conn, user)
		if err != nil {
			return nil, terror.Error(err)
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	return user, nil

}

func ProcessValues(sups string, inputValue string, inputDecimalStr string) (decimal.Decimal, decimal.Decimal, int, error) {
	outputAmt, err := decimal.NewFromString(sups)
	if err != nil {
		return decimal.Zero, decimal.Zero, 0, err
	}
	bigOutputAmt := outputAmt.Shift(1 * api.SUPSDecimals)
	inputAmt, err := decimal.NewFromString(inputValue)
	if err != nil {
		return decimal.Zero, decimal.Zero, 0, err
	}
	// Default decimal for natives
	tokenDecimal := 18
	if inputDecimalStr != "" {
		tokenDecimal, err = strconv.Atoi(inputDecimalStr)
		if err != nil {
			return decimal.Zero, decimal.Zero, 0, fmt.Errorf("tokendecimal %s: %w", inputDecimalStr, err)
		}
	}
	bigInputAmt := inputAmt.Shift(1 * int32(tokenDecimal))
	return bigInputAmt, bigOutputAmt, tokenDecimal, nil
}

func StoreRecord(ctx context.Context, toUser *passport.User, ucm *api.UserCacheMap, record *Record) error {
	input, output, tokenDecimals, err := ProcessValues(record.Sups, record.Value, record.JSON.TokenDecimal)
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("purchased %s SUPS for %s [%s]", output.Shift(-1*api.SUPSDecimals).StringFixed(4), input.Shift(-1*int32(tokenDecimals)).StringFixed(4), strings.ToUpper(record.Symbol))

	trans := &passport.NewTransaction{
		To:                   toUser.ID,
		From:                 passport.XsynSaleUserID,
		Amount:               *output.BigInt(),
		TransactionReference: passport.TransactionReference(record.TxHash),
		Description:          msg,
		Safe:                 true,
	}

	_, _, _, err = ucm.Process(trans)
	if err != nil {
		return fmt.Errorf("create tx entry for tx %s: %w", record.TxHash, err)
	}
	return nil
}

func get(sym Symbol) ([]*Record, error) {
	resp, err := http.Get(fmt.Sprintf("http://139.180.182.245:3001/api/%s", sym))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := []*Record{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func enoughConfirmations(requiredConfirmations int, receivedConfirms string) bool {
	confirms, err := decimal.NewFromString(receivedConfirms)
	if err != nil {
		fmt.Println("process confirms:", err)
		return false
	}

	if confirms.LessThan(decimal.NewFromInt(int64(requiredConfirmations))) {
		return false
	}

	return true
}

func BNB(requiredConfirmations int) ([]*Record, error) {
	records, err := get(BNBSymbol)
	if err != nil {
		return nil, err
	}
	results := []*Record{}
	for _, record := range records {
		if !enoughConfirmations(requiredConfirmations, record.JSON.Confirmations) {
			continue
		}
		results = append(results, record)
	}
	return results, nil
}

func BUSD(requiredConfirmations int) ([]*Record, error) {
	contractAddr := "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
	records, err := get(BUSDSymbol)
	if err != nil {
		return nil, err
	}

	results := []*Record{}
	for _, record := range records {
		if !enoughConfirmations(requiredConfirmations, record.JSON.Confirmations) {
			continue
		}
		if common.HexToAddress(record.JSON.ContractAddress).Hex() != common.HexToAddress(contractAddr).Hex() {
			continue
		}
		results = append(results, record)
	}
	return results, nil
}

func USDC(requiredConfirmations int) ([]*Record, error) {
	contractAddr := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"

	records, err := get(USDCSymbol)
	if err != nil {
		return nil, err
	}

	results := []*Record{}
	for _, record := range records {
		if !enoughConfirmations(requiredConfirmations, record.JSON.Confirmations) {
			continue
		}
		if common.HexToAddress(record.JSON.ContractAddress).Hex() != common.HexToAddress(contractAddr).Hex() {
			continue
		}
		results = append(results, record)
	}
	return results, nil
}

func ETH(requiredConfirmations int) ([]*Record, error) {

	records, err := get(ETHSymbol)
	if err != nil {
		return nil, err
	}
	results := []*Record{}
	for _, record := range records {
		if !enoughConfirmations(requiredConfirmations, record.JSON.Confirmations) {
			continue
		}
		results = append(results, record)
	}
	return results, nil
}
