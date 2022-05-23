package db

import (
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/sqlboiler/v4/types"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	xsynTypes "xsyn-services/types"

	"xsyn-services/passport/rpcclient"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type OnChainStatus string

const MINTABLE OnChainStatus = "MINTABLE"
const STAKABLE OnChainStatus = "STAKABLE"
const UNSTAKABLE OnChainStatus = "UNSTAKABLE"

func PurchasedItemSetOnChainStatus(purchasedItemID uuid.UUID, status OnChainStatus) error {
	// TODO: Vinnie FIX
	//item, err := boiler.FindPurchasedItemsOld(passdb.StdConn, purchasedItemID.String())
	//if err != nil {
	//	return err
	//}
	//item.OnChainStatus = string(status)
	//_, err = item.Update(passdb.StdConn, boil.Whitelist(boiler.PurchasedItemsOldColumns.OnChainStatus))
	//if err != nil {
	//	return err
	//}
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
	// TODO: Vinnie FIX
	//tx, err := passdb.StdConn.Begin()
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//defer tx.Rollback()
	//item, err := boiler.FindPurchasedItemsOld(tx, itemID.String())
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//item.UnlockedAt = time.Now().Add(5 * time.Minute)
	//_, err = item.Update(tx, boil.Infer())
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//err = tx.Commit()
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	return nil, nil
}

func PurchasedItemIsMinted(collectionAddr common.Address, tokenID int) (bool, error) {
	item, err := PurchasedItemByMintContractAndTokenID(collectionAddr, tokenID)
	if err != nil {
		return false, terror.Error(err)
	}
	return item.OnChainStatus != string(MINTABLE), nil
}

func PurchasedItemByMintContractAndTokenID(contractAddr common.Address, tokenID int) (*boiler.PurchasedItemsOld, error) {
	// TODO: Vinnie FIX
	//passlog.L.Trace().Str("fn", "PurchasedItemByMintContractAndTokenID").Strs("args", []string{contractAddr.Hex(), strconv.Itoa(tokenID)}).Msg("db func")
	//collection, err := CollectionByMintAddress(contractAddr)
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//item, err := boiler.PurchasedItemsOlds(
	//	boiler.PurchasedItemsOldWhere.CollectionID.EQ(collection.ID),
	//	boiler.PurchasedItemsOldWhere.ExternalTokenID.EQ(tokenID),
	//).One(passdb.StdConn)
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	return nil, nil
}
func PurchasedItemByHash(hash string) (*boiler.PurchasedItemsOld, error) {
	// TODO: Vinnie FIX
	//passlog.L.Trace().Str("fn", "PurchasedItemByHash").Msg("db func")
	//item, err := boiler.PurchasedItemsOlds(boiler.PurchasedItemsOldWhere.Hash.EQ(hash)).One(passdb.StdConn)
	//if err != nil && err != sql.ErrNoRows {
	//	return nil, terror.Error(err)
	//}
	//if err != nil {
	//	passlog.L.Error().Err(err).Msgf("unable to retrieve hash: %s", hash)
	//	return nil, terror.Error(err)
	//}
	//item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	return nil, nil
}
//func PurchasedItemsByOwnerID(ownerID string, limit int, afterExternalTokenID *int, includeAssetIDs, excludeAssetIDs []uuid.UUID) ([]*boiler.PurchasedItemsOld, error) {
//	passlog.L.Trace().Str("fn", "PurchasedItemsByOwnerID").Msg("db func")
//
//	orderBy := boiler.PurchasedItemsOldColumns.ExternalTokenID + " ASC"
//	orderByArgs := []interface{}{}
//	queryMods := []qm.QueryMod{
//		boiler.PurchasedItemsOldWhere.OwnerID.EQ(ownerID),
//		qm.Limit(limit),
//	}
//	if afterExternalTokenID != nil {
//		queryMods = append(queryMods, boiler.PurchasedItemsOldWhere.ExternalTokenID.GT(*afterExternalTokenID))
//	}
//	if len(includeAssetIDs) > 0 {
//		queuePositions := []string{}
//		for i, assetID := range includeAssetIDs {
//			queuePositions = append(queuePositions, fmt.Sprintf("WHEN ? THEN %d", i))
//			orderByArgs = append(orderByArgs, assetID.String())
//		}
//
//		orderBy = fmt.Sprintf(
//			`(
//				CASE %s
//					%s
//				END
//			), %s`,
//			boiler.PurchasedItemsOldColumns.ID,
//			strings.Join(queuePositions, "\n"),
//			orderBy,
//		)
//	}
//	if len(excludeAssetIDs) > 0 {
//		args := []string{}
//		for _, assetID := range excludeAssetIDs {
//			args = append(args, assetID.String())
//		}
//		queryMods = append(queryMods, boiler.PurchasedItemsOldWhere.ID.NIN(args))
//	}
//	queryMods = append(queryMods, qm.OrderBy(orderBy, orderByArgs...))
//
//	items, err := boiler.PurchasedItemsOlds(queryMods...).All(passdb.StdConn)
//	if err != nil {
//		return nil, terror.Error(err)
//	}
//	result := []*boiler.PurchasedItemsOld{}
//	for _, item := range items {
//		item, err = getPurchasedItem(uuid.Must(uuid.FromString(item.ID)))
//		if err != nil {
//			return nil, terror.Error(err)
//		}
//		result = append(result, item)
//	}
//	return result, nil
//}
//
//func PurchasedItemsByOwnerIDAndTier(ownerID string, tier string) (int, error) {
//	count, err := boiler.PurchasedItemsOlds(
//		boiler.PurchasedItemsOldWhere.OwnerID.EQ(ownerID),
//		boiler.PurchasedItemsOldWhere.Tier.EQ(tier),
//	).Count(passdb.StdConn)
//	if err != nil {
//		return 0, err
//	}
//	return int(count), nil
//}

//// PurchasedItems for admin only
//func PurchasedItems() ([]*boiler.PurchasedItemsOld, error) {
//	result, err := boiler.PurchasedItemsOlds().All(passdb.StdConn)
//	if err != nil {
//		return nil, terror.Error(err)
//	}
//	return result, nil
//}

func PurchasedItemRegister(storeItemID uuid.UUID, ownerID uuid.UUID) ([]*xsynTypes.UserAsset, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemRegister").Msg("db func")
	req := rpcclient.TemplateRegisterReq{TemplateID: storeItemID, OwnerID: ownerID}
	resp := &rpcclient.TemplateRegisterResp{}
	err := rpcclient.Client.Call("S.TemplateRegister", req, resp)
	if err != nil {
		return nil, terror.Error(err,  "communication to supremacy has failed")
	}
	var newItems []*xsynTypes.UserAsset
	// for each asset, assign it on our database
	for _, itm := range resp.Assets {
		// get collection
		collection, err := CollectionBySlug(itm.CollectionSlug)
		if err != nil {
			return nil, terror.Error(err)
		}

		var jsonAtrribs types.JSON
		err = jsonAtrribs.Marshal(itm.Attributes)
		if err != nil {
			return nil, terror.Error(err)
		}

		boilerAsset := &boiler.UserAsset{
			CollectionID:    collection.ID,
			ID:              itm.ID,
			TokenID: itm.TokenID,
			Tier:            itm.Tier,
			Hash:            itm.Hash,
			OwnerID:         itm.OwnerID,
			Data:            itm.Data,
			Attributes:      jsonAtrribs,
			Name:            itm.Name,
			ImageURL:        itm.ImageURL,
			ExternalURL:     itm.ExternalURL,
			Description:     itm.Description,
			BackgroundColor: itm.BackgroundColor,
			AnimationURL: itm.AnimationURL,
			YoutubeURL: itm.YoutubeURL,
			UnlockedAt: itm.UnlockedAt,
			MintedAt: itm.MintedAt,
			OnChainStatus: itm.OnChainStatus,
			XsynLocked: itm.XsynLocked,
			DataRefreshedAt: time.Now(),
		}

		err = boilerAsset.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			passlog.L.Error().Interface("req", req).Err(err).Msg("failed to register new asset - can't insert asset")
			return nil, err
		}

		newItems = append(newItems, xsynTypes.UserAssetFromBoiler(boilerAsset))
	}

	return newItems, nil
}
//func PurchasedItemSetName(purchasedItemID uuid.UUID, name string) (*boiler.PurchasedItemsOld, error) {
//	passlog.L.Trace().Str("fn", "PurchasedItemSetName").Msg("db func")
//	req := rpcclient.MechSetNameReq{MechID: purchasedItemID, Name: name}
//	resp := &rpcclient.MechSetNameResp{}
//	err := rpcclient.Client.Call("S.MechSetName", req, resp)
//	if err != nil {
//		return nil, terror.Error(err)
//	}
//	refreshedItem, err := refreshItem(purchasedItemID, true)
//	if err != nil {
//		return nil, terror.Error(err)
//	}
//	return refreshedItem, nil
//}
//func PurchasedItemSetOwner(purchasedItemID uuid.UUID, ownerID uuid.UUID) (*boiler.PurchasedItemsOld, error) {
//	passlog.L.Trace().Str("fn", "PurchasedItemSetOwner").Msg("db func")
//	req := rpcclient.MechSetOwnerReq{MechID: purchasedItemID, OwnerID: ownerID}
//	resp := &rpcclient.MechSetOwnerResp{}
//	err := rpcclient.Client.Call("S.MechSetOwner", req, resp)
//	if err != nil {
//		return nil, terror.Error(err)
//	}
//	refreshedItem, err := refreshItem(purchasedItemID, true)
//	if err != nil {
//		return nil, terror.Error(err)
//	}
//	return refreshedItem, nil
//}

func refreshItem(itemID uuid.UUID, force bool) (*boiler.PurchasedItemsOld, error) {
	// TODO: Vinnie FIX
	//// TODO: Vinnie - refactor to refresh any item
	//passlog.L.Trace().Str("fn", "refreshItem").Msg("db func")
	//if itemID == uuid.Nil {
	//	return nil, terror.Error(terror.ErrNilUUID)
	//}
	//tx, err := passdb.StdConn.Begin()
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//defer tx.Rollback()
	//
	//dbitem, err := boiler.FindPurchasedItemsOld(tx, itemID.String())
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//
	//if !force {
	//	if dbitem.RefreshesAt.After(time.Now()) {
	//		return dbitem, nil
	//	}
	//}
	//
	//resp := &rpcclient.AssetResp{}
	//err = rpcclient.Client.Call("S.UserAsset", rpcclient.AssetReq{AssetID: itemID}, resp)
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//
	//b, err := json.Marshal(resp.Asset)
	//if err != nil {
	//	return nil, terror.Error(err)
	//}
	//
	//dbitem.OwnerID = resp.Asset.OwnerID
	//dbitem.Data = b
	//dbitem.RefreshesAt = time.Now().Add(RefreshDuration)
	//dbitem.UpdatedAt = time.Now()
	//
	//_, err = dbitem.Update(tx, boil.Infer())
	//if err != nil {
	//	passlog.L.Error().Err(err).
	//		Interface("dbitem", dbitem).
	//		Interface("resp", resp).
	//		Interface("b", b).
	//		Msg("issue updating item")
	//	return nil, terror.Error(err)
	//}
	//
	//err = tx.Commit()
	//if err != nil {
	//	return nil, terror.Error(err)
	//}

	return nil, nil
}


func ChangeStoreItemsTemplateID(oldID, newID string) error {
	query := `
		WITH old AS (
			UPDATE store_items SET id = $1
			WHERE id =  $2
			RETURNING $1::uuid as new, $2::uuid as old
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

	return nil
}
