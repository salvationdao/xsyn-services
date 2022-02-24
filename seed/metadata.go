package seed

import (
	"context"
	"passport"
	"passport/db"

	"github.com/ninja-software/terror/v2"
)

func (s *Seeder) SeedItemMetadata(ctx context.Context) (warMachines, abilities, weapons, utility []*passport.XsynMetadata, error error) {
	supremacyCollection, err := db.CollectionGet(ctx, s.Conn, "supremacy")
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	utility, err = s.SeedUtility(ctx, supremacyCollection)
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	weapons, err = s.SeedWeapons(ctx, supremacyCollection)
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	abilities, err = s.SeedWarMachineAbility(ctx, supremacyCollection)
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	warMachines, err = s.SeedWarMachine(ctx, weapons, abilities, supremacyCollection)
	if err != nil {
		return nil, nil, nil, nil, terror.Error(err)
	}

	return
}

func (s *Seeder) SeedWarMachine(ctx context.Context, weapons, abilities []*passport.XsynMetadata, collection *passport.Collection) ([]*passport.XsynMetadata, error) {
	newNFTs := []*passport.XsynMetadata{
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "",
			Description:        "A big ass War Machine - links to attached nfts",
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
					TokenID:   weapons[0].TokenID,
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
			Description:  "Average size War Machine - links to attached nfts",
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
			Description:  "A iddy biddy tiny War Machine - links to attached nfts",
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
	newNFT := []*passport.XsynMetadata{
		// pulse rifles
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Pulse Rifle",
			Description:        "A rifle that shoots pulses",
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
			Description:        "A rifle that shoots pulses",
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
			Description:        "A cannon that shoots projectiles",
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
			Description:        "A cannon that shoots projectiles",
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
			Description:        "A shoulder weapon that fires rockets",
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
			Description:        "A shoulder weapon that fires rockets",
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
			Description:        "A shoulder weapon that fires rockets quickly",
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
			Description:        "A shoulder weapon that fires rockets quickly",
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
	newNFT := []*passport.XsynMetadata{
		// large shield
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Large Shield",
			Description:        "A large shield",
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
			Description:        "A large shield",
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
			Description:        "A Medium shield",
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
			Description:        "A Medium shield",
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
			Description:        "A small shield",
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
			Description:        "A small shield",
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
			Description:        "A Medium shield",
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
			Description:        "A Medium shield",
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

func (s *Seeder) SeedWarMachineAbility(ctx context.Context, collection *passport.Collection) ([]*passport.XsynMetadata, error) {
	newNFT := []*passport.XsynMetadata{
		{
			CollectionID:       collection.ID,
			GameObject:         nil,
			Name:               "Overload",
			Description:        "Boost war machine's stat",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Name",
					Value:     "Overload",
				},
				{
					TraitType: "Asset Type",
					Value:     "Ability",
				},
				{
					TraitType: string(passport.AbilityAttFieldAbilityCost),
					Value:     "100000000000000000000",
				},
				{
					TraitType: string(passport.AbilityAttFieldAbilityID),
					Value:     11,
				},
				{
					TraitType: string(passport.AbilityAttFieldRequiredSlot),
					Value:     "Ability",
				},
				{
					TraitType:   string(passport.AbilityAttFieldRequiredPowerGrid),
					Value:       50,
					DisplayType: passport.Number,
				},
				{
					TraitType:   string(passport.AbilityAttFieldRequiredCPU),
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
