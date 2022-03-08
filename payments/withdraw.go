package payments

import (
	"context"
	"fmt"
	"passport/api"
	"passport/passdb"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

var latestWithdrawBlock = 0

func GetWithdraws(testnet bool) ([]*Record, error) {
	records, err := get("sups_withdraw_txs", latestWithdrawBlock, testnet)
	if err != nil {
		return nil, fmt.Errorf("get withdraw txes: %w", err)
	}
	latestBlock := latestBlockFromRecords(records)
	latestWithdrawBlock = latestBlock
	return records, nil
}
func ProcessWithdraws(records []*Record, ucm *api.UserCacheMap) (int, int, error) {
	ctx := context.Background()
	success := 0
	skipped := 0
	for _, record := range records {
		u, err := CreateOrGetUser(ctx, passdb.Conn, record.FromAddress)
		if err != nil {
			skipped++
			log.Error().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("user by address")
			continue
		}
		value, err := decimal.NewFromString(record.JSON.Value)
		if err != nil {
			skipped++
			log.Error().Str("txid", record.TxHash).Err(err).Msg("process decimal")
			continue
		}

		fmt.Printf("[TESTNET] WITHDRAW AMOUNT BY %s: %s\n", u.PublicAddress.String, value.Shift(-1*api.SUPSDecimals).StringFixed(4))

		// HANDLE RECORD HERE

		success++
	}
	return success, skipped, nil
}
