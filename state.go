package passport

import "github.com/shopspring/decimal"

type State struct {
	LatestEthBlock uint64          `json:"latest_eth_block" db:"latest_eth_block"`
	LatestBscBlock uint64          `json:"latest_bsc_block" db:"latest_bsc_block" `
	ETHtoUSD       decimal.Decimal `json:"eth_to_usd" db:"eth_to_usd"`
	BNBtoUSD       decimal.Decimal `json:"bnb_to_usd" db:"bnb_to_usd"`
	SUPtoUSD       decimal.Decimal `json:"sup_to_usd" db:"sup_to_usd"`
}
