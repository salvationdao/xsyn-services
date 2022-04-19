package comms

import (

"github.com/ninja-software/terror/v2"
"xsyn-services/boiler"
"xsyn-services/passport/passdb""xsyn-services/passport/passlog"


)types2 "xsyn-services/types"


	"github.com/ninja-software/terror/v2"
)



type AssetOnChainStatusReq struct {
	AssetHash string `json:"asset_hash"`
}

//MINTABLE', 'STAKABLE', 'UNSTAKABLE

type AssetOnChainStatusResp struct {
	OnChainStatus string `json:"on_chain_status"`
}

func (s *S) AssetOnChainStatus(req AssetOnChainStatusReq, resp *AssetOnChainStatusResp) error {
	item, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.Hash.EQ(req.AssetHash)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetHash", req.AssetHash).Err(err).Msg("failed to get asset")
		return terror.Error(err)
	}

	resp.OnChainStatus = item.OnChainStatus
	return nil
}
