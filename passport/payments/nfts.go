package payments

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"strconv"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/asset"
	"xsyn-services/passport/db"
	"xsyn-services/passport/nft1155"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/supremacy_rpcclient"
	"xsyn-services/types"
)

type NFTOwnerStatus struct {
	Collection     common.Address
	Owner          common.Address
	OnChainStatus  db.OnChainStatus
	TxHash         string
	BlockNumber    int
	BlockTimestamp time.Time
}

func UpdateOwners(nftStatuses map[int]*NFTOwnerStatus, collection *boiler.Collection, environment types.Environment) (int, int, error) {
	l := passlog.L.With().Str("svc", "avant_nft_ownership_update").Logger()

	updated := 0
	skipped := 0

	l.Debug().Int("records", len(nftStatuses)).Msg("processing new owners for NFT")

	for tokenID, nftStatus := range nftStatuses {
		l.Debug().
			Int("token_id", tokenID).
			Str("collection", nftStatus.Collection.Hex()).
			Str("owner", nftStatus.Owner.Hex()).
			Str("on_chain_status", string(nftStatus.OnChainStatus)).
			Msg("processing new owner for NFT")

		// if tx exists, continue
		txExists, err := boiler.ItemOnchainTransactions(
			boiler.ItemOnchainTransactionWhere.TXID.EQ(nftStatus.TxHash),
			boiler.ItemOnchainTransactionWhere.ExternalTokenID.EQ(tokenID),
		).Exists(passdb.StdConn)
		if err != nil {
			return 0, 0, fmt.Errorf("get purchased item: %w", err)
		}
		if txExists {
			continue
		}

		userAsset, err := boiler.UserAssets(
			boiler.UserAssetWhere.CollectionID.EQ(collection.ID),
			boiler.UserAssetWhere.TokenID.EQ(int64(tokenID)),
		).One(passdb.StdConn)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			l.Debug().Err(err).Str("collection_addr", collection.MintContract.String).Int("external_token_id", tokenID).Msg("item not found")
			skipped++
			continue
		} else if err != nil {
			return 0, 0, fmt.Errorf("get purchased item: %w", err)
		}

		onChainStatusObject, err := boiler.UserAssetOnChainStatuses(
			boiler.UserAssetOnChainStatusWhere.CollectionID.EQ(userAsset.CollectionID),
			boiler.UserAssetOnChainStatusWhere.AssetHash.EQ(userAsset.Hash),
		).One(passdb.StdConn)
		if err != nil {
			l.Debug().Err(err).Interface("userAsset", userAsset).Msg("assets on chain status not found")
			skipped++
			continue
		}

		// if a newer tx exists, insert the tx and continue
		// this is for when nodes don't align and avoids assets being able to bounce back due to delayed/stale nodes
		newerExists, err := boiler.ItemOnchainTransactions(
			boiler.ItemOnchainTransactionWhere.CollectionID.EQ(collection.ID),
			boiler.ItemOnchainTransactionWhere.ExternalTokenID.EQ(tokenID),
			boiler.ItemOnchainTransactionWhere.BlockTimestamp.GT(nftStatus.BlockTimestamp), // if timestamp greater than new tx timestamp
		).Exists(passdb.StdConn)
		if err != nil {
			return 0, 0, fmt.Errorf("get purchased item: %w", err)
		}
		if newerExists {
			// insert older and continue
			newItemOnChainTransaction := &boiler.ItemOnchainTransaction{
				CollectionID:    collection.ID,
				ExternalTokenID: tokenID,
				TXID:            nftStatus.TxHash,
				ContractAddr:    collection.MintContract.String,
				FromAddr:        userAsset.OwnerID,
				ToAddr:          nftStatus.Owner.Hex(),
				BlockNumber:     nftStatus.BlockNumber,
				BlockTimestamp:  nftStatus.BlockTimestamp,
			}
			err = newItemOnChainTransaction.Insert(passdb.StdConn, boil.Infer())
			if err != nil {
				passlog.L.Error().Err(err).Interface("newItemOnChainTransaction", newItemOnChainTransaction).Msg("failed to insert new on chain tx history")
			}

			continue
		}

		// on chain user may not exist in our db
		onChainOwner, err := CreateOrGetUser(nftStatus.Owner, environment)
		if err != nil {
			return 0, 0, fmt.Errorf("get or create onchain user: %w", err)
		}

		// of chain user has to exist, it is the current owner
		offChainOwner, err := boiler.FindUser(passdb.StdConn, userAsset.OwnerID)
		if err != nil {
			return 0, 0, fmt.Errorf("get offchain user: %w", err)
		}

		offChainAddr := common.HexToAddress(offChainOwner.PublicAddress.String)
		onChainAddr := common.HexToAddress(onChainOwner.PublicAddress.String)

		l.Debug().
			Str("off_chain_user", offChainAddr.Hex()).
			Str("on_chain_user", onChainAddr.Hex()).
			Bool("matches", offChainAddr.Hex() == onChainAddr.Hex()).
			Msg("check if nft owners match")

		updatedBool := false

		// if the owner is different, transfer asset to new owner
		if offChainAddr.Hex() != onChainAddr.Hex() {
			l.Debug().
				Str("new_owner", onChainOwner.ID).
				Str("old_owner", offChainOwner.ID).
				Str("item_id", userAsset.ID).
				Msg("setting new nft owner")

			userAsset, _, err = asset.TransferAsset(
				userAsset.Hash,
				offChainOwner.ID,
				onChainOwner.ID,
				"",
				true,
				null.String{},
				func(te *boiler.AssetTransferEvent) {
					otherAssets, _ := supremacy_rpcclient.SupremacyAssetTransferEvent(&types.TransferEvent{
						TransferEventID: te.ID,
						AssetHash:       te.UserAssetHash,
						FromUserID:      te.FromUserID,
						ToUserID:        te.ToUserID,
						TransferredAt:   te.TransferredAt,
						TransferTXID:    te.TransferTXID,
					})
					for _, othAsstHash := range otherAssets {
						_, _, err = asset.TransferAsset(
							othAsstHash,
							offChainOwner.ID,
							onChainOwner.ID,
							types.SupremacyGameUserID.String(),
							false, // we don't want to change the service id
							null.String{},
							nil,
						)
						if err != nil {
							passlog.L.Error().Err(err).
								Str("othAsstHash", othAsstHash).
								Str("offChainOwner.ID", offChainOwner.ID).
								Str("onChainOwner.ID", onChainOwner.ID).Msg("failed to transfer attached assets")
						}
					}
				},
			)
			if err != nil {
				passlog.L.Error().Err(err).
					Str("userAsset.Hash", userAsset.Hash).
					Str("offChainOwner.ID", offChainOwner.ID).
					Str("onChainOwner.ID", onChainOwner.ID).
					Msg("failed to transfer asset - UpdateOwners")
				return 0, 0, fmt.Errorf("set new nft owner: %w", err)
			}
			updatedBool = true
		}

		if string(nftStatus.OnChainStatus) != onChainStatusObject.OnChainStatus {
			onChainStatusObject.OnChainStatus = string(nftStatus.OnChainStatus)
			_, err = onChainStatusObject.Update(passdb.StdConn, boil.Infer())
			if err != nil {
				return 0, 0, err
			}
			_, err = userAsset.Update(passdb.StdConn, boil.Infer())
			if err != nil {
				return 0, 0, err
			}
			updatedBool = true
		}
		if updatedBool {
			updated++
		}

		// insert older and continue
		newItemOnChainTransaction := &boiler.ItemOnchainTransaction{
			CollectionID:    collection.ID,
			ExternalTokenID: tokenID,
			TXID:            nftStatus.TxHash,
			ContractAddr:    collection.MintContract.String,
			FromAddr:        userAsset.OwnerID,
			ToAddr:          nftStatus.Owner.Hex(),
			BlockNumber:     nftStatus.BlockNumber,
			BlockTimestamp:  nftStatus.BlockTimestamp,
		}
		err = newItemOnChainTransaction.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			passlog.L.Error().Err(err).Interface("newItemOnChainTransaction", newItemOnChainTransaction).Msg("failed to insert new on chain tx history")
		}
	}

	return updated, skipped, nil
}

func UpdateSuccessful1155WithdrawalsWithTxHash(records []*NFT1155TransferRecord, contract string) (int, int) {
	l := passlog.L.With().Str("svc", "avant_pending_refund_set_tx_hash").Logger()

	skipped := 0
	success := 0

	for _, record := range records {
		// Null address is equal to mint
		if !strings.EqualFold(record.FromAddress, "0x0000000000000000000000000000000000000000") {
			skipped++
			continue
		}

		val, err := strconv.Atoi(record.ValueInt)
		if err != nil {
			l.Warn().
				Str("user_addr", record.ToAddress).
				Str("tx_hash", record.TxHash).
				Err(err).
				Msg("convert to decimal failed")
			skipped++
			continue
		}

		u, err := users.PublicAddress(common.HexToAddress(record.ToAddress))
		if err != nil {
			skipped++
			continue
		}

		filter := []qm.QueryMod{
			boiler.Pending1155RollbackWhere.UserID.EQ(u.ID),
			boiler.Pending1155RollbackWhere.Count.EQ(val),
			boiler.Pending1155RollbackWhere.IsRefunded.EQ(false),
			boiler.Pending1155RollbackWhere.RefundCanceledAt.IsNull(), // Not cancelled yet
			boiler.Pending1155RollbackWhere.DeletedAt.IsNull(),
			boiler.Pending1155RollbackWhere.TXHash.EQ(""),
			boiler.Pending1155RollbackWhere.TXHash.NEQ(record.TxHash), // Ignore tx hash if already assigned to another pending refund
		}

		count, err := boiler.Pending1155Rollbacks(filter...).Count(passdb.StdConn)
		if err != nil {
			l.Warn().Err(err).Msg("failed to get count")
			skipped++
			continue
		}
		if count <= 0 {
			//is this even an error? do we need to be warned about this?
			//l.Warn().Err(err).Msg("user does not have any pending refunds matching the value")
			skipped++
			continue
		}

		// Get pending refunds for user that are ready to be confirmed as on chain
		filter = append(filter, qm.OrderBy("created_at ASC")) // Sort so we get the oldest one
		pendingRefund, err := boiler.Pending1155Rollbacks(filter...).One(passdb.StdConn)
		if err != nil {
			l.Warn().Err(err).Msg("could not get matching single pending refund")
			skipped++
			continue
		}
		pendingRefund.TXHash = record.TxHash
		pendingRefund.RefundCanceledAt = null.TimeFrom(time.Now())

		_, err = pendingRefund.Update(passdb.StdConn, boil.Whitelist(boiler.PendingRefundColumns.TXHash, boiler.PendingRefundColumns.RefundCanceledAt))
		if err != nil {
			l.Warn().Err(err).Msg("failed to update user pending refund with tx hash")
			skipped++
			continue
		}

		//l.Info().Msg("successfully set tx hash, cancel refund")
		success++
	}

	return success, skipped
}

// ReverseFailed1155 Rollback stale 1155 (dangerous if buggy, check very, very carefully)
func ReverseFailed1155(enabled1155Rollback bool) (int, int, error) {
	l := passlog.L.
		With().
		Str("svc", "avant_rollback_1155").
		Bool("enable_1155_rollback", enabled1155Rollback).
		Logger()

	success := 0
	skipped := 0

	// Get refunds that can be marked as failed withdraws
	filter := []qm.QueryMod{
		boiler.Pending1155RollbackWhere.RefundedAt.LT(time.Now()),
		boiler.Pending1155RollbackWhere.RefundCanceledAt.IsNull(),
		boiler.Pending1155RollbackWhere.IsRefunded.EQ(false),
		boiler.Pending1155RollbackWhere.DeletedAt.IsNull(),
		boiler.Pending1155RollbackWhere.TXHash.EQ(""),
		qm.Load(boiler.Pending1155RollbackRels.Asset, qm.Select(boiler.UserAssets1155Columns.ID, boiler.UserAssets1155Columns.Count)),
	}

	refundsToProcess, err := boiler.Pending1155Rollbacks(filter...).All(passdb.StdConn)
	if err != nil {
		return success, skipped, err
	}

	for _, refund := range refundsToProcess {
		l = l.With().
			Str("asset_id", refund.R.Asset.ID).
			Int("count_from", refund.R.Asset.Count).
			Int("count_to", refund.R.Asset.Count+refund.Count).
			Logger()

		refund.R.Asset.Count += refund.Count

		refund.IsRefunded = true
		refund.RefundCanceledAt = null.TimeFrom(time.Now())

		_, err := refund.Update(passdb.StdConn, boil.Infer())
		if err != nil {
			l.Warn().Err(err).Msg("failed to rollback 1155 asset")
			skipped++
			continue
		}

		_, err = refund.R.Asset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.ID, boiler.UserAssets1155Columns.Count))
		if err != nil {
			l.Warn().Err(err).Msg("failed to rollback 1155 asset")
		}

		l.Info().Msg("successfully 1155 asset rollback")
		success++
	}

	return success, skipped, nil
}

func Process1155Deposits(records []*NFT1155TransferRecord, collectionSlug string, purchaseAddress common.Address, environment types.Environment) (int, int, error) {
	l := passlog.L.With().Str("svc", "avant_1155deposit_processor").Logger()
	success := 0
	skipped := 0

	l.Info().Int("records", len(records)).Msg("processing deposits")
	for _, record := range records {
		if !strings.EqualFold(record.ToAddress, purchaseAddress.Hex()) {
			skipped++
			continue
		}
		user, err := CreateOrGetUser(common.HexToAddress(record.FromAddress), environment)
		if err != nil {
			skipped++
			l.Error().Str("txid", record.TxHash).Str("user_addr", record.FromAddress).Err(err).Msg("create or get user")
			continue
		}

		count, err := strconv.Atoi(record.ValueInt)
		if err != nil {
			skipped++
			l.Error().Str("txid", record.TxHash).Err(err).Msg("process decimal")
			continue
		}

		asset, err := nft1155.CreateOrGet1155Asset(record.TokenID, user, collectionSlug)
		if err != nil {
			l.Error().Str("txid", record.TxHash).Err(err).Msg("failed creating or getting asset")
			skipped++
			continue
		}

		asset.Count += count

		_, err = asset.Update(passdb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Interface("asset", asset).Err(err).Msg("failed creating or getting asset")
			skipped++
			continue
		}

		depositAssetTransaction, err := boiler.DepositAsset1155Transactions(
			boiler.DepositAsset1155TransactionWhere.CollectionSlug.EQ(collectionSlug),
			boiler.DepositAsset1155TransactionWhere.TXHash.EQ(record.TxHash),
		).One(passdb.StdConn)
		if err != nil {
			l.Error().Interface("asset", asset).Err(err).Msg("failed to find asset transaction history")
			skipped++
			continue
		}

		depositAssetTransaction.Status = "confirmed"
		depositAssetTransaction.UpdatedAt = null.TimeFrom(time.Now())

		_, err = depositAssetTransaction.Update(passdb.StdConn, boil.Infer())
		if err != nil {
			l.Error().Interface("asset", asset).Err(err).Msg("failed to update asset transaction history")
			skipped++
			continue
		}

		success++

	}
	l.Info().
		Int("success", success).
		Int("skipped", skipped).
		Msg("synced 1155 deposits")

	return success, skipped, nil
}
