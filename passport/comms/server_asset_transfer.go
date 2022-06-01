package comms

import (
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
)

type AssetTransferOwnershipResp struct {

}

type AssetTransferOwnershipReq struct {
	ApiKey         string `json:"api_key,omitempty"`
	FromOwnerID        string `json:"from_owner_id,omitempty"`
	ToOwnerID        string `json:"to_owner_id,omitempty"`
	Hash           string `json:"hash,omitempty"`
}


// AssetTransferOwnershipHandler request an ownership transfer of an asset
func (s *S) AssetTransferOwnershipHandler(req AssetTransferOwnershipReq, resp *AssetTransferOwnershipResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to service lock asst")
		return err
	}

	// get asset
	userAsset, err := boiler.UserAssets(
		boiler.UserAssetWhere.Hash.EQ(req.Hash),
		boiler.UserAssetWhere.OwnerID.EQ(req.FromOwnerID),
		).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Interface("req", req).Msg("failed to get user asset - AssetTransferOwnershipHandler")
		return err
	}

	if userAsset.LockedToService.Valid || userAsset.LockedToService.String != serviceID {
		err := fmt.Errorf("cannot transfer asset the service doesn't control")
		passlog.L.Error().Err(err).
			Interface("req", req).
			Interface("userAsset",userAsset).
			Str("serviceID",serviceID).
			Msg("failed to transfer asset ownership - AssetTransferOwnershipHandler")
		return err
	}

	// TODO: insert transfer into events table

	userAsset.OwnerID = req.ToOwnerID

	_, err = userAsset.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Interface("userAsset", userAsset).Msg("failed to update asset ownership - AssetTransferOwnershipHandler")
		return err
	}


	return nil
}
