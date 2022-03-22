package payments

import (
	"context"
	"fmt"
	"passport"
	"passport/db"
	"passport/passdb"
	"passport/passlog"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

func GetDeposits(testnet bool) ([]*Record, error) {
	latestDepositBlock := db.GetInt(db.KeyLatestDepositBlock)
	records, err := get("sups_deposit_txs", latestDepositBlock, testnet)
	if err != nil {
		return nil, err
	}
	db.PutInt(db.KeyLatestDepositBlock, latestBlockFromRecords(latestDepositBlock, records))
	return records, nil
}
func ProcessDeposits(records []*Record, ucm UserCacheMap) (int, int, error) {
	l := passlog.L.With().Str("svc", "avant_deposit_processor").Logger()
	ctx := context.Background()
	success := 0
	skipped := 0
	for _, record := range records {
		user, err := CreateOrGetUser(ctx, passdb.Conn, common.HexToAddress(record.FromAddress))
		if err != nil {
			skipped++
			l.Error().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("create or get user")
			continue
		}

		value, err := decimal.NewFromString(record.JSON.Value)
		if err != nil {
			skipped++
			l.Error().Str("txid", record.TxHash).Err(err).Msg("process decimal")
			continue
		}

		msg := fmt.Sprintf("deposited %s SUPS", value.Shift(-1*passport.SUPSDecimals).StringFixed(4))
		l.Debug().Str("msg", msg).Str("txid", record.TxHash).Msg("insert deposit tx")
		trans := &passport.NewTransaction{
			To:                   user.ID,
			From:                 passport.XsynTreasuryUserID,
			Amount:               value,
			TransactionReference: passport.TransactionReference(record.TxHash),
			Description:          msg,
			Group:                passport.TransactionGroupStore,
		}

		_, _, _, err = ucm.Process(trans)
		if err != nil {
			l.Err(err).Str("txid", record.TxHash).Msg("failed to create tx entry for deposit")
			skipped++
			continue
		}
		success++
	}

	return success, skipped, nil
}
