package comms

import (
	"github.com/volatiletech/null/v8"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/asset"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
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

	_, transferID, err := asset.TransferAsset(req.Hash, req.FromOwnerID, req.ToOwnerID, serviceID,req.RelatedTransactionID)
	if err != nil {
		passlog.L.Error().Err(err).Interface("req", req).Msg("failed to transfer asset - AssetTransferOwnershipHandler")
		return err
	}

	resp.TransferEventID = transferID
	return nil
}

type TransferEvent struct {
	TransferEventID int64       `json:"transfer_event_id"`
	AssetHast       string      `json:"asset_hast,omitempty"`
	FromUserID      string      `json:"from_user_id,omitempty"`
	ToUserID        string      `json:"to_user_id,omitempty"`
	TransferredAt   time.Time   `json:"transferred_at"`
	TransferTXID    null.String `json:"transfer_tx_id"`
}

type GetAssetTransferEventsResp struct {
	TransferEvents []*TransferEvent `json:"transfer_events"`
}

type GetAssetTransferEventsReq struct {
	ApiKey      string `json:"api_key"`
	FromEventID int64 `json:"from_event_id"`
}

// GetAssetTransferEventsHandler request all asset events since given int64
func (s *S) GetAssetTransferEventsHandler(req GetAssetTransferEventsReq, resp *GetAssetTransferEventsResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to get service id - GetAssetTransferEventsHandler")
		return err
	}

	transferEvents, err := boiler.AssetTransferEvents(boiler.AssetTransferEventWhere.ID.GTE(req.FromEventID)).All(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).
			Interface("req", req).
			Msg("failed to get events - GetAssetTransferEventsHandler")
		return err
	}

	var events []*TransferEvent
	for _, te := range transferEvents {
		events = append(events, &TransferEvent{
			TransferEventID: te.ID,
			AssetHast:       te.UserAssetHash,
			FromUserID:      te.FromUserID,
			ToUserID:        te.ToUserID,
			TransferredAt:   te.TransferredAt,
			TransferTXID:    te.TransferTXID,
		})
	}

	resp.TransferEvents = events
	return nil
}
