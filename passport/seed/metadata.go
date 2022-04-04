package seed

import (
	"context"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

func (s *Seeder) SeedItemMetadata(ctx context.Context) (warMachines, abilities, weapons, utility []*types.XsynMetadata, error error) {
	supremacyCollectionBoiler, err := db.CollectionBySlug(ctx, s.Conn, "supremacy")
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	collection := &types.Collection{
		ID:            types.CollectionID(uuid.Must(uuid.FromString(supremacyCollectionBoiler.ID))),
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

func (s *Seeder) SeedWarMachine(ctx context.Context, weapons, abilities []*types.XsynMetadata, collection *types.Collection) ([]*types.XsynMetadata, error) {
	des := "A big ass War Machine - links to attached nfts"

	newNFTs := []*types.XsynMetadata{
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "",
			Description:        &des,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Power Grid",
					Value:       170,
					DisplayType: types.Number,
				},
				{
					TraitType:   "CPU",
					Value:       100,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       2,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       2,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Speed",
					Value:       4,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Power Grid",
					Value:       140,
					DisplayType: types.Number,
				},
				{
					TraitType:   "CPU",
					Value:       120,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Turret Hardpoints",
					Value:       1,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       1,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Speed",
					Value:       6,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Power Grid",
					Value:       140,
					DisplayType: types.Number,
				},
				{
					TraitType:   "CPU",
					Value:       150,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Weapon Hardpoints",
					Value:       2,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Utility Slots",
					Value:       2,
					DisplayType: types.Number,
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

func (s *Seeder) SeedWeapons(ctx context.Context, collection *types.Collection) ([]*types.XsynMetadata, error) {
	pulseRifleDes := "A rifle that shoots pulses"
	autoCannonDes := "A cannon that shoots projectiles"
	rocketLauncherDes := "A shoulder weapon that fires rockets"
	rapidRocketLauncherDes := "A shoulder weapon that fires rockets quickly"

	newNFT := []*types.XsynMetadata{
		// pulse rifles
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Pulse Rifle",
			Description:        &pulseRifleDes,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       6,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       6,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       40,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       20,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       15,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       40,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       20,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       15,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Weapon Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       25,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       120,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       10,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shoots per minute",
					Value:       120,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Magazine Size",
					Value:       10,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Reload Time",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Turret Hardpoint",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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

func (s *Seeder) SeedUtility(ctx context.Context, collection *types.Collection) ([]*types.XsynMetadata, error) {
	largeShieldDesc := "A large Shield"
	mediumShieldDesc := "A Medium shield"
	smallShieldDesc := "A small shield"

	newNFT := []*types.XsynMetadata{
		// large shield
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Large Shield",
			Description:        &largeShieldDesc,
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       5,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       100,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       50,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       5,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       30,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       100,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       50,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       8,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       20,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       70,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       40,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       8,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       20,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       70,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       40,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       15,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       10,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Regen Per Second",
					Value:       15,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Shield Recharge Cooldown",
					Value:       10,
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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
			Attributes: []*types.Attribute{
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
					DisplayType: types.Number,
				},
				{
					TraitType: "Required Slot",
					Value:     "Utility",
				},
				{
					TraitType:   "Required Power Grid",
					Value:       50,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Required CPU",
					Value:       20,
					DisplayType: types.Number,
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
