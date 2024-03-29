package supremacy_rpcclient

import (
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"
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
		TokenID:        tokenID,
	}
	resp := &GenesisOrLimitedMechResp{}
	err := SupremacyClient.Call("S.GenesisOrLimitedMechHandler", req, resp)
	if err != nil {
		return nil, terror.Error(err, "communication to supremacy has failed")
	}

	return resp.Asset, nil
}

type NFT1155DetailsReq struct {
	TokenID        int    `json:"token_id"`
	CollectionSlug string `json:"collection_slug"`
}

type NFT1155DetailsResp struct {
	Label        string      `json:"label"`
	Description  string      `json:"description"`
	ImageURL     string      `json:"image_url"`
	AnimationUrl null.String `json:"animation_url"`
	Group        string      `json:"group"`
	Syndicate    null.String `json:"syndicate"`
}

func Get1155Details(tokenID int, collectionSlug string) (*NFT1155DetailsResp, error) {
	req := &NFT1155DetailsReq{
		CollectionSlug: collectionSlug,
		TokenID:        tokenID,
	}
	resp := &NFT1155DetailsResp{}
	err := SupremacyClient.Call("S.Get1155Details", req, resp)
	if err != nil {
		return nil, terror.Error(err, "communication to supremacy has failed")
	}

	return resp, nil
}

type AssetTransferReq struct {
	TransferEvent *types.TransferEvent `json:"transfer_event"`
}

type AssetTransferResp struct {
	OtherTransferredAssetHashes []string `json:"other_transferred_asset_hashes"`
}

func SupremacyAssetTransferEvent(TransferEvent *types.TransferEvent) ([]string, error) {
	req := &AssetTransferReq{
		TransferEvent: TransferEvent,
	}
	resp := &AssetTransferResp{}
	err := SupremacyClient.Call("S.AssetTransferHandler", req, resp)
	if err != nil {
		return resp.OtherTransferredAssetHashes, terror.Error(err, "communication to supremacy has failed")
	}

	return resp.OtherTransferredAssetHashes, nil
}

type AssetReq struct {
	AssetHash string `json:"asset_hash"`
}

type AssetResp struct {
	Asset *XsynAsset `json:"asset"`
}

func AssetGet(assetHash string) (*XsynAsset, error) {
	req := &AssetReq{
		assetHash,
	}
	resp := &AssetResp{}
	err := SupremacyClient.Call("S.AssetHandler", req, resp)
	if err != nil {
		return nil, terror.Error(err, "communication to supremacy has failed")
	}

	return resp.Asset, nil
}
