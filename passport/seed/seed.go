package seed

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"xsyn-services/passport/api"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/gosimple/slug"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"syreclabs.com/go/faker"
)

type Seeder struct {
	Conn             *pgxpool.Pool
	TxConn           *sql.DB
	TransactionCache *api.TransactionCache
	Log              *zerolog.Logger
	PassportHostUrl  string
	ExternalURL      string
}

// NewSeeder returns a new Seeder
func NewSeeder(conn *pgxpool.Pool, txConn *sql.DB, passportHostUrl string, externalUrl string) *Seeder {
	log := log_helpers.LoggerInitZero("seed", "DebugLevel")
	tc := api.NewTransactionCache(txConn, log)
	s := &Seeder{Conn: conn, TxConn: txConn, TransactionCache: tc, PassportHostUrl: passportHostUrl}
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
	_, err = s.factions(ctx)
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

	fmt.Println("Seeding Early Contributors")
	err = s.EarlyContributors(ctx)
	if err != nil {
		return terror.Error(err, "seed early contributors failed")
	}

	fmt.Println("Seeding initial store items")
	err = s.SeedInitialStoreItems(ctx, s.PassportHostUrl)
	if err != nil {
		return terror.Error(err, "seed users failed")
	}

	fmt.Println("Seed initial state")

	fmt.Println("Seed complete")
	return nil
}

func (s *Seeder) factions(ctx context.Context) ([]*types.Faction, error) {
	factions := []*types.Faction{}
	for _, faction := range types.Factions {
		var err error
		logoBlob := &types.Blob{}
		backgroundBlob := &types.Blob{}

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

		factions = append(factions, faction)
	}

	// build faction mvp material view
	err := db.FactionMvpMaterialisedViewCreate(ctx, s.Conn)
	if err != nil {
		return nil, terror.Error(err)
	}

	return factions, nil
}

func (s *Seeder) factionLogo(ctx context.Context, filename string) (*types.Blob, error) {
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

	blob := &types.Blob{
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

func (s *Seeder) factionBackground(ctx context.Context, filename string) (*types.Blob, error) {
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

	blob := &types.Blob{
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
	for _, perm := range types.AllPerm {
		allPerms = append(allPerms, string(perm))
	}
	// Off world/OnChain Account role
	offWorldRole := &types.Role{
		ID:          types.UserRoleOffChain,
		Name:        "Off World Role",
		Permissions: allPerms,
		Tier:        1,
	}
	err := db.RoleCreateReserved(ctx, s.Conn, offWorldRole)
	if err != nil {
		return terror.Error(err)
	}

	// Xsyn Treasury Account role
	xsynTreasuryRole := &types.Role{
		ID:          types.UserRoleXsynTreasury,
		Name:        "Xsyn Treasury",
		Permissions: allPerms,
		Tier:        1,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, xsynTreasuryRole)
	if err != nil {
		return terror.Error(err)
	}

	xsynSaleTreasuryRole := &types.Role{
		ID:          types.UserRoleXsynSaleTreasury,
		Name:        "Xsyn Sale Treasury",
		Permissions: allPerms,
		Tier:        1,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, xsynSaleTreasuryRole)
	if err != nil {
		return terror.Error(err)
	}

	// Game Treasury Account role
	gameTreasuryRole := &types.Role{
		ID:          types.UserRoleGameAccount,
		Name:        "Game Treasury",
		Permissions: allPerms,
		Tier:        1,
	}
	err = db.RoleCreateReserved(ctx, s.Conn, gameTreasuryRole)
	if err != nil {
		return terror.Error(err)
	}

	// Member
	memberRole := &types.Role{
		ID:          types.UserRoleMemberID,
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
func (s *Seeder) Organisations(ctx context.Context) ([]*types.Organisation, error) {
	organisations := []*types.Organisation{}

	org := &types.Organisation{
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

		org := &types.Organisation{
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
