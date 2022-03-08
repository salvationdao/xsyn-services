package payments

import (
	"fmt"
	"passport/api"
	"passport/passlog"

	"github.com/davecgh/go-spew/spew"
)

var latestWithdrawBlock = 0

func GetWithdraws(testnet bool) ([]*Record, error) {
	records, err := get("sups_withdraw_txs", latestWithdrawBlock, testnet)
	if err != nil {
		return nil, fmt.Errorf("get withdraw txes: %w", err)
	}
	latestBlock := latestBlockFromRecords(records)
	latestWithdrawBlock = latestBlock
	spew.Dump(records)
	return records, nil
}
func ProcessWithdraws(records []*Record, ucm *api.UserCacheMap) (int, int, error) {
	success := 0
	skipped := 0
	for _, record := range records {
		in, out, decimal, err := ProcessValues(record.Sups, record.JSON.Input, record.JSON.TokenDecimal)
		if err != nil {
			skipped++
			passlog.L.Err(err).Msg("get withdraw decimal values")
		}
		spew.Dump(in, out, decimal)

		// HANDLE RECORD HERE

		success++
	}
	return success, skipped, nil
}
