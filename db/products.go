package db

import (
	"context"
	"fmt"
	"passport"
	"strings"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/gosimple/slug"
	"github.com/ninja-software/terror/v2"
)

type ProductColumn string

const (
	ProductColumnID          ProductColumn = "id"
	ProductColumnSlug        ProductColumn = "slug"
	ProductColumnName        ProductColumn = "name"
	ProductColumnDescription ProductColumn = "description"
	ProductColumnImageID     ProductColumn = "imageID"

	ProductColumnDeletedAt ProductColumn = "deleted_at"
	ProductColumnUpdatedAt ProductColumn = "updated_at"
	ProductColumnCreatedAt ProductColumn = "created_at"
)

func (ic ProductColumn) IsValid() error {
	switch ic {
	case ProductColumnID,
		ProductColumnSlug,
		ProductColumnName,
		ProductColumnDescription,
		ProductColumnImageID,
		ProductColumnDeletedAt,
		ProductColumnUpdatedAt,
		ProductColumnCreatedAt:
		return nil
	}
	return terror.Error(fmt.Errorf("invalid product column type"))
}

const ProductGetQuery string = `--sql
SELECT 
	id, slug, name, description, image_id, deleted_at, updated_at, created_at
FROM products
`

// ProductGet returns a product by given ID
func ProductGet(ctx context.Context, conn Conn, productID passport.ProductID) (*passport.Product, error) {
	product := &passport.Product{}
	q := ProductGetQuery + ` WHERE id = $1`
	err := pgxscan.Get(ctx, conn, product, q, productID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return product, nil
}

// ProductGetBySlug returns a product by given slug
func ProductGetBySlug(ctx context.Context, conn Conn, slug string) (*passport.Product, error) {
	product := &passport.Product{}
	q := ProductGetQuery + ` WHERE slug = $1`
	err := pgxscan.Get(ctx, conn, product, q, slug)
	if err != nil {
		return nil, terror.Error(err)
	}
	return product, nil
}

// ProductCreate will create a new product
func ProductCreate(ctx context.Context, conn Conn, product *passport.Product) error {
	slug := slug.Make(product.Name)
	q := `--sql
		INSERT INTO products (slug, name, description, image_id)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id, slug, name, description, image_id, deleted_at, updated_at, created_at`
	err := pgxscan.Get(ctx,
		conn,
		product,
		q,
		slug,
		product.Name,
		product.Description,
		product.ImageID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ProductUpdate will update an existing product
func ProductUpdate(ctx context.Context, conn Conn, product *passport.Product) error {
	slug := slug.Make(product.Name)
	q := `--sql
		UPDATE products
		SET 
			slug = $2, name = $3, description = $4, image_id = $5
		WHERE id = $1`
	_, err := conn.Exec(ctx,
		q,
		product.ID,
		slug,
		product.Name,
		product.Description,
		product.ImageID,
	)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ProductArchiveUpdate will update an existing product as archived or unarchived
func ProductArchiveUpdate(ctx context.Context, conn Conn, id passport.ProductID, archived bool) error {
	var deletedAt *time.Time
	if archived {
		now := time.Now()
		deletedAt = &now
	}
	q := `--sql
		UPDATE products
		SET deleted_at = $2
		WHERE id = $1`
	_, err := conn.Exec(ctx, q, id, deletedAt)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// ProductList will grab a list of products in offset pagination format
func ProductList(ctx context.Context,
	conn Conn,
	result *[]*passport.Product,
	search string,
	archived bool,
	filter *ListFilterRequest,
	offset int,
	pageSize int,
	sortBy ProductColumn,
	sortDir SortByDir,
) (int, error) {
	// Prepare Filters
	var args []interface{}
	filterConditionsString := ""

	if filter != nil {
		filterConditions := []string{}
		for i, f := range filter.Items {
			column := ProductColumn(f.ColumnField)
			err := column.IsValid()
			if err != nil {
				return 0, terror.Error(err)
			}

			condition, value := GenerateListFilterSQL(f.ColumnField, f.Value, f.OperatorValue, i+1)
			if condition != "" {
				filterConditions = append(filterConditions, condition)
				args = append(args, value)
			}
		}

		if len(filterConditions) > 0 {
			filterConditionsString = " AND (" + strings.Join(filterConditions, " "+string(filter.LinkOperator)+" ") + ")"
		}
	}

	archiveCondition := "IS NULL"
	if archived {
		archiveCondition = "IS NOT NULL"
	}

	searchCondition := ""
	if search != "" {
		xsearch := ParseQueryText(search, true)
		if len(xsearch) > 0 {
			args = append(args, xsearch)
			searchCondition = fmt.Sprintf(" AND keywords @@ to_tsquery($%d)", len(args))
		}
	}

	// Get Total Found
	q := `--sql
		SELECT COUNT(*)
		FROM products
		WHERE deleted_at ` + archiveCondition + filterConditionsString + searchCondition
	var totalRows int
	err := pgxscan.Get(ctx, conn, &totalRows, q, args...)
	if err != nil {
		return 0, terror.Error(err)
	}
	if totalRows == 0 {
		return totalRows, nil
	}

	// Order and Limit
	orderBy := " ORDER BY created_at desc"
	if sortBy != "" {
		err := sortBy.IsValid()
		if err != nil {
			return 0, terror.Error(err)
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", sortBy, sortDir)
	}
	limit := ""
	if pageSize > 0 {
		limit = fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)
	}

	// Get Paginated Result
	q = fmt.Sprintf(
		ProductGetQuery+`--sql
		WHERE deleted_at %s
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
	err = pgxscan.Select(ctx, conn, result, q, args...)
	if err != nil {
		return 0, terror.Error(err)
	}
	return totalRows, nil
}
