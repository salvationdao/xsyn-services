package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"passport/db"
	"strconv"
	"time"

	"github.com/ninja-syndicate/hub/ext/messagebus"

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
	user, err := db.UserByPublicAddress(r.Context(), api.Conn, address)
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

	err = json.NewEncoder(w).Encode(struct {
		MessageSignature string `json:"messageSignature"`
		Expiry           int64  `json:"expiry"`
	}{
		MessageSignature: hexutil.Encode(messageSig),
		Expiry:           expiry.Unix(),
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	return http.StatusOK, nil
}

// MintAsset
// Flow to mint asset
// get nonce from nft contract (front end)
// send nonce, token id and user wallet addr to server (front end)
// server validates they own that token id (here)
// server generates a sig and returns it (here)
// server locks that asset, so it cannot be used or traded on world
// submit that sig to eft contract signedMint func
// listen on server for update
func (api *API) MintAsset(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	nonce := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	tokenID := chi.URLParam(r, "externalTokenID")
	if tokenID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tokenID"), "Missing tokenID.")
	}

	collectionSlug := chi.URLParam(r, "collectionSlug")
	if collectionSlug == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing collection slug"), "Missing Collection slug.")
	}

	tokenIDuint64, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to convert token id.")
	}

	//toAddress := common.HexToAddress(address)

	nonceBigInt := new(big.Int)
	_, ok := nonceBigInt.SetString(nonce, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse nonce to big int"), "Invalid nonce.")
	}

	// check user owns this asset
	user, err := db.UserByPublicAddress(r.Context(), api.Conn, address)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	// get collection details
	collection, err := db.CollectionGet(r.Context(), api.Conn, collectionSlug)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection.")
	}

	asset, err := db.AssetGetFromContractAndID(r.Context(), api.Conn, collection.MintContract, tokenIDuint64)
	if err != nil {
		fmt.Println("1")
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset.")
	}
	if asset == nil {
		fmt.Println("2")
		return http.StatusBadRequest, terror.Warn(err, "Asset doesn't exist")
	}

	if asset.Minted {
		fmt.Println("3")
		return http.StatusBadRequest, terror.Warn(err, "Asset already minted")
	}

	if asset.UserID != nil && *asset.UserID != user.ID {
		fmt.Println("4")
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("unable to validate ownership of asset"), "Unable to validate ownership of asset.")
	}

	if !asset.IsUsable() {
		fmt.Println("5")
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("asset is locked"), "Asset is locked.")
	}

	tokenAsBigInt := big.NewInt(0)
	tokenAsBigInt.SetUint64(tokenIDuint64)

	//  sign it
	expiry := time.Now().Add(5 * time.Minute)
	signer := bridge.NewSigner(api.BridgeParams.SignerPrivateKey)
	fmt.Println()
	fmt.Println(common.HexToAddress(address))
	fmt.Println(common.HexToAddress(collection.MintContract))
	fmt.Println(tokenAsBigInt.String())
	fmt.Println(nonceBigInt.String())
	fmt.Println(big.NewInt(expiry.Unix()).String())
	fmt.Println()
	_, messageSig, err := signer.GenerateSignatureWithExpiryAndCollection(common.HexToAddress(address), common.HexToAddress(collection.MintContract), tokenAsBigInt, nonceBigInt, big.NewInt(expiry.Unix()))
	if err != nil {
		fmt.Println("6")
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	// mint lock asset
	err = db.XsynAssetMintLock(r.Context(), api.Conn, asset.Hash, hexutil.Encode(messageSig))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to generate signature, please try again.")
	}

	// get updated asset
	asset, err = db.AssetGetFromContractAndID(r.Context(), api.Conn, collection.MintContract, tokenIDuint64)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset.")

	}
	if asset == nil {
		return http.StatusBadRequest, terror.Warn(err, "Asset doesn't exist")
	}
	go api.MessageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetSubscribe, tokenID)), asset)

	err = json.NewEncoder(w).Encode(struct {
		MessageSignature string `json:"messageSignature"`
		Expiry           int64  `json:"expiry"`
	}{
		MessageSignature: hexutil.Encode(messageSig),
		Expiry:           expiry.Unix(),
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	return http.StatusOK, nil
}
