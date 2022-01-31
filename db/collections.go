package db

import (
	"context"
	"passport"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/ninja-software/terror/v2"
)

const CollectionGetQuery string = `
SELECT 
collections.id,
collections.name,
collections.deleted_at,
collections.updated_at,
collections.created_at
` + CollectionGetQueryFrom

const CollectionGetQueryFrom = `
FROM collections
`

// CollectionGet returns a collection by name
func CollectionGet(ctx context.Context, conn Conn, name string) (*passport.Collection, error) {
	collection := &passport.Collection{}
	q := CollectionGetQuery + ` WHERE collections.name = $1`

	err := pgxscan.Get(ctx, conn, collection, q, name)
	if err != nil {
		return nil, terror.Error(err, "Issue getting collection.")
	}
	return collection, nil
}
