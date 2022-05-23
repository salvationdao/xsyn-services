package db

import (
	"fmt"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
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
	//passlog.L.Trace().Str("fn", "SyncStoreItems").Msg("db func")
	//
	//tx, err := passdb.StdConn.Begin()
	//if err != nil {
	//	return err
	//}
	//defer tx.Rollback()
	//templateResp := &rpcclient.TemplatesResp{}
	//err = rpcclient.Client.Call("S.Templates", rpcclient.TemplatesReq{}, templateResp)
	//if err != nil {
	//	return err
	//}
	//for _, template := range templateResp.TemplateContainers {
	//	if template.Template.ID == uuid.Nil.String() {
	//		return errors.New("nil template ID")
	//	}
	//	exists, err := boiler.StoreItemExists(tx, template.Template.ID)
	//	if err != nil {
	//		return err
	//	}
	//	passlog.L.Debug().Str("id", template.Template.ID).Msg("sync store item")
	//	if !exists {
	//		data, err := json.Marshal(template)
	//		if err != nil {
	//			return err
	//		}
	//		var collection *boiler.Collection
	//		var collectionSlug string
	//		if !template.Template.CollectionSlug.Valid {
	//			return fmt.Errorf("template collection slug not valid")
	//		}
	//
	//		collectionSlug = template.Template.CollectionSlug.String
	//		collection, err = CollectionBySlug(collectionSlug)
	//		if err != nil {
	//			return err
	//		}
	//		if template.Template.IsDefault {
	//			collection, err = AICollection()
	//			if err != nil {
	//				return err
	//			}
	//		}
	//
	//		if template.Template.ID == "" {
	//			return fmt.Errorf("template.Template.ID invalid")
	//		}
	//		if collection.ID == "" {
	//			return fmt.Errorf("collection.ID invalid")
	//		}
	//		if template.Template.FactionID == "" {
	//			return fmt.Errorf("template.Template.FactionID invalid")
	//		}
	//		restrictionGroup, ok := RestrictionMap[template.Template.Tier]
	//		if !ok {
	//			return fmt.Errorf("restriction not found for %s", template.Template.Tier)
	//		}
	//
	//		// Golds are prizes only, not purchasable
	//		if template.BlueprintChassis.Skin == "Gold" {
	//			restrictionGroup = RestrictionGroupPrize
	//		}
	//		if template.BlueprintChassis.Skin == "Slava Ukraini" {
	//			restrictionGroup = RestrictionGroupPrize
	//		}
	//		amountAvailable, ok := AmountMap[template.Template.Tier]
	//		if !ok {
	//			return fmt.Errorf("amountAvailable not found for %s", template.Template.Tier)
	//		}
	//		priceCents, ok := PriceCentsMap[template.Template.Tier]
	//		if !ok {
	//			return fmt.Errorf("amountAvailable not found for %s", template.Template.Tier)
	//		}
	//		count, err := StoreItemPurchasedCount(uuid.Must(uuid.FromString(template.Template.ID)))
	//		if err != nil {
	//			return fmt.Errorf("get purchase count: %w", err)
	//		}
	//		newStoreItem := &boiler.StoreItem{
	//			ID:               template.Template.ID,
	//			CollectionID:     collection.ID,
	//			FactionID:        template.Template.FactionID,
	//			UsdCentCost:      priceCents,
	//			Tier:             template.Template.Tier,
	//			IsDefault:        template.Template.IsDefault,
	//			AmountSold:       count,
	//			AmountAvailable:  amountAvailable,
	//			RestrictionGroup: restrictionGroup,
	//			Data:             data,
	//			RefreshesAt:      time.Now().Add(RefreshDuration),
	//		}
	//		passlog.L.Info().Str("id", template.Template.ID).Interface("data", newStoreItem).Msg("inserting new store item")
	//		err = newStoreItem.Insert(tx, boil.Infer())
	//		if err != nil {
	//			return fmt.Errorf("insert new store item: %w", err)
	//		}
	//	} else {
	//		passlog.L.Info().Str("id", template.Template.ID).Msg("updating existing store item")
	//		_, err = refreshStoreItem(uuid.Must(uuid.FromString(template.Template.ID)), true)
	//		if err != nil {
	//
	//			return err
	//		}
	//	}
	//}
	//err = tx.Commit()
	//if err != nil {
	//	return err
	//}

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
func StoreItemsAvailable() ([]*types.FactionSaleAvailable, error) {
	// TODO: Vinnie - rework this
	collection, err := GenesisCollection()
	if err != nil {
		return nil, err
	}
	factions, err := boiler.Factions().All(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	result := []*types.FactionSaleAvailable{}

	for _, faction := range factions {
		theme := &types.FactionTheme{}
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
		record := &types.FactionSaleAvailable{
			ID:            types.FactionID(uuid.Must(uuid.FromString(faction.ID))),
			Label:         faction.Label,
			LogoBlobID:    types.BlobID(uuid.Must(uuid.FromString(faction.LogoBlobID))),
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
	passlog.L.Trace().Str("fn", "StoreItem").Msg("db func")
	return getStoreItem(storeItemID)
}
func StoreItemPurchasedCount(templateID uuid.UUID) (int, error) {
	// TODO: Vinnie - see if we still need and refactor if needed
	//passlog.L.Trace().Str("fn", "StoreItemPurchasedCount").Msg("db func")
	//resp := &rpcclient.TemplatePurchasedCountResp{}
	//err := rpcclient.Client.Call("S.TemplatePurchasedCount", rpcclient.TemplatePurchasedCountReq{TemplateID: templateID}, resp)
	//if err != nil {
	//	return 0, err
	//}
	//return resp.Count, nil
	return 0,nil
}

func StoreItemsByFactionIDAndRestrictionGroup(factionID uuid.UUID, restrictionGroup string) ([]*boiler.StoreItem, error) {
	result, err := boiler.StoreItems(
		boiler.StoreItemWhere.FactionID.EQ(factionID.String()),
		boiler.StoreItemWhere.RestrictionGroup.EQ(restrictionGroup),
	).All(passdb.StdConn)
	return result, err
}

func StoreItemsByFactionID(factionID uuid.UUID) ([]*boiler.StoreItem, error) {
	passlog.L.Trace().Str("fn", "StoreItemsByFactionID").Msg("db func")
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
	passlog.L.Trace().Str("fn", "refreshStoreItem").Msg("db func")
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

	//resp := &rpcclient.TemplateResp{}
	//err = rpcclient.Client.Call("S.Template", rpcclient.TemplateReq{TemplateID: storeItemID}, resp)
	//if err != nil {
	//	return nil, err
	//}
	//
	//if resp.TemplateContainer.Template.CollectionSlug.Valid {
	//	collection, err := CollectionBySlug(resp.TemplateContainer.Template.CollectionSlug.String)
	//	if err != nil {
	//		return nil, err
	//	}
	//	dbitem.CollectionID = collection.ID
	//}
	//
	//b, err := json.Marshal(resp.TemplateContainer)
	//if err != nil {
	//	return nil, err
	//}
	//
	//dbitem.Data = b
	//dbitem.FactionID = resp.TemplateContainer.Template.FactionID
	//dbitem.RefreshesAt = time.Now().Add(RefreshDuration)
	//dbitem.UpdatedAt = time.Now()
	//
	//// Not done on a refresh as it should never be updated via sync after initial insertion
	//// dbitem.RestrictionGroup = restrictionGroup
	//// dbitem.AmountAvailable = amountAvailable
	//// dbitem.UsdCentCost = priceCents
	//// dbitem.AmountSold = count
	//
	//dbitem.Tier = resp.TemplateContainer.Template.Tier
	//dbitem.IsDefault = resp.TemplateContainer.Template.IsDefault
	//
	//passlog.L.Info().Str("id", dbitem.ID).Interface("data", dbitem).Msg("updating store item")
	//
	//_, err = dbitem.Update(tx, boil.Infer())
	//if err != nil {
	//	return nil, err
	//}
	//
	//err = tx.Commit()
	//if err != nil {
	//	return nil, err
	//}

	return dbitem, nil

}

// getStoreItem fetches the item, obeying TTL
func getStoreItem(storeItemID uuid.UUID) (*boiler.StoreItem, error) {
	passlog.L.Trace().Str("fn", "getStoreItem").Msg("db func")
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

type StoreItemColumn string

const (
	StoreItemExternalTokenID StoreItemColumn = "external_token_id"
	StoreItemDeletedAt       StoreItemColumn = "deleted_at"
	StoreItemUpdatedAt       StoreItemColumn = "updated_at"
	StoreItemCreatedAt       StoreItemColumn = "created_at"
	StoreItemHash            StoreItemColumn = "hash"
	StoreItemUsername        StoreItemColumn = "username"
	StoreItemCollectionID    StoreItemColumn = "collection_id"
	StoreItemAssetType       StoreItemColumn = "asset_type"
	StoreItemFactionID       StoreItemColumn = "faction_id"
)

func (ic StoreItemColumn) IsValid() error {
	switch ic {
	case
		StoreItemExternalTokenID,
		StoreItemDeletedAt,
		StoreItemUpdatedAt,
		StoreItemCreatedAt,
		StoreItemHash,
		StoreItemUsername,
		StoreItemCollectionID,
		StoreItemFactionID,
		StoreItemAssetType:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid asset column type %s", ic))
}

const StoreItemGetQuery string = `
SELECT 
row_to_json(c) as collection,
store_items.id,
store_items.tier,
store_items.is_default,
store_items.restriction_group,
store_items.deleted_at,
store_items.updated_at,
store_items.created_at
` + StoreItemGetFrom

const StoreItemGetFrom = `
FROM store_items 
INNER JOIN (
	SELECT  id,
			name,
			logo_blob_id as logoBlobID,
			keywords,
			slug,
			deleted_at as deletedAt,  
			mint_contract as "mintContract",
			stake_contract as "stakeContract"
	FROM collections _c
) c ON store_items.collection_id = c.id
`

//  StoreItemsList gets a list of store items depending on the filters
func StoreItemsList(
	search string,
	archived bool,
	filter *ListFilterRequest,
	attributeFilter *AttributeFilterRequest,
	offset int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
) (int, []*types.StoreItem, error) {

	// Prepare Filters
	var args []interface{}

	filterConditionsString := ""
	argIndex := 0
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := StoreItemColumn(f.Column)
			err := column.IsValid()
			if err != nil {
				return 0, nil, err
			}

			argIndex += 1
			if f.Column == string("collection_id") {
				f.Column = fmt.Sprintf("collection_id")
			}
			condition, value := GenerateListFilterSQL(f.Column, f.Value, f.Operator, argIndex)
			if condition != "" {
				filterConditions = append(filterConditions, condition)
				args = append(args, value)
			}
		}
		if len(filterConditions) > 0 {
			filterConditionsString = "AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	if attributeFilter != nil {
		filterConditions := []string{}
		for _, f := range attributeFilter.Items {
			column := TraitType(f.Trait)
			err := column.IsValid()
			if err != nil {
				return 0, nil, err
			}
			condition := GenerateDataFilterSQL(f.Trait, f.Value, argIndex, "store_items")
			filterConditions = append(filterConditions, condition)
		}
		if len(filterConditions) > 0 {
			filterConditionsString += "AND (" + strings.Join(filterConditions, " "+string(attributeFilter.LinkOperator)+" ") + ")"
		}
	}

	archiveCondition := "IS NULL"
	if archived {
		archiveCondition = "IS NOT NULL"
	}

	searchCondition := ""
	if search != "" {
		if len(search) > 0 {
			searchValueLabel, conditionLabel := GenerateDataSearchStoreItemsSQL("label", search, argIndex+1, "store_items")
			searchValueName, conditionName := GenerateDataSearchStoreItemsSQL("name", search, argIndex+2, "store_items")
			searchValueType, conditionType := GenerateDataSearchStoreItemsSQL("asset_type", search, argIndex+3, "store_items")
			searchValueTier, conditionTier := GenerateDataSearchStoreItemsSQL("tier", search, argIndex+4, "store_items")
			args = append(
				args,
				"%"+searchValueLabel+"%",
				"%"+searchValueName+"%",
				"%"+searchValueType+"%",
				"%"+searchValueTier+"%")
			searchCondition = " AND " + fmt.Sprintf("(%s OR %s OR %s OR %s)", conditionLabel, conditionName, conditionType, conditionTier)

		}

	}

	// Filter by Megas for now
	filterByMegasCondition := fmt.Sprintf(
		`AND store_items.restriction_group != $%d AND store_items.tier = $%d AND store_items.is_default = false`,
		len(args)+1,
		len(args)+2,
	)
	args = append(args, RestrictionGroupPrize, TierMega)

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT store_items.id)
		%s
		WHERE store_items.deleted_at %s
			%s
			%s
			%s
		`,
		StoreItemGetFrom,
		archiveCondition,
		filterConditionsString,
		filterByMegasCondition,
		searchCondition,
	)

	var totalRows int
	err := passdb.StdConn.QueryRow(countQ, args...).Scan(&totalRows)
	if err != nil {
		return 0, nil, err
	}
	if totalRows == 0 {
		return 0, make([]*types.StoreItem, 0), nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"
	if sortBy != "" {
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(
		StoreItemGetQuery+`--sql
		WHERE store_items.deleted_at %s
		%s
		%s
		%s
		%s
		%s`,
		archiveCondition,
		filterConditionsString,
		filterByMegasCondition,
		searchCondition,
		orderBy,
		limit,
	)

	result := make([]*types.StoreItem, 0)

	r, err := passdb.StdConn.Query(q, args...)
	if err != nil {
		return 0, nil, err
	}

	for r.Next() {
		si := &types.StoreItem{}
		err = r.Scan(
			&si.Collection,
			&si.ID,
			&si.Tier,
			&si.IsDefault,
			&si.RestrictionGroup,
			&si.DeletedAt,
			&si.UpdatedAt,
			&si.CreatedAt,
		)
		if err != nil {
			return 0, nil, err
		}

		result = append(result, si)
	}

	return totalRows, result, nil
}
