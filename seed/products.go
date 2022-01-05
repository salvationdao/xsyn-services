package seed

import (
	"passport"
	"passport/db"
	"context"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/ninja-software/terror/v2"
	"syreclabs.com/go/faker"
)

func (s *Seeder) Products(ctx context.Context) error {
	photoIDs := []int{}
	productSlugs := []string{}
	count := 0
	for i := 0; i < 80; i++ {
		// Get unique name/slug
		var productName string
		var productSlug string

		for {
			productName = faker.Commerce().ProductName()
			productSlug = slug.Make(productName)

			exists := false
			for _, slug := range productSlugs {
				if productSlug == slug {
					exists = true
					break
				}
			}
			if !exists {
				break
			}
		}
		productSlugs = append(productSlugs, productSlug)

		// Get image
		var photoID int
		for {
			photoID = faker.RandomInt(0, 992)
			exists := false
			for _, id := range photoIDs {
				if photoID == id {
					exists = true
					break
				}
			}
			if !exists {
				break
			}
		}
		photoIDs = append(photoIDs, photoID)
		image, err := passport.BlobFromURL(
			fmt.Sprintf("https://ninjasoftware-static-media.s3.ap-southeast-2.amazonaws.com/passport-seeding/images/400x250/%d.jpg", photoID),
			productSlug+".jpg",
		)
		if err != nil {
			continue
		}

		// Insert image
		err = db.BlobInsert(ctx, s.Conn, image)
		if err != nil {
			return terror.Error(err)
		}

		// Create product
		product := &passport.Product{
			Name:        productName,
			Slug:        productSlug,
			Description: faker.Lorem().Paragraph(faker.RandomInt(1, 4)),
			ImageID:     &image.ID,
		}
		err = db.ProductCreate(ctx, s.Conn, product)
		if err != nil {
			return terror.Error(err)
		}

		count++
	}
	fmt.Printf(" Seeded %d products\n", count)

	return nil
}
