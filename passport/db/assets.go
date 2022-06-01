package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/supremacy_rpcclient"
	xsynTypes "xsyn-services/types"
)

func IsUserAssetColumn(col string) bool {
	switch col {
	case boiler.UserAssetColumns.ID,
		boiler.UserAssetColumns.CollectionID,
		boiler.UserAssetColumns.TokenID,
		boiler.UserAssetColumns.Tier,
		boiler.UserAssetColumns.Hash,
		boiler.UserAssetColumns.OwnerID,
		boiler.UserAssetColumns.Data,
		boiler.UserAssetColumns.Attributes,
		boiler.UserAssetColumns.Name,
		boiler.UserAssetColumns.ImageURL,
		boiler.UserAssetColumns.ExternalURL,
		boiler.UserAssetColumns.Description,
		boiler.UserAssetColumns.BackgroundColor,
		boiler.UserAssetColumns.AnimationURL,
		boiler.UserAssetColumns.YoutubeURL,
		boiler.UserAssetColumns.UnlockedAt,
		boiler.UserAssetColumns.MintedAt,
		boiler.UserAssetColumns.OnChainStatus,
		boiler.UserAssetColumns.DeletedAt,
		boiler.UserAssetColumns.DataRefreshedAt:
		return true
	default:
		return false
	}
}

func IsUserAsset1155Column(col string) bool {
	switch col {
	case boiler.UserAssets1155Columns.ID,
		boiler.UserAssets1155Columns.CollectionID,
		boiler.UserAssets1155Columns.Label,
		boiler.UserAssets1155Columns.Description,
		boiler.UserAssets1155Columns.OwnerID,
		boiler.UserAssets1155Columns.Attributes,
		boiler.UserAssets1155Columns.ImageURL,
		boiler.UserAssets1155Columns.AnimationURL,
		boiler.UserAssets1155Columns.ServiceID:
		return true
	default:
		return false
	}
}

type AssetListOpts struct {
	UserID          xsynTypes.UserID
	Sort            *ListSortRequest
	Filter          *ListFilterRequest
	AttributeFilter *AttributeFilterRequest
	AssetType       string
	Search          string
	PageSize        int
	Page            int
}

func AssetList721(opts *AssetListOpts) (int64, []*xsynTypes.UserAsset, error) {
	var queryMods []qm.QueryMod

	// create the where owner id = clause
	queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
		Table:    boiler.TableNames.UserAssets,
		Column:   boiler.UserAssetColumns.OwnerID,
		Operator: OperatorValueTypeEquals,
		Value:    opts.UserID.String(),
	}, 0, ""))

	// Filters // TODO: filtering
	//if opts.Filter != nil {
	//	// if we have filter
	//	for i, f := range opts.Filter.Items {
	//		// validate it is the right table and valid column
	//		if f.Table == boiler.TableNames.UserAssets && IsAssetColumn(f.Column) {
	//			queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
	//		}
	//
	//	}
	//}

	if opts.AssetType != "" {
		queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
			Table:    boiler.TableNames.UserAssets,
			Column:   boiler.UserAssetColumns.AssetType,
			Operator: OperatorValueTypeEquals,
			Value:    opts.AssetType,
		}, 0, ""))
	}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?))",
					boiler.TableNames.UserAssets,
					boiler.UserAssetColumns.Name,
				),
					xSearch,
				))
		}
	}

	total, err := boiler.UserAssets(
		queryMods...,
	).Count(passdb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	// Sort
	if opts.Sort != nil && opts.Sort.Table == boiler.TableNames.UserAssets && IsUserAssetColumn(opts.Sort.Column) && opts.Sort.Direction.IsValid() {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.UserAssets, opts.Sort.Column, opts.Sort.Direction)))
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s desc", boiler.TableNames.UserAssets, boiler.UserAssetColumns.Name)))
	}

	// Limit/Offset
	if opts.PageSize > 0 {
		queryMods = append(queryMods, qm.Limit(opts.PageSize))
	}
	if opts.Page > 0 {
		queryMods = append(queryMods, qm.Offset(opts.PageSize*(opts.Page-1)))
	}

	boilerAssets, err := boiler.UserAssets(queryMods...).All(passdb.StdConn)
	if err != nil {
		return 0, nil, err
	}

	return total, xsynTypes.UserAssets721FromBoiler(boilerAssets), nil
}

func AssetList1155(opts *AssetListOpts) (int64, []*xsynTypes.User1155Asset, error) {
	var queryMods []qm.QueryMod

	// create the where owner id = clause
	queryMods = append(queryMods, GenerateListFilterQueryMod(ListFilterRequestItem{
		Table:    boiler.TableNames.UserAssets1155,
		Column:   boiler.UserAssets1155Columns.OwnerID,
		Operator: OperatorValueTypeEquals,
		Value:    opts.UserID.String(),
	}, 0, ""))

	// Filters // TODO: filtering
	//if opts.Filter != nil {
	//	// if we have filter
	//	for i, f := range opts.Filter.Items {
	//		// validate it is the right table and valid column
	//		if f.Table == boiler.TableNames.UserAssets && IsAssetColumn(f.Column) {
	//			queryMods = append(queryMods, GenerateListFilterQueryMod(*f, i+1, opts.Filter.LinkOperator))
	//		}
	//
	//	}
	//}

	// Search
	if opts.Search != "" {
		xSearch := ParseQueryText(opts.Search, true)
		if len(xSearch) > 0 {
			queryMods = append(queryMods,
				qm.And(fmt.Sprintf(
					"((to_tsvector('english', %[1]s.%[2]s) @@ to_tsquery(?))",
					boiler.TableNames.UserAssets1155,
					boiler.UserAssets1155Columns.Label,
				),
					xSearch,
				))
		}
	}

	total, err := boiler.UserAssets1155S(
		queryMods...,
	).Count(passdb.StdConn)
	if err != nil {
		return 0, nil, err
	}
	// Sort
	if opts.Sort != nil && opts.Sort.Table == boiler.TableNames.UserAssets1155 && IsUserAsset1155Column(opts.Sort.Column) && opts.Sort.Direction.IsValid() {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s %s", boiler.TableNames.UserAssets1155, opts.Sort.Column, opts.Sort.Direction)))
	} else {
		queryMods = append(queryMods, qm.OrderBy(fmt.Sprintf("%s.%s desc", boiler.TableNames.UserAssets1155, boiler.UserAssets1155Columns.Label)))
	}

	// Limit/Offset
	if opts.PageSize > 0 {
		queryMods = append(queryMods, qm.Limit(opts.PageSize))
	}
	if opts.Page > 0 {
		queryMods = append(queryMods, qm.Offset(opts.PageSize*(opts.Page-1)))
	}

	queryMods = append(queryMods, qm.Load(boiler.UserAssets1155Rels.Collection, qm.Select(boiler.CollectionColumns.Slug, boiler.CollectionColumns.ID, boiler.CollectionColumns.MintContract)))
	queryMods = append(queryMods, boiler.UserAssets1155Where.Count.GT(0))
	boilerAssets, err := boiler.UserAssets1155S(queryMods...).All(passdb.StdConn)
	if err != nil {
		return 0, nil, err
	}

	return total, xsynTypes.UserAssets1155FromBoiler(boilerAssets), nil
}

func PurchasedItemRegister(storeItemID uuid.UUID, ownerID uuid.UUID) ([]*xsynTypes.UserAsset, error) {
	passlog.L.Trace().Str("fn", "PurchasedItemRegister").Msg("db func")
	req := supremacy_rpcclient.TemplateRegisterReq{TemplateID: storeItemID, OwnerID: ownerID}
	resp := &supremacy_rpcclient.TemplateRegisterResp{}
	err := supremacy_rpcclient.SupremacyClient.Call("S.TemplateRegisterHandler", req, resp)
	if err != nil {
		return nil, terror.Error(err, "communication to supremacy has failed")
	}
	var newItems []*xsynTypes.UserAsset
	// for each asset, assign it on our database
	for _, itm := range resp.Assets {
		userAsset, err := RegisterUserAsset(itm, xsynTypes.SupremacyGameUserID.String())
		if err != nil {
			return nil, terror.Error(err, "Failed to register new user asset.")
		}

		newItems = append(newItems, xsynTypes.UserAssetFromBoiler(userAsset))
	}

	return newItems, nil
}

func RegisterUserAsset(itm *supremacy_rpcclient.XsynAsset, serviceID string) (*boiler.UserAsset, error) {
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
		TokenID:         itm.TokenID,
		Tier:            itm.Tier,
		Hash:            itm.Hash,
		OwnerID:         itm.OwnerID,
		Data:            itm.Data,
		Attributes:      jsonAtrribs,
		Name:            itm.Name,
		AssetType:       itm.AssetType,
		ImageURL:        itm.ImageURL,
		ExternalURL:     itm.ExternalURL,
		Description:     itm.Description,
		BackgroundColor: itm.BackgroundColor,
		AnimationURL:    itm.AnimationURL,
		YoutubeURL:      itm.YoutubeURL,
		UnlockedAt:      itm.UnlockedAt,
		AvatarURL:       itm.AvatarURL,
		MintedAt:        itm.MintedAt,
		OnChainStatus:   itm.OnChainStatus,
		DataRefreshedAt: time.Now(),
		LockedToService: null.StringFrom(serviceID),
	}

	// see if old asset exists
	oldAsset, err := boiler.PurchasedItemsOlds(
		boiler.PurchasedItemsOldWhere.CollectionID.EQ(collection.ID),
		boiler.PurchasedItemsOldWhere.ExternalTokenID.EQ(int(itm.TokenID)),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err)
	}

	if oldAsset != nil {
		boilerAsset.MintedAt = oldAsset.MintedAt
		boilerAsset.OnChainStatus = oldAsset.OnChainStatus
		boilerAsset.UnlockedAt = oldAsset.UnlockedAt
	}

	// if minted tell gameserver item is xsyn locked
	if boilerAsset.OnChainStatus == "STAKABLE" {
		boilerAsset.LockedToService = null.String{}
		err := supremacy_rpcclient.AssetUnlockFromSupremacy(xsynTypes.UserAssetFromBoiler(boilerAsset), 0)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	// if staked tell gameserver item is market locked
	if boilerAsset.OnChainStatus == "UNSTAKABLE" {
		err := supremacy_rpcclient.AssetLockToSupremacy(xsynTypes.UserAssetFromBoiler(boilerAsset), 0, false)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	// if staked tell gameserver item is market locked
	// UNSTAKABLE_OLD = still staked on old contract, not market tradable
	if boilerAsset.OnChainStatus == "UNSTAKABLE_OLD" {
		err := supremacy_rpcclient.AssetLockToSupremacy(xsynTypes.UserAssetFromBoiler(boilerAsset), 0, true)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	err = boilerAsset.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Interface("itm", itm).Err(err).Msg("failed to register new asset - can't insert asset")
		return nil, err
	}

	return boilerAsset, nil
}
