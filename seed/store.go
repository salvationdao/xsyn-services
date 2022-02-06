package seed

import (
	"context"
	"passport"
	"passport/db"
	"time"

	"github.com/ninja-software/terror/v2"
)

func (s *Seeder) SeedInitialStoreItems(ctx context.Context) error {
	supremacyCollection, err := db.CollectionGet(ctx, s.Conn, "Supremacy")
	if err != nil {
		return terror.Error(err)
	}

	initialStoreItems := []*passport.StoreItem{
		{
			Name:            "",
			FactionID:       passport.ZaibatsuFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     200000,
			AmountAvailable: 10,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Legendary",
				},
				{
					TraitType: "Brand",
					Value:     "Zaibatsu",
				},
				{
					TraitType: "Model",
					Value:     "WREX",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Neon",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: passport.Number,
				},
				//{
				//	TraitType:   "Power Grid",
				//	Value:       170,
				//	DisplayType: passport.Number,
				//},
				//{
				//	TraitType:   "CPU",
				//	Value:       100,
				//	DisplayType: passport.Number,
				//},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Sniper Rifle",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Laser Sword",
				},
				{
					TraitType: "Turret One",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Turret Two",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		{
			Name:            "",
			FactionID:       passport.ZaibatsuFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     30000,
			AmountAvailable: 200,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Rare",
				},
				{
					TraitType: "Brand",
					Value:     "Zaibatsu",
				},
				{
					TraitType: "Model",
					Value:     "WREX",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Blue",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: passport.Number,
				},
				//{
				//	TraitType:   "Power Grid",
				//	Value:       170,
				//	DisplayType: passport.Number,
				//},
				//{
				//	TraitType:   "CPU",
				//	Value:       100,
				//	DisplayType: passport.Number,
				//},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Sniper Rifle",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Laser Sword",
				},
				{
					TraitType: "Turret One",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Turret Two",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		{
			Name:            "",
			FactionID:       passport.ZaibatsuFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     100,
			AmountAvailable: 20000,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Common",
				},
				{
					TraitType: "Brand",
					Value:     "Zaibatsu",
				},
				{
					TraitType: "Model",
					Value:     "WREX",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Grey",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: passport.Number,
				},
				//{
				//	TraitType:   "Power Grid",
				//	Value:       170,
				//	DisplayType: passport.Number,
				//},
				//{
				//	TraitType:   "CPU",
				//	Value:       100,
				//	DisplayType: passport.Number,
				//},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Sniper Rifle",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Laser Sword",
				},
				{
					TraitType: "Turret One",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Turret Two",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		// red mountain
		{
			Name:            "",
			FactionID:       passport.RedMountainFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     100000,
			AmountAvailable: 10,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Legendary",
				},
				{
					TraitType: "Brand",
					Value:     "Red Mountain",
				},
				{
					TraitType: "Model",
					Value:     "BXSD",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Neon",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       1750,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Turret One",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Turret Two",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		{
			Name:            "",
			FactionID:       passport.RedMountainFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     20000,
			AmountAvailable: 100,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Rare",
				},
				{
					TraitType: "Brand",
					Value:     "Red Mountain",
				},
				{
					TraitType: "Model",
					Value:     "BXSD",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Red",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       1750,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Turret One",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Turret Two",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		{
			Name:            "",
			FactionID:       passport.RedMountainFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     100,
			AmountAvailable: 50000,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Common",
				},
				{
					TraitType: "Brand",
					Value:     "Red Mountain",
				},
				{
					TraitType: "Model",
					Value:     "BXSD",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Black",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       1750,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Turret One",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Turret Two",
					Value:     "Rocket Pod",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		// boston
		{
			Name:            "",
			FactionID:       passport.BostonCyberneticsFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     100000,
			AmountAvailable: 10,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Legendary",
				},
				{
					TraitType: "Brand",
					Value:     "Boston Cybernetics",
				},
				{
					TraitType: "Model",
					Value:     "XFVS",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Neon",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1200,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1200,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2750,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Plasma Rifle",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Sword",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		{
			Name:            "",
			FactionID:       passport.BostonCyberneticsFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     20000,
			AmountAvailable: 500,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Rare",
				},
				{
					TraitType: "Brand",
					Value:     "Boston Cybernetics",
				},
				{
					TraitType: "Model",
					Value:     "XFVS",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Red",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1200,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1200,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2750,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Plasma Rifle",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Sword",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
		{
			Name:            "",
			FactionID:       passport.BostonCyberneticsFactionID,
			CollectionID:    supremacyCollection.ID,
			Description:     "",
			Image:           "",
			UsdCentCost:     100,
			AmountAvailable: 10000,
			//SoldAfter: time,
			SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC),
			Attributes: []*passport.Attribute{
				{
					TraitType: "Rarity",
					Value:     "Common",
				},
				{
					TraitType: "Brand",
					Value:     "Boston Cybernetics",
				},
				{
					TraitType: "Model",
					Value:     "XFVS",
				},
				{
					TraitType: "Name",
					Value:     "",
				},
				{
					TraitType: "SubModel",
					Value:     "Blue",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1200,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1200,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2750,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "Plasma Rifle",
				},
				{
					TraitType: "Weapon Two",
					Value:     "Sword",
				},
				{
					TraitType: "Utility One",
					Value:     "Shield",
				},
			},
		},
	}

	for _, si := range initialStoreItems {
		err := db.AddItemToStore(ctx, s.Conn, si)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}

func (s *Seeder) SeedCollections(ctx context.Context) ([]*passport.Collection, error) {
	collectionNames := []string{"Supremacy"}
	collections := []*passport.Collection{}
	for _, name := range collectionNames {
		collection := &passport.Collection{
			Name: name,
		}

		err := db.CollectionInsert(ctx, s.Conn, collection)
		if err != nil {
			return nil, terror.Error(err)
		}

		collections = append(collections, collection)
	}
	return collections, nil
}
