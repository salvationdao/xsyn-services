package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ninja-software/log_helpers"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
)

// SupController holds handlers for as
type SupController struct {
	Log *zerolog.Logger
	API *API
	cc  *ChainClients
}

// NewSupController creates the sup hub
func NewSupController(log *zerolog.Logger, api *API, cc *ChainClients) *SupController {
	supHub := &SupController{
		Log: log_helpers.NamedLogger(log, "sup_hub"),
		API: api,
		cc:  cc,
	}

	api.SecureCommand(HubKeyWithdrawSups, supHub.WithdrawSupHandler)
	api.SecureCommand(HubKeyDepositSups, supHub.DepositSupHandler)
	api.SecureCommand(HubKeyDepositTransactionList, supHub.DepositTransactionListHandler)
	return supHub
}

// DepositTransactionListResponse is the response from DepositTransactionList
type DepositTransactionListResponse struct {
	Total        int                            `json:"total"`
	Transactions boiler.DepositTransactionSlice `json:"transactions"`
}

const HubKeyDepositTransactionList = "SUPS:DEPOSIT:LIST"

func (sc *SupController) DepositTransactionListHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue getting deposit transaction list, try again or contact support."

	dtxs, err := boiler.DepositTransactions(boiler.DepositTransactionWhere.UserID.EQ(user.ID), qm.Limit(10), qm.OrderBy("created_at DESC")).All(passdb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if dtxs == nil {
		dtxs = make(boiler.DepositTransactionSlice, 0)
	}

	resp := &DepositTransactionListResponse{
		len(dtxs),
		dtxs,
	}

	reply(resp)
	return nil
}

type SupDepositRequest struct {
	Payload struct {
		TransactionHash string          `json:"transaction_hash"`
		Amount          decimal.Decimal `json:"amount"`
	} `json:"payload"`
}

const HubKeyDepositSups = "SUPS:DEPOSIT"

func (sc *SupController) DepositSupHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing SUPs deposit transaction, try again or contact support."

	req := &SupDepositRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.TransactionHash == "" {
		passlog.L.Error().Str("func", "DepositSupHandler").Msg("deposit transaction hash was not provided")
		return terror.Error(fmt.Errorf("transaction hash was not provided"), errMsg)
	}

	if req.Payload.Amount.LessThan(decimal.NewFromInt(0)) {
		passlog.L.Error().Str("func", "DepositSupHandler").Msg("deposit transaction amount is lower than the minimum required amount")
		return terror.Error(fmt.Errorf("deposit transaction amount is lower than the minimum required amount"), "Deposit transaction amount is lower than the minimum required amount.")
	}

	dtx := boiler.DepositTransaction{
		UserID: user.ID,
		TXHash: req.Payload.TransactionHash,
		Amount: req.Payload.Amount,
	}
	err = dtx.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Str("func", "DepositSupHandler").Msg("failed to create deposit transaction in db")
		return terror.Error(err, errMsg)
	}

	reply(true)
	return nil
}

type SupWithdrawRequest struct {
	Payload struct {
		Amount string `json:"amount"`
	} `json:"payload"`
}

const HubKeyWithdrawSups = "SUPS:WITHDRAW"

func (sc *SupController) WithdrawSupHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue processing SUPs withdrawal transaction, try again or contact support."

	req := &SupWithdrawRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if sc.cc.SUPS == nil {
		return terror.Error(fmt.Errorf("sups controller not initalized"), "Internal error, try again or contact support.")
	}

	withdrawAmount, err := decimal.NewFromString(req.Payload.Amount)
	if err != nil {
		return terror.Error(fmt.Errorf("failed to create decimal from amount: %v", err), errMsg)
	}

	if !user.PublicAddress.Valid || user.PublicAddress.String == "" {
		return terror.Error(fmt.Errorf("user has no public address"), "Account missing public address.")
	}

	//checks block_withdraw table and returns if user's connected wallet is found
	blockedExists, err := boiler.BlockWithdraws(
		boiler.BlockWithdrawWhere.PublicAddress.EQ(user.PublicAddress.String),
		boiler.BlockWithdrawWhere.BlockNFTWithdraws.GTE(time.Now()),
	).Exists(passdb.StdConn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if blockedExists {
		return terror.Error(fmt.Errorf("user is blocked from withdrawing sups"), "The address connected to this account may not withdraw SUPS.")
	}

	// if balance too low
	if user.Sups.Cmp(withdrawAmount) < 0 {
		return terror.Error(fmt.Errorf("user tried to withdraw without enough funds"), "Insufficient funds.")
	}
	// check withdraw wallet sups
	balanceBI, err := sc.cc.SUPS.Balance()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	balance := decimal.NewFromBigInt(balanceBI, 0)

	// if wallet sups balance too low
	if balance.LessThan(withdrawAmount) {
		return terror.Error(fmt.Errorf("not enough funds in our withdraw wallet"), "Insufficient in funds in withdraw wallet at this time.")
	}
	// check withdraw wallet gas
	pendingBalance, err := sc.cc.BscClient.PendingBalanceAt(ctx, sc.cc.SUPS.PublicAddress)
	if err != nil {
		errMsg := "Issue creating SUPs deposit transaction, try again or contact support."
		return terror.Error(err, errMsg)
	}

	suggestGasPrice, err := sc.cc.BscClient.SuggestGasPrice(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	if pendingBalance.Cmp(suggestGasPrice) < 0 {
		return terror.Error(fmt.Errorf("not enough gas funds in our withdraw wallet"), "Insufficient gas in withdraw wallet at this time.")
	}
	txID := uuid.Must(uuid.NewV4())

	txRef := fmt.Sprintf("sup|withdraw|%s", txID)

	trans := &types.NewTransaction{
		To:                   types.OnChainUserID,
		From:                 types.UserID(uuid.Must(uuid.FromString(user.ID))),
		NotSafe:              true,
		Amount:               withdrawAmount,
		TransactionReference: types.TransactionReference(txRef),
		Description:          "Withdraw of SUPS.",
		Group:                types.TransactionGroupWithdrawal,
	}

	_, err = sc.API.userCacheMap.Transact(trans)
	if err != nil {
		return terror.Error(err, errMsg)
	}


	// refund callback
	refund := func(reason string) {
		trans := &types.NewTransaction{
			To:                   types.UserID(uuid.Must(uuid.FromString(user.ID))),
			NotSafe:              true,
			From:                 types.OnChainUserID,
			Amount:               withdrawAmount,
			TransactionReference: types.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of Withdraw of SUPS.",
			Group:                types.TransactionGroupWithdrawal,
		}

		_, err := sc.API.userCacheMap.Transact(trans)
		if err != nil {
			sc.API.Log.Err(fmt.Errorf("failed to process user fund"))
			return
		}
	}
	tx, err := sc.cc.SUPS.Transfer(ctx, common.HexToAddress(user.PublicAddress.String), withdrawAmount.BigInt())
	if err != nil {
		refund(err.Error())
		return terror.Error(err, "Withdraw failed: %s. Try again or contact support.", txID.String())
	}
	errChan := make(chan error)
	attemptsChan := make(chan int)
	// we check every 5 seconds on updates to their transaction
	go func() {
		attempts := 0
		for {
			time.Sleep(5 * time.Second)

			attempts++
			// get tx
			rawTx, isPending, err := sc.cc.BscClient.TransactionByHash(ctx, tx.Hash())
			if err != nil {
				attemptsChan <- attempts
				continue
			}

			if isPending {
				attemptsChan <- attempts
				continue
			}
			// if not pending get the tx receipt and check it status
			txReceipt, err := sc.cc.BscClient.TransactionReceipt(ctx, rawTx.Hash())
			if err != nil {
				attemptsChan <- attempts
				continue
			}

			if txReceipt.Status == 0 {
				errChan <- fmt.Errorf("transaction recepit status == 0")
				return
			}

			errChan <- nil
			return
		}
	}()
	// wait for either 60 seconds or errChan
	// if errChan == nil the transaction was successful
	// if errChan wasn't nil, refund
	// if attempts hit 12 (60 seconds) then return true
	// if it does fail after 60 seconds they will need to contact support
	for {
		select {
		case attempts := <-attemptsChan:
			if attempts == 12 {
				reply(true)
				return nil
			}
		case err := <-errChan:
			if err != nil {
				refund(err.Error())
				return terror.Warn(err, "Transaction failed: %s. Try again or contact support.", err.Error())
			}
			reply(true)
			return nil
		}
	}
}
