package api

import (
	"encoding/json"
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"math/big"
	"net/http"
	"strconv"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

func (api *API) NFTRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/check", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/owner_address/{owner_address}/nonce/{nonce}/collection_slug/{collection_slug}/token_id/{external_token_id}", WithError(api.MintAsset))
	r.Post("/owner_address/{owner_address}/collection_slug/{collection_slug}/token_id/{external_token_id}", WithError(api.LockNFT))
	r.Post("/unstake/owner_address/{owner_address}/collection_slug/{collection_slug}/token_id/{external_token_id}", WithError(api.LockNFT))
	return r
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
	address := chi.URLParam(r, "owner_address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	nonce := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	tokenID := chi.URLParam(r, "external_token_id")
	if tokenID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tokenID"), "Missing tokenID.")
	}

	collectionSlug := chi.URLParam(r, "collection_slug")
	if collectionSlug == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing collection slug"), "Missing Collection slug.")
	}

	tokenIDuint64, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to convert token id.")
	}

	nonceBigInt := new(big.Int)
	_, ok := nonceBigInt.SetString(nonce, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse nonce to big int"), "Invalid nonce.")
	}

	// check user owns this asset
	user, err := users.PublicAddress(common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	isLocked := user.CheckUserIsLocked("minting")
	if isLocked {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user: %s, attempting to mint while account is locked.", user.ID), "Minting assets is locked, contact support to unlock.")
	}

	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(collectionSlug)).One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection.")
	}

	item, err := boiler.UserAssets(
		boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
		boiler.UserAssetWhere.TokenID.EQ(int64(tokenIDuint64)),
		qm.Load(boiler.UserAssetRels.Collection),
		qm.Load(
			boiler.UserAssetRels.Owner,
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		),
		qm.Load(boiler.UserAssetRels.LockedToServiceUser),
		).One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset.")
	}
	if item.OwnerID != user.ID {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("unable to validate ownership of asset"), "Unable to validate ownership of asset.")
	}

	if item.OnChainStatus != string(db.MINTABLE) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("unable to mint asset with status %s", item.OnChainStatus), "Failed to mint asset.")
	}

	if item.UnlockedAt.After(time.Now()) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("asset is locked"), "Asset is locked.")
	}

	if item.LockedToService.Valid {
		service, err := boiler.FindUser(passdb.StdConn, item.LockedToService.String)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to lock asset")
		}
		err = fmt.Errorf("unable to lock asset owned by another service, please transistion from %s first", service.Username)
		return http.StatusBadRequest, terror.Error(err)
	}

	tokenAsBigInt := big.NewInt(0)
	tokenAsBigInt.SetUint64(tokenIDuint64)

	//  sign it
	expiry := time.Now().Add(5 * time.Minute)
	signer := bridge.NewSigner(api.BridgeParams.SignerPrivateKey)

	_, messageSig, err := signer.GenerateSignatureWithExpiryAndCollection(common.HexToAddress(address), common.HexToAddress(collection.MintContract.String), tokenAsBigInt, nonceBigInt, big.NewInt(expiry.Unix()))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	item.UnlockedAt = time.Now().Add(5 * time.Minute)
	_, err = item.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
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

func (api *API) LockNFT(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "owner_address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	// check user owns this asset
	user, err := users.PublicAddress(common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	collectionSlug := chi.URLParam(r, "collection_slug")
	if collectionSlug == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing collection slug"), "Missing Collection slug.")
	}

	tokenIDStr := chi.URLParam(r, "external_token_id")
	if tokenIDStr == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tokenID"), "Missing tokenID.")
	}


	tokenIDuint64, err := strconv.ParseUint(tokenIDStr, 10, 64)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to convert token id.")
	}

	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(collectionSlug)).One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection.")
	}

	item, err := boiler.UserAssets(
		boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
		boiler.UserAssetWhere.TokenID.EQ(int64(tokenIDuint64)),
		qm.Load(boiler.UserAssetRels.Collection),
		qm.Load(
			boiler.UserAssetRels.Owner,
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		),
		qm.Load(boiler.UserAssetRels.LockedToServiceUser),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset.")
	}
	if item.OwnerID != user.ID {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("unable to validate ownership of asset"), "Unable to validate ownership of asset.")
	}

	if item.LockedToService.Valid {
		service, err := boiler.FindUser(passdb.StdConn, item.LockedToService.String)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to lock asset")
		}
		err = fmt.Errorf("unable to lock asset owned by another service, please transistion from %s first", service.Username)
		return http.StatusBadRequest, terror.Error(err)
	}

	item.UnlockedAt = time.Now().Add(5 * time.Minute)
	_, err = item.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}

	return http.StatusOK, nil
}

// UnstakeNFT
// Flow to unstake asset
// get nonce from nft contract (front end)
// send collection address, token id and user wallet addr to server (front end)
// server validates they own that token id (here)
// server generates a sig and returns it (here)
// server locks that asset, so it cannot be used or traded on world
// submit that sig to eft contract signedMint func
// listen on server for update
func (api *API) UnstakeNFT(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "owner_address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	nonce := chi.URLParam(r, "nonce")
	if nonce == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing nonce"), "Missing nonce.")
	}

	tokenID := chi.URLParam(r, "external_token_id")
	if tokenID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tokenID"), "Missing tokenID.")
	}

	collectionSlug := chi.URLParam(r, "collection_slug")
	if collectionSlug == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing collection slug"), "Missing Collection slug.")
	}

	tokenIDuint64, err := strconv.ParseUint(tokenID, 10, 64)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to convert token id.")
	}

	nonceBigInt := new(big.Int)
	_, ok := nonceBigInt.SetString(nonce, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse nonce to big int"), "Invalid nonce.")
	}

	// check user owns this asset
	user, err := users.PublicAddress(common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	isLocked := user.CheckUserIsLocked("minting")
	if isLocked {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("user: %s, attempting to mint while account is locked.", user.ID), "Minting assets is locked, contact support to unlock.")
	}

	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(collectionSlug)).One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection.")
	}

	item, err := boiler.UserAssets(
		boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
		boiler.UserAssetWhere.TokenID.EQ(int64(tokenIDuint64)),
		qm.Load(boiler.UserAssetRels.Collection),
		qm.Load(
			boiler.UserAssetRels.Owner,
			qm.Select(
				boiler.UserColumns.ID,
				boiler.UserColumns.Username,
			),
		),
		qm.Load(boiler.UserAssetRels.LockedToServiceUser),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset.")
	}
	if item.OwnerID != user.ID {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("unable to validate ownership of asset"), "Unable to validate ownership of asset.")
	}

	if item.LockedToService.Valid {
		service, err := boiler.FindUser(passdb.StdConn, item.LockedToService.String)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to lock asset")
		}
		err = fmt.Errorf("unable to lock asset owned by another service, please transistion from %s first", service.Username)
		return http.StatusBadRequest, terror.Error(err)
	}

	if item.OnChainStatus != string(db.UNSTAKABLE) {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("unable to unstake asset with status %s", item.OnChainStatus), "Failed to mint asset.")
	}

	if item.UnlockedAt.After(time.Now()) {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("asset is locked"), "Asset is locked.")
	}

	tokenAsBigInt := big.NewInt(0)
	tokenAsBigInt.SetUint64(tokenIDuint64)

	//  sign it
	expiry := time.Now().Add(5 * time.Minute)
	signer := bridge.NewSigner(api.BridgeParams.SignerPrivateKey)

	_, messageSig, err := signer.GenerateSignatureWithExpiryAndCollection(common.HexToAddress(address), common.HexToAddress(collection.MintContract.String), tokenAsBigInt, nonceBigInt, big.NewInt(expiry.Unix()))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to create withdraw signature, please try again or contact support.")
	}

	item.UnlockedAt = time.Now().Add(5 * time.Minute)
	_, err = item.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
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
