package payments

import (
	"fmt"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type DepositTransactionStatus string

const (
	DepositTransactionStatusPending   DepositTransactionStatus = "pending"
	DepositTransactionStatusConfirmed DepositTransactionStatus = "confirmed"
)

func ProcessDeposits(records []*SUPTransferRecord, ucm UserCacheMap, purchaseAddress common.Address, environment types.Environment) (int, int, error) {
	l := passlog.L.With().Str("svc", "avant_deposit_processor").Logger()
	success := 0
	skipped := 0

	l.Info().Int("records", len(records)).Msg("processing deposits")
	for _, record := range records {
		if strings.EqualFold(record.FromAddress, purchaseAddress.String()) {
			skipped++
			continue
		}

		exists, err := db.TransactionReferenceExists(record.TxHash)
		if err != nil {
			skipped++
			continue
		}

		if exists {
			skipped++
			continue
		}

		user, err := CreateOrGetUser(common.HexToAddress(record.FromAddress), environment)
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
			skipped++
			continue
		}

		msg := fmt.Sprintf("deposited %s SUPS", value.Shift(-1*types.SUPSDecimals).StringFixed(4))

		debitor, err := boiler.FindUser(passdb.StdConn, types.OnChainUserID.String())
		if err != nil {
			skipped++
			continue
		}

		trans := &types.NewTransaction{
			CreditAccountID:      user.AccountID,
			DebitAccountID:       debitor.AccountID,
			Amount:               value,
			TransactionReference: types.TransactionReference(record.TxHash),
			Description:          msg,
			Group:                types.TransactionGroupStore,
		}
		_, err = ucm.Transact(trans)
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
