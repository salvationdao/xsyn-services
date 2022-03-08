package payments

import (
	"passport/api"

	"github.com/davecgh/go-spew/spew"
	"github.com/ninja-software/terror/v2"
)

var latestDepositBlock = 0

func GetDeposits(testnet bool) ([]*Record, error) {
	records, err := get("sups_deposit_txs", latestDepositBlock, testnet)
	if err != nil {
		return nil, err
	}
	latestBlock := latestBlockFromRecords(records)
	latestDepositBlock = latestBlock
	spew.Dump(records)
	return records, nil
}
func ProcessDeposits(records []*Record, ucm *api.UserCacheMap) (int, int, error) {

	return 0, 0, terror.ErrNotImplemented
}
