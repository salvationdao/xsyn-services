package comms

import (
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
)

type AssetLockToServiceResp struct {

}

type AssetLockToServiceReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}


// AssetLockToServiceHandler request a service lock of an asset
func (s *S) AssetLockToServiceHandler(req AssetLockToServiceReq, resp *AssetLockToServiceResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to service lock asst")
		return err
	}

	// get collection
	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(req.CollectionSlug)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Str("req.CollectionSlug",req.CollectionSlug).Msg("failed to get collection - AssetLockToServiceHandler")
		return err
	}

	// get asset
	userAsset, err := boiler.UserAssets(
		boiler.UserAssetWhere.OwnerID.EQ(req.OwnerID),
		boiler.UserAssetWhere.TokenID.EQ(req.TokenID),
		boiler.UserAssetWhere.Hash.EQ(req.Hash),
		boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
		).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Interface("req", req).Msg("failed to get user asset - AssetLockToServiceHandler")
		return err
	}

	// check if asset is lockable
	// already locked by this service
	if userAsset.LockedToService.Valid && userAsset.LockedToService.String == serviceID {
		passlog.L.Warn().Msg("asset is already locked by this service - AssetUnlockFromServiceHandler")
		return nil
	}
	// already locked by different service
	if userAsset.LockedToService.Valid && userAsset.LockedToService.String != serviceID {
		err := fmt.Errorf("attempted to lock user asset that is locked to a different service")
		passlog.L.Error().Err(err).Str("userAsset.LockedToService.String", userAsset.LockedToService.String).Str("serviceID", serviceID).Msg("asset is locked by another service - AssetLockToServiceHandler")
		return err
	}

	return nil
}

type AssetUnlockToServiceResp struct {
}

type AssetUnlockToServiceReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	CollectionSlug string `json:"collection_slug,omitempty"`
	TokenID        int64  `json:"token_id,omitempty"`
	OwnerID        string `json:"owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}

// AssetUnlockFromServiceHandler request a service unlock of an asset
func (s *S) AssetUnlockFromServiceHandler(req AssetUnlockToServiceReq, resp *AssetUnlockToServiceResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to service unlock asset")
		return err
	}


	// get collection
	collection, err := boiler.Collections(boiler.CollectionWhere.Slug.EQ(req.CollectionSlug)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Str("req.CollectionSlug",req.CollectionSlug).Msg("failed to get collection - AssetUnlockFromServiceHandler")
		return err
	}

	// get asset
	userAsset, err := boiler.UserAssets(
		boiler.UserAssetWhere.OwnerID.EQ(req.OwnerID),
		boiler.UserAssetWhere.TokenID.EQ(req.TokenID),
		boiler.UserAssetWhere.Hash.EQ(req.Hash),
		boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
	).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Interface("req", req).Msg("failed to get user asset - AssetUnlockFromServiceHandler")
		return err
	}

	if !userAsset.LockedToService.Valid {
		err := fmt.Errorf("attempted to unlock user asset that is already unlocked")
		passlog.L.Error().Err(err).Str("userAsset.ID", userAsset.ID).Msg("asset is already unlocked - AssetUnlockFromServiceHandler")
		return err
	}

	if userAsset.LockedToService.String != serviceID {
		err := fmt.Errorf("attempted to unlock user asset that is locked to a different service")
		passlog.L.Error().Err(err).Str("userAsset.LockedToService.String", userAsset.LockedToService.String).Str("serviceID", serviceID).Msg("asset is locked by another service - AssetUnlockFromServiceHandler")
		return err
	}

	// unlock!
	userAsset.LockedToService = null.NewString("", false)
	_, err = userAsset.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Interface("req", req).Msg("failed to update user asset lock - AssetUnlockFromServiceHandler")
		return err
	}

	return nil
}
