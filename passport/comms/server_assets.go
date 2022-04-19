package comms

import (
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"

	"github.com/ninja-software/terror/v2"
)

type AssetOnChainStatusReq struct {
	AssetID string `json:"asset_ID"`
}

type AssetOnChainStatusResp struct {
	OnChainStatus string `json:"on_chain_status"`
}

func (s *S) AssetOnChainStatusHandler(req AssetOnChainStatusReq, resp *AssetOnChainStatusResp) error {
	item, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.ID.EQ(req.AssetID)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetID", req.AssetID).Err(err).Msg("failed to get asset")
		return terror.Error(err)
	}

	resp.OnChainStatus = item.OnChainStatus
	return nil
}

type AssetsOnChainStatusReq struct {
	AssetIDs []string `json:"asset_IDs"`
}

type AssetsOnChainStatusResp struct {
	OnChainStatuses map[string]string `json:"on_chain_statuses"`
}

func (s *S) AssetsOnChainStatusHandler(req AssetsOnChainStatusReq, resp *AssetsOnChainStatusResp) error {
	items, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.ID.IN(req.AssetIDs)).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetIDs", strings.Join(req.AssetIDs, ", ")).Err(err).Msg("failed to get assets")
		return terror.Error(err)
	}

	assetMap := make(map[string]string)
	for _, asset := range items {
		assetMap[asset.ID] = asset.OnChainStatus
	}

	resp.OnChainStatuses = assetMap
	return nil
}
