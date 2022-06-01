package comms

import (
	"fmt"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"time"
	"xsyn-services/boiler"
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
			Interface("userAsset", userAsset).
			Str("serviceID", serviceID).
			Msg("failed to transfer asset ownership - AssetTransferOwnershipHandler")
		return err
	}

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		passlog.L.Error().Err(err).Interface("req", req).Msg("failed to begin tx - AssetTransferOwnershipHandler")
		return err
	}

	userAsset.OwnerID = req.ToOwnerID

	_, err = userAsset.Update(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Interface("userAsset", userAsset).Msg("failed to update asset ownership - AssetTransferOwnershipHandler")
		return err
	}

	transferEvent := &boiler.AssetTransferEvent{
		UserAssetID:   userAsset.ID,
		UserAssetHash: userAsset.Hash,
		FromUserID:    req.FromOwnerID,
		ToUserID:      req.ToOwnerID,
		InitiatedFrom: serviceID,
		TransferTXID:  req.RelatedTransactionID,
	}

	err = transferEvent.Insert(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).
			Interface("transferEvent", transferEvent).
			Interface("userAsset", userAsset).
			Interface("req", req).
			Msg("failed to insert transferEvent - AssetTransferOwnershipHandler")
		return err
	}

	err = tx.Commit()
	if err != nil {
		passlog.L.Error().Err(err).
			Interface("transferEvent", transferEvent).
			Interface("userAsset", userAsset).
			Interface("req", req).
			Msg("failed to commit tx - AssetTransferOwnershipHandler")
		return err
	}

	resp.TransferEventID = transferEvent.ID
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
