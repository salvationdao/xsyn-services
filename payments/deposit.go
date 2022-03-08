package payments

import (
	"passport/api"
	"time"
)

type DepositRecord struct {
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

func GetDeposits() error                          { return nil }
func InsertDeposits() error                       { return nil }
func ProcessDeposits(ucm *api.UserCacheMap) error { return nil }
