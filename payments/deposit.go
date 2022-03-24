package payments

import (
	"context"
	"fmt"
	"passport"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DepositTransactionStatus string

const (
	DepositTransactionStatusPending   DepositTransactionStatus = "pending"
	DepositTransactionStatusConfirmed DepositTransactionStatus = "confirmed"
)

func ProcessDeposits(records []*SUPTransferRecord, ucm UserCacheMap) (int, int, error) {
	l := passlog.L.With().Str("svc", "avant_deposit_processor").Logger()
	ctx := context.Background()
	success := 0
	skipped := 0

	l.Info().Int("records", len(records)).Msg("processing deposits")
	for _, record := range records {
		exists, err := boiler.Transactions(boiler.TransactionWhere.TransactionReference.EQ(record.TxHash)).Exists(passdb.StdConn)
		if err != nil {
			skipped++
			l.Debug().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("check if tx exists")
			continue
		}

		if exists {
			skipped++
			l.Debug().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("tx already exists")
			continue
		}
		user, err := CreateOrGetUser(ctx, passdb.Conn, common.HexToAddress(record.FromAddress))
		if err != nil {
			skipped++
			l.Error().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("create or get user")
			continue
		}

		value, err := decimal.NewFromString(record.ValueInt)
		if err != nil {
			skipped++
			l.Error().Str("txid", record.TxHash).Err(err).Msg("process decimal")
			continue
		}

		if value.Equal(decimal.Zero) {
			l.Debug().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("skipping zero value deposit")
			skipped++
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

		// Update deposit transaction's status in db from pending to success
		dtx, err := boiler.DepositTransactions(boiler.DepositTransactionWhere.TXHash.EQ(record.TxHash)).One(passdb.StdConn)
		if err != nil {
			l.Err(err).Str("txid", record.TxHash).Msg("failed to find tx entry for deposit in db")
			continue
		}

		dtx.Status = string(DepositTransactionStatusConfirmed)
		dtx.UpdatedAt = time.Now()
		_, err = dtx.Update(passdb.StdConn, boil.Infer())
		if err != nil {
			l.Err(err).Str("txid", record.TxHash).Msg("failed to update tx entry for deposit in db")
			continue
		}
		l.Info().Str("txid", record.TxHash).Msg("successfully updated status of tx entry for deposit in db")
	}
	l.Info().
		Int("success", success).
		Int("skipped", skipped).
		Msg("synced deposits")

	return success, skipped, nil
}
