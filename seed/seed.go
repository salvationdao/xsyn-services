package seed

import (
	"context"
	"database/sql"
	"fmt"
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
	},
	{
		ID:    passport.FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2"))),
		Label: "Boston Cybernetics",
		Theme: &passport.FactionTheme{
			Primary:    "#428EC1",
			Secondary:  "#FFFFFF",
			Background: "#050A12",
		},
	},
	{
		ID:    passport.FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d"))),
		Label: "Zaibatsu Heavy Industries",
		Theme: &passport.FactionTheme{
			Primary:    "#FFFFFF",
			Secondary:  "#FFFFFF",
			Background: "#0D0D0D",
		},
	},
}

func (s *Seeder) factions(ctx context.Context) error {
	for _, faction := range Factions {
		err := db.FactionCreate(ctx, s.Conn, faction)
		if err != nil {
			return terror.Error(err)
		}
	}
	return nil
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
