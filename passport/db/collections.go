package db

import (
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func RogueCollection() (*boiler.Collection, error) {
	collection, err := boiler.Collections(
		boiler.CollectionWhere.Name.EQ("Supremacy"),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return collection, nil
}
func AICollection() (*boiler.Collection, error) {
	collection, err := boiler.Collections(
		boiler.CollectionWhere.Name.EQ("Supremacy AI"),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return collection, nil
}
func GenesisCollection() (*boiler.Collection, error) {
	collection, err := boiler.Collections(
		boiler.CollectionWhere.Name.EQ("Supremacy Genesis"),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return collection, nil
}
func LimitedReleaseCollection() (*boiler.Collection, error) {
	collection, err := boiler.Collections(
		boiler.CollectionWhere.Name.EQ("Supremacy Limited Release"),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return collection, nil
}
func Collection(id uuid.UUID) (*boiler.Collection, error) {
	collection, err := boiler.FindCollection(
		passdb.StdConn,
		id.String(),
	)
	if err != nil {
		return nil, err
	}
	return collection, nil
}
func CollectionByMintAddress(mintAddr common.Address) (*boiler.Collection, error) {
	collection, err := boiler.Collections(
		boiler.CollectionWhere.MintContract.EQ(
			null.StringFrom(
				mintAddr.Hex(),
			),
		),
	).One(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return collection, nil
}

// CollectionBySlug returns a collection by slug
func CollectionBySlug(slug string) (*boiler.Collection, error) {
	collection, err := boiler.Collections(
		boiler.CollectionWhere.Slug.EQ(slug),
	).One(passdb.StdConn)
	if err != nil {
		return nil, terror.Error(err, "Issue getting collection.")
	}
	return collection, nil
}

// CollectionsList gets a list of collections depending on the filters
func CollectionsList() ([]*boiler.Collection, error) {
	collections, err := boiler.Collections(
		boiler.CollectionWhere.Slug.NEQ("supremacy-ai"),
		qm.And("slug != ?", "supremacy"),
	).All(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return collections, nil
}

// CollectionsList gets a list of collections depending on the filters
func CollectionsVisibleList() ([]*boiler.Collection, error) {
	collections, err := boiler.Collections(
		boiler.CollectionWhere.Slug.NEQ("supremacy-ai"),
		qm.And("slug != ?", "supremacy"),
		boiler.CollectionWhere.IsVisible.EQ(null.BoolFrom(true)),
	).All(passdb.StdConn)
	if err != nil {
		return nil, err
	}
	return collections, nil
}
