package api

import (
	"encoding/json"
	"net/http"
	"time"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/ninja-software/terror/v2"
	"github.com/ethereum/go-ethereum/common"
)

type Transfer struct {
	TriggerName string `json:"triggerName"`
	Object      struct {
		FromAddress    string    `json:"from_address"`
		Hash           string    `json:"hash"`
		ToAddress      string    `json:"to_address"`
		CreatedAt      time.Time `json:"createdAt"`
		BlockHash      string    `json:"block_hash"`
		BlockNumber    uint64       `json:"block_number"`
		BlockTimestamp struct {
			Type string    `json:"__type"`
			Iso  time.Time `json:"iso"`
		} `json:"block_timestamp"`
		Confirmed bool `json:"confirmed"`
		Decimal   struct {
			Type  string `json:"__type"`
			Value string `json:"value"`
		} `json:"decimal"`
		Gas                      int       `json:"gas"`
		GasPrice                 int64     `json:"gas_price"`
		Input                    string    `json:"input"`
		Nonce                    int       `json:"nonce"`
		ReceiptCumulativeGasUsed int       `json:"receipt_cumulative_gas_used"`
		ReceiptGasUsed           int       `json:"receipt_gas_used"`
		ReceiptStatus            int       `json:"receipt_status"`
		TransactionIndex         int       `json:"transaction_index"`
		Value                    string    `json:"value"`
		UpdatedAt                time.Time `json:"updatedAt"`
		ObjectID                 string    `json:"objectId"`
		ClassName                string    `json:"className"`
	} `json:"object"`
	Master bool `json:"master"`
	Log    struct {
		Options struct {
			JSONLogs    bool   `json:"jsonLogs"`
			LogsFolder  string `json:"logsFolder"`
			Verbose     bool   `json:"verbose"`
			MaxLogFiles int    `json:"maxLogFiles"`
		} `json:"options"`
		AppID string `json:"appId"`
	} `json:"log"`
	Headers struct {
		Accept              string `json:"accept"`
		ContentType         string `json:"content-type"`
		XParseApplicationID string `json:"x-parse-application-id"`
		XParseSessionToken  string `json:"x-parse-session-token"`
		UserAgent           string `json:"user-agent"`
		ContentLength       string `json:"content-length"`
		Host                string `json:"host"`
		Connection          string `json:"connection"`
	} `json:"headers"`
	IP       string `json:"ip"`
	Original struct {
		FromAddress    string    `json:"from_address"`
		Hash           string    `json:"hash"`
		ToAddress      string    `json:"to_address"`
		CreatedAt      time.Time `json:"createdAt"`
		BlockHash      string    `json:"block_hash"`
		BlockNumber    int       `json:"block_number"`
		BlockTimestamp struct {
			Type string    `json:"__type"`
			Iso  time.Time `json:"iso"`
		} `json:"block_timestamp"`
		Confirmed bool `json:"confirmed"`
		Decimal   struct {
			Type  string `json:"__type"`
			Value struct {
				NumberDecimal string `json:"$numberDecimal"`
			} `json:"value"`
		} `json:"decimal"`
		Gas                      int       `json:"gas"`
		GasPrice                 int64     `json:"gas_price"`
		Input                    string    `json:"input"`
		Nonce                    int       `json:"nonce"`
		ReceiptCumulativeGasUsed int       `json:"receipt_cumulative_gas_used"`
		ReceiptGasUsed           int       `json:"receipt_gas_used"`
		ReceiptStatus            int       `json:"receipt_status"`
		TransactionIndex         int       `json:"transaction_index"`
		Value                    string    `json:"value"`
		UpdatedAt                time.Time `json:"updatedAt"`
		ObjectID                 string    `json:"objectId"`
		ClassName                string    `json:"className"`
	} `json:"original"`
	Context struct {
	} `json:"context"`
	User struct {
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		ACL       struct {
			RoleCoreservices struct {
				Read  bool `json:"read"`
				Write bool `json:"write"`
			} `json:"role:coreservices"`
			NjOTfRODukXQg5KckhwSXfPc struct {
				Read  bool `json:"read"`
				Write bool `json:"write"`
			} `json:"njOTfRODukXQg5kckhwSXfPc"`
		} `json:"ACL"`
		SessionToken string `json:"sessionToken"`
		ObjectID     string `json:"objectId"`
	} `json:"user"`
}

func (cc *ChainClients) TransactionWebhook(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &Transfer{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}

	chain := ""


	if req.Object.ClassName == "EthTransactions" {
		chain = "eth"
	}


	if req.Object.ClassName == "BscTransactions" {
		chain = "bsc"
	}

	if chain == "" {
		cc.Log.Err(terror.Error(fmt.Errorf("invalid chain classname %s %s", chain, req.Object.ClassName))).Msg("")
		return http.StatusOK, nil
	}

	record, err := GetNativeTX(common.HexToHash(req.Object.Hash), chain)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Missing Tx.")
	}

	cc.Log.Info().
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
