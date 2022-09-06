package db

import (
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"strconv"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
)

type OnChainStatus string

const MINTABLE OnChainStatus = "MINTABLE"
const STAKABLE OnChainStatus = "STAKABLE"
const UNSTAKABLE OnChainStatus = "UNSTAKABLE"
const UNSTAKABLEOLD OnChainStatus = "UNSTAKABLE_OLD"

func PurchasedItemIsMintedDEPRECATE(collectionAddr common.Address, tokenID int) (bool, error) {
	item, err := PurchasedItemByMintContractAndTokenIDDEPRECATE(collectionAddr, tokenID)
	if err != nil {
		return false, terror.Error(err)
	}
	return item.OnChainStatus != string(MINTABLE), nil
}

func PurchasedItemSetOnChainStatusDEPRECATE(purchasedItemID uuid.UUID, status OnChainStatus) error {
	item, err := boiler.FindPurchasedItemsOld(passdb.StdConn, purchasedItemID.String())
	if err != nil {
		return err
	}
	item.OnChainStatus = string(status)
	_, err = item.Update(passdb.StdConn, boil.Whitelist(boiler.PurchasedItemsOldColumns.OnChainStatus))
	if err != nil {
		return err
	}
	return nil
}

const RefreshDuration = 24 * time.Hour

// PurchasedItemLockDEPRECATED lock for five minutes after user receives a mint signature to prevent on-world/off-world split brain
func PurchasedItemLockDEPRECATED(itemID uuid.UUID) (*boiler.PurchasedItemsOld, error) {
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}
	defer tx.Rollback()

	item, err := boiler.FindUserAsset(tx, itemID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	item.UnlockedAt = time.Now().Add(5 * time.Minute)
	_, err = item.Update(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}
	return nil, nil
}

func PurchasedItemByHashDEPRECATE(hash string) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemByHashDEPRECATE").Msg("db func")
	item, err := boiler.PurchasedItemsOlds(boiler.PurchasedItemsOldWhere.Hash.EQ(hash)).One(passdb.StdConn)
	if err != nil && err != sql.ErrNoRows {
		return nil, terror.Error(err)
	}
	if err != nil {
		passlog.L.Error().Err(err).Msgf("unable to retrieve hash: %s", hash)
		return nil, terror.Error(err)
	}

	return item, nil
}

func PurchasedItemsByOwnerIDAndTierDEPRECATE(ownerID string, tier string) (int, error) {
	count, err := boiler.PurchasedItemsOlds(
		boiler.PurchasedItemsOldWhere.OwnerID.EQ(ownerID),
		boiler.PurchasedItemsOldWhere.Tier.EQ(tier),
	).Count(passdb.StdConn)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func PurchasedItemByMintContractAndTokenIDDEPRECATE(contractAddr common.Address, tokenID int) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemByMintContractAndTokenIDDEPRECATE").Strs("args", []string{contractAddr.Hex(), strconv.Itoa(tokenID)}).Msg("db func")
	collection, err := CollectionByMintAddress(contractAddr)
	if err != nil {
		return nil, terror.Error(err)
	}
	item, err := boiler.PurchasedItemsOlds(
		boiler.PurchasedItemsOldWhere.CollectionID.EQ(collection.ID),
		boiler.PurchasedItemsOldWhere.ExternalTokenID.EQ(tokenID),
	).One(passdb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}
	item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
	if err != nil {
		return nil, terror.Error(err)
	}
	return item, nil
}

// getPurchasedItem fetches the item, obeying TTL
func getPurchasedItem(itemID uuid.UUID) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "getPurchasedItem").Msg("db func")
	item, err := boiler.FindPurchasedItemsOld(passdb.StdConn, itemID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	return item, nil
}

// PurchasedItemsDEPRECATE for admin only
func PurchasedItemsDEPRECATE() ([]*boiler.PurchasedItemsOld, error) {
	result, err := boiler.PurchasedItemsOlds().All(passdb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

func ChangeStoreItemsTemplateID(oldID, newID string) error {
	query := `
		WITH old AS (
			UPDATE store_items SET id = $1
			WHERE id =  $2
			RETURNING $1::uuid AS new, $2::uuid AS old
		) UPDATE purchased_items_old
		SET store_item_id = old.new
		FROM old
		WHERE store_item_id = old.old;
		`
	_, err := passdb.StdConn.Exec(query, newID, oldID)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to update store item id")
		return terror.Error(err)
	}

	count, err := boiler.StoreItems(boiler.StoreItemWhere.ID.EQ(newID)).Count(passdb.StdConn)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to update store item id")
		return terror.Error(err)
	}
	if count != 1 {
		err = fmt.Errorf("new id didn't update correctly")
		passlog.L.Error().Err(err).Msg("failed to update store item id")
		return terror.Error(err)
	}

	return nil
}
