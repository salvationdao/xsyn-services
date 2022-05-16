package comms

import (
	"fmt"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/gofrs/uuid"

	"github.com/volatiletech/null/v8"

	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
)

func (s *S) RefundTransaction(req RefundTransactionReq, resp *RefundTransactionResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	transaction, err := db.TransactionGet(req.TransactionID)
	if err != nil {
		passlog.L.Error().
			Err(err).
			Str("func", "RefundTransaction").
			Str("transaction_id", req.TransactionID).
			Msg("error finding transaction for refund")
		return terror.Error(err, "Failed to find transaction.")
	}

	if transaction.RelatedTransactionID.Valid && transaction.RelatedTransactionID.String != "" {
		return terror.Warn(fmt.Errorf("transaction already has related transaction id"), "Transaction already has related transaction ID.")
	}

	if !transaction.ServiceID.Valid || serviceID != transaction.ServiceID.String {
		passlog.L.Error().
			Err(err).
			Str("func", "RefundTransaction").
			Str("transaction_id", req.TransactionID).
			Str("service id", serviceID).
			Str("transaction_service_id", transaction.ServiceID.String).
			Msg("service trying to refund transaction from another service")
		return terror.Error(terror.ErrForbidden, "You can only refund transactions you made.")
	}

	tx := &types.NewTransaction{
		From:                 types.UserID(uuid.FromStringOrNil(transaction.Credit)),
		To:                   types.UserID(uuid.FromStringOrNil(transaction.Debit)),
		TransactionReference: types.TransactionReference(fmt.Sprintf("REFUND - %s", transaction.TransactionReference)),
		Description:          fmt.Sprintf("Reverse transaction - %s", transaction.Description),
		Amount:               transaction.Amount,
		Group:                types.TransactionGroup(transaction.Group),
		SubGroup:             transaction.SubGroup.String,
		ServiceID:            types.UserID(uuid.FromStringOrNil(transaction.ServiceID.String)),
		RelatedTransactionID: null.StringFrom(transaction.ID),
	}

	_, _, txID, err := s.UserCacheMap.Transact(tx)
	if err != nil {
		passlog.L.Error().
			Err(err).
			Str("func", "RefundTransaction").
			Str("transaction_id", req.TransactionID).
			Msg("refund failed")
		return terror.Error(err, "Failed to process refund.")
	}

	// mark the original transaction as refunded
	err = db.TransactionAddRelatedTransaction(transaction.ID, txID)
	if err != nil {
		passlog.L.Error().
			Err(err).
			Str("func", "RefundTransaction").
			Str("original_transaction_id", transaction.ID).
			Str("new_related_trasaction_id", txID).
			Msg("failed to add related transaction id.")
	}

	resp.TransactionID = txID
	return nil
}

func (s *S) SupremacySpendSupsHandler(req SpendSupsReq, resp *SpendSupsResp) error {
	serviceID, err := IsSupremacyClient(req.ApiKey)
	if err != nil {
		return err
	}
	amt, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return err
	}

	user, err := users.UUID(req.FromUserID)
	if err != nil {
		return err
	}

	isLocked := user.CheckUserIsLocked("account")
	if isLocked {
		return terror.Error(fmt.Errorf("user: %s attempting to purchase on Supremacy while locked", user.ID), "This account is locked, contact support to unlock.")
	}

	if amt.LessThan(decimal.Zero) {
		return terror.Error(terror.ErrInvalidInput, "Sups amount can not be negative")
	}

	serviceAsUUID := uuid.FromStringOrNil(serviceID)
	if serviceAsUUID.IsNil() {
		return terror.Error(fmt.Errorf("service uuid is nil"))
	}

	tx := &types.NewTransaction{
		From:                 types.UserID(req.FromUserID),
		To:                   types.UserID(req.ToUserID),
		TransactionReference: req.TransactionReference,
		Description:          req.Description,
		Amount:               amt,
		Group:                req.Group,
		SubGroup:             req.SubGroup,
		ServiceID:            types.UserID(serviceAsUUID),
	}

	if req.NotSafe {
		tx.NotSafe = true
	}

	_, _, txID, err := s.UserCacheMap.Transact(tx)
	if err != nil {
		return terror.Error(err, "failed to process sups")
	}

	tx.ID = txID

	resp.TransactionID = txID
	return nil
}
