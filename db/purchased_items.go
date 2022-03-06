package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"time"

	"passport/rpcclient"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

const RefreshDuration = 1 * time.Minute

// SyncPurchasedItems against gameserver
func SyncPurchasedItems() error {
	passlog.L.Debug().Str("fn", "SyncPurchasedItems").Msg("db func")
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return fmt.Errorf("start tx: %w", err)
	}
	defer tx.Rollback()
	mechResp := &rpcclient.MechsResp{}
	err = rpcclient.Client.Call("S.Mechs", rpcclient.MechsReq{}, mechResp)
	if err != nil {
		return fmt.Errorf("call rpc: %w", err)
	}
	for _, item := range mechResp.MechContainers {
		exists, err := boiler.PurchasedItemExists(tx, item.Mech.ID)
		if err != nil {
			return fmt.Errorf("check purchased item exists: %w", err)
		}
		if !exists {
			data, err := json.Marshal(item)
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			collection, err := GenesisCollection()
			if err != nil {
				return fmt.Errorf("get genesis collection: %w", err)
			}
			if item.Mech.IsDefault {
				collection, err = AICollection()
				if err != nil {
					return fmt.Errorf("get ai collection: %w", err)
				}
			}

			newItem := &boiler.PurchasedItem{
				ID:           item.Mech.ID,
				CollectionID: collection.ID,
				StoreItemID:  item.Mech.TemplateID,
				OwnerID:      item.Mech.OwnerID,
				Hash:         item.Mech.Hash,
				Data:         data,
				RefreshesAt:  time.Now().Add(RefreshDuration),
			}
			err = newItem.Insert(tx, boil.Infer())
			if err != nil {
				return fmt.Errorf("insert new item: %w", err)
			}
		} else {
			_, err = refreshItem(uuid.Must(uuid.FromString(item.Mech.ID)), true)
			if err != nil {
				return fmt.Errorf("refresh item: %w", err)
			}
		}

	}

	tx.Commit()

	return nil
}

func PurchasedItemsByOwnerID(ownerID uuid.UUID) ([]*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemsByOwnerID").Msg("db func")
	items, err := boiler.PurchasedItems(boiler.PurchasedItemWhere.OwnerID.EQ(ownerID.String())).All(passdb.StdConn)
	if err != nil {
		return nil, fmt.Errorf("list purchased items: %w", err)
	}
	result := []*boiler.PurchasedItem{}
	for _, item := range items {
		item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
		if err != nil {
			return nil, fmt.Errorf("get purchased item: %w", err)
		}
		result = append(result, item)
	}
	return result, nil
}

func PurchasedItemRegister() (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemRegister").Msg("db func")
	req := rpcclient.MechRegisterReq{}
	resp := &rpcclient.MechRegisterResp{}
	err := rpcclient.Client.Call("S.MechRegister", req, resp)
	if err != nil {
		return nil, fmt.Errorf("rpc call: %w", err)
	}

	data, err := json.Marshal(resp.MechContainer)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	collection, err := GenesisCollection()
	if err != nil {
		return nil, fmt.Errorf("get genesis collection: %w", err)
	}
	if resp.MechContainer.Mech.IsDefault {
		collection, err = AICollection()
		if err != nil {
			return nil, fmt.Errorf("get ai collection: %w", err)
		}
	}
	newItem := &boiler.PurchasedItem{
		ID:           resp.MechContainer.Mech.ID,
		CollectionID: collection.ID,
		OwnerID:      resp.MechContainer.Mech.OwnerID,
		Data:         data,
		RefreshesAt:  time.Now().Add(RefreshDuration),
	}
	newItem, err = setPurchasedItem(newItem)
	if err != nil {
		return nil, fmt.Errorf("set purchased item: %w", err)
	}

	storeItem, err := boiler.FindStoreItem(passdb.StdConn, resp.MechContainer.Mech.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("find store item: %w", err)
	}
	newCount, err := StoreItemPurchasedCount(uuid.Must(uuid.FromString(resp.MechContainer.Mech.TemplateID)))
	if err != nil {
		return nil, fmt.Errorf("get purchase count: %w", err)
	}
	storeItem.AmountSold = newCount
	_, err = storeItem.Update(passdb.StdConn, boil.Whitelist(boiler.StoreItemColumns.AmountSold))
	if err != nil {
		return nil, fmt.Errorf("update store item: %w", err)
	}
	return newItem, nil
}
func PurchasedItemSetName(mechID uuid.UUID, name string) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemSetName").Msg("db func")
	req := rpcclient.MechSetNameReq{MechID: mechID, Name: name}
	resp := &rpcclient.MechSetNameResp{}
	err := rpcclient.Client.Call("S.MechSetName", req, resp)
	if err != nil {
		return nil, fmt.Errorf("rpc call: %w", err)
	}
	refreshedItem, err := refreshItem(mechID, true)
	if err != nil {
		return nil, fmt.Errorf("refresh item: %w", err)
	}
	return refreshedItem, nil
}
func PurchasedItemSetOwner(mechID uuid.UUID, ownerID uuid.UUID) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItemSetOwner").Msg("db func")
	req := rpcclient.MechSetOwnerReq{MechID: mechID, OwnerID: ownerID}
	resp := &rpcclient.MechSetOwnerResp{}
	err := rpcclient.Client.Call("S.MechSetOwner", req, resp)
	if err != nil {
		return nil, fmt.Errorf("rpc call: %w", err)
	}
	refreshedItem, err := refreshItem(mechID, true)
	if err != nil {
		return nil, fmt.Errorf("refresh item: %w", err)
	}
	return refreshedItem, nil
}

func refreshItem(itemID uuid.UUID, force bool) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "refreshItem").Msg("db func")
	if itemID == uuid.Nil {
		return nil, errors.New("nil UUID")
	}
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, fmt.Errorf("start tx: %w", err)
	}
	defer tx.Rollback()

	dbitem, err := boiler.FindPurchasedItem(tx, itemID.String())
	if err != nil {
		return nil, fmt.Errorf("find item: %w", err)
	}

	if !force {
		if dbitem.RefreshesAt.After(time.Now()) {
			return dbitem, nil
		}
	}

	resp := &rpcclient.MechResp{}
	err = rpcclient.Client.Call("S.Mech", rpcclient.MechReq{MechID: itemID}, resp)
	if err != nil {
		return nil, fmt.Errorf("rpc call: %w", err)
	}

	b, err := json.Marshal(resp.MechContainer)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}

	dbitem.Data = b
	dbitem.RefreshesAt = time.Now().Add(RefreshDuration)
	dbitem.UpdatedAt = time.Now()

	_, err = dbitem.Update(tx, boil.Infer())
	if err != nil {
		return nil, fmt.Errorf("update item: %w", err)
	}

	tx.Commit()

	return dbitem, nil

}

// setPurchasedItem sets the item, inserting it on the fly if it doesn't exist
// Does not obey TTL, can be heavy to run
func setPurchasedItem(item *boiler.PurchasedItem) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "setPurchasedItem").Msg("db func")
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()
	exists, err := boiler.PurchasedItemExists(passdb.StdConn, item.ID)
	if err != nil {
		return nil, fmt.Errorf("check item exists: %w", err)
	}
	if !exists {
		err = item.Insert(tx, boil.Infer())
		if err != nil {
			return nil, fmt.Errorf("insert item: %w", err)
		}
	}
	item, err = refreshItem(uuid.Must(uuid.FromString(item.ID)), true)
	if err != nil {
		return nil, fmt.Errorf("refresh item: %w", err)
	}
	tx.Commit()
	return item, nil
}

// getPurchasedItem fetches the item, obeying TTL
func getPurchasedItem(itemID uuid.UUID) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "getPurchasedItem").Msg("db func")
	item, err := boiler.FindPurchasedItem(passdb.StdConn, itemID.String())
	if err != nil {
		return nil, fmt.Errorf("find item: %w", err)
	}
	item, err = refreshItem(uuid.Must(uuid.FromString(item.ID)), true)
	if err != nil {
		return nil, fmt.Errorf("refresh item: %w", err)
	}
	return item, nil
}

// PurchasedItem fetches the item, obeying TTL
func PurchasedItem(itemID uuid.UUID) (*boiler.PurchasedItem, error) {
	passlog.L.Debug().Str("fn", "PurchasedItem").Msg("db func")
	return getPurchasedItem(itemID)
}
