package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"passport"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"
	"passport/rpcclient"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var RestrictionMap = map[string]string{
	TierColossal:       RestrictionGroupLootbox,
	TierDeusEx:         RestrictionGroupLootbox,
	TierEliteLegendary: RestrictionGroupLootbox,
	TierExotic:         RestrictionGroupLootbox,
	TierGuardian:       RestrictionGroupLootbox,
	TierLegendary:      RestrictionGroupLootbox,
	TierMega:           RestrictionGroupNone,
	TierMythic:         RestrictionGroupLootbox,
	TierRare:           RestrictionGroupLootbox,
	TierUltraRare:      RestrictionGroupLootbox,
}

const RestrictionGroupLootbox = "LOOTBOX"
const RestrictionGroupNone = "NONE"
const RestrictionGroupPrize = "PRIZE"

const TierMega = "MEGA"
const TierColossal = "COLOSSAL"
const TierRare = "RARE"
const TierLegendary = "LEGENDARY"
const TierEliteLegendary = "ELITE_LEGENDARY"
const TierUltraRare = "ULTRA_RARE"
const TierExotic = "EXOTIC"
const TierGuardian = "GUARDIAN"
const TierMythic = "MYTHIC"
const TierDeusEx = "DEUS_EX"

var AmountMap = map[string]int{
	TierColossal:       400,
	TierDeusEx:         3,
	TierEliteLegendary: 100,
	TierExotic:         40,
	TierGuardian:       20,
	TierLegendary:      200,
	TierMega:           500,
	TierMythic:         10,
	TierRare:           300,
	TierUltraRare:      60,
}
var PriceCentsMap = map[string]int{
	TierColossal:       100000,
	TierDeusEx:         100000,
	TierEliteLegendary: 100000,
	TierExotic:         100000,
	TierGuardian:       100000,
	TierLegendary:      100000,
	TierMega:           100,
	TierMythic:         100000,
	TierRare:           100000,
	TierUltraRare:      100000,
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

			// Golds are prizes only, not purchasable
			if template.BlueprintChassis.Skin == "Gold" {
				restrictionGroup = RestrictionGroupPrize
			}
			if template.BlueprintChassis.Skin == "Slava Ukraini" {
				restrictionGroup = RestrictionGroupPrize
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
				Tier:             template.Template.Tier,
				IsDefault:        template.Template.IsDefault,
				AmountSold:       count,
				AmountAvailable:  amountAvailable,
				RestrictionGroup: restrictionGroup,
				Data:             data,
				RefreshesAt:      time.Now().Add(RefreshDuration),
			}
			passlog.L.Info().Str("id", template.Template.ID).Msg("inserting new store item")
			err = newStoreItem.Insert(tx, boil.Infer())
			if err != nil {
				return fmt.Errorf("insert new store item: %w", err)
			}
		} else {
			passlog.L.Info().Str("id", template.Template.ID).Msg("updating existing store item")
			_, err = refreshStoreItem(uuid.Must(uuid.FromString(template.Template.ID)), true)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

func StoreItemsRemainingByFactionIDAndRestrictionGroup(collectionID uuid.UUID, factionID uuid.UUID, restrictionGroup string) (int, error) {
	items, err := boiler.StoreItems(
		boiler.StoreItemWhere.FactionID.EQ(factionID.String()),
		boiler.StoreItemWhere.RestrictionGroup.EQ(restrictionGroup),
		boiler.StoreItemWhere.IsDefault.EQ(false),
	).All(passdb.StdConn)
	count := 0
	for _, item := range items {
		count = count + item.AmountAvailable - item.AmountSold
	}
	return count, err
}
func StoreItemsRemainingByFactionIDAndTier(collectionID uuid.UUID, factionID uuid.UUID, tier string) (int, error) {
	items, err := boiler.StoreItems(
		boiler.StoreItemWhere.FactionID.EQ(factionID.String()),
		boiler.StoreItemWhere.Tier.EQ(tier),
		boiler.StoreItemWhere.RestrictionGroup.NEQ(RestrictionGroupPrize),
		boiler.StoreItemWhere.IsDefault.EQ(false),
	).All(passdb.StdConn)
	count := 0
	for _, item := range items {
		count = count + item.AmountAvailable - item.AmountSold
	}
	return count, err
}

// StoreItemsAvailable return the total of available war machine in each faction
func StoreItemsAvailable() ([]*passport.FactionSaleAvailable, error) {
	collection, err := GenesisCollection()
	if err != nil {
		return nil, err
	}
	factions, err := boiler.Factions().All(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	result := []*passport.FactionSaleAvailable{}

	for _, faction := range factions {
		theme := &passport.FactionTheme{}
		err = faction.Theme.Unmarshal(theme)
		if err != nil {
			return nil, err
		}
		megaAmount, err := StoreItemsRemainingByFactionIDAndTier(uuid.Must(uuid.FromString(collection.ID)), uuid.Must(uuid.FromString(faction.ID)), TierMega)
		if err != nil {
			return nil, err
		}
		lootboxAmount, err := StoreItemsRemainingByFactionIDAndRestrictionGroup(uuid.Must(uuid.FromString(collection.ID)), uuid.Must(uuid.FromString(faction.ID)), RestrictionGroupLootbox)
		if err != nil {
			return nil, err
		}
		record := &passport.FactionSaleAvailable{
			ID:            passport.FactionID(uuid.Must(uuid.FromString(faction.ID))),
			Label:         faction.Label,
			LogoBlobID:    passport.BlobID(uuid.Must(uuid.FromString(faction.LogoBlobID))),
			Theme:         theme,
			MegaAmount:    int64(megaAmount),
			LootboxAmount: int64(lootboxAmount),
		}
		result = append(result, record)
	}
	return result, nil
}

// StoreItems for admin only
func StoreItems() ([]*boiler.StoreItem, error) {
	result, err := boiler.StoreItems().All(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return result, nil
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

func StoreItemsByFactionIDAndRestrictionGroup(factionID uuid.UUID, restrictionGroup string) ([]*boiler.StoreItem, error) {
	result, err := boiler.StoreItems(
		boiler.StoreItemWhere.FactionID.EQ(factionID.String()),
		boiler.StoreItemWhere.RestrictionGroup.EQ(restrictionGroup),
	).All(passdb.StdConn)
	return result, err
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

	if resp.TemplateContainer.BlueprintChassis.Skin == "Slava Ukraini" {
		restrictionGroup = RestrictionGroupPrize
	}

	// Golds are prizes only, not purchasable
	if resp.TemplateContainer.BlueprintChassis.Skin == "Gold" {
		restrictionGroup = RestrictionGroupPrize
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
	dbitem.Tier = resp.TemplateContainer.Template.Tier
	dbitem.IsDefault = resp.TemplateContainer.Template.IsDefault

	_, err = dbitem.Update(tx, boil.Infer())
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

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
