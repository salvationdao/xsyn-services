package seed

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"passport"
	"passport/db"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"syreclabs.com/go/faker"
)

type Seeder struct {
	Conn   *pgxpool.Pool
	TxConn *sql.DB
}

// NewSeeder returns a new Seeder
func NewSeeder(conn *pgxpool.Pool, txConn *sql.DB) *Seeder {
	s := &Seeder{Conn: conn, TxConn: txConn}
	return s
}

// Run for database spinup
func (s *Seeder) Run(isProd bool) error {
	ctx := context.Background()
	fmt.Println("Seeding collections")
	_, err := s.SeedCollections(ctx)
	if err != nil {
		return terror.Error(err, "seed collections failed")
	}

	fmt.Println("Seeding roles")
	err = s.Roles(ctx)
	if err != nil {
		return terror.Error(err, "seed roles failed")
	}

	fmt.Println("Seeding factions")
	factions, err = s.factions(ctx)
	if err != nil {
		return terror.Error(err, "seed factions")
	}

	fmt.Println("Seeding Off world / On chain User")
	_, err = s.ETHChainUser(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	fmt.Println("Seeding Treasury User")
	_, err = s.XsynTreasuryUser(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	fmt.Println("Seeding XSYN Sale Treasury User")
	_, err = s.XsynSaleUser(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	fmt.Println("Seeding Supremacy User")
	_, err = s.SupremacyUser(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	fmt.Println("Seeding Supremacy Battle User")
	_, err = s.SupremacyBattleUser(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	fmt.Println("Seeding Supremacy Sups Pool User")
	_, err = s.SupremacySupPoolUser(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	fmt.Println("Seeding Supremacy Faction User")
	_, err = s.SupremacyFactionUsers(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	//fmt.Println("Seeding organisations")
	//organisations, err := s.Organisations(ctx)
	//if err != nil {
	//	return terror.Error(err, "seed organisations failed")
	//}

	fmt.Println("Seeding initial store items")
	err = s.SeedInitialStoreItems(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	if !isProd {
		fmt.Println("Seeding xsyn item metadata")
		_, abilities, _, _, err := s.SeedItemMetadata(ctx)
		if err != nil {
			return terror.Error(err, "seed nfts failed")
		}
		fmt.Println("Seeding users")
		err = s.Users(ctx, nil)
		if err != nil {
			return terror.Error(err, "seed users failed")
		}
		// err = s.AndAssignNftToMember(ctx)
		// if err != nil {
		// 	return terror.Error(err, "seed users failed")
		// }

		// seed ability to zaibatsu war machines
		fmt.Println("Seeding assign ability to war machine")
		err = s.zaibatsuWarMachineAbilitySet(ctx, abilities)
		if err != nil {
			return terror.Error(err, "unable to seed zaibatsu abilities")
		}

	}
	fmt.Println("Seed initial state")

	q := `INSERT INTO state (latest_eth_block,
                             latest_bsc_block)
			VALUES(6359098, 16654769);`

	_, err = s.Conn.Exec(ctx, q)
	if err != nil {
		return terror.Error(err, "unable to seed state")
	}

	fmt.Println("Seed complete")
	return nil
}

func (s *Seeder) zaibatsuWarMachineAbilitySet(ctx context.Context, abilities []*passport.XsynMetadata) error {
	// get Zaibatsu war machines
	warMachines, err := db.WarMachineGetByUserID(ctx, s.Conn, passport.SupremacyZaibatsuUserID)
	if err != nil {
		return terror.Error(err)
	}

	for _, wm := range warMachines {
		abilityMetadata := &passport.AbilityMetadata{}
		passport.ParseAbilityMetadata(abilities[0], abilityMetadata)

		err := db.WarMachineAbilitySet(ctx, s.Conn, wm.TokenID, abilityMetadata.TokenID, passport.WarMachineAttFieldAbility01)
		if err != nil {
			return terror.Error(err)
		}

		err = db.WarMachineAbilityCostUpsert(ctx, s.Conn, wm.TokenID, abilityMetadata.TokenID, abilityMetadata.SupsCost)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}

var factions = []*passport.Faction{
	{
		ID:    passport.RedMountainFactionID,
		Label: "Red Mountain Offworld Mining Corporation",
		Theme: &passport.FactionTheme{
			Primary:    "#C24242",
			Secondary:  "#FFFFFF",
			Background: "#120E0E",
		},
		// NOTE: change content
		Description: "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
	},
	{
		ID:    passport.BostonCyberneticsFactionID,
		Label: "Boston Cybernetics",
		Theme: &passport.FactionTheme{
			Primary:    "#428EC1",
			Secondary:  "#FFFFFF",
			Background: "#080C12",
		},
		// NOTE: change content
		Description: "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
	},
	{
		ID:    passport.ZaibatsuFactionID,
		Label: "Zaibatsu Heavy Industries",
		Theme: &passport.FactionTheme{
			Primary:    "#FFFFFF",
			Secondary:  "#000000",
			Background: "#0D0D0D",
		},
		// NOTE: change content
		Description: "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
	},
}

func (s *Seeder) factions(ctx context.Context) ([]*passport.Faction, error) {
	for _, faction := range factions {
		var err error
		logoBlob := &passport.Blob{}
		backgroundBlob := &passport.Blob{}

		switch faction.Label {
		case "Red Mountain Offworld Mining Corporation":
			logoBlob, err = s.factionLogo(ctx, "red_mountain_logo")
			if err != nil {
				return nil, terror.Error(err)
			}
			backgroundBlob, err = s.factionBackground(ctx, "red_mountain_bg")
			if err != nil {
				return nil, terror.Error(err)
			}
		case "Boston Cybernetics":
			logoBlob, err = s.factionLogo(ctx, "boston_cybernetics_logo")
			if err != nil {
				return nil, terror.Error(err)
			}
			backgroundBlob, err = s.factionBackground(ctx, "boston_cybernetics_bg")
			if err != nil {
				return nil, terror.Error(err)
			}
		case "Zaibatsu Heavy Industries":
			logoBlob, err = s.factionLogo(ctx, "zaibatsu_logo")
			if err != nil {
				return nil, terror.Error(err)
			}
			backgroundBlob, err = s.factionBackground(ctx, "zaibatsu_bg")
			if err != nil {
				return nil, terror.Error(err)
			}
		}

		faction.LogoBlobID = logoBlob.ID
		faction.BackgroundBlobID = backgroundBlob.ID

		err = db.FactionCreate(ctx, s.Conn, faction)
		if err != nil {
			return nil, terror.Error(err)
		}
	}

	// build faction mvp material view
	err := db.FactionMvpMaterialisedViewCreate(ctx, s.Conn)
	if err != nil {
		return nil, terror.Error(err)
	}

	return factions, nil
}

func (s *Seeder) factionLogo(ctx context.Context, filename string) (*passport.Blob, error) {
	// get read file from asset
	factionLogo, err := os.ReadFile(fmt.Sprintf("./asset/%s.svg", filename))
	if err != nil {
		return nil, terror.Error(err)
	}

	// Get hash
	hasher := md5.New()
	_, err = hasher.Write(factionLogo)
	if err != nil {
		return nil, terror.Error(err, "hash error")
	}
	hashResult := hasher.Sum(nil)
	hash := hex.EncodeToString(hashResult)

	blob := &passport.Blob{
		FileName:      filename,
		MimeType:      "image/svg+xml",
		Extension:     "svg",
		FileSizeBytes: int64(len(factionLogo)),
		File:          factionLogo,
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

func (s *Seeder) factionBackground(ctx context.Context, filename string) (*passport.Blob, error) {
	// get read file from asset
	factionLogo, err := os.ReadFile(fmt.Sprintf("./asset/%s.webp", filename))
	if err != nil {
		return nil, terror.Error(err)
	}

	// Get hash
	hasher := md5.New()
	_, err = hasher.Write(factionLogo)
	if err != nil {
		return nil, terror.Error(err, "hash error")
	}
	hashResult := hasher.Sum(nil)
	hash := hex.EncodeToString(hashResult)

	blob := &passport.Blob{
		FileName:      filename,
		MimeType:      "image/webp",
		Extension:     "webp",
		FileSizeBytes: int64(len(factionLogo)),
		File:          factionLogo,
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

// Roles for database spin-up
func (s *Seeder) Roles(ctx context.Context) error {

	var allPerms []string
	for _, perm := range passport.AllPerm {
		allPerms = append(allPerms, string(perm))
	}
	// Off world/OnChain Account role
	offWorldRole := &passport.Role{
		ID:          passport.UserRoleOffChain,
		Name:        "Off World Role",
		Permissions: allPerms,
		Tier:        1,
	}
	err := db.RoleCreateReserved(ctx, s.Conn, offWorldRole)
	if err != nil {
		return terror.Error(err)
	}

	// Xsyn Treasury Account role
	xsynTreasuryRole := &passport.Role{
		ID:          passport.UserRoleXsynTreasury,
		Name:        "Xsyn Treasury",
		Permissions: allPerms,
		Tier:        1,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, xsynTreasuryRole)
	if err != nil {
		return terror.Error(err)
	}

	xsynSaleTreasuryRole := &passport.Role{
		ID:          passport.UserRoleXsynSaleTreasury,
		Name:        "Xsyn Sale Treasury",
		Permissions: allPerms,
		Tier:        1,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, xsynSaleTreasuryRole)
	if err != nil {
		return terror.Error(err)
	}

	// Game Treasury Account role
	gameTreasuryRole := &passport.Role{
		ID:          passport.UserRoleGameAccount,
		Name:        "Game Treasury",
		Permissions: allPerms,
		Tier:        1,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, gameTreasuryRole)
	if err != nil {
		return terror.Error(err)
	}

	// Member
	memberRole := &passport.Role{
		ID:          passport.UserRoleMemberID,
		Name:        "Member",
		Permissions: []string{
			// TODO: possible add perms?
			//string(passport.PermUserRead),
			//string(passport.PermOrganisationRead),
			//string(passport.PermProductList),
			//string(passport.PermProductRead),
		},
		Tier: 3,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, memberRole)
	if err != nil {
		return terror.Error(err)
	}

	return nil
}

// Organisations for database spinup
func (s *Seeder) Organisations(ctx context.Context) ([]*passport.Organisation, error) {
	organisations := []*passport.Organisation{}

	org := &passport.Organisation{
		Name: "Ninja Software",
		Slug: "ninja-software",
	}
	err := db.OrganisationCreate(ctx, s.Conn, org)
	if err != nil {
		return nil, terror.Error(err)
	}
	organisations = append(organisations, org)

	for i := 0; i < 5; i++ {
		orgName := faker.Company().Name()
		orgSlug := slug.Make(orgName)

		org := &passport.Organisation{
			Name: orgName,
			Slug: orgSlug,
		}
		err := db.OrganisationCreate(ctx, s.Conn, org)
		if err != nil {
			return nil, terror.Error(err)
		}
		organisations = append(organisations, org)
	}

	return organisations, nil
}
