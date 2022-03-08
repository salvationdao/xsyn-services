package payments_test

import (
	"encoding/json"
	"passport/payments"
	"testing"

	"github.com/shopspring/decimal"
)

func TestProcessValues(t *testing.T) {
	record := `
	{
		"symbol": "eth",
		"usd_rate": "2718.825",
		"sups": "0.022656875000000000000000000000000000",
		"value": "0.000001000000000000",
		"tx_hash": "0x729c49ceae31895a822b173b1396be4ea6061c9c59e1198cc0c5ecdb03c696e6",
		"time": 1645607749.0,
		"to_address": "0x52b38626d3167e5357fe7348624352b7062fe271",
		"from_address": "0x03469fdba3e9f4880e8e9dd7b74d61851afc02f3",
		"json": {
		  "to": "0x52b38626d3167e5357fe7348624352b7062fe271",
		  "gas": "21000",
		  "from": "0x03469fdba3e9f4880e8e9dd7b74d61851afc02f3",
		  "hash": "0x729c49ceae31895a822b173b1396be4ea6061c9c59e1198cc0c5ecdb03c696e6",
		  "input": "0x",
		  "nonce": "2",
		  "value": "1000000000000",
		  "gasUsed": "21000",
		  "isError": "0",
		  "gasPrice": "34061670065",
		  "blockHash": "0xe63a8d815c7c3d2b46f26d6314a964578a0d7b2c81a9d3bf9323bba9ed41ad34",
		  "timeStamp": "1645607749",
		  "blockNumber": "14261476",
		  "confirmations": "4083",
		  "contractAddress": "",
		  "transactionIndex": "149",
		  "txreceipt_status": "1",
		  "cumulativeGasUsed": "16892173"
		}
	  }
	  
	`
	result := &payments.Record{}
	err := json.Unmarshal([]byte(record), result)
	if err != nil {
		t.Error(err)
	}
	input, output, decimals, err := payments.ProcessValues(result.Sups, result.Value, result.JSON.TokenDecimal)
	if err != nil {
		t.Error(err)
	}
	if decimals == 0 {
		t.Error("decimals is zero")
	}
	if input.Equal(decimal.Zero) {
		t.Error("zero input")
	}
	if output.Equal(decimal.Zero) {
		t.Error("zero output")
	}

}
