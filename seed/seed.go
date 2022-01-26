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

	"github.com/gofrs/uuid"
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

	fmt.Println("Seeding roles")
	err := s.Roles(ctx)
	if err != nil {
		return terror.Error(err, "seed roles failed")
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

	fmt.Println("Seeding Supremacy User")
	_, err = s.SupremacyUser(ctx)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	//fmt.Println("Seeding organisations")
	//organisations, err := s.Organisations(ctx)
	//if err != nil {
	//	return terror.Error(err, "seed organisations failed")
	//}

	if !isProd {
		fmt.Println("Seeding xsyn NFTs")
		_, _, _, err := s.SeedNFTS(ctx)
		if err != nil {
			return terror.Error(err, "seed nfts failed")
		}

		fmt.Println("Seeding users")
		err = s.Users(ctx, nil)
		if err != nil {
			return terror.Error(err, "seed users failed")
		}

	}

	fmt.Println("Seeding factions")
	err = s.factions(ctx)
	if err != nil {
		return terror.Error(err, "seed factions")
	}

	//fmt.Println("Seeding products")
	//err = s.Products(ctx)
	//if err != nil {
	//	return terror.Error(err, "seed products failed")
	//}

	//s.Conn.
	fmt.Println("Seed complete")
	return nil
}

var Factions = []*passport.Faction{
	{
		ID:    passport.FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060"))),
		Label: "Red Mountain Offworld Mining Corporation",
		Theme: &passport.FactionTheme{
			Primary:    "#C24242",
			Secondary:  "#FFFFFF",
			Background: "#0D0404",
		},
		// TODO: remove from prod
		Description:   "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
		Velocity:      51,
		SharePercent:  39,
		RecruitNumber: 16554,
		WinCount:      1916,
		LossCount:     1337,
		KillCount:     4810,
		DeathCount:    3418,
		MVP:           "test user",
	},
	{
		ID:    passport.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))),
		Label: "Boston Cybernetics",
		Theme: &passport.FactionTheme{
			Primary:    "#428EC1",
			Secondary:  "#FFFFFF",
			Background: "#050A12",
		},
		// TODO: remove from prod
		Description:   "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
		Velocity:      51,
		SharePercent:  39,
		RecruitNumber: 16554,
		WinCount:      1916,
		LossCount:     1337,
		KillCount:     4810,
		DeathCount:    3418,
		MVP:           "test user",
	},
	{
		ID:    passport.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))),
		Label: "Zaibatsu Heavy Industries",
		Theme: &passport.FactionTheme{
			Primary:    "#FFFFFF",
			Secondary:  "#000000",
			Background: "#0D0D0D",
		},
		// TODO: remove from prod
		Description:   "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
		Velocity:      51,
		SharePercent:  39,
		RecruitNumber: 16554,
		WinCount:      1916,
		LossCount:     1337,
		KillCount:     4810,
		DeathCount:    3418,
		MVP:           "test user",
	},
}

func (s *Seeder) factions(ctx context.Context) error {
	for _, faction := range Factions {
		var err error
		logoBlob := &passport.Blob{}
		backgroundBlob := &passport.Blob{}

		switch faction.Label {
		case "Red Mountain Offworld Mining Corporation":
			logoBlob, err = s.factionLogo(ctx, "red_mountain_logo")
			if err != nil {
				return terror.Error(err)
			}
			backgroundBlob, err = s.factionBackground(ctx, "red_mountain_bg")
			if err != nil {
				return terror.Error(err)
			}
		case "Boston Cybernetics":
			logoBlob, err = s.factionLogo(ctx, "boston_cybernetics_logo")
			if err != nil {
				return terror.Error(err)
			}
			backgroundBlob, err = s.factionBackground(ctx, "boston_cybernetics_bg")
			if err != nil {
				return terror.Error(err)
			}
		case "Zaibatsu Heavy Industries":
			logoBlob, err = s.factionLogo(ctx, "zaibatsu_logo")
			if err != nil {
				return terror.Error(err)
			}
			backgroundBlob, err = s.factionBackground(ctx, "zaibatsu_bg")
			if err != nil {
				return terror.Error(err)
			}
		}

		faction.LogoBlobID = logoBlob.ID
		faction.BackgroundBlobID = backgroundBlob.ID

		err = db.FactionCreate(ctx, s.Conn, faction)
		if err != nil {
			return terror.Error(err)
		}
	}
	return nil
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

	// Game Treasury Account role
	gameTreasuryRole := &passport.Role{
		ID:          passport.UserRoleGameTreasury,
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
