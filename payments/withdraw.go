package payments

import (
	"fmt"
	"math/big"
	"passport"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-software/terror/v2"
)

// InsertPendingRefund inserts a pending refund to the pending_refunds table
func InsertPendingRefund(ucm UserCacheMap, userID passport.UserID, amount big.Int, expiry time.Time) (string, error) {
	txRef := passport.TransactionReference(fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond()))
	// remove sups

	tx := &passport.NewTransaction{
		To:                   passport.OnChainUserID,
		From:                 userID,
		Amount:               amount,
		TransactionReference: txRef,
		Description:          fmt.Sprintf("Withdraw of %s SUPS", decimal.NewFromBigInt(&amount, -18)),
		Group:                passport.TransactionGroupWithdrawal,
	}

	_, _, _, err := ucm.Process(tx)
	if err != nil {
		return "", terror.Error(err)
	}

	txHold := boiler.PendingRefund{
		RefundedAt:           expiry.Add(1 * time.Minute),
		TransactionReference: string(txRef),
	}

	err = txHold.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return "", terror.Error(err)
	}

	return txHold.ID, nil
}

var latestWithdrawBlock = 0

func GetWithdraws(testnet bool) ([]*Record, error) {
	records, err := get("sups_withdraw_txs", latestWithdrawBlock, testnet)
	if err != nil {
		return nil, fmt.Errorf("get withdraw txes: %w", err)
	}
	latestWithdrawBlock = latestBlockFromRecords(latestWithdrawBlock, records)
	return records, nil
}

func ProcessWithdraws(records []*Record) (int, int, error) {
	// ctx := context.Background()
	success := 0
	skipped := 0
	for _, record := range records {
		value, err := decimal.NewFromString(record.JSON.Value)
		if err != nil {
			skipped++
			passlog.L.Err(err).Str("txid", record.TxHash).Err(err).Msg("process decimal")
			continue
		}
		// find out pending refund and mark it as refunded
		result, err := boiler.PendingRefunds(
			qm.Where("tx_hash ILIKE ?", record.TxHash),
		).One(passdb.StdConn)
		if err != nil {
			skipped++
			passlog.L.Warn().Err(err).Str("tx_id", record.TxHash).Str("amount", value.Shift(-1*SUPDecimals).StringFixed(4)).Msg("could not find matching tx_id in refund table")
			continue
		}
		// check it hasn't expired
		if result.RefundedAt.Before(time.Now()) || result.IsRefunded {
			skipped++
			continue
		}
		// check it hasn't been cancelled
		if result.RefundCanceledAt.Valid {
			skipped++
			continue
		}
		// check it isn't deleted
		if result.DeletedAt.Valid {
			skipped++
			passlog.L.Warn().Err(fmt.Errorf("refund deleted_at not null")).Str("tx_id", record.TxHash).Str("amount", value.Shift(-1*SUPDecimals).StringFixed(4)).Msg("refund has been deleted")
			continue
		}

		result.RefundCanceledAt = null.TimeFrom(time.Now())

		_, err = result.Update(passdb.StdConn, boil.Whitelist(boiler.PendingRefundColumns.RefundCanceledAt))
		if err != nil {
			skipped++
			passlog.L.Err(err).Str("tx_id", record.TxHash).Str("amount", value.Shift(-1*SUPDecimals).StringFixed(4)).Msg("updating pending refund table")
			continue
		}

		success++
	}

	return success, skipped, nil
}

func ProcessPendingRefunds(ucm UserCacheMap) (int, int, error) {
	success := 0
	skipped := 0
	refundsToProcess, err := boiler.PendingRefunds(
		boiler.PendingRefundWhere.RefundedAt.LT(time.Now()),
		qm.And("refund_canceled_at IS NULL"),
		qm.And("is_refunded = false"),
		qm.And("deleted_at IS NULL"),
		qm.Load(boiler.PendingRefundRels.TransactionReferenceTransaction),
	).All(passdb.StdConn)
	if err != nil {
		return success, skipped, err
	}

	for _, refund := range refundsToProcess {
		userUUID, err := uuid.FromString(refund.R.TransactionReferenceTransaction.Debit)
		if err != nil {
			skipped++
			passlog.L.Err(err).Msg("failed to convert to user uuid")
			continue
		}
		newTx := &passport.NewTransaction{
			To:                   passport.UserID(userUUID),
			From:                 passport.OnChainUserID,
			Amount:               *refund.R.TransactionReferenceTransaction.Amount.BigInt(),
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s", refund.R.TransactionReferenceTransaction.TransactionReference)),
			Description:          fmt.Sprintf("REFUND %s", refund.R.TransactionReferenceTransaction.Description),
			Group:                passport.TransactionGroup(refund.R.TransactionReferenceTransaction.Group),
		}

		_, _, _, err = ucm.Process(newTx)
		if err != nil {
			skipped++
			passlog.L.Err(err).Msg("failed to process refund")
			continue
		}
		success++
	}

	return success, skipped, nil
}
