package supremacy_rpcclient

import (
	"github.com/ninja-software/terror/v2"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"
)

type AssetLockToSupremacyResp struct {
}

type AssetLockToSupremacyReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
	TransferEventID int64 `json:"transfer_event_id"`
	MarketLocked bool `json:"market_locked"`
}

// AssetLockToSupremacy requests an asset to be locked on supremacy
func AssetLockToSupremacy(assetToLock *types.UserAsset, collectionSlug string, transferEventID int64, marketLocked bool) error {
	req := &AssetLockToSupremacyReq{
		CollectionSlug: collectionSlug,
		OwnerID:        assetToLock.OwnerID,
		Hash:           assetToLock.Hash,
		TransferEventID: transferEventID,
		MarketLocked: marketLocked,
	}
	resp := &AssetLockToSupremacyResp{}
	err := SupremacyClient.Call("S.AssetLockToSupremacyHandler", req, resp)
	if err != nil {
		passlog.L.Error().Err(err).Interface("assetToLock", assetToLock).Str("collectionSlug", collectionSlug).Msg("failed to lock asset on supremacy")
		return  terror.Error(err, "communication to supremacy has failed")
	}

	return nil
}

type AssetUnlockFromSupremacyResp struct {
}

type AssetUnlockFromSupremacyReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
	TransferEventID int64 `json:"transfer_event_id"`
}

// AssetUnlockFromSupremacy request an un-lock of an asset
func AssetUnlockFromSupremacy(assetToUnlock *types.UserAsset, collectionSlug string, transferEventID int64) error {
	req := &AssetUnlockFromSupremacyReq{
		CollectionSlug: collectionSlug,
		OwnerID:        assetToUnlock.OwnerID,
		Hash:           assetToUnlock.Hash,
		TransferEventID: transferEventID,
	}

	resp := &GenesisOrLimitedMechResp{}
	err := SupremacyClient.Call("S.AssetUnlockFromSupremacyHandler", req, resp)
	if err != nil {
		passlog.L.Error().Err(err).Interface("assetToUnlock", assetToUnlock).Str("collectionSlug", collectionSlug).Msg("failed to unlock asset on supremacy")
		return  terror.Error(err, "communication to supremacy has failed")
	}

	return nil
}
