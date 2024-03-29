package asset

import (
	"database/sql"
	"fmt"
	"github.com/friendsofgo/errors"
	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
)

func TransferAsset(
	assetHash,
	fromID,
	toID,
	serviceID string,
	updateServiceID bool,
	relatedTransactionID null.String,
	assetTransferNotify func(te *boiler.AssetTransferEvent),
) (*boiler.UserAsset, int64, error) {
	// get asset
	userAsset, err := boiler.UserAssets(
		boiler.UserAssetWhere.Hash.EQ(assetHash),
		boiler.UserAssetWhere.OwnerID.EQ(fromID),
	).One(passdb.StdConn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userAsset, 0, fmt.Errorf("asset not exist")
		}
		passlog.L.Error().Err(err).
			Str("assetHash", assetHash).
			Str("fromID", fromID).
			Str("toID", toID).
			Str("serviceID", serviceID).
			Msg("failed to get user asset - TransferAsset")
		return userAsset, 0, err
	}

	if updateServiceID && serviceID != "" && userAsset.LockedToService.Valid && userAsset.LockedToService.String != serviceID {
		err := fmt.Errorf("cannot transfer asset the service doesn't control")
		passlog.L.Error().Err(err).
			Str("assetHash", assetHash).
			Str("fromID", fromID).
			Str("toID", toID).
			Str("serviceID", serviceID).
			Msg("failed to transfer asset ownership - TransferAsset")
		return userAsset, 0, err
	}

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetHash).
			Str("fromID", fromID).
			Str("toID", toID).
			Str("serviceID", serviceID).
			Msg("failed to begin tx - TransferAsset")
		return userAsset, 0, err
	}

	userAsset.OwnerID = toID
	if updateServiceID {
		userAsset.LockedToService = null.String{}
		if serviceID != "" {
			userAsset.LockedToService = null.StringFrom(serviceID)
		}
	}

	_, err = userAsset.Update(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetHash).
			Str("fromID", fromID).
			Str("toID", toID).
			Str("serviceID", serviceID).
			Interface("userAsset", userAsset).
			Msg("failed to update asset ownership - TransferAsset")
		return userAsset, 0, err
	}

	transferEvent := &boiler.AssetTransferEvent{
		UserAssetID:   userAsset.ID,
		UserAssetHash: userAsset.Hash,
		FromUserID:    fromID,
		ToUserID:      toID,
		InitiatedFrom: serviceID,
		TransferTXID:  relatedTransactionID,
	}

	err = transferEvent.Insert(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetHash).
			Str("fromID", fromID).
			Str("toID", toID).
			Str("serviceID", serviceID).
			Interface("transferEvent", transferEvent).
			Msg("failed to insert transferEvent - TransferAsset")
		return userAsset, 0, err
	}

	err = tx.Commit()
	if err != nil {
		passlog.L.Error().Err(err).
			Interface("transferEvent", transferEvent).
			Interface("userAsset", userAsset).
			Str("assetHash", assetHash).
			Str("fromID", fromID).
			Str("toID", toID).
			Str("serviceID", serviceID).
			Msg("failed to commit tx - TransferAsset")
		return userAsset, 0, err
	}

	if assetTransferNotify != nil {
		assetTransferNotify(transferEvent)
	}
	return userAsset, transferEvent.ID, nil
}

// TransferAssetADMIN is used for admins to transfer assets, ignore service id and previous owner
func TransferAssetADMIN(assetID, toID uuid.UUID) (int64, error) {
	// get asset
	userAsset, err := boiler.FindUserAsset(passdb.StdConn, assetID.String())
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetID.String()).
			Str("toID", toID.String()).
			Msg("failed to get user asset - TransferAssetADMIN")
		return 0, err
	}

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetID.String()).
			Str("toID", toID.String()).
			Msg("failed to begin tx - TransferAssetADMIN")
		return 0, err
	}
	oldOwner := userAsset.OwnerID
	userAsset.OwnerID = toID.String()

	_, err = userAsset.Update(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetID.String()).
			Str("toID", toID.String()).
			Interface("userAsset", userAsset).
			Msg("failed to update asset ownership - TransferAssetADMIN")
		return 0, err
	}

	transferEvent := &boiler.AssetTransferEvent{
		UserAssetID:   userAsset.ID,
		UserAssetHash: userAsset.Hash,
		FromUserID:    oldOwner,
		ToUserID:      toID.String(),
	}

	err = transferEvent.Insert(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetID.String()).
			Str("toID", toID.String()).
			Interface("userAsset", userAsset).
			Interface("transferEvent", transferEvent).
			Msg("failed to insert transferEvent - TransferAssetADMIN")
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		passlog.L.Error().Err(err).
			Str("assetHash", assetID.String()).
			Str("toID", toID.String()).
			Interface("userAsset", userAsset).
			Interface("transferEvent", transferEvent).
			Msg("failed to commit tx - TransferAssetADMIN")
		return 0, err
	}

	return transferEvent.ID, nil
}
