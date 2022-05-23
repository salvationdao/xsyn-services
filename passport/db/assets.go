package db

import (
	"fmt"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"
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
		boiler.UserAssetColumns.XsynLocked,
		boiler.UserAssetColumns.DeletedAt,
		boiler.UserAssetColumns.DataRefreshedAt:
		return true
	default:
		return false
	}
}


type AssetListOpts struct {
	UserID              types.UserID
	Sort              *ListSortRequest
	Filter              *ListFilterRequest
	AttributeFilter     *AttributeFilterRequest
	AssetType           string
	Search              string
	PageSize            int
	Page                int
}

func AssetList(opts *AssetListOpts) (int64, []*types.UserAsset, error) {
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

	return total, types.UserAssetsFromBoiler(boilerAssets), nil
}
