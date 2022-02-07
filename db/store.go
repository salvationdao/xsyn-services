package db

import (
	"context"
	"passport"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/ninja-software/terror/v2"
)

// AddItemToStore added the object to the xsyn nft store table
func AddItemToStore(ctx context.Context, conn Conn, storeObject *passport.StoreItem) error {
	q := `INSERT INTO xsyn_store (faction_id,
                                      name,
                                      collection_id,
                                      description,
                                      image,
                                      attributes,
                                      usd_cent_cost,
                                      amount_sold,
                                      amount_available,
                                      sold_after,
                                      sold_before)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			RETURNING id`

	_, err := conn.Exec(ctx, q,
		storeObject.FactionID,
		storeObject.Name,
		storeObject.CollectionID,
		storeObject.Description,
		storeObject.Image,
		storeObject.Attributes,
		storeObject.UsdCentCost,
		storeObject.AmountSold,
		storeObject.AmountAvailable,
		storeObject.SoldAfter,
		storeObject.SoldAfter,
	)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// StoreItemByID get store item by id
func StoreItemByID(ctx context.Context, conn Conn, storeItemID passport.StoreItemID) (*passport.StoreItem, error) {
	storeItem := &passport.StoreItem{}
	q := `SELECT 	id, 
					faction_id,
					name,
					collection_id,
					description,
					image,
					attributes,
					usd_cent_cost,
					amount_sold,
					amount_available,
					sold_after,
					sold_before
			FROM xsyn_store WHERE id = $1`

	err := pgxscan.Get(ctx, conn, storeItem, q, storeItemID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return storeItem, nil
}

// StoreItemPurchased bumps a store items amount sold up
func StoreItemPurchased(ctx context.Context, conn Conn, storeItem *passport.StoreItem) error {
	q := `UPDATE xsyn_store SET amount_sold = amount_sold + 1 WHERE id = $1 RETURNING amount_sold`

	err := pgxscan.Get(ctx, conn, storeItem, q, storeItem.ID)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// StoreItemListByCollectionAndFaction returns a list of store item IDs
func StoreItemListByCollectionAndFaction(ctx context.Context, conn Conn,
	collectionID passport.CollectionID,
	factionID passport.FactionID,
	page int,
	pageSize int,
) ([]*passport.StoreItemID, error) {
	storeItems := []*passport.StoreItemID{}

	q := `	SELECT 	id
			FROM xsyn_store
			WHERE collection_id = $1
			AND faction_id = $2
			LIMIT $3 OFFSET $4`

	err := pgxscan.Select(ctx, conn, &storeItems, q, collectionID, factionID, pageSize, page*pageSize)
	if err != nil {
		return nil, terror.Error(err)
	}

	return storeItems, nil
}
