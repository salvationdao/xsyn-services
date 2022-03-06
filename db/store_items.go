package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"passport/rpcclient"
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var RestrictionMap = map[string]string{
	"COLOSSAL":        "LOOTBOX",
	"DEUS_EX":         "LOOTBOX",
	"ELITE_LEGENDARY": "LOOTBOX",
	"EXOTIC":          "LOOTBOX",
	"GUARDIAN":        "LOOTBOX",
	"LEGENDARY":       "LOOTBOX",
	"MEGA":            "NONE",
	"MYTHIC":          "LOOTBOX",
	"RARE":            "LOOTBOX",
	"ULTRA_RARE":      "LOOTBOX",
}
var AmountMap = map[string]int{
	"COLOSSAL":        400,
	"DEUS_EX":         3,
	"ELITE_LEGENDARY": 100,
	"EXOTIC":          40,
	"GUARDIAN":        20,
	"LEGENDARY":       200,
	"MEGA":            500,
	"MYTHIC":          10,
	"RARE":            300,
	"ULTRA_RARE":      60,
}
var PriceCentsMap = map[string]int{
	"COLOSSAL":        100000,
	"DEUS_EX":         100000,
	"ELITE_LEGENDARY": 100000,
	"EXOTIC":          100000,
	"GUARDIAN":        100000,
	"LEGENDARY":       100000,
	"MEGA":            100,
	"MYTHIC":          100000,
	"RARE":            100000,
	"ULTRA_RARE":      100000,
}

func SyncStoreItems() error {
	passlog.L.Debug().Str("fn", "SyncStoreItems").Msg("db func")
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	templateResp := &rpcclient.TemplatesResp{}
	err = rpcclient.Client.Call("S.Templates", rpcclient.TemplatesReq{}, templateResp)
	if err != nil {
		return err
	}
	for _, template := range templateResp.TemplateContainers {
		if template.Template.ID == uuid.Nil.String() {
			return errors.New("nil template ID")
		}
		exists, err := boiler.StoreItemExists(tx, template.Template.ID)
		if err != nil {
			return err
		}
		passlog.L.Debug().Str("id", template.Template.ID).Msg("sync store item")
		if !exists {
			data, err := json.Marshal(template)
			if err != nil {
				return err
			}
			collection, err := GenesisCollection()
			if err != nil {
				return err
			}
			if template.Template.IsDefault {
				collection, err = AICollection()
				if err != nil {
					return err
				}
			}

			if template.Template.ID == "" {
				return fmt.Errorf("template.Template.ID invalid")
			}
			if collection.ID == "" {
				return fmt.Errorf("collection.ID invalid")
			}
			if template.Template.FactionID == "" {
				return fmt.Errorf("template.Template.FactionID invalid")
			}
			restrictionGroup, ok := RestrictionMap[template.Template.Tier]
			if !ok {
				return fmt.Errorf("restriction not found for %s", template.Template.Tier)
			}
			amountAvailable, ok := AmountMap[template.Template.Tier]
			if !ok {
				return fmt.Errorf("amountAvailable not found for %s", template.Template.Tier)
			}
			priceCents, ok := PriceCentsMap[template.Template.Tier]
			if !ok {
				return fmt.Errorf("amountAvailable not found for %s", template.Template.Tier)
			}
			count, err := StoreItemPurchasedCount(uuid.Must(uuid.FromString(template.Template.ID)))
			if err != nil {
				return fmt.Errorf("get purchase count: %w", err)
			}
			newStoreItem := &boiler.StoreItem{
				ID:               template.Template.ID,
				CollectionID:     collection.ID,
				FactionID:        template.Template.FactionID,
				UsdCentCost:      priceCents,
				AmountSold:       count,
				AmountAvailable:  amountAvailable,
				RestrictionGroup: restrictionGroup,
				Data:             data,
				RefreshesAt:      time.Now().Add(RefreshDuration),
			}
			err = newStoreItem.Insert(tx, boil.Infer())
			if err != nil {
				return fmt.Errorf("insert new store item: %w", err)
			}
		} else {
			_, err = refreshStoreItem(uuid.Must(uuid.FromString(template.Template.ID)), true)
			if err != nil {
				return err
			}
		}
	}

	tx.Commit()

	return nil
}
func StoreItem(storeItemID uuid.UUID) (*boiler.StoreItem, error) {
	passlog.L.Debug().Str("fn", "StoreItem").Msg("db func")
	return getStoreItem(storeItemID)
}
func StoreItemPurchasedCount(templateID uuid.UUID) (int, error) {
	passlog.L.Debug().Str("fn", "StoreItemPurchasedCount").Msg("db func")
	resp := &rpcclient.TemplatePurchasedCountResp{}
	err := rpcclient.Client.Call("S.TemplatePurchasedCount", rpcclient.TemplatePurchasedCountReq{TemplateID: templateID}, resp)
	if err != nil {
		return 0, err
	}
	return resp.Count, nil
}
func StoreItemsByFactionID(factionID uuid.UUID) ([]*boiler.StoreItem, error) {
	passlog.L.Debug().Str("fn", "StoreItemsByFactionID").Msg("db func")
	storeItems, err := boiler.StoreItems(boiler.StoreItemWhere.FactionID.EQ(factionID.String())).All(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	result := []*boiler.StoreItem{}
	for _, storeItem := range storeItems {
		storeItem, err = getStoreItem(uuid.Must(uuid.FromString(storeItem.ID)))
		if err != nil {
			return nil, err
		}
		result = append(result, storeItem)
	}
	return result, nil
}

func refreshStoreItem(storeItemID uuid.UUID, force bool) (*boiler.StoreItem, error) {
	passlog.L.Debug().Str("fn", "refreshStoreItem").Msg("db func")
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	dbitem, err := boiler.FindStoreItem(tx, storeItemID.String())
	if err != nil {
		return nil, err
	}

	if !force {
		if dbitem.RefreshesAt.After(time.Now()) {
			return dbitem, nil
		}
	}

	resp := &rpcclient.TemplateResp{}
	err = rpcclient.Client.Call("S.Template", rpcclient.TemplateReq{TemplateID: storeItemID}, resp)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(resp.TemplateContainer)
	if err != nil {
		return nil, err
	}

	restrictionGroup, ok := RestrictionMap[resp.TemplateContainer.Template.Tier]
	if !ok {
		return nil, fmt.Errorf("restriction not found for %s", resp.TemplateContainer.Template.Tier)
	}
	amountAvailable, ok := AmountMap[resp.TemplateContainer.Template.Tier]
	if !ok {
		return nil, fmt.Errorf("amountAvailable not found for %s", resp.TemplateContainer.Template.Tier)
	}
	priceCents, ok := PriceCentsMap[resp.TemplateContainer.Template.Tier]
	if !ok {
		return nil, fmt.Errorf("amountAvailable not found for %s", resp.TemplateContainer.Template.Tier)
	}
	count, err := StoreItemPurchasedCount(uuid.Must(uuid.FromString(resp.TemplateContainer.Template.ID)))
	if err != nil {
		return nil, fmt.Errorf("get purchase count: %w", err)
	}
	dbitem.Data = b
	dbitem.FactionID = resp.TemplateContainer.Template.FactionID
	dbitem.RefreshesAt = time.Now().Add(RefreshDuration)
	dbitem.UpdatedAt = time.Now()
	dbitem.RestrictionGroup = restrictionGroup
	dbitem.AmountAvailable = amountAvailable
	dbitem.UsdCentCost = priceCents
	dbitem.AmountSold = count

	_, err = dbitem.Update(tx, boil.Infer())
	if err != nil {
		return nil, err
	}

	tx.Commit()

	return dbitem, nil

}

// getStoreItem fetches the item, obeying TTL
func getStoreItem(storeItemID uuid.UUID) (*boiler.StoreItem, error) {
	passlog.L.Debug().Str("fn", "getStoreItem").Msg("db func")
	item, err := boiler.FindStoreItem(passdb.StdConn, storeItemID.String())
	if err != nil {
		return nil, err
	}
	refreshedItem, err := refreshStoreItem(uuid.Must(uuid.FromString(item.ID)), true)
	if err != nil {
		passlog.L.Err(err).Str("store_item_id", item.ID).Msg("could not refresh store item from gameserver, using cached store item")
		return item, nil
	}
	return refreshedItem, nil
}
