package db

import (
	"database/sql"
	"encoding/json"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"time"

	"passport/rpcclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type OnChainStatus string

const MINTABLE OnChainStatus = "MINTABLE"
const STAKABLE OnChainStatus = "STAKABLE"
const UNSTAKABLE OnChainStatus = "UNSTAKABLE"

func PurchasedItemSetOnChainStatus(purchasedItemID uuid.UUID, status OnChainStatus) error {
	item, err := boiler.FindPurchasedItem(passdb.StdConn, purchasedItemID.String())
	if err != nil {
		return err
	}
	item.OnChainStatus = string(status)
	_, err = item.Update(passdb.StdConn, boil.Whitelist(boiler.PurchasedItemColumns.OnChainStatus))
	if err != nil {
		return err
	}
	return nil
}

const RefreshDuration = 1 * time.Minute

// SyncPurchasedItems against gameserver
func SyncPurchasedItems() error {
	passlog.L.Debug().Str("fn", "SyncPurchasedItems").Msg("db func")
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return terror.Error(err)
	}
	defer tx.Rollback()
	mechResp := &rpcclient.MechsResp{}
	err = rpcclient.Client.Call("S.Mechs", rpcclient.MechsReq{}, mechResp)
	if err != nil {
		return terror.Error(err)
	}
	for _, item := range mechResp.MechContainers {
		exists, err := boiler.PurchasedItemExists(tx, item.Mech.ID)
		if err != nil {
			passlog.L.Err(err).Str("id", item.Mech.ID).Msg("check if mech exists")
			return terror.Error(err)
		}
		if !exists {
			data, err := json.Marshal(item)
			if err != nil {
				return terror.Error(err)
			}
			collection, err := GenesisCollection()
			if err != nil {
				return terror.Error(err)
			}
			if item.Mech.IsDefault {
				collection, err = AICollection()
				if err != nil {
					return terror.Error(err)
				}
			}

			if item.Mech.Hash == "k8zlb6Yl1L" {
				collection, err = RogueCollection()
				if err != nil {
					return terror.Error(err)
				}
			}

			newItem := &boiler.PurchasedItem{
				ID:              item.Mech.ID,
				CollectionID:    collection.ID,
				StoreItemID:     item.Mech.TemplateID,
				OwnerID:         item.Mech.OwnerID,
				ExternalTokenID: item.Mech.ExternalTokenID,
				Tier:            item.Mech.Tier,
				Hash:            item.Mech.Hash,
				Data:            data,
				RefreshesAt:     time.Now().Add(RefreshDuration),
			}
			passlog.L.Info().Str("id", item.Mech.ID).
				Str("collection_id", collection.ID).
				Str("store_item_id", item.Mech.TemplateID).
				Str("owner_id", item.Mech.OwnerID).
				Int("external_token_id", item.Mech.ExternalTokenID).
				Msg("creating new mech")
			err = newItem.Insert(tx, boil.Infer())
			if err != nil {
				return terror.Error(err)
			}
		} else {
			passlog.L.Info().Str("id", item.Mech.ID).Msg("updating existing mech")
			_, err = refreshItem(uuid.Must(uuid.FromString(item.Mech.ID)), true)
			if err != nil {
				return terror.Error(err)
			}
		}

	}

	tx.Commit()

	return nil
}

// PurchasedItemLock lock for five minutes after user receives a mint signature to prevent on-world/off-world split brain
func PurchasedItemLock(itemID uuid.UUID) (*boiler.PurchasedItem, error) {
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}
	defer tx.Rollback()
	item, err := boiler.FindPurchasedItem(tx, itemID.String())
	if err != nil {
		return nil, terror.Error(err)
	}
	item.UnlockedAt = time.Now().Add(5 * time.Minute)
	_, err = item.Update(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}
	tx.Commit()
	return item, nil
}
func PurchasedItemIsOnWorld()  {}
func PurchasedItemIsOffWorld() {}
func PurchasedItemIsMinted(collectionAddr common.Address, tokenID int) (bool, error) {
	collection, err := CollectionByMintAddress(collectionAddr)
	if err != nil {
		return false, terror.Error(err)
	}
	count, err := boiler.ItemOnchainTransactions(
		boiler.ItemOnchainTransactionWhere.CollectionID.EQ(collection.ID),
		boiler.ItemOnchainTransactionWhere.ExternalTokenID.EQ(tokenID),
	).Count(passdb.StdConn)
	if err != nil {
		return false, terror.Error(err)
	}
	return count > 0, nil
}

func PurchasedItemByMintContractAndTokenID(contractAddr common.Address, tokenID int) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemByMintContractAndTokenID").Msg("db func")
	collection, err := CollectionByMintAddress(contractAddr)
	if err != nil {
		return nil, terror.Error(err)
	}
	item, err := boiler.PurchasedItems(
		boiler.PurchasedItemWhere.CollectionID.EQ(collection.ID),
		boiler.PurchasedItemWhere.ExternalTokenID.EQ(tokenID),
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
func PurchasedItemByHash(hash string) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemByHash").Msg("db func")
	item, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.Hash.EQ(hash)).One(passdb.StdConn)
	if err != nil && err != sql.ErrNoRows {
		return nil, terror.Error(err)
	}
	item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
	if err != nil {
		return nil, terror.Error(err)
	}
	return item, nil
}
func PurchasedItemsByOwnerID(ownerID uuid.UUID) ([]*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemsByOwnerID").Msg("db func")
	items, err := boiler.PurchasedItems(
		boiler.PurchasedItemWhere.OwnerID.EQ(ownerID.String()),
		qm.OrderBy("external_token_id ASC"),
	).All(passdb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}
	result := []*boiler.PurchasedItem{}
	for _, item := range items {
		item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
		if err != nil {
			return nil, terror.Error(err)
		}
		result = append(result, item)
	}
	return result, nil
}

func PurchasedItemsbyOwnerIDAndTier(ownerID uuid.UUID, tier string) (int, error) {
	count, err := boiler.PurchasedItems(
		boiler.PurchasedItemWhere.OwnerID.EQ(ownerID.String()),
		boiler.PurchasedItemWhere.Tier.EQ(tier),
	).Count(passdb.StdConn)
	if err != nil {
		return 0, terror.Error(err)
	}
	return int(count), nil
}

// PurchasedItems for admin only
func PurchasedItems() ([]*boiler.PurchasedItem, error) {
	result, err := boiler.PurchasedItems().All(passdb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

func PurchasedItemRegister(storeItemID uuid.UUID, ownerID uuid.UUID) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemRegister").Msg("db func")
	req := rpcclient.MechRegisterReq{TemplateID: storeItemID, OwnerID: ownerID}
	resp := &rpcclient.MechRegisterResp{}
	err := rpcclient.Client.Call("S.MechRegister", req, resp)
	if err != nil {
		return nil, terror.Error(err)
	}

	data, err := json.Marshal(resp.MechContainer)
	if err != nil {
		return nil, terror.Error(err)
	}
	collection, err := GenesisCollection()
	if err != nil {
		return nil, terror.Error(err)
	}
	if resp.MechContainer.Mech.IsDefault {
		collection, err = AICollection()
		if err != nil {
			return nil, terror.Error(err)
		}
	}
	newItem := &boiler.PurchasedItem{
		ID:              resp.MechContainer.Mech.ID,
		StoreItemID:     resp.MechContainer.Mech.TemplateID,
		ExternalTokenID: resp.MechContainer.Mech.ExternalTokenID,
		Hash:            resp.MechContainer.Mech.Hash,
		Tier:            resp.MechContainer.Mech.Tier,
		CollectionID:    collection.ID,
		OwnerID:         resp.MechContainer.Mech.OwnerID,
		Data:            data,
		RefreshesAt:     time.Now().Add(RefreshDuration),
	}
	newItem, err = setPurchasedItem(newItem)
	if err != nil {
		return nil, terror.Error(err)
	}

	storeItem, err := boiler.FindStoreItem(passdb.StdConn, resp.MechContainer.Mech.TemplateID)
	if err != nil {
		return nil, terror.Error(err)
	}
	newCount, err := StoreItemPurchasedCount(uuid.Must(uuid.FromString(resp.MechContainer.Mech.TemplateID)))
	if err != nil {
		return nil, terror.Error(err)
	}
	storeItem.AmountSold = newCount
	_, err = storeItem.Update(passdb.StdConn, boil.Whitelist(boiler.StoreItemColumns.AmountSold))
	if err != nil {
		return nil, terror.Error(err)
	}
	return newItem, nil
}
func PurchasedItemSetName(purchasedItemID uuid.UUID, name string) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemSetName").Msg("db func")
	req := rpcclient.MechSetNameReq{MechID: purchasedItemID, Name: name}
	resp := &rpcclient.MechSetNameResp{}
	err := rpcclient.Client.Call("S.MechSetName", req, resp)
	if err != nil {
		return nil, terror.Error(err)
	}
	refreshedItem, err := refreshItem(purchasedItemID, true)
	if err != nil {
		return nil, terror.Error(err)
	}
	return refreshedItem, nil
}
func PurchasedItemSetOwner(purchasedItemID uuid.UUID, ownerID uuid.UUID) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemSetOwner").Msg("db func")
	req := rpcclient.MechSetOwnerReq{MechID: purchasedItemID, OwnerID: ownerID}
	resp := &rpcclient.MechSetOwnerResp{}
	err := rpcclient.Client.Call("S.MechSetOwner", req, resp)
	if err != nil {
		return nil, terror.Error(err)
	}
	refreshedItem, err := refreshItem(purchasedItemID, true)
	if err != nil {
		return nil, terror.Error(err)
	}
	return refreshedItem, nil
}

func refreshItem(itemID uuid.UUID, force bool) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "refreshItem").Msg("db func")
	if itemID == uuid.Nil {
		return nil, terror.Error(terror.ErrNilUUID)
	}
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}
	defer tx.Rollback()

	dbitem, err := boiler.FindPurchasedItem(tx, itemID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	if !force {
		if dbitem.RefreshesAt.After(time.Now()) {
			return dbitem, nil
		}
	}

	resp := &rpcclient.MechResp{}
	err = rpcclient.Client.Call("S.Mech", rpcclient.MechReq{MechID: itemID}, resp)
	if err != nil {
		return nil, terror.Error(err)
	}

	b, err := json.Marshal(resp.MechContainer)
	if err != nil {
		return nil, terror.Error(err)
	}

	dbitem.Data = b
	dbitem.RefreshesAt = time.Now().Add(RefreshDuration)
	dbitem.UpdatedAt = time.Now()

	_, err = dbitem.Update(tx, boil.Infer())
	if err != nil {
		return nil, terror.Error(err)
	}

	tx.Commit()

	return dbitem, nil

}

// setPurchasedItem sets the item, inserting it on the fly if it doesn't exist
// Does not obey TTL, can be heavy to run
func setPurchasedItem(item *boiler.PurchasedItem) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "setPurchasedItem").Msg("db func")
	exists, err := boiler.PurchasedItemExists(passdb.StdConn, item.ID)
	if err != nil {
		return nil, terror.Error(err)
	}
	if !exists {
		err = item.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return nil, terror.Error(err)
		}
	}
	item, err = refreshItem(uuid.Must(uuid.FromString(item.ID)), true)
	if err != nil {
		return nil, terror.Error(err)
	}

	return item, nil
}

// getPurchasedItem fetches the item, obeying TTL
func getPurchasedItem(itemID uuid.UUID) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "getPurchasedItem").Msg("db func")
	item, err := boiler.FindPurchasedItem(passdb.StdConn, itemID.String())
	if err != nil {
		return nil, terror.Error(err)
	}
	refreshedItem, err := refreshItem(uuid.Must(uuid.FromString(item.ID)), true)
	if err != nil {
		passlog.L.Err(err).Str("purchased_item_id", item.ID).Msg("could not refresh purchased item from gameserver, using cached purchased item")
		return item, nil
	}
	return refreshedItem, nil
}

// PurchasedItem fetches the item, with the db as a fallback cache
func PurchasedItem(itemID uuid.UUID) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItem").Msg("db func")
	purchasedItem, err := getPurchasedItem(itemID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return purchasedItem, nil
}
