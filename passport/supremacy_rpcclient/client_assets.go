package supremacy_rpcclient

import (
	"github.com/ninja-software/terror/v2"
	"xsyn-services/passport/passlog"
)

type GenesisOrLimitedMechReq struct {
	CollectionSlug string
	TokenID        int
}

type GenesisOrLimitedMechResp struct {
	Asset *XsynAsset
}

func GenesisOrLimitedMech(collectionSlug string, tokenID int) (*XsynAsset, error) {
	passlog.L.Trace().Str("fn", "GenesisOrLimitedMech").Msg("db func")
	req := &GenesisOrLimitedMechReq{
		CollectionSlug: collectionSlug,
		TokenID: tokenID,
	}
	resp := &GenesisOrLimitedMechResp{}
	err := SupremacyClient.Call("S.GenesisOrLimitedMechHandler", req, resp)
	if err != nil {
		return nil, terror.Error(err, "communication to supremacy has failed")
	}

	return resp.Asset, nil
}
