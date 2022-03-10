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

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"
)

func (api *API) NFTRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/check", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Get("/owner_address/{owner_address}/nonce/{nonce}/collection_slug/{collection_slug}/token_id/{external_token_id}", WithError(api.MintAsset))
	r.Post("/owner_address/{owner_address}/collection_slug/{collection_slug}/token_id/{external_token_id}", WithError(api.LockNFT))
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
// r.Get("/mint-nft/{owner_address}/{nonce}/{collection_slug}/{external_token_id}", WithError(api.MintAsset))
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

	//toAddress := common.HexToAddress(address)

	nonceBigInt := new(big.Int)
	_, ok := nonceBigInt.SetString(nonce, 10)
	if !ok {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("failed to parse nonce to big int"), "Invalid nonce.")
	}

	// check user owns this asset
	user, err := db.UserByPublicAddress(context.Background(), api.Conn, common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	// get collection details
	collection, err := db.CollectionBySlug(context.Background(), api.Conn, collectionSlug)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection.")
	}
	isMinted, err := db.PurchasedItemIsMinted(common.HexToAddress(collection.MintContract.String), int(tokenIDuint64))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to check mint status.")
	}
	if isMinted {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("already minted: %s %s", collection.MintContract.String, tokenID), "NFT already minted")
	}

	item, err := db.PurchasedItemByMintContractAndTokenID(common.HexToAddress(collection.MintContract.String), int(tokenIDuint64))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset.")
	}
	if item.OwnerID != user.ID.String() {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("unable to validate ownership of asset"), "Unable to validate ownership of asset.")
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

	// Lock item for 5 minutes
	item, err = db.PurchasedItemLock(uuid.Must(uuid.FromString(item.ID)))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not lock item.")
	}

	go api.MessageBus.Send(context.Background(), messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyAssetSubscribe, item.Hash)), &AssetUpdatedSubscribeResponse{
		CollectionSlug: collectionSlug,
		PurchasedItem:  item,
		OwnerUsername:  address,
	})

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

func (api *API) LockNFT(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "owner_address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing address"), "Missing address.")
	}

	collectionSlug := chi.URLParam(r, "collection_slug")
	if collectionSlug == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing collection slug"), "Missing Collection slug.")
	}

	tokenIDStr := chi.URLParam(r, "external_token_id")
	if tokenIDStr == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing tokenID"), "Missing tokenID.")
	}

	collection, err := db.CollectionBySlug(context.Background(), api.Conn, collectionSlug)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection.")
	}

	tokenID, err := strconv.Atoi(tokenIDStr)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to convert token id.")
	}

	item, err := db.PurchasedItemByMintContractAndTokenID(common.HexToAddress(collection.MintContract.String), tokenID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get asset.")
	}

	// Lock item for 5 minutes
	item, err = db.PurchasedItemLock(uuid.Must(uuid.FromString(item.ID)))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Could not lock item.")
	}

	// check user owns this asset
	user, err := db.UserByPublicAddress(context.Background(), api.Conn, common.HexToAddress(address))
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to find user with this wallet address.")
	}

	go api.MessageBus.Send(r.Context(), messagebus.BusKey(fmt.Sprintf("%s:%v", HubKeyAssetSubscribe, item.Hash)), &AssetUpdatedSubscribeResponse{
		PurchasedItem:  item,
		OwnerUsername:  user.PublicAddress.String,
		CollectionSlug: collection.Slug,
	})
	return http.StatusOK, nil
}