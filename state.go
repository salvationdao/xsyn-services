package passport

import "github.com/shopspring/decimal"

type State struct {
	LatestEthBlock uint64          `db:"latest_eth_block"`
	LatestBscBlock uint64          `db:"latest_bsc_block"`
	ETHtoUSD       decimal.Decimal `json:"ETHtoUSD" db:"eth_to_usd"`
	BNBtoUSD       decimal.Decimal `json:"BNBtoUSD" db:"bnb_to_usd"`
	SUPtoUSD       decimal.Decimal `json:"SUPtoUSD" db:"sup_to_usd"`
}
