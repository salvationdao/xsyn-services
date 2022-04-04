package seed

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/georgysavva/scany/pgxscan"

	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog/log"
)

var ZaibatsuDefaultWeapons = []*types.Attribute{
	{
		TraitType: "Name",
		Value:     "",
	},
	{
		TraitType:   "Speed",
		Value:       2500,
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
		Value:       1,
		DisplayType: types.Number,
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

var RedMountainDefaultWeapons = []*types.Attribute{
	{
		TraitType: "Name",
		Value:     "",
	},
	{
		TraitType:   "Speed",
		Value:       1750,
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
		Value:       1,
		DisplayType: types.Number,
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

var BostonDefaultWeapons = []*types.Attribute{
	{
		TraitType: "Name",
		Value:     "",
	},
	{
		TraitType:   "Speed",
		Value:       2750,
		DisplayType: types.Number,
	},
	{
		TraitType:   "Weapon Hardpoints",
		Value:       2,
		DisplayType: types.Number,
	},
	{
		TraitType:   "Utility Slots",
		Value:       1,
		DisplayType: types.Number,
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
	supremacyCollectionID := types.CollectionID{}
	q := `select id from collections where name ILIKE $1`
	err := pgxscan.Get(ctx, s.Conn, &supremacyCollectionID, q, "supremacy genesis")
	if err != nil {
		return terror.Error(err)
	}

	initialStoreItems := []*types.StoreItem{
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Crystal Blue", UsdCentCost: 100000, AmountAvailable: 20, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Crystal Blue"}, {TraitType: "Rarity", Value: "Guardian"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Rust Bucket", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Rust Bucket"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Dune", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Dune"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Dynamic Yellow", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Dynamic Yellow"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Molten", UsdCentCost: 100000, AmountAvailable: 10, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Molten"}, {TraitType: "Rarity", Value: "Mythic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Mystermech", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Mystermech"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Nebula", UsdCentCost: 100000, AmountAvailable: 3, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Nebula"}, {TraitType: "Rarity", Value: "Deus ex"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Sleek", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Sleek"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Vintage", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Vintage"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Blue White", UsdCentCost: 100000, AmountAvailable: 100, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Blue White"}, {TraitType: "Rarity", Value: "Elite Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "BioHazard", UsdCentCost: 100000, AmountAvailable: 40, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "BioHazard"}, {TraitType: "Rarity", Value: "Exotic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Cyber", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Cyber"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Gold", UsdCentCost: 100000, AmountAvailable: 200, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Gold"}, {TraitType: "Rarity", Value: "Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.BostonCyberneticsFactionID, CollectionID: supremacyCollectionID, Description: "Light Blue Police", UsdCentCost: 100000, AmountAvailable: 60, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Boston Cybernetics"}, {TraitType: "Model", Value: "Law Enforcer X-1000"}, {TraitType: "SubModel", Value: "Light Blue Police"}, {TraitType: "Rarity", Value: "Ultra Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Vintage", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Vintage"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Red White", UsdCentCost: 100000, AmountAvailable: 3, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Red White"}, {TraitType: "Rarity", Value: "Deus ex"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Red Hex", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Red Hex"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Gold", UsdCentCost: 100000, AmountAvailable: 200, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Gold"}, {TraitType: "Rarity", Value: "Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Desert", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Desert"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Navy", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Navy"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Nautical", UsdCentCost: 100000, AmountAvailable: 60, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Nautical"}, {TraitType: "Rarity", Value: "Ultra Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Military", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Military"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Irradiated", UsdCentCost: 100000, AmountAvailable: 10, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Irradiated"}, {TraitType: "Rarity", Value: "Mythic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Eva", UsdCentCost: 100000, AmountAvailable: 40, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Eva"}, {TraitType: "Rarity", Value: "Exotic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Beetle", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Beetle"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Villain", UsdCentCost: 100000, AmountAvailable: 100, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Villain"}, {TraitType: "Rarity", Value: "Elite Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Green Yellow", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Green Yellow"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.RedMountainFactionID, CollectionID: supremacyCollectionID, Description: "Red Blue", UsdCentCost: 100000, AmountAvailable: 20, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Red Mountain"}, {TraitType: "Model", Value: "Olympus Mons LY07"}, {TraitType: "SubModel", Value: "Red Blue"}, {TraitType: "Rarity", Value: "Guardian"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "White Gold", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "White Gold"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Vector", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Vector"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Cherry Blossom", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Cherry Blossom"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Warden", UsdCentCost: 100000, AmountAvailable: 300, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Warden"}, {TraitType: "Rarity", Value: "Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Gundam", UsdCentCost: 100000, AmountAvailable: 60, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Gundam"}, {TraitType: "Rarity", Value: "Ultra Rare"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "White Gold Pattern", UsdCentCost: 100000, AmountAvailable: 400, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "White Gold Pattern"}, {TraitType: "Rarity", Value: "Colossal"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Evangelion", UsdCentCost: 100000, AmountAvailable: 20, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Evangelion"}, {TraitType: "Rarity", Value: "Guardian"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Chalky Neon", UsdCentCost: 100000, AmountAvailable: 3, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Chalky Neon"}, {TraitType: "Rarity", Value: "Deus ex"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Black Digi", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Black Digi"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Purple Haze", UsdCentCost: 100000, AmountAvailable: 10, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Purple Haze"}, {TraitType: "Rarity", Value: "Mythic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Destroyer", UsdCentCost: 100000, AmountAvailable: 40, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Destroyer"}, {TraitType: "Rarity", Value: "Exotic"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Static", UsdCentCost: 100, AmountAvailable: 500, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Static"}, {TraitType: "Rarity", Value: "Mega"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Neon", UsdCentCost: 100000, AmountAvailable: 100, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Neon"}, {TraitType: "Rarity", Value: "Elite Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
		{Restriction: "LOOTBOX", Name: "", FactionID: types.ZaibatsuFactionID, CollectionID: supremacyCollectionID, Description: "Gold", UsdCentCost: 100000, AmountAvailable: 200, SoldBefore: time.Date(2022, 03, 01, 0, 0, 0, 0, time.UTC), Attributes: []*types.Attribute{{TraitType: "Brand", Value: "Zaibatsu"}, {TraitType: "Model", Value: "Tenshi Mk1"}, {TraitType: "SubModel", Value: "Gold"}, {TraitType: "Rarity", Value: "Legendary"}, {TraitType: "Asset Type", Value: "War Machine"}, {TraitType: "Max Structure Hit Points", Value: 1000, DisplayType: types.Number}, {TraitType: "Max Shield Hit Points", Value: 1000, DisplayType: types.Number}}},
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
		case types.RedMountainFactionID:
			si.Attributes = append(si.Attributes, RedMountainDefaultWeapons...)
		case types.ZaibatsuFactionID:
			si.Attributes = append(si.Attributes, ZaibatsuDefaultWeapons...)
		case types.BostonCyberneticsFactionID:
			si.Attributes = append(si.Attributes, BostonDefaultWeapons...)
		}

		q := `INSERT INTO xsyn_store (faction_id,
										  name,
										  collection_id,
										  description,
										  image,
										  animation_url,
										  attributes,
										  usd_cent_cost,
										  amount_sold,
										  amount_available,
										  sold_after,
										  sold_before)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

		_, err = s.Conn.Exec(ctx, q,
			si.FactionID,
			si.Name,
			si.CollectionID,
			si.Description,
			si.Image,
			si.AnimationURL,
			si.Attributes,
			si.UsdCentCost,
			si.AmountSold,
			si.AmountAvailable,
			si.SoldAfter,
			si.SoldAfter,
		)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}

func (s *Seeder) SeedCollections(ctx context.Context) ([]*types.Collection, error) {
	collections := []*types.Collection{{
		Name: "Supremacy",
	}, {
		Name: "Supremacy Genesis",
	}}
	for _, col := range collections {
		q := `INSERT INTO collections (name) VALUES($1)`
		_, err := s.Conn.Exec(ctx, q, col.Name)
		if err != nil {
			return nil, terror.Error(err)
		}
	}
	return collections, nil
}

func (s *Seeder) storeImages(ctx context.Context, filename string) (*types.Blob, error) {
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

	blob := &types.Blob{
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

func (s *Seeder) storeWebM(ctx context.Context, filename string) (*types.Blob, error) {
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

	blob := &types.Blob{
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
