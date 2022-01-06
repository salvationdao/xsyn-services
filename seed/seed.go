package seed

import (
	"context"
	"fmt"
	"passport"
	"passport/db"

	"github.com/gofrs/uuid"
	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"syreclabs.com/go/faker"
)

// ID constants to keep cookies constant between re-seeds (so you don't have to keep logging back in)
var (
	userSuperAdminID = passport.UserID(uuid.Must(uuid.FromString("88a825b9-dae2-40dd-9848-73db7870c9d5")))
	userAdminID      = passport.UserID(uuid.Must(uuid.FromString("639cb314-50c5-4a27-bd9f-f85bcd16fc3e")))
	userMemberID     = passport.UserID(uuid.Must(uuid.FromString("ce4363e1-f522-45a3-93a1-216974304e75")))
)

// MaxMembersPerOrganisation is the default amount of member users per organisation (also includes non-organisation users)
const MaxMembersPerOrganisation = 40

// MaxTestUsers is the default amount of user for the test organisation account (first organisation has X reserved test users)
const MaxTestUsers = 3

type Seeder struct {
	Conn *pgxpool.Pool
}

// NewSeeder returns a new Seeder
func NewSeeder(conn *pgxpool.Pool) *Seeder {
	s := &Seeder{conn}
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

	//fmt.Println("Seeding organisations")
	//organisations, err := s.Organisations(ctx)
	//if err != nil {
	//	return terror.Error(err, "seed organisations failed")
	//}

	fmt.Println("Seeding users")
	err = s.Users(ctx, nil)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	//fmt.Println("Seeding products")
	//err = s.Products(ctx)
	//if err != nil {
	//	return terror.Error(err, "seed products failed")
	//}

	fmt.Println("Seed complete")
	return nil
}

// Roles for database spinup
func (s *Seeder) Roles(ctx context.Context) error {
	// Super Admin
	allPerms := []string{}
	for _, perm := range passport.AllPerm {
		allPerms = append(allPerms, string(perm))
	}
	superAdminRole := &passport.Role{
		ID:          passport.UserRoleSuperAdminID,
		Name:        "Super Admin",
		Permissions: allPerms,
		Tier:        1,
	}
	err := db.RoleCreateReserved(ctx, s.Conn, superAdminRole)
	if err != nil {
		return terror.Error(err)
	}

	// Admin
	adminRole := &passport.Role{
		ID:   passport.UserRoleAdminID,
		Name: "Admin",
		Permissions: []string{
			// Users
			string(passport.PermUserList),
			string(passport.PermUserCreate),
			string(passport.PermUserRead),
			string(passport.PermUserUpdate),
			string(passport.PermUserArchive),
			string(passport.PermUserUnarchive),
			// Organisations
			string(passport.PermOrganisationList),
			string(passport.PermOrganisationCreate),
			string(passport.PermOrganisationRead),
			string(passport.PermOrganisationUpdate),
			string(passport.PermOrganisationArchive),
			string(passport.PermOrganisationUnarchive),
			// Products
			string(passport.PermProductList),
			string(passport.PermProductCreate),
			string(passport.PermProductRead),
			string(passport.PermProductUpdate),
			string(passport.PermProductArchive),
			string(passport.PermProductUnarchive),
			// Other
			string(passport.PermAdminPortal),
			string(passport.PermUserActivityList),
		},
		Tier: 2,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, adminRole)
	if err != nil {
		return terror.Error(err)
	}

	// Member
	memberRole := &passport.Role{
		ID:   passport.UserRoleMemberID,
		Name: "Member",
		Permissions: []string{
			string(passport.PermUserRead),
			string(passport.PermOrganisationRead),
			string(passport.PermProductList),
			string(passport.PermProductRead),
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
