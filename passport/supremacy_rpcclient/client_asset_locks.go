package supremacy_rpcclient

import (
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"strings"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"
)

type AssetLockToSupremacyResp struct {
}

type AssetLockToSupremacyReq struct {
	ApiKey          string `json:"api_key,omitempty"`
	OwnerID         string `json:"owner_id,omitempty"`
	Hash            string `json:"hash,omitempty"`
	TransferEventID int64  `json:"transfer_event_id"`
	MarketLocked    bool   `json:"market_locked"`
}

// AssetLockToSupremacy requests an asset to be locked on supremacy
func AssetLockToSupremacy(assetToLock *types.UserAsset, transferEventID int64, marketLocked bool) error {
	req := &AssetLockToSupremacyReq{
		OwnerID:         assetToLock.OwnerID,
		Hash:            assetToLock.Hash,
		TransferEventID: transferEventID,
		MarketLocked:    marketLocked,
	}
	resp := &AssetLockToSupremacyResp{}
	err := SupremacyClient.Call("S.AssetLockToSupremacyHandler", req, resp)
	if err != nil {
		passlog.L.Error().Err(err).Interface("assetToLock", assetToLock).Bool("marketLocked", marketLocked).Msg("failed to lock asset on supremacy")
		return terror.Error(err, "communication to supremacy has failed")
	}

	return nil
}

type AssetUnlockFromSupremacyResp struct {
}

type AssetUnlockFromSupremacyReq struct {
	ApiKey          string `json:"api_key,omitempty"`
	OwnerID         string `json:"owner_id,omitempty"`
	Hash            string `json:"hash,omitempty"`
	TransferEventID int64  `json:"transfer_event_id"`
}

// AssetUnlockFromSupremacy request an un-lock of an asset
func AssetUnlockFromSupremacy(assetToUnlock *types.UserAsset, transferEventID int64) error {
	req := &AssetUnlockFromSupremacyReq{
		OwnerID:         assetToUnlock.OwnerID,
		Hash:            assetToUnlock.Hash,
		TransferEventID: transferEventID,
	}

	resp := &GenesisOrLimitedMechResp{}
	err := SupremacyClient.Call("S.AssetUnlockFromSupremacyHandler", req, resp)
	if err != nil {
		if strings.Contains(err.Error(), "asset is equipped to another object") {
			return terror.Warn(err)
		}
		passlog.L.Error().Err(err).Interface("assetToUnlock", assetToUnlock).Msg("failed to unlock asset on supremacy")
		return terror.Error(err, "communication to supremacy has failed")
	}

	return nil
}

type Asset1155LockToSupremacyReq struct {
	ApiKey          string `json:"api_key,omitempty"`
	OwnerID         string `json:"owner_id,omitempty"`
	Amount          int    `json:"amount"`
	TokenID         int    `json:"token_id"`
	TransferEventID int64  `json:"transfer_event_id"`
}

// KeycardsTransferToSupremacy request an un-lock of an asset
func KeycardsTransferToSupremacy(asset *types.User1155Asset, transferEventID int64, amount int) error {
	req := &Asset1155LockToSupremacyReq{
		OwnerID:         asset.OwnerID,
		Amount:          amount,
		TokenID:         asset.ExternalTokenID,
		TransferEventID: transferEventID,
	}

	err := SupremacyClient.Call("S.KeycardTransferToSupremacyHandler", req, &Asset1155FromSupremacyResp{})
	if err != nil {
		passlog.L.Error().Err(err).Interface("asset", asset).Msg("failed to unlock asset on supremacy")
		return terror.Error(err, "communication to supremacy has failed")
	}

	return nil
}

type Asset1155FromSupremacyResp struct {
	Label        string
	Description  string
	ImageURL     string
	AnimationURL null.String
	KeycardGroup string
	Syndicate    null.String
	Count        int
}

// KeycardsTransferFromSupremacy request an un-lock of an asset
func KeycardsTransferFromSupremacy(asset *types.User1155Asset, transferEventID int64, amount int) (*Asset1155FromSupremacyResp, error) {
	req := &Asset1155LockToSupremacyReq{
		OwnerID:         asset.OwnerID,
		Amount:          amount,
		TokenID:         asset.ExternalTokenID,
		TransferEventID: transferEventID,
	}

	resp := &Asset1155FromSupremacyResp{}
	err := SupremacyClient.Call("S.KeycardTransferToXsynHandler", req, resp)
	if err != nil {
		passlog.L.Error().Err(err).Interface("asset", asset).Msg("failed to unlock asset on supremacy")
		return nil, terror.Error(err, "communication to supremacy has failed")
	}

	return resp, nil
}
