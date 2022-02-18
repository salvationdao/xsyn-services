package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"passport"
	"passport/db"
	"passport/log_helpers"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"

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

	//// sups list
	//api.Command(HubKeySupList, supHub.SupListHandler)
	//
	//// sup subscribe
	//api.SubscribeCommand(HubKeySupSubscribe, supHub.SupUpdatedSubscribeHandler)
	//
	//// sup set name
	//api.SecureCommand(HubKeySupUpdateName, supHub.SupUpdateNameHandler)
	//
	api.SecureCommand(HubKeyWithdrawSups, supHub.WithdrawSupHandler)
	//api.SecureCommand(HubKeySupQueueLeave, supHub.LeaveQueueHandler)
	//api.SecureCommand(HubKeySupInsurancePay, supHub.PaySupInsuranceHandler)
	//api.SecureUserSubscribeCommand(HubKeySupQueueContractReward, supHub.SupQueueContractRewardSubscriber)

	return supHub
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

	fmt.Println("here123123123")
	if sc.cc.SUPS == nil {
		return terror.Error(fmt.Errorf("sups controller not initalized"), "Internal error, try again or contact support.")
	}

	fmt.Println("1")
	withdrawAmount := big.NewInt(0)
	if _, ok := withdrawAmount.SetString(req.Payload.Amount, 10); !ok {
		return terror.Error(fmt.Errorf("failed to create big int from amount"), "Issue getting amount.")
	}
	fmt.Println("2")
	userID := passport.UserID(uuid.FromStringOrNil(hubc.Identifier()))
	if userID.IsNil() {
		return terror.Error(terror.ErrForbidden)
	}
	fmt.Println("3")
	user, err := db.UserGet(ctx, sc.Conn, userID)
	if userID.IsNil() {
		return terror.Error(err)
	}
	fmt.Println("4")
	if !user.PublicAddress.Valid || user.PublicAddress.String == "" {
		return terror.Error(fmt.Errorf("user has no public address"), "Account missing public address.")
	}
	fmt.Println("5")
	// if balance too low
	if user.Sups.Cmp(withdrawAmount) < 0 {
		return terror.Error(fmt.Errorf("user tried to withdraw without enough funds"), "Insufficient funds.")
	}
	fmt.Println("6")
	// check withdraw wallet sups
	balance, err := sc.cc.SUPS.Balance()
	if err != nil {
		return terror.Error(err)
	}
	fmt.Println("7")

	// if wallet sups balance too low
	if balance.Cmp(withdrawAmount) < 0 {
		return terror.Error(fmt.Errorf("not enough funds in our withdraw wallet"), "Insufficient in funds in withdraw wallet at this time.")
	}
	fmt.Println("8")
	// check withdraw wallet gas
	pendingBalance, err := sc.cc.BscClient.PendingBalanceAt(ctx, sc.cc.SUPS.PublicAddress)
	if err != nil {
		return terror.Error(err)
	}

	fmt.Println("9")
	suggestGasPrice, err := sc.cc.BscClient.SuggestGasPrice(ctx)
	if err != nil {
		return terror.Error(err)
	}
	fmt.Println("10")
	if pendingBalance.Cmp(suggestGasPrice) < 0 {
		return terror.Error(fmt.Errorf("not enough gas funds in our withdraw wallet"), "Insufficient gas in withdraw wallet at this time.")
	}
	fmt.Println("11")
	txID := uuid.Must(uuid.NewV4())

	txRef := fmt.Sprintf("sup|withdraw|%s", txID)
	fmt.Println("12")
	resultChan := make(chan *passport.TransactionResult, 1)
	sc.API.transaction <- &passport.NewTransaction{
		To:                   passport.OnChainUserID,
		From:                 userID,
		Amount:               *withdrawAmount,
		TransactionReference: passport.TransactionReference(txRef),
		Description:          "Withdraw of SUPS.",
		ResultChan:           resultChan,
	}
	fmt.Println("13")
	result := <-resultChan
	fmt.Println("14")
	if result.Error != nil {
		return terror.Error(result.Error, "Withdraw failed: %s", result.Error.Error())
	}
	fmt.Println("15")
	if result.Transaction.Status != passport.TransactionSuccess {
		return terror.Error(fmt.Errorf("withdraw failed: %s", result.Transaction.Reason), fmt.Sprintf("Withdraw failed: %s.", result.Transaction.Reason))
	}
	fmt.Println("16")
	// refund callback
	refund := func(reason string) {
		sc.API.transaction <- &passport.NewTransaction{
			To:                   userID,
			From:                 passport.OnChainUserID,
			Amount:               *withdrawAmount,
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND %s - %s", reason, txRef)),
			Description:          "Refund of Withdraw of SUPS.",
		}
	}
	fmt.Println("17")
	tx, err := sc.cc.SUPS.Transfer(ctx, common.HexToAddress(user.PublicAddress.String), withdrawAmount)
	if err != nil {
		refund(err.Error())
		return terror.Error(err, "Withdraw failed: %s", result.Error.Error())
	}
	fmt.Println("18")
	errChan := make(chan error)
	attemptsChan := make(chan int)
	// we check every 5 seconds on updates to their transaction
	go func() {
		attempts := 0
		for {
			fmt.Printf("attempt: %d \n", attempts)
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
	fmt.Println("19")
	// wait for either 60 seconds or errChan
	// if errChan == nil the transaction was successful
	// if errChan wasn't nil, refund
	// if attempts hit 12 (60 seconds) then return true
	// if it does fail after 60 seconds they will need to contact support
	for {
		select {
		case attempts := <-attemptsChan:
			if attempts == 12 {
				fmt.Println("20")
				reply(true)
				return nil
			}
		case err := <-errChan:
			if err != nil {
				fmt.Println("21")
				refund(err.Error())
				return terror.Warn(err, "Transaction failed: %s", err.Error())
			}
			fmt.Println("22")
			reply(true)
			return nil
		}
	}
}
