package seed

import (
	"context"
	"passport"
	"passport/db"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

func (s *Seeder) SeedItemMetadata(ctx context.Context) (warMachines, abilities, weapons, utility []*passport.XsynMetadata, error error) {
	supremacyCollectionBoiler, err := db.CollectionBySlug(ctx, s.Conn, "supremacy")
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	collection := &passport.Collection{
		ID:            passport.CollectionID(uuid.Must(uuid.FromString(supremacyCollectionBoiler.ID))),
		Name:          supremacyCollectionBoiler.Name,
		Slug:          supremacyCollectionBoiler.Slug,
		DeletedAt:     &supremacyCollectionBoiler.DeletedAt.Time,
		UpdatedAt:     supremacyCollectionBoiler.UpdatedAt,
		CreatedAt:     supremacyCollectionBoiler.CreatedAt,
		MintContract:  supremacyCollectionBoiler.MintContract.String,
		StakeContract: supremacyCollectionBoiler.StakeContract.String,
	}

	utility, err = s.SeedUtility(ctx, collection)
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	weapons, err = s.SeedWeapons(ctx, collection)
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	warMachines, err = s.SeedWarMachine(ctx, weapons, abilities, collection)
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	return
}

func (s *Seeder) SeedWarMachine(ctx context.Context, weapons, abilities []*passport.XsynMetadata, collection *passport.Collection) ([]*passport.XsynMetadata, error) {
	des := "A big ass War Machine - links to attached nfts"

	newNFTs := []*passport.XsynMetadata{
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "",
			Description:        &des,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Big War Machine",
				},
				{
					TraitType: "Brand",
					Value:     "Placeholder Brand",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType: "War Machine Win Rank",
					Value:     1,
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Power Grid",
					Value:       170,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "CPU",
					Value:       100,
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
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "none",
					TokenID:   weapons[0].ExternalTokenID,
				},
				{
					TraitType: "Weapon Two",
					Value:     "none",
				},
				{
					TraitType: "Turret One",
					Value:     "none",
				},
				{
					TraitType: "Turret Two",
					Value:     "none",
				},
				{
					TraitType: "Utility One",
					Value:     "none",
				},
				{
					TraitType: "Utility Two",
					Value:     "none",
				},
				{
					TraitType: "Ability One",
					Value:     "none",
				},
				{
					TraitType: "Ability Two",
					Value:     "none",
				},
			},
		},
		{
			CollectionID: collection.ID,
			Name:         "Medium War Machine",
			Description:  &des,
			ExternalUrl:  "",
			Image:        "",
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Medium War Machine",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType: "War Machine Win Rank",
					Value:     3,
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       300,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       4,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Power Grid",
					Value:       140,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "CPU",
					Value:       120,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "none",
				},
				{
					TraitType: "Weapon Two",
					Value:     "none",
				},
				{
					TraitType: "Turret One",
					Value:     "none",
				},
				{
					TraitType: "Utility One",
					Value:     "none",
				},
				{
					TraitType: "Ability One",
					Value:     "none",
				},
				{
					TraitType: "Ability Two",
					Value:     "none",
				},
			},
		},
		{
			CollectionID: collection.ID,
			Name:         "Small War Machine",
			Description:  &des,
			ExternalUrl:  "",
			Image:        "",
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Small War Machine",
				},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType: "War Machine Win Rank",
					Value:     2,
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       250,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       6,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Power Grid",
					Value:       140,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "CPU",
					Value:       150,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       2,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Weapon One",
					Value:     "none",
				},
				{
					TraitType: "Weapon Two",
					Value:     "none",
				},
				{
					TraitType: "Utility One",
					Value:     "none",
				},
				{
					TraitType: "Utility Two",
					Value:     "none",
				},
				{
					TraitType: "Ability One",
				},
				{
					TraitType: "Ability Two",
				},
			},
		},
	}

	for _, nft := range newNFTs {
		err := db.XsynMetadataInsert(ctx, s.Conn, nft, s.ExternalURL)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	return newNFTs, nil
}

func (s *Seeder) SeedWeapons(ctx context.Context, collection *passport.Collection) ([]*passport.XsynMetadata, error) {
	pulseRifleDes := "A rifle that shoots pulses"
	autoCannonDes := "A cannon that shoots projectiles"
	rocketLauncherDes := "A shoulder weapon that fires rockets"
	rapidRocketLauncherDes := "A shoulder weapon that fires rockets quickly"

	newNFT := []*passport.XsynMetadata{
		// pulse rifles
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Pulse Rifle",
			Description:        &pulseRifleDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Pulse Rifle",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     1,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       6,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Pulse Rifle",
			Description:        &pulseRifleDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Pulse Rifle",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     0,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       6,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: passport.Number,
				},
			},
		},
		// auto cannons
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Auto Cannon",
			Description:        &autoCannonDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     4,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       12,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       40,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       20,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       15,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Auto Cannon",
			Description:        &autoCannonDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Auto Cannon",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     4,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       12,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       40,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       20,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       15,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: passport.Number,
				},
			},
		},
		// rocket launchers
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Rocket Launcher",
			Description:        &rocketLauncherDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Rocket Launcher",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     2,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       150,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Rocket Launcher",
			Description:        &rocketLauncherDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Rocket Launcher",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     2,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       150,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
		// rapid rocket launchers
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Rapid Rocket Launcher",
			Description:        &rapidRocketLauncherDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Rapid Rocket Launcher",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     2,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       25,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       120,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       10,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Rapid Rocket Launcher",
			Description:        &rapidRocketLauncherDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Rapid Rocket Launcher",
				},
				{
					TraitType: "Asset Type",
					Value:     "Weapon",
				},
				{
					TraitType: "Kills",
					Value:     2,
				},
				{
					TraitType:   "Damage Per shot",
					Value:       25,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       120,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       10,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
	}

	for _, nft := range newNFT {
		err := db.XsynMetadataInsert(ctx, s.Conn, nft, s.ExternalURL)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	return newNFT, nil
}

func (s *Seeder) SeedUtility(ctx context.Context, collection *passport.Collection) ([]*passport.XsynMetadata, error) {
	largeShieldDesc := "A large Shield"
	mediumShieldDesc := "A Medium shield"
	smallShieldDesc := "A small shield"

	newNFT := []*passport.XsynMetadata{
		// large shield
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Large Shield",
			Description:        &largeShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Large Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Shield Hit Points",
					Value:       500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       5,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       100,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       50,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Large Shield",
			Description:        &largeShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Large Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Shield Hit Points",
					Value:       500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       5,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       30,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       100,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       50,
					DisplayType: passport.Number,
				},
			},
		},
		// med shield
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        &mediumShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Medium Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Shield Hit Points",
					Value:       300,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       8,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       20,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       70,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       40,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        &mediumShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Medium Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Shield Hit Points",
					Value:       300,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       8,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       20,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       70,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       40,
					DisplayType: passport.Number,
				},
			},
		},
		// small shields
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Small Shield",
			Description:        &smallShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Small Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Shield Hit Points",
					Value:       150,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       15,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       10,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Small Shield",
			Description:        &smallShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Small Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Shield Hit Points",
					Value:       150,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       15,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       10,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
		// healing drone
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        &mediumShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Medium Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Health Per Second",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        &mediumShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Medium Shield",
				},
				{
					TraitType: "Asset Type",
					Value:     "Utility",
				},
				{
					TraitType:   "Health Per Second",
					Value:       1,
					DisplayType: passport.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: passport.Number,
				},
			},
		},
	}

	for _, nft := range newNFT {
		err := db.XsynMetadataInsert(ctx, s.Conn, nft, s.ExternalURL)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	return newNFT, nil
}
