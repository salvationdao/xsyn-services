package seed

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"passport"
	"passport/db"
	"strings"
	"time"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog/log"
)

var ZaibatsuDefaultWeapons = []*passport.Attribute{
	{
		TraitType: "Name",
		Value:     "",
	},
	{
		TraitType:   "Speed",
		Value:       2500,
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
}

var RedMountainDefaultWeapons = []*passport.Attribute{
	{
		TraitType: "Name",
		Value:     "",
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
}

var BostonDefaultWeapons = []*passport.Attribute{
	{
		TraitType: "Name",
		Value:     "",
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
}

func (s *Seeder) SeedInitialStoreItems(ctx context.Context, passportURL string) error {
	supremacyCollection, err := db.CollectionGet(ctx, s.Conn, "Supremacy")
	if err != nil {
		return terror.Error(err)
	}

	initialStoreItems := []*passport.StoreItem{
		{
			Restriction: "LOOTBOX",
			Name:        "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Crystal Blue", UsdCentCost: 100000, AmountAvailable: 20, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Crystal Blue"}, {TraitType: "Rarity", Value: "Guardian"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Rust Bucket", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Rust Bucket"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Dune", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Dune"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Dynamic Yellow", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Dynamic Yellow"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Molten", UsdCentCost: 100000, AmountAvailable: 10, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Molten"}, {TraitType: "Rarity", Value: "Mythic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Mystermech", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Mystermech"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Nebula", UsdCentCost: 100000, AmountAvailable: 3, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Nebula"}, {TraitType: "Rarity", Value: "Deus ex"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Sleek", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Sleek"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Vintage", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Vintage"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "White Blue", UsdCentCost: 100000, AmountAvailable: 100, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "White Blue"}, {TraitType: "Rarity", Value: "Elite Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "BioHazard", UsdCentCost: 100000, AmountAvailable: 40, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "BioHazard"}, {TraitType: "Rarity", Value: "Exotic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Cyber", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Cyber"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Gold", UsdCentCost: 100000, AmountAvailable: 200, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Gold"}, {TraitType: "Rarity", Value: "Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.BostonCyberneticsFactionID, CollectionID: supremacyCollection.ID, Description: "Light Blue Police", UsdCentCost: 100000, AmountAvailable: 60, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Light Blue Police"}, {TraitType: "Rarity", Value: "Ultra Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Vintage", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Vintage"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Red White", UsdCentCost: 100000, AmountAvailable: 3, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Red White"}, {TraitType: "Rarity", Value: "Deus ex"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Red Hex", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Red Hex"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Gold", UsdCentCost: 100000, AmountAvailable: 200, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Gold"}, {TraitType: "Rarity", Value: "Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Desert", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Desert"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Navy", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Navy"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Nautical", UsdCentCost: 100000, AmountAvailable: 60, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Nautical"}, {TraitType: "Rarity", Value: "Ultra Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Military", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Military"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Irradiated", UsdCentCost: 100000, AmountAvailable: 10, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Irradiated"}, {TraitType: "Rarity", Value: "Mythic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Eva", UsdCentCost: 100000, AmountAvailable: 40, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Eva"}, {TraitType: "Rarity", Value: "Exotic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Beetle", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Beetle"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Villain", UsdCentCost: 100000, AmountAvailable: 100, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Villain"}, {TraitType: "Rarity", Value: "Elite Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Green Yellow", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Green Yellow"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.RedMountainFactionID, CollectionID: supremacyCollection.ID, Description: "Red Blue", UsdCentCost: 100000, AmountAvailable: 20, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Red Blue"}, {TraitType: "Rarity", Value: "Guardian"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "White Gold", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "White Gold"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Vector", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Vector"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Cherry Blossom", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Cherry Blossom"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Warden", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Warden"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Gundam", UsdCentCost: 100000, AmountAvailable: 60, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Gundam"}, {TraitType: "Rarity", Value: "Ultra Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Gold Pattern", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Gold Pattern"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Evangelion", UsdCentCost: 100000, AmountAvailable: 20, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Evangelion"}, {TraitType: "Rarity", Value: "Guardian"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Chalky Neon", UsdCentCost: 100000, AmountAvailable: 3, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Chalky Neon"}, {TraitType: "Rarity", Value: "Deus ex"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Black Digi", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Black Digi"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Purple Haze", UsdCentCost: 100000, AmountAvailable: 10, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Purple Haze"}, {TraitType: "Rarity", Value: "Mythic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Destroyer", UsdCentCost: 100000, AmountAvailable: 40, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Destroyer"}, {TraitType: "Rarity", Value: "Exotic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Static", UsdCentCost: 100000, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Static"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "White Neon", UsdCentCost: 100000, AmountAvailable: 100, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "White Neon"}, {TraitType: "Rarity", Value: "Elite Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
		{Name: "", FactionID: passport.ZaibatsuFactionID, CollectionID: supremacyCollection.ID, Description: "Gold", UsdCentCost: 100000, AmountAvailable: 200, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*passport.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Gold"}, {TraitType: "Rarity", Value: "Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: passport.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: passport.Number}}},
	}

	for _, si := range initialStoreItems {
		brand := strings.ReplaceAll(fmt.Sprintf("%v", si.Attributes[0].Value), " ", "-")
		model := strings.ReplaceAll(fmt.Sprintf("%v", si.Attributes[1].Value), " ", "-")
		subModel := strings.ReplaceAll(fmt.Sprintf("%v", si.Attributes[2].Value), " ", "-")
		si.AnimationURL = ""
		image, err := s.storeImages(ctx, fmt.Sprintf("%s_%s_%s", brand, model, subModel))
		if err != nil {
			log.Warn().Str("Image", fmt.Sprintf("%s_%s_%s", brand, model, subModel)).Err(err).Msg("Couldn't find image")
			return terror.Error(err)
		} else {
			si.Image = fmt.Sprintf("%s/api/files/%s", passportURL, image.ID.String())
		}
		webM, err := s.storeWebM(ctx, fmt.Sprintf("%s_%s_%s", brand, model, subModel))
		if err != nil {
			log.Warn().Str("WebM", fmt.Sprintf("%s_%s_%s", brand, model, subModel)).Err(err).Msg("Couldn't find webm")
			return terror.Error(err)
		} else {
			si.AnimationURL = fmt.Sprintf("%s/api/files/%s", passportURL, webM.ID.String())
		}
		switch si.FactionID {
		case passport.RedMountainFactionID:
			si.Attributes = append(si.Attributes, RedMountainDefaultWeapons...)
		case passport.ZaibatsuFactionID:
			si.Attributes = append(si.Attributes, ZaibatsuDefaultWeapons...)
		case passport.BostonCyberneticsFactionID:
			si.Attributes = append(si.Attributes, BostonDefaultWeapons...)
		}

		err = db.AddItemToStore(ctx, s.Conn, si)
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

func (s *Seeder) storeImages(ctx context.Context, filename string) (*passport.Blob, error) {
	// get read file from asset
	storeImage, err := os.ReadFile(fmt.Sprintf("./asset/store/images/%s.png", strings.ToLower(filename)))
	if err != nil {
		return nil, terror.Error(err)
	}

	// Get hash
	hasher := md5.New()
	_, err = hasher.Write(storeImage)
	if err != nil {
		return nil, terror.Error(err, "hash error")
	}
	hashResult := hasher.Sum(nil)
	hash := hex.EncodeToString(hashResult)

	blob := &passport.Blob{
		FileName:      filename,
		MimeType:      "image/png",
		Extension:     "png",
		FileSizeBytes: int64(len(storeImage)),
		File:          storeImage,
		Hash:          &hash,
		Public:        true,
	}

	// insert blob
	err = db.BlobInsert(ctx, s.Conn, blob)
	if err != nil {
		return nil, terror.Error(err)
	}

	return blob, nil
}

func (s *Seeder) storeWebM(ctx context.Context, filename string) (*passport.Blob, error) {
	// get read file from asset
	storeWebM, err := os.ReadFile(fmt.Sprintf("./asset/store/webm/%s.webm", strings.ToLower(filename)))
	if err != nil {
		return nil, terror.Error(err)
	}

	// Get hash
	hasher := md5.New()
	_, err = hasher.Write(storeWebM)
	if err != nil {
		return nil, terror.Error(err, "hash error")
	}
	hashResult := hasher.Sum(nil)
	hash := hex.EncodeToString(hashResult)

	blob := &passport.Blob{
		FileName:      filename,
		MimeType:      "video/webm",
		Extension:     "webm",
		FileSizeBytes: int64(len(storeWebM)),
		File:          storeWebM,
		Hash:          &hash,
		Public:        true,
	}

	// insert blob
	err = db.BlobInsert(ctx, s.Conn, blob)
	if err != nil {
		return nil, terror.Error(err)
	}

	return blob, nil
}
