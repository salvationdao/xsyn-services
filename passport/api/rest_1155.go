package api

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/big"
	"net/http"
	"strconv"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
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
	tokenInt, err := strconv.Atoi(tokenID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing external token id"), "Missing external token id.")
	}

	nonce := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	amount := chi.URLParam(r, "amount")
	if amount == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing amount"), "Missing amount.")
	}

	// check balance
	user, err := users.PublicAddress(common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	isLocked := user.CheckUserIsLocked("minting")
	if isLocked {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user: %s, attempting to withdraw while account is locked.", user.ID), "Withdrawals is locked, contact support to unlock.")
	}

	userAsset, err := boiler.UserAssets1155S(
		boiler.UserAssets1155Where.OwnerID.EQ(user.ID),
		boiler.UserAssets1155Where.ExternalTokenID.EQ(tokenInt),
		boiler.UserAssets1155Where.ServiceID.IsNull(),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to get user asset info")
	}

	nonceInt, err := strconv.Atoi(nonce)
	if err != nil {
		return http.StatusBadRequest, err
	}

	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return http.StatusBadRequest, err
	}

	total := userAsset.Count - amountInt

	if total < 0 {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("amount total after withdraw is below 0"), "Amount total after withdraw is less than 0")
	}

	signer := bridge.NewSigner("0x9878e47371dc28d434b8e5a2e36a5ac2fad84af4ebcd8ea34470b2417590e087")
	_, messageSig, err := signer.GenerateSignature(toAddress, big.NewInt(int64(tokenInt)), big.NewInt(int64(nonceInt)))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	err = db.Withdraw1155AssetWithPendingRollback(amountInt, tokenInt, user.ID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to process withdrawal. Please contract support or try again")
	}

	err = json.NewEncoder(w).Encode(struct {
		MessageSignature string `json:"messageSignature"`
	}{
		MessageSignature: hexutil.Encode(messageSig),
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to encode json. Please try again or contact support")
	}
	return http.StatusOK, nil
}

func (api *API) Get1155Contracts(w http.ResponseWriter, r *http.Request) (int, error) {
	contracts, err := boiler.Collections(
		boiler.CollectionWhere.ContractType.EQ(null.String{
			String: "EIP-1155",
			Valid:  true,
		}),
		qm.Select(boiler.CollectionColumns.MintContract),
	).All(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get EIP-1555 contracts. Please try again or contract support")
	}
	var allMintContract []string

	for _, contract := range contracts {
		if contract.MintContract.Valid {
			allMintContract = append(allMintContract, contract.MintContract.String)
		}
	}

	err = json.NewEncoder(w).Encode(struct {
		Contracts []string `json:"contracts"`
	}{
		Contracts: allMintContract,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to encode json. Please try again or contact support")
	}

	return http.StatusOK, nil
}
