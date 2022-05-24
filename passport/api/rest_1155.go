package api

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
	"math/big"
	"net/http"
	"time"
	"xsyn-services/passport/api/users"
)

func (api *API) Withdraw1155(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	toAddress := common.HexToAddress(address)

	tokenID := chi.URLParam(r, "token_id")
	if tokenID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing external token id"), "Missing external token id.")
	}

	nonce := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	amount := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	// check balance
	user, err := users.PublicAddress(common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	isLocked := user.CheckUserIsLocked("withdrawals")
	if isLocked {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user: %s, attempting to withdraw while account is locked.", user.ID), "Withdrawals is locked, contact support to unlock.")
	}

	amountBigInt := new(big.Int)
	_, ok := amountBigInt.SetString(amount, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse amount to big int"), "Invalid amount.")
	}

	nonceBigInt := new(big.Int)
	_, ok = amountBigInt.SetString(nonce, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse amount to big int"), "Invalid amount.")
	}

	expiry := time.Now().Add(5 * time.Minute)
	signer := bridge.NewSigner(api.BridgeParams.SignerPrivateKey)
	_, messageSig, err := signer.GenerateSignatureWithExpiry(toAddress, amountBigInt, nonceBigInt, big.NewInt(expiry.Unix()))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	err = json.NewEncoder(w).Encode(struct {
		MessageSignature string `json:"messageSignature"`
		Expiry           int64  `json:"expiry"`
	}{
		MessageSignature: hexutil.Encode(messageSig),
		Expiry:           expiry.Unix(),
	})
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
