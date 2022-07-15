package comms

import (
	"xsyn-services/boiler"
	"xsyn-services/passport/asset"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	xsynTypes "xsyn-services/types"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type AssetTransferOwnershipResp struct {
	TransferEventID int64 `json:"transfer_event_id"`
}

type AssetTransferOwnershipReq struct {
	ApiKey               string      `json:"api_key,omitempty"`
	FromOwnerID          string      `json:"from_owner_id,omitempty"`
	ToOwnerID            string      `json:"to_owner_id,omitempty"`
	Hash                 string      `json:"hash,omitempty"`
	RelatedTransactionID null.String `json:"related_transaction_id"`
}

// AssetTransferOwnershipHandler request an ownership transfer of an asset
func (s *S) AssetTransferOwnershipHandler(req AssetTransferOwnershipReq, resp *AssetTransferOwnershipResp) error {
	serviceID, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - AssetTransferOwnershipHandler")
		return err
	}

	_, transferID, err := asset.TransferAsset(req.Hash, req.FromOwnerID, req.ToOwnerID, serviceID, req.RelatedTransactionID, nil)
	if err != nil {
		passlog.L.Error().Err(err).Interface("req", req).Msg("failed to transfer asset - AssetTransferOwnershipHandler")
		return err
	}

	resp.TransferEventID = transferID
	return nil
}

type GetAssetTransferEventsResp struct {
	TransferEvents []*xsynTypes.TransferEvent `json:"transfer_events"`
}

type GetAssetTransferEventsReq struct {
	ApiKey      string `json:"api_key"`
	FromEventID int64  `json:"from_event_id"`
}

// GetAssetTransferEventsHandler request all asset events since given int64
func (s *S) GetAssetTransferEventsHandler(req GetAssetTransferEventsReq, resp *GetAssetTransferEventsResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - GetAssetTransferEventsHandler")
		return err
	}

	transferEvents, err := boiler.AssetTransferEvents(
		boiler.AssetTransferEventWhere.ID.GT(req.FromEventID),
		qm.Load(boiler.AssetTransferEventRels.UserAsset),
	).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).
			Interface("req", req).
			Msg("failed to get events - GetAssetTransferEventsHandler")
		return err
	}

	var events []*xsynTypes.TransferEvent
	for _, te := range transferEvents {
		evt := &xsynTypes.TransferEvent{
			TransferEventID: te.ID,
			AssetHash:       te.UserAssetHash,
			FromUserID:      te.FromUserID,
			ToUserID:        te.ToUserID,
			TransferredAt:   te.TransferredAt,
			TransferTXID:    te.TransferTXID,
		}

		if te.R != nil && te.R.UserAsset != nil {
			evt.OwnedService = te.R.UserAsset.LockedToService
		}
		events = append(events, evt)
	}

	resp.TransferEvents = events
	return nil
}
