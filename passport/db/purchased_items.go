package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"strconv"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"xsyn-services/passport/rpcclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

type OnChainStatus string

const MINTABLE OnChainStatus = "MINTABLE"
const STAKABLE OnChainStatus = "STAKABLE"
const UNSTAKABLE OnChainStatus = "UNSTAKABLE"

func PurchasedItemSetOnChainStatus(purchasedItemID uuid.UUID, status OnChainStatus) error {
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

// SyncPurchasedItems against gameserver
func SyncPurchasedItems() error {
	passlog.L.Debug().Str("fn", "SyncPurchasedItems").Msg("db func")
	// TODO: Vinnie fix
	//tx, err := passdb.StdConn.Begin()
	//if err != nil {
	//	return terror.Error(err)
	//}
	//defer tx.Rollback()
	//mechResp := &rpcclient.MechsResp{}
	//err = rpcclient.Client.Call("S.Mechs", rpcclient.MechsReq{}, mechResp)
	//if err != nil {
	//	return terror.Error(err)
	//}
	//for _, item := range mechResp.MechContainers {
	//	exists, err := boiler.PurchasedItemsOldExists(tx, item.Mech.ID)
	//	if err != nil {
	//		passlog.L.Err(err).Str("id", item.Mech.ID).Msg("check if mech exists")
	//		return terror.Error(err)
	//	}
	//	if !exists {
	//		data, err := json.Marshal(item)
	//		if err != nil {
	//			return terror.Error(err)
	//		}
	//		var collection *boiler.Collection
	//		var collectionSlug string
	//		if !item.Mech.CollectionSlug.Valid {
	//			return terror.Error(fmt.Errorf("mech collection slug not valid"), "Mech collection slug not valid")
	//		}
	//
	//		collectionSlug = item.Mech.CollectionSlug.String
	//		collection, err = CollectionBySlug(context.Background(), passdb.Conn, collectionSlug)
	//		if err != nil {
	//			return terror.Error(err)
	//		}
	//		if item.Mech.IsDefault {
	//			collection, err = AICollection()
	//			if err != nil {
	//				return terror.Error(err)
	//			}
	//		}
	//
	//		if item.Mech.Hash == "k8zlb6Yl1L" {
	//			collection, err = RogueCollection()
	//			if err != nil {
	//				return terror.Error(err)
	//			}
	//		}
	//
	//		newItem := &boiler.PurchasedItemsOld{
	//			ID:              item.Mech.ID,
	//			CollectionID:    collection.ID,
	//			StoreItemID:     item.Mech.TemplateID,
	//			OwnerID:         item.Mech.OwnerID,
	//			ExternalTokenID: item.Mech.ExternalTokenID,
	//			IsDefault:       item.Mech.IsDefault,
	//			Tier:            item.Mech.Tier,
	//			Hash:            item.Mech.Hash,
	//			Data:            data,
	//			RefreshesAt:     time.Now().Add(RefreshDuration),
	//		}
	//		passlog.L.Info().Str("id", item.Mech.ID).
	//			Str("collection_id", collection.ID).
	//			Str("store_item_id", item.Mech.TemplateID).
	//			Str("owner_id", item.Mech.OwnerID).
	//			Int("external_token_id", item.Mech.ExternalTokenID).
	//			Msg("creating new mech")
	//		err = newItem.Insert(tx, boil.Infer())
	//		if err != nil {
	//			return terror.Error(err)
	//		}
	//	} else {
	//		passlog.L.Info().Str("id", item.Mech.ID).Msg("updating existing mech")
	//		_, err = refreshItem(uuid.Must(uuid.FromString(item.Mech.ID)), true)
	//		if err != nil {
	//			return terror.Error(err)
	//		}
	//	}
	//
	//}
	//
	//err = tx.Commit()
	//if err != nil {
	//	return terror.Error(err)
	//}
	return nil
}

// PurchasedItemLock lock for five minutes after user receives a mint signature to prevent on-world/off-world split brain
func PurchasedItemLock(itemID uuid.UUID) (*boiler.PurchasedItemsOld, error) {
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}
	defer tx.Rollback()
	item, err := boiler.FindPurchasedItemsOld(tx, itemID.String())
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
	return item, nil
}
func PurchasedItemIsOnWorld()  {}
func PurchasedItemIsOffWorld() {}
func PurchasedItemIsMinted(collectionAddr common.Address, tokenID int) (bool, error) {
	item, err := PurchasedItemByMintContractAndTokenID(collectionAddr, tokenID)
	if err != nil {
		return false, terror.Error(err)
	}
	return item.OnChainStatus != string(MINTABLE), nil
}

func PurchasedItemByMintContractAndTokenID(contractAddr common.Address, tokenID int) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemByMintContractAndTokenID").Strs("args", []string{contractAddr.Hex(), strconv.Itoa(tokenID)}).Msg("db func")
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
func PurchasedItemByHash(hash string) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemByHash").Msg("db func")
	item, err := boiler.PurchasedItemsOlds(boiler.PurchasedItemsOldWhere.Hash.EQ(hash)).One(passdb.StdConn)
	if err != nil && err != sql.ErrNoRows {
		return nil, terror.Error(err)
	}
	if err != nil {
		passlog.L.Error().Err(err).Msgf("unable to retrieve hash: %s", hash)
		return nil, terror.Error(err)
	}
	item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
	if err != nil {
		return nil, terror.Error(err)
	}
	return item, nil
}
func PurchasedItemsByOwnerID(ownerID string, limit int, afterExternalTokenID *int, includeAssetIDs, excludeAssetIDs []uuid.UUID) ([]*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemsByOwnerID").Msg("db func")

	orderBy := boiler.PurchasedItemsOldColumns.ExternalTokenID + " ASC"
	orderByArgs := []interface{}{}
	queryMods := []qm.QueryMod{
		boiler.PurchasedItemsOldWhere.OwnerID.EQ(ownerID),
		qm.Limit(limit),
	}
	if afterExternalTokenID != nil {
		queryMods = append(queryMods, boiler.PurchasedItemsOldWhere.ExternalTokenID.GT(*afterExternalTokenID))
	}
	if len(includeAssetIDs) > 0 {
		queuePositions := []string{}
		for i, assetID := range includeAssetIDs {
			queuePositions = append(queuePositions, fmt.Sprintf("WHEN ? THEN %d", i))
			orderByArgs = append(orderByArgs, assetID.String())
		}

		orderBy = fmt.Sprintf(
			`(
				CASE %s
					%s
				END
			), %s`,
			boiler.PurchasedItemsOldColumns.ID,
			strings.Join(queuePositions, "\n"),
			orderBy,
		)
	}
	if len(excludeAssetIDs) > 0 {
		args := []string{}
		for _, assetID := range excludeAssetIDs {
			args = append(args, assetID.String())
		}
		queryMods = append(queryMods, boiler.PurchasedItemsOldWhere.ID.NIN(args))
	}
	queryMods = append(queryMods, qm.OrderBy(orderBy, orderByArgs...))

	items, err := boiler.PurchasedItemsOlds(queryMods...).All(passdb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}
	result := []*boiler.PurchasedItemsOld{}
	for _, item := range items {
		item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
		if err != nil {
			return nil, terror.Error(err)
		}
		result = append(result, item)
	}
	return result, nil
}

func PurchasedItemsByOwnerIDAndTier(ownerID string, tier string) (int, error) {
	count, err := boiler.PurchasedItemsOlds(
		boiler.PurchasedItemsOldWhere.OwnerID.EQ(ownerID),
		boiler.PurchasedItemsOldWhere.Tier.EQ(tier),
	).Count(passdb.StdConn)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// PurchasedItems for admin only
func PurchasedItems() ([]*boiler.PurchasedItemsOld, error) {
	result, err := boiler.PurchasedItemsOlds().All(passdb.StdConn)
	if err != nil {
		return nil, terror.Error(err)
	}
	return result, nil
}

func PurchasedItemRegister(storeItemID uuid.UUID, ownerID uuid.UUID) ([]*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemRegister").Msg("db func")
	req := rpcclient.TemplateRegisterReq{TemplateID: storeItemID, OwnerID: ownerID}
	resp := &rpcclient.TemplateRegisterResp{}
	err := rpcclient.Client.Call("S.TemplateRegister", req, resp)
	if err != nil {
		return nil, terror.Error(err,  "communication to supremacy has failed")
	}
	var newItems []*boiler.PurchasedItemsOld
	// for each asset, assign it on our database
	for _, itm := range resp.Assets {
		data, err := json.Marshal(itm.Data)
		if err != nil {
			return nil, terror.Error(err)
		}

		// get collection
		collection, err := CollectionBySlug(itm.CollectionSlug)
		if err != nil {
			return nil, terror.Error(err)
		}

		newItem := &boiler.PurchasedItemsOld{
			ID:              itm.ID,
			ExternalTokenID: itm.TokenID,
			Hash:            itm.Hash,
			Tier:            itm.Tier,
			CollectionID:    collection.ID,
			OwnerID:         itm.OwnerID,
			Data:            data,
			RefreshesAt:     time.Now().Add(RefreshDuration),
		}
		newItem, err = setPurchasedItem(newItem)
		if err != nil {
			passlog.L.Error().Err(err).Interface("newItem", newItem).Msg("failed to purchase")
			return nil, terror.Error(err)
		}

		newItems = append(newItems, newItem)

		// TODO: Vinnie - figure how we want to handle store counts, separate store counts? joint?
		//storeItem, err := boiler.FindStoreItem(passdb.StdConn, resp.MechContainer.Mech.TemplateID)
		//if err != nil {
		//	return nil, terror.Error(err)
		//}
		//newCount, err := StoreItemPurchasedCount(uuid.Must(uuid.FromString(resp.MechContainer.Mech.TemplateID)))
		//if err != nil {
		//	return nil, terror.Error(err)
		//}
		//storeItem.AmountSold = newCount
		//_, err = storeItem.Update(passdb.StdConn, boil.Whitelist(boiler.StoreItemColumns.AmountSold))
		//if err != nil {
		//	return nil, terror.Error(err)
		//}
	}

	return newItems, nil
}
func PurchasedItemSetName(purchasedItemID uuid.UUID, name string) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemSetName").Msg("db func")
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
func PurchasedItemSetOwner(purchasedItemID uuid.UUID, ownerID uuid.UUID) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemSetOwner").Msg("db func")
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

func refreshItem(itemID uuid.UUID, force bool) (*boiler.PurchasedItemsOld, error) {
	// TODO: Vinnie - refactor to refresh any item
	passlog.L.Trace().Str("fn", "refreshItem").Msg("db func")
	if itemID == uuid.Nil {
		return nil, terror.Error(terror.ErrNilUUID)
	}
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return nil, terror.Error(err)
	}
	defer tx.Rollback()

	dbitem, err := boiler.FindPurchasedItemsOld(tx, itemID.String())
	if err != nil {
		return nil, terror.Error(err)
	}

	if !force {
		if dbitem.RefreshesAt.After(time.Now()) {
			return dbitem, nil
		}
	}

	resp := &rpcclient.AssetResp{}
	err = rpcclient.Client.Call("S.Asset", rpcclient.AssetReq{AssetID: itemID}, resp)
	if err != nil {
		return nil, terror.Error(err)
	}

	b, err := json.Marshal(resp.Asset)
	if err != nil {
		return nil, terror.Error(err)
	}

	dbitem.OwnerID = resp.Asset.OwnerID
	dbitem.Data = b
	dbitem.RefreshesAt = time.Now().Add(RefreshDuration)
	dbitem.UpdatedAt = time.Now()

	_, err = dbitem.Update(tx, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).
			Interface("dbitem", dbitem).
			Interface("resp", resp).
			Interface("b", b).
			Msg("issue updating item")
		return nil, terror.Error(err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, terror.Error(err)
	}

	return dbitem, nil
}

// setPurchasedItem sets the item, inserting it on the fly if it doesn't exist
// Does not obey TTL, can be heavy to run
func setPurchasedItem(item *boiler.PurchasedItemsOld) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "setPurchasedItem").Msg("db func")
	exists, err := boiler.PurchasedItemsOldExists(passdb.StdConn, item.ID)
	if err != nil {
		return item, terror.Error(err)
	}
	if !exists {
		err = item.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return item, terror.Error(err)
		}
	}
	item, err = refreshItem(uuid.Must(uuid.FromString(item.ID)), true)
	if err != nil {
		return item, terror.Error(err)
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
	refreshedItem, err := refreshItem(uuid.Must(uuid.FromString(item.ID)), false)
	if err != nil {
		passlog.L.Err(err).Str("purchased_item_id", item.ID).Msg("could not refresh purchased item from gameserver, using cached purchased item")
		return item, nil
	}
	return refreshedItem, nil
}

// PurchasedItem fetches the item, with the db as a fallback cache
func PurchasedItem(itemID uuid.UUID) (*boiler.PurchasedItemsOld, error) {
	passlog.L.Trace().Str("fn", "PurchasedItem").Msg("db func")
	purchasedItem, err := getPurchasedItem(itemID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return purchasedItem, nil
}

type PurchasedItemColumn string

const (
	PurchasedItemExternalTokenID PurchasedItemColumn = "external_token_id"
	PurchasedItemDeletedAt       PurchasedItemColumn = "deleted_at"
	PurchasedItemUpdatedAt       PurchasedItemColumn = "updated_at"
	PurchasedItemCreatedAt       PurchasedItemColumn = "created_at"
	PurchasedItemHash            PurchasedItemColumn = "hash"
	PurchasedItemUsername        PurchasedItemColumn = "username"
	PurchasedItemCollectionID    PurchasedItemColumn = "collection_id"
	PurchasedItemAssetType       PurchasedItemColumn = "asset_type"
	PurchasedItemOnChainStatus   PurchasedItemColumn = "on_chain_status"
)

func (ic PurchasedItemColumn) IsValid() error {
	switch ic {
	case
		PurchasedItemExternalTokenID,
		PurchasedItemDeletedAt,
		PurchasedItemUpdatedAt,
		PurchasedItemCreatedAt,
		PurchasedItemUsername,
		PurchasedItemCollectionID,
		PurchasedItemAssetType,
		PurchasedItemOnChainStatus,
		PurchasedItemHash:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid asset column type %s", ic))
}

const PurchaseGetQuery string = `
SELECT 
row_to_json(c) as collection,
purchased_items.external_token_id,
purchased_items.deleted_at,
purchased_items.updated_at,
purchased_items.created_at,
purchased_items.hash,
COALESCE(u.username, '') as username
` + PurchaseGetFrom

const PurchaseGetFrom = `
FROM purchased_items 
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
) c ON purchased_items.collection_id = c.id
INNER JOIN users u ON purchased_items.owner_id = u.id
`

// PurchaseItemsList gets a list of purchased items depending on the filters
func PurchaseItemsList(
	search string,
	archived bool,
	filter *ListFilterRequest,
	attributeFilter *AttributeFilterRequest,
	offset int,
	pageSize int,
	sortBy string,
	sortDir SortByDir,
) (int, []*types.PurchasedItem, error) {

	// Prepare Filters
	var args []interface{}

	filterConditionsString := ""
	argIndex := 0
	if filter != nil {
		filterConditions := []string{}
		for _, f := range filter.Items {
			column := PurchasedItemColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, nil, terror.Error(err)
			}

			argIndex += 1
			if f.ColumnField == string("collection_id") {
				f.ColumnField = fmt.Sprintf("purchased_items.%s", "collection_id")
			}
			condition, value := GenerateListFilterSQL(f.ColumnField, f.Value, f.OperatorValue, argIndex)
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
				return 0, nil, terror.Error(err)
			}
			condition := GenerateDataFilterSQL(f.Trait, f.Value, argIndex, "purchased_items")
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
			searchValueLabel, conditionLabel := GenerateDataSearchSQL("label", search, argIndex+1, "purchased_items")
			searchValueName, conditionName := GenerateDataSearchSQL("name", search, argIndex+2, "purchased_items")
			searchValueType, conditionType := GenerateDataSearchSQL("asset_type", search, argIndex+3, "purchased_items")
			searchValueTier, conditionTier := GenerateDataSearchSQL("tier", search, argIndex+4, "purchased_items")
			args = append(
				args,
				"%"+searchValueLabel+"%",
				"%"+searchValueName+"%",
				"%"+searchValueType+"%",
				"%"+searchValueTier+"%")
			searchCondition = " AND " + fmt.Sprintf("(%s OR %s OR %s OR %s)", conditionLabel, conditionName, conditionType, conditionTier)
		}
	}

	// Get Total Found
	countQ := fmt.Sprintf(`--sql
		SELECT COUNT(DISTINCT purchased_items.external_token_id)
		%s
		WHERE purchased_items.deleted_at %s
			%s
			%s
		`,
		PurchaseGetFrom,
		archiveCondition,
		filterConditionsString,
		searchCondition,
	)

	var totalRows int
	err := passdb.StdConn.QueryRow(countQ, args...).Scan(&totalRows)
	if err != nil {
		return 0, nil, err
	}
	if totalRows == 0 {
		return 0, make([]*types.PurchasedItem, 0), nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"

	if sortBy != "" {
		if sortBy == "name" {
			orderBy = fmt.Sprintf(" ORDER BY purchased_items.data->'mech'->>'name' %s", sortDir)
		} else {
			orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
		}
	}

	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q := fmt.Sprintf(
		PurchaseGetQuery+`--sql
		WHERE purchased_items.deleted_at %s
			%s
			%s
		%s
		%s`,
		archiveCondition,
		filterConditionsString,
		searchCondition,
		orderBy,
		limit,
	)

	result := []*types.PurchasedItem{}
	r, err := passdb.StdConn.Query(q, args...)
	if err != nil {

		return 0, nil, err
	}

	for r.Next() {
		pi := &types.PurchasedItem{}

		err = r.Scan(
			&pi.Collection,
			&pi.ExternalTokenID,
			&pi.DeletedAt,
			&pi.UpdatedAt,
			&pi.CreatedAt,
			&pi.Hash,
			&pi.Username,
		)
		if err != nil {
			return 0, nil, err
		}

		result = append(result, pi)
	}

	return totalRows, result, nil
}


func ChangePurchasedItemID(oldID, newID string) error {
	query := `
		UPDATE purchased_items SET id = $1 WHERE id = $2
		`

	_, err := passdb.StdConn.Exec(query, newID, oldID)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to update purchased item id")
		return terror.Error(err)
	}

	return nil
}

func ChangeStoreItemsTemplateID(oldID, newID string) error {
	query := `
		WITH old AS (
			UPDATE store_items SET id = $1
			WHERE id =  $2
			RETURNING $1::uuid as new, $2::uuid as old
		) UPDATE purchased_items
		SET store_item_id = old.new
		FROM old
		WHERE store_item_id = old.old;
		`

	_, err := passdb.StdConn.Exec(query, newID, oldID)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to update store item id")
		return terror.Error(err)
	}

	return nil
}
