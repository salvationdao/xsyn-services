package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"
	"passport/payments"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

// WithdrawSups
// Flow to withdraw sups
// get nonce from withdraw contract
// send nonce, amount and user wallet addr to server,
// server validates they have enough sups
// server generates a sig and returns it
// submit that sig to withdraw contract withdrawSups func
// listen on backend for update
func (api *API) WithdrawSups(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	nonce := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	amount := chi.URLParam(r, "amount")
	if amount == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing amount"), "Missing amount.")
	}

	toAddress := common.HexToAddress(address)

	amountBigInt := new(big.Int)
	_, ok := amountBigInt.SetString(amount, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse amount to big int"), "Invalid amount.")
	}

	nonceBigInt := new(big.Int)
	_, ok = nonceBigInt.SetString(nonce, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse nonce to big int"), "Invalid nonce.")
	}

	// check balance
	user, err := db.UserByPublicAddress(context.Background(), api.Conn, common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	userSups, err := api.userCacheMap.Get(user.ID.String())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get users SUP balance.")
	}
	if userSups.Cmp(amountBigInt) < 0 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user has insufficient funds: %s, %s", userSups.String(), amountBigInt), "Insufficient funds.")
	}

	//  sign it
	expiry := time.Now().Add(5 * time.Minute)
	signer := bridge.NewSigner(api.BridgeParams.SignerPrivateKey)
	_, messageSig, err := signer.GenerateSignatureWithExpiry(toAddress, amountBigInt, nonceBigInt, big.NewInt(expiry.Unix()))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	refundID, err := payments.InsertPendingRefund(api.userCacheMap, user.ID, *amountBigInt, expiry)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	err = json.NewEncoder(w).Encode(struct {
		MessageSignature string `json:"messageSignature"`
		Expiry           int64  `json:"expiry"`
		RefundID         string `json:"refundID"`
	}{
		MessageSignature: hexutil.Encode(messageSig),
		Expiry:           expiry.Unix(),
		RefundID:         refundID,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	return http.StatusOK, nil
}

func (api *API) UpdatePendingRefund(w http.ResponseWriter, r *http.Request) (int, error) {
	refundID := chi.URLParam(r, "refundID")
	if refundID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing refund ID"), "Missing address.")
	}

	txHash := chi.URLParam(r, "txHash")
	if txHash == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tx hash"), "Missing nonce.")
	}

	pendingRefund, err := boiler.FindPendingRefund(passdb.StdConn, refundID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tx hash"), "Missing nonce.")
	}

	pendingRefund.TXHash = txHash
	_, err = pendingRefund.Update(passdb.StdConn, boil.Whitelist(boiler.PendingRefundColumns.TXHash))
	if err != nil {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tx hash"), "Missing nonce.")
	}
	return http.StatusOK, nil
}
