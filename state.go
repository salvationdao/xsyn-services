package passport

type State struct {
	LatestEthBlock uint64 `db:"latest_eth_block"`
	LatestBscBlock uint64 `db:"latest_bsc_block"`
}
