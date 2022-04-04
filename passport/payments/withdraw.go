package payments

import (
	"context"
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/volatiletech/null/v8"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-software/terror/v2"
)

// InsertPendingRefund inserts a pending refund to the pending_refunds table
func InsertPendingRefund(ucm UserCacheMap, userID types.UserID, amount decimal.Decimal, expiry time.Time) (string, error) {
	txRef := types.TransactionReference(fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond()))
	// remove sups

	newTx := &types.NewTransaction{
		To:                   types.OnChainUserID,
		From:                 userID,
		Amount:               amount,
		TransactionReference: txRef,
		Description:          fmt.Sprintf("Withdraw of %s SUPS", amount.Shift(-18).StringFixed(4)),
		Group:                types.TransactionGroupWithdrawal,
	}

	_, _, _, err := ucm.Transact(newTx)
	if err != nil {
		return "", terror.Error(err)
	}

	tx, err := boiler.Transactions(boiler.TransactionWhere.TransactionReference.EQ(string(newTx.TransactionReference))).One(passdb.StdConn)
	if err != nil {
		return "", terror.Error(err)
	}

	amountString, err := decimal.NewFromString(amount.String())
	if err != nil {
		return "", terror.Error(err)
	}

	txHold := boiler.PendingRefund{
		UserID:                userID.String(),
		RefundedAt:            expiry.Add(10 * time.Minute),
		TransactionReference:  string(txRef),
		AmountSups:            amountString,
		WithdrawTransactionID: null.StringFrom(tx.ID),
	}

	err = txHold.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return "", terror.Error(err)
	}

	return txHold.ID, nil
}

func UpdateSuccessfulWithdrawsWithTxHash(records []*SUPTransferRecord) (int, int) {
	l := passlog.L.With().Str("svc", "avant_pending_refund_set_tx_hash").Logger()

	skipped := 0
	success := 0
	for _, record := range records {
		val, err := decimal.NewFromString(record.ValueInt)
		if err != nil {
			l.Warn().
				Str("user_addr", record.ToAddress).
				Str("tx_hash", record.TxHash).
				Err(err).
				Msg("convert to decimal failed")
			skipped++
			continue
		}

		u, err := db.UserByPublicAddress(context.Background(), passdb.Conn, common.HexToAddress(record.ToAddress))
		if err != nil {
			skipped++
			continue
		}

		filter := []qm.QueryMod{
			boiler.PendingRefundWhere.UserID.EQ(u.ID.String()),
			boiler.PendingRefundWhere.AmountSups.EQ(val),
			boiler.PendingRefundWhere.IsRefunded.EQ(false),
			boiler.PendingRefundWhere.RefundCanceledAt.IsNull(), // Not cancelled yet
			boiler.PendingRefundWhere.DeletedAt.IsNull(),
			boiler.PendingRefundWhere.TXHash.EQ(""),
			boiler.PendingRefundWhere.TXHash.NEQ(record.TxHash), // Ignore tx hash if already assigned to another pending refund
		}

		count, err := boiler.PendingRefunds(filter...).Count(passdb.StdConn)
		if err != nil {
			l.Warn().Err(err).Msg("failed to get count")
			skipped++
			continue
		}
		if count <= 0 {
			//is this even an error? do we need to be warned about this?
			//l.Warn().Err(err).Msg("user does not have any pending refunds matching the value")
			skipped++
			continue
		}

		// Get pending refunds for user that are ready to be confirmed as on chain
		filter = append(filter, qm.OrderBy("created_at ASC")) // Sort so we get the oldest one
		pendingRefund, err := boiler.PendingRefunds(filter...).One(passdb.StdConn)
		if err != nil {
			l.Warn().Err(err).Msg("could not get matching single pending refund")
			skipped++
			continue
		}
		pendingRefund.TXHash = record.TxHash
		pendingRefund.RefundCanceledAt = null.TimeFrom(time.Now())

		_, err = pendingRefund.Update(passdb.StdConn, boil.Whitelist(boiler.PendingRefundColumns.TXHash, boiler.PendingRefundColumns.RefundCanceledAt))
		if err != nil {
			l.Warn().Err(err).Msg("failed to update user pending refund with tx hash")
			skipped++
			continue
		}

		//l.Info().Msg("successfully set tx hash, cancel refund")
		success++
	}

	return success, skipped
}

// Rollback stale withdraws (dangerous if buggy, check very, very carefully)
func ReverseFailedWithdraws(ucm UserCacheMap, enableWithdrawRollback bool) (int, int, error) {
	l := passlog.L.
		With().
		Str("svc", "avant_rollback_withdraw").
		Bool("enable_withdraw_rollback", enableWithdrawRollback).
		Logger()

	success := 0
	skipped := 0

	// Get refunds that can be marked as failed withdraws
	filter := []qm.QueryMod{
		boiler.PendingRefundWhere.RefundedAt.LT(time.Now()),
		boiler.PendingRefundWhere.RefundCanceledAt.IsNull(),
		boiler.PendingRefundWhere.IsRefunded.EQ(false),
		boiler.PendingRefundWhere.DeletedAt.IsNull(),
		boiler.PendingRefundWhere.TXHash.EQ(""),
		qm.Load(boiler.PendingRefundRels.TransactionReferenceTransaction),
	}

	refundsToProcess, err := boiler.PendingRefunds(filter...).All(passdb.StdConn)
	if err != nil {
		return success, skipped, err
	}

	for _, refund := range refundsToProcess {
		userUUID, err := uuid.FromString(refund.R.TransactionReferenceTransaction.Debit)
		if err != nil {
			skipped++
			l.Warn().Err(err).Msg("failed to convert to user uuid")
			continue
		}

		txRef := types.TransactionReference(fmt.Sprintf("REFUND %s", refund.R.TransactionReferenceTransaction.TransactionReference))
		newTx := &types.NewTransaction{
			To:                   types.UserID(userUUID),
			From:                 types.OnChainUserID,
			Amount:               refund.R.TransactionReferenceTransaction.Amount,
			TransactionReference: txRef,
			Description:          fmt.Sprintf("REFUND %s", refund.R.TransactionReferenceTransaction.Description),
			Group:                types.TransactionGroup(refund.R.TransactionReferenceTransaction.Group),
			RelatedTransactionID: refund.R.TransactionReferenceTransaction.RelatedTransactionID,
		}

		refund.RefundCanceledAt = null.TimeFrom(time.Now())
		refund.IsRefunded = true

		l = l.With().
			Str("refund.refund_id", refund.ID).
			Str("refund.user_id", refund.UserID).
			Str("refund.amount_sups", refund.AmountSups.Shift(-18).StringFixed(4)).
			Str("refund.refunded_at", refund.RefundedAt.Format(time.RFC3339)).
			Str("refund.refund_canceled_at", refund.RefundCanceledAt.Time.Format(time.RFC3339)).
			Str("refund.tx_hash", refund.TXHash).
			Str("refund.transaction_reference", refund.TransactionReference).
			Bool("refund.is_refunded", refund.IsRefunded).
			Str("reverse_tx.to", newTx.To.String()).
			Str("reverse_tx.from", newTx.From.String()).
			Str("reverse_tx.amount", newTx.Amount.String()).
			Str("reverse_tx.transaction_reference", string(newTx.TransactionReference)).
			Str("reverse_tx.description", newTx.Description).
			Str("reverse_tx.group", string(newTx.Group)).
			Logger()

		if enableWithdrawRollback {
			_, _, txID, err := ucm.Transact(newTx)
			if err != nil {
				skipped++
				l.Warn().Err(err).Msg("failed to process refund")
				continue
			}

			// Link withdrawal to transaction ID
			refund.ReversalTransactionID = null.StringFrom(txID)

			_, err = refund.Update(passdb.StdConn, boil.Infer())
			if err != nil {
				skipped++
				l.Warn().Err(err).Msg("failed to process refund")
				continue
			}
			l.Info().Msg("successfully reversed withdraw")
		} else {
			l.Info().Msg("successfully reversed withdraw (dry run)")
		}

		success++
	}

	return success, skipped, nil
}