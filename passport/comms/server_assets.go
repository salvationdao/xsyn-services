package comms

import (
	"strings"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
)

func (s *S) AssetOnChainStatusHandler(req AssetOnChainStatusReq, resp *AssetOnChainStatusResp) error {
	item, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.ID.EQ(req.AssetID)).One(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetID", req.AssetID).Err(err).Msg("failed to get asset")
		return err
	}

	resp.OnChainStatus = item.OnChainStatus
	return nil
}

func (s *S) AssetsOnChainStatusHandler(req AssetsOnChainStatusReq, resp *AssetsOnChainStatusResp) error {
	items, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.ID.IN(req.AssetIDs)).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Str("req.AssetIDs", strings.Join(req.AssetIDs, ", ")).Err(err).Msg("failed to get assets")
		return err
	}

	assetMap := make(map[string]string)
	for _, asset := range items {
		assetMap[asset.ID] = asset.OnChainStatus
	}

	resp.OnChainStatuses = assetMap
	return nil
}
