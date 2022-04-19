package comms

import (

"github.com/ninja-software/terror/v2"
"strings"
"xsyn-services/boiler"
"xsyn-services/passport/passdb"
"xsyn-services/passport/passlog"
)"xsyn-services/passport/passlog"


)types2 "xsyn-services/types"


	"github.com/ninja-software/terror/v2"
)



type AssetOnChainStatusReq struct {
	AssetHash string `json:"asset_hash"`
}


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


type AssetsOnChainStatusReq struct {
	AssetHashes []string `json:"asset_hashes"`
}


type AssetsOnChainStatusResp struct {
	OnChainStatuses map[string]string `json:"on_chain_statuses"`
}

func (s *S) AssetsOnChainStatus(req AssetsOnChainStatusReq, resp *AssetsOnChainStatusResp) error {
	items, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.Hash.IN(req.AssetHashes)).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetHashes", strings.Join(req.AssetHashes, ", ")).Err(err).Msg("failed to get assets")
		return terror.Error(err)
	}

	assetMap := make(map[string]string)
	for _, asset := range items {
		assetMap[asset.Hash] = asset.OnChainStatus
	}

	resp.OnChainStatuses = assetMap
	return nil
}
