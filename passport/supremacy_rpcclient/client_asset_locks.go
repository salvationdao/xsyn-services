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
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

// AssetLockToSupremacy requests an asset to be locked on supremacy
func AssetLockToSupremacy(assetToLock *types.UserAsset, collectionSlug string) error {
	req := &AssetLockToSupremacyReq{
		CollectionSlug: collectionSlug,
		TokenID:        assetToLock.TokenID,
		OwnerID:        assetToLock.OwnerID,
		Hash:           assetToLock.Hash,
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
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

// AssetUnlockFromSupremacy request a unlock of an asset
func AssetUnlockFromSupremacy(assetToUnlock *types.UserAsset, collectionSlug string) error {
	req := &AssetUnlockFromSupremacyReq{
		CollectionSlug: collectionSlug,
		TokenID:        assetToUnlock.TokenID,
		OwnerID:        assetToUnlock.OwnerID,
		Hash:           assetToUnlock.Hash,
	}

	resp := &GenesisOrLimitedMechResp{}
	err := SupremacyClient.Call("S.AssetUnlockFromSupremacyHandler", req, resp)
	if err != nil {
		passlog.L.Error().Err(err).Interface("assetToUnlock", assetToUnlock).Str("collectionSlug", collectionSlug).Msg("failed to unlock asset on supremacy")
		return  terror.Error(err, "communication to supremacy has failed")
	}

	return nil
}
