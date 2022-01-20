package seed

import (
	"context"
	"passport"
	"passport/db"

	"github.com/ninja-software/terror/v2"
)

func (s *Seeder) SeedNFTS(ctx context.Context) (warMachines, weapons, utility []*passport.XsynNftMetadata, error error) {
	warMachines, err := s.SeedWarMachine(ctx)
	if err != nil {
		return nil, nil, nil, terror.Error(err)
	}
	weapons, err = s.SeedWeapons(ctx)
	if err != nil {
		return nil, nil, nil, terror.Error(err)
	}
	utility, err = s.SeedUtility(ctx)
	if err != nil {
		return nil, nil, nil, terror.Error(err)
	}

	return
}

func (s *Seeder) SeedWarMachine(ctx context.Context) ([]*passport.XsynNftMetadata, error) {
	newNFT := []*passport.XsynNftMetadata{
		{
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Big War Machine",
			Description:        "A big ass War Machine - links to attached nfts",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			},
		},
		{
			Collection:  "SUPREMACY",
			Name:        "Medium War Machine",
			Description: "Average size War Machine - links to attached nfts",
			ExternalUrl: "",
			Image:       "",
			Attributes: []*passport.Attribute{
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
			},
		},
		{
			Collection:  "SUPREMACY",
			Name:        "Small War Machine",
			Description: "A iddy biddy tiny War Machine - links to attached nfts",
			ExternalUrl: "",
			Image:       "",
			Attributes: []*passport.Attribute{
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
			},
		},
	}

	for _, nft := range newNFT {
		err := db.XsynNftMetadataInsert(ctx, s.Conn, nft)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	return newNFT, nil
}

func (s *Seeder) SeedWeapons(ctx context.Context) ([]*passport.XsynNftMetadata, error) {
	newNFT := []*passport.XsynNftMetadata{
		// pulse rifles
		{
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Pulse Rifle",
			Description:        "A rifle that shoots pulses",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Pulse Rifle",
			Description:        "A rifle that shoots pulses",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Auto Cannon",
			Description:        "A cannon that shoots projectiles",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Auto Cannon",
			Description:        "A cannon that shoots projectiles",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Rocket Launcher",
			Description:        "A shoulder weapon that fires rockets",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Rocket Launcher",
			Description:        "A shoulder weapon that fires rockets",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Rapid Rocket Launcher",
			Description:        "A shoulder weapon that fires rockets quickly",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Rapid Rocket Launcher",
			Description:        "A shoulder weapon that fires rockets quickly",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
		err := db.XsynNftMetadataInsert(ctx, s.Conn, nft)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	return newNFT, nil
}

func (s *Seeder) SeedUtility(ctx context.Context) ([]*passport.XsynNftMetadata, error) {
	newNFT := []*passport.XsynNftMetadata{
		// large shield
		{
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Large Shield",
			Description:        "A large shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Large Shield",
			Description:        "A large shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        "A Medium shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        "A Medium shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Small Shield",
			Description:        "A small shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Small Shield",
			Description:        "A small shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        "A Medium shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
			Collection:         "SUPREMACY",
			GameObject:         nil,
			Name:               "Medium Shield",
			Description:        "A Medium shield",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
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
		err := db.XsynNftMetadataInsert(ctx, s.Conn, nft)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	return newNFT, nil
}
