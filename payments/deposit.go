package payments

import (
	"context"
	"fmt"
	"passport"
	"passport/api"
	"passport/passdb"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

var latestDepositBlock = 0

func GetDeposits(testnet bool) ([]*Record, error) {
	records, err := get("sups_deposit_txs", latestDepositBlock, testnet)
	if err != nil {
		return nil, err
	}
	latestBlock := latestBlockFromRecords(records)
	latestDepositBlock = latestBlock
	return records, nil
}
func ProcessDeposits(records []*Record, ucm *api.UserCacheMap) (int, int, error) {
	ctx := context.Background()
	success := 0
	skipped := 0
	for _, record := range records {
		user, err := CreateOrGetUser(ctx, passdb.Conn, record.FromAddress)
		if err != nil {
			skipped++
			log.Error().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("create or get user")
			continue
		}

		value, err := decimal.NewFromString(record.JSON.Value)
		if err != nil {
			skipped++
			log.Error().Str("txid", record.TxHash).Err(err).Msg("process decimal")
			continue
		}

		msg := fmt.Sprintf("deposited %s SUPS", value.Shift(-1*api.SUPSDecimals).StringFixed(4))

		trans := &passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               *value.BigInt(),
			TransactionReference: passport.TransactionReference(record.TxHash),
			Description:          msg,
			Group:                passport.TransactionGroupStore,
		}

		_, _, _, err = ucm.Process(trans)
		if err != nil {
			return success, skipped, fmt.Errorf("create tx entry for tx %s: %w", record.TxHash, err)
		}
		success++
	}

	return success, skipped, nil
}
