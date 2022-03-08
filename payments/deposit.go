package payments

import (
	"context"
	"passport"
	"passport/api"
	"passport/passdb"

	"github.com/davecgh/go-spew/spew"
	"github.com/rs/zerolog/log"
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
	ctx := context.Background()
	success := 0
	skipped := 0
	for _, r := range records {
		user, err := CreateOrGetUser(ctx, passdb.Conn, r.FromAddress)
		if err != nil {
			skipped++
			log.Error().Str("sym", r.Symbol).Str("txid", r.TxHash).Err(err).Msg("store record")
			continue
		}
		err = StoreRecord(ctx, passport.XsynTreasuryUserID, user.ID, ucm, r, false)
		if err != nil {
			skipped++
			log.Error().Str("sym", r.Symbol).Str("txid", r.TxHash).Err(err).Msg("store record")
			continue
		}
		success++
	}

	return success, skipped, nil
}
