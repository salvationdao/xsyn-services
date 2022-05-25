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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/big"
	"net/http"
	"strconv"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
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
	asset, err := boiler.UserAssets1155S(
		boiler.UserAssets1155Where.ExternalTokenID.EQ(tokenInt),
		boiler.UserAssets1155Where.ServiceID.IsNull(),
		qm.Load(
			boiler.UserAssets1155Rels.Owner,
			boiler.UserWhere.PublicAddress.EQ(null.StringFrom(common.HexToAddress(address).Hex())),
		),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("can't find user asset"), "Can't find asset detils for user")
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

	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to convert amount. Please contract support or try again")
	}

	asset.Count -= amountInt
	if asset.Count < 0 {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("user amount will become less then 0 after update"), "Amount cannot be less than 0 after withdraw")
	}
	_, err = asset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.Count))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update asset details. Please contract support or try again")
	}

	newPendingRollback := boiler.Pending1155Rollback{
		UserID:     user.ID,
		AssetID:    asset.ID,
		Count:      amountInt,
		RefundedAt: time.Now().Add(10 * time.Minute),
	}

	err = newPendingRollback.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to insert pending rollback.")
	}

	err = json.NewEncoder(w).Encode(struct {
		MessageSignature string `json:"messageSignature"`
		Expiry           int64  `json:"expiry"`
	}{
		MessageSignature: hexutil.Encode(messageSig),
		Expiry:           expiry.Unix(),
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
		Contracts []string
	}{
		Contracts: allMintContract,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to encode json. Please try again or contact support")
	}

	return http.StatusOK, nil
}
