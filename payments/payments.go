package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"passport"
	"passport/db"
	"passport/passlog"
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

type UserCacheMap interface {
	Process(nt *passport.NewTransaction) (decimal.Decimal, decimal.Decimal, string, error)
}

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

func latestBlockFromRecords(currentLatestBlock int, records []*Record) int {
	latestBlock := currentLatestBlock
	for _, record := range records {
		recordBlockNumber, err := strconv.Atoi(record.JSON.BlockNumber)
		if err != nil {
			passlog.L.Err(err).Msg("could not parse blocknumber")
			return 0
		}
		if recordBlockNumber > latestBlock {
			latestBlock = recordBlockNumber
		}
	}
	return latestBlock
}

func CreateOrGetUser(ctx context.Context, conn *pgxpool.Pool, userAddr common.Address) (*passport.User, error) {
	var user *passport.User
	var err error
	user, err = db.UserByPublicAddress(ctx, conn, userAddr)
	if errors.Is(err, pgx.ErrNoRows) {
		user = &passport.User{}
		user.Username = userAddr.Hex()
		user.PublicAddress = null.NewString(userAddr.Hex(), true)
		user.RoleID = passport.UserRoleMemberID
		err := db.UserCreate(ctx, conn, user)
		if err != nil {
			return nil, terror.Error(err)
		}
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, terror.Error(err)
	}
	return user, nil
}

func ProcessValues(sups string, inputValue string, inputDecimalStr string) (decimal.Decimal, decimal.Decimal, int, error) {
	outputAmt, err := decimal.NewFromString(sups)
	if err != nil {
		return decimal.Zero, decimal.Zero, 0, err
	}
	bigOutputAmt := outputAmt.Shift(1 * passport.SUPSDecimals)
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

func StoreRecord(ctx context.Context, fromUserID passport.UserID, toUserID passport.UserID, ucm UserCacheMap, record *Record, isPurchase bool) error {
	input, output, tokenDecimals, err := ProcessValues(record.Sups, record.Value, record.JSON.TokenDecimal)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("purchased %s SUPS for %s [%s]", output.Shift(-1*passport.SUPSDecimals).StringFixed(4), input.Shift(-1*int32(tokenDecimals)).StringFixed(4), strings.ToUpper(record.Symbol))
	if !isPurchase {
		msg = fmt.Sprintf("deposited %s SUPS", output.Shift(-1*passport.SUPSDecimals).StringFixed(4))
	}

	trans := &passport.NewTransaction{
		To:                   toUserID,
		From:                 fromUserID,
		Amount:               output,
		TransactionReference: passport.TransactionReference(record.TxHash),
		Description:          msg,
		Group:                passport.TransactionGroupStore,
	}

	_, _, _, err = ucm.Process(trans)
	if err != nil {
		return fmt.Errorf("create tx entry for tx %s: %w", record.TxHash, err)
	}
	return nil
}

const baseURL = "http://v2.supremacy-api.avantdata.com:3001"

func get(sym Symbol, latestBlock int, testnet bool) ([]*Record, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/%s", baseURL, sym), nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("since_block", strconv.Itoa(latestBlock))
	if testnet {
		q.Add("is_testnet", "true")
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response for %s: %d", req.URL.String(), resp.StatusCode)
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

func BUSD() ([]*Record, error) {
	contractAddr := "0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56"
	LastBUSDBlock := db.GetInt(db.KeyLatestBUSDBlock)
	records, err := get(BUSDSymbol, LastBUSDBlock, false)
	if err != nil {
		return nil, err
	}

	results := []*Record{}
	latestBlock := 0
	for _, record := range records {
		if common.HexToAddress(record.JSON.ContractAddress).Hex() != common.HexToAddress(contractAddr).Hex() {
			continue
		}
		blockNumber, err := strconv.Atoi(record.JSON.BlockNumber)
		if err != nil {
			fmt.Println("avant_scraper: could not parse blocknumber")
			continue
		}
		if blockNumber > latestBlock {
			latestBlock = blockNumber
		}
		results = append(results, record)
	}
	if latestBlock > LastBUSDBlock {
		db.PutInt(db.KeyLatestBUSDBlock, latestBlock)
	}

	return results, nil
}

func USDC() ([]*Record, error) {
	contractAddr := "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	LastUSDCBlock := db.GetInt(db.KeyLatestUSDCBlock)
	records, err := get(USDCSymbol, LastUSDCBlock, false)
	if err != nil {
		return nil, err
	}

	results := []*Record{}
	latestBlock := 0
	for _, record := range records {
		if common.HexToAddress(record.JSON.ContractAddress).Hex() != common.HexToAddress(contractAddr).Hex() {
			continue
		}
		blockNumber, err := strconv.Atoi(record.JSON.BlockNumber)
		if err != nil {
			fmt.Println("avant_scraper: could not parse blocknumber")
			continue
		}
		if blockNumber > latestBlock {
			latestBlock = blockNumber
		}
		results = append(results, record)
	}
	if latestBlock > LastUSDCBlock {
		db.PutInt(db.KeyLatestUSDCBlock, latestBlock)
	}
	return results, nil
}

func ETH() ([]*Record, error) {
	LastETHBlock := db.GetInt(db.KeyLatestETHBlock)
	records, err := get(ETHSymbol, LastETHBlock, false)
	if err != nil {
		return nil, err
	}
	results := []*Record{}
	latestBlock := 0
	for _, record := range records {
		blockNumber, err := strconv.Atoi(record.JSON.BlockNumber)
		if err != nil {
			fmt.Println("avant_scraper: could not parse blocknumber")
			continue
		}
		if blockNumber > latestBlock {
			latestBlock = blockNumber
		}
		results = append(results, record)
	}
	if latestBlock > LastETHBlock {
		db.PutInt(db.KeyLatestETHBlock, latestBlock)
	}
	return results, nil
}

func BNB() ([]*Record, error) {
	LastBNBBlock := db.GetInt(db.KeyLatestBNBBlock)
	records, err := get(BNBSymbol, LastBNBBlock, false)
	if err != nil {
		return nil, err
	}
	results := []*Record{}
	latestBlock := 0
	for _, record := range records {
		blockNumber, err := strconv.Atoi(record.JSON.BlockNumber)
		if err != nil {
			fmt.Println("avant_scraper: could not parse blocknumber")
			continue
		}
		if blockNumber > latestBlock {
			latestBlock = blockNumber
		}
		results = append(results, record)
	}
	if latestBlock > LastBNBBlock {
		db.PutInt(db.KeyLatestBNBBlock, latestBlock)
	}
	return results, nil
}
