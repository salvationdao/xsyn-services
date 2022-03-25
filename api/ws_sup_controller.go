package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"time"

	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ninja-software/log_helpers"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

// SupController holds handlers for as
type SupController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
	cc   *ChainClients
}

// NewSupController creates the sup hub
func NewSupController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, cc *ChainClients) *SupController {
	supHub := &SupController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "sup_hub"),
		API:  api,
		cc:   cc,
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

const HubKeyDepositTransactionList hub.HubCommandKey = "SUPS:DEPOSIT:LIST"

func (sc *SupController) DepositTransactionListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err)
	}

	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err)
	}

	userID := passport.UserID(uid)

	dtxs, err := boiler.DepositTransactions(boiler.DepositTransactionWhere.UserID.EQ(userID.String()), qm.Limit(10)).All(passdb.StdConn)
	if err != nil {
		return terror.Error(err)
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
	*hub.HubCommandRequest
	Payload struct {
		TransactionHash string          `json:"transaction_hash"`
		Amount          decimal.Decimal `json:"amount"`
	} `json:"payload"`
}

const HubKeyDepositSups hub.HubCommandKey = "SUPS:DEPOSIT"

func (sc *SupController) DepositSupHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupDepositRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.TransactionHash == "" {
		passlog.L.Error().Str("func", "DepositSupHandler").Msg("deposit transaction hash was not provided")
		return terror.Error(fmt.Errorf("transaction hash was not provided"))
	}

	if req.Payload.Amount.LessThan(decimal.NewFromInt(0)) {
		passlog.L.Error().Str("func", "DepositSupHandler").Msg("deposit transaction amount is lower than the minimum required amount")
		return terror.Error(fmt.Errorf("deposit transaction amount is lower than the minimum required amount"))
	}

	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err)
	}

	userID := passport.UserID(uid)

	dtx := boiler.DepositTransaction{
		UserID: userID.String(),
		TXHash: req.Payload.TransactionHash,
		Amount: req.Payload.Amount,
	}
	err = dtx.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Str("func", "DepositSupHandler").Msg("failed to create deposit transaction in db")
		return terror.Error(err)
	}

	reply(true)
	return nil
}

type SupWithdrawRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Amount string `json:"amount"`
	} `json:"payload"`
}

const HubKeyWithdrawSups hub.HubCommandKey = "SUPS:WITHDRAW"

func (sc *SupController) WithdrawSupHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &SupWithdrawRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if sc.cc.SUPS == nil {
		return terror.Error(fmt.Errorf("sups controller not initalized"), "Internal error, try again or contact support.")
	}

	withdrawAmount := big.NewInt(0)
	if _, ok := withdrawAmount.SetString(req.Payload.Amount, 10); !ok {
		return terror.Error(fmt.Errorf("failed to create big int from amount"), "Issue getting amount.")
	}
	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}
	user, err := db.UserGet(ctx, sc.Conn, userID)
	if userID.IsNil() {
		return terror.Error(err)
	}
	if !user.PublicAddress.Valid || user.PublicAddress.String == "" {
		return terror.Error(fmt.Errorf("user has no public address"), "Account missing public address.")
	}
	// if balance too low
	if user.Sups.Cmp(withdrawAmount) < 0 {
		return terror.Error(fmt.Errorf("user tried to withdraw without enough funds"), "Insufficient funds.")
	}
	// check withdraw wallet sups
	balance, err := sc.cc.SUPS.Balance()
	if err != nil {
		return terror.Error(err)
	}

	// if wallet sups balance too low
	if balance.Cmp(withdrawAmount) < 0 {
		return terror.Error(fmt.Errorf("not enough funds in our withdraw wallet"), "Insufficient in funds in withdraw wallet at this time.")
	}
	// check withdraw wallet gas
	pendingBalance, err := sc.cc.BscClient.PendingBalanceAt(ctx, sc.cc.SUPS.PublicAddress)
	if err != nil {
		return terror.Error(err)
	}

	suggestGasPrice, err := sc.cc.BscClient.SuggestGasPrice(ctx)
	if err != nil {
		return terror.Error(err)
	}
	if pendingBalance.Cmp(suggestGasPrice) < 0 {
		return terror.Error(fmt.Errorf("not enough gas funds in our withdraw wallet"), "Insufficient gas in withdraw wallet at this time.")
	}
	txID := uuid.Must(uuid.NewV4())

	txRef := fmt.Sprintf("sup|withdraw|%s", txID)

	trans := &passport.NewTransaction{
		To:                   passport.OnChainUserID,
		From:                 userID,
		NotSafe:              true,
		Amount:               decimal.NewFromBigInt(withdrawAmount, 0),
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Withdraw of SUPS.",
		Group:                passport.TransactionGroupWithdrawal,
	}

	nfb, ntb, _, err := sc.API.userCacheMap.Transact(trans)
	if err != nil {
		return terror.Error(err, "failed to process user fund")
	}

	if !trans.From.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, trans.From)), nfb.String())
	}
	if !trans.To.IsSystemUser() {
		go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, trans.To)), ntb.String())
	}

	// refund callback
	refund := func(reason string) {
		trans := &passport.NewTransaction{
			To:                   userID,
			NotSafe:              true,
			From:                 passport.OnChainUserID,
			Amount:               decimal.NewFromBigInt(withdrawAmount, 0),
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of Withdraw of SUPS.",
			Group:                passport.TransactionGroupWithdrawal,
		}

		_, _, _, err := sc.API.userCacheMap.Transact(trans)
		if err != nil {
			sc.API.Log.Err(fmt.Errorf("failed to process user fund"))
			return
		}

		if !trans.From.IsSystemUser() {
			go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, trans.From)), nfb.String())
		}
		if !trans.To.IsSystemUser() {
			go sc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, trans.To)), ntb.String())
		}

	}
	tx, err := sc.cc.SUPS.Transfer(ctx, common.HexToAddress(user.PublicAddress.String), withdrawAmount)
	if err != nil {
		refund(err.Error())
		return terror.Error(err, "Withdraw failed: %s", txID.String())
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
				return terror.Warn(err, "Transaction failed: %s", err.Error())
			}
			reply(true)
			return nil
		}
	}
}
