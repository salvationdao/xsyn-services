package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"passport"
	"passport/api"
	"passport/crypto"
	"passport/db"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"syreclabs.com/go/faker"
)

// Users for database spinup
func (s *Seeder) Users(ctx context.Context, organisations []*passport.Organisation) error {
	// Seed Random Users (use constant ids for first 4 users)
	randomUsers, err := s.RandomUsers(
		ctx,
		10,
		passport.UserRoleMemberID,
		nil)
	if err != nil {
		return terror.Error(err, "generate random users")
	}
	if len(randomUsers) == 0 {
		return terror.Error(terror.ErrWrongLength, "Random Users return wrong length")
	}

	passwordHash := crypto.HashPassword("NinjaDojo_!")

	fmt.Println(" - set member user 1")
	user := randomUsers[0]
	user.Email = passport.NewString("member@example.com")
	user.RoleID = passport.UserRoleMemberID
	err = db.UserUpdate(ctx, s.Conn, user)
	if err != nil {
		return terror.Error(err)
	}
	err = db.AuthSetPasswordHash(ctx, s.Conn, user.ID, passwordHash)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

// RandomUsers grabs a list of seeded users
func (s *Seeder) RandomUsers(
	ctx context.Context,
	amount int,
	roleID passport.RoleID,
	org *passport.Organisation,
	// optional ids to use for the seeded users (will use until ran out if less than `amount`)
	ids ...passport.UserID,
) ([]*passport.User, error) {
	// Get random user data from randomuser.me (only au, us and gb nationalities to avoid non-alphanumeric names)
	r, err := http.Get(fmt.Sprintf("https://randomuser.me/api/?results=%d&inc=picture,name,email&nat=au,us,gb&noinfo", amount))
	if err != nil {
		return nil, terror.Error(err, "get random avatar")
	}

	var result struct {
		Results []struct {
			Name struct {
				First string
				Last  string
			}
			Email   string
			Picture struct {
				Medium    string
				Large     string
				Thumbnail string
			}
		}
	}

	// Decode json
	err = json.NewDecoder(r.Body).Decode(&result)
	if err != nil {
		return nil, terror.Error(err, "decode json result")
	}
	if len(result.Results) == 0 {
		return nil, terror.Error(fmt.Errorf("no results"))
	}

	// Loop over results
	users := []*passport.User{}
	for i, result := range result.Results {
		// Get avatar
		avatar, err := passport.BlobFromURL(result.Picture.Large, uuid.Must(uuid.NewV4()).String()+".jpg")
		if err != nil {
			return nil, terror.Error(err, "get image")
		}
		avatar.Public = true

		// Insert avatar
		err = db.BlobInsert(ctx, s.Conn, avatar)
		if err != nil {
			return nil, terror.Error(err)
		}

		// Create user
		u := &passport.User{
			FirstName: result.Name.First,
			LastName:  result.Name.Last,
			Email:     passport.NewString(result.Email),
			Verified:  true,
			AvatarID:  &avatar.ID,
			RoleID:    roleID,
		}

		u.Username = fmt.Sprintf("%s%s", u.FirstName, u.LastName)

		if len(ids) > i {
			u.ID = ids[i]
		}

		// Insert
		err = db.UserCreate(ctx, s.Conn, u)
		if err != nil {
			return nil, terror.Error(err)
		}

		passwordHash := crypto.HashPassword(faker.Internet().Password(8, 20))
		err = db.AuthSetPasswordHash(ctx, s.Conn, u.ID, passwordHash)
		if err != nil {
			return nil, terror.Error(err)
		}

		// Set Organisation
		//if org != nil {
		//	err = db.UserSetOrganisations(ctx, s.Conn, u.ID, org.ID)
		//	if err != nil {
		//		return nil, terror.Error(err)
		//	}
		//}

		users = append(users, u)
	}

	return users, nil
}

func (s *Seeder) ETHChainUser(ctx context.Context) (*passport.User, error) {
	// Create user
	u := &passport.User{
		ID:       passport.OnChainUserID,
		Username: passport.OnChainUsername,
		RoleID:   passport.UserRoleOffChain,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) XsynTreasuryUser(ctx context.Context) (*passport.User, error) {
	// Create user
	u := &passport.User{
		ID:       passport.XsynTreasuryUserID,
		Username: passport.XsynTreasuryUsername,
		RoleID:   passport.UserRoleXsynTreasury,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	// add 30mil
	amount, ok := big.NewInt(0).SetString("30000000000000000000000123", 0)
	if !ok {
		return nil, terror.Error(fmt.Errorf("invalid string for big int"))

	}

	// create treasury opening balance (30mil sups)
	_, err = api.CreateTransactionEntry(s.TxConn,
		*amount,
		u.ID,
		passport.OnChainUserID,
		"",
		"",
	)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SupremacyUser(ctx context.Context) (*passport.User, error) {
	// Create user
	u := &passport.User{
		ID:       passport.SupremacyGameUserID,
		Username: passport.SupremacyGameUsername,
		RoleID:   passport.UserRoleGameAccount,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	// add 12mil
	amount, ok := big.NewInt(0).SetString("1200000000000000432500000", 0)
	if !ok {
		return nil, terror.Error(fmt.Errorf("invalid string for big int"))

	}

	// create treasury opening balance (30mil sups)
	_, err = api.CreateTransactionEntry(s.TxConn,
		*amount,
		u.ID,
		passport.XsynTreasuryUserID,
		"",
		"",
	)
	if err != nil {
		return nil, terror.Error(err)
	}
	return u, nil
}

func (s *Seeder) SupremacyBattleUser(ctx context.Context) (*passport.User, error) {
	// Create user
	u := &passport.User{
		ID:       passport.SupremacyBattleUserID,
		Username: passport.SupremacyBattleUsername,
		RoleID:   passport.UserRoleGameAccount,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SupremacyFactionUsers(ctx context.Context) (*passport.User, error) {
	// Create user
	u := &passport.User{
		ID:        passport.SupremacyZaibatsuUserID,
		Username:  passport.SupremacyZaibatsuUsername,
		RoleID:    passport.UserRoleGameAccount,
		Verified:  true,
		FactionID: &passport.ZaibatsuFactionID,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	// Create user
	u = &passport.User{
		ID:        passport.SupremacyBostonCyberneticsUserID,
		Username:  passport.SupremacyBostonCyberneticsUsername,
		RoleID:    passport.UserRoleGameAccount,
		Verified:  true,
		FactionID: &passport.BostonCyberneticsFactionID,
	}

	// Insert
	err = db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	// Create user
	u = &passport.User{
		ID:        passport.SupremacyRedMountainUserID,
		Username:  passport.SupremacyRedMountainUsername,
		RoleID:    passport.UserRoleGameAccount,
		Verified:  true,
		FactionID: &passport.RedMountainFactionID,
	}

	// Insert
	err = db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = s.SeedAndAssignZaibatsu(ctx)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = s.SeedAndAssignRedMountain(ctx)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = s.SeedAndAssignBoston(ctx)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SeedAndAssignZaibatsu(ctx context.Context) error {
	newNFT := []*passport.XsynNftMetadata{
		{
			Collection:         "SUPREMACY",
			Name:               "Tenshi Mk1",
			Description:        "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: passport.Number,
				},
				//{
				//	TraitType:   "Power Grid",
				//	Value:       170,
				//	DisplayType: passport.Number,
				//},
				//{
				//	TraitType:   "CPU",
				//	Value:       100,
				//	DisplayType: passport.Number,
				//},
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
			},
		},
		{
			Collection:         "SUPREMACY",
			Name:               "Tenshi Mk1 B",
			Description:        "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: passport.Number,
				},
				//{
				//	TraitType:   "Power Grid",
				//	Value:       170,
				//	DisplayType: passport.Number,
				//},
				//{
				//	TraitType:   "CPU",
				//	Value:       100,
				//	DisplayType: passport.Number,
				//},
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
			},
		},
	}

	for _, nft := range newNFT {
		err := db.XsynNftMetadataInsert(ctx, s.Conn, nft)
		if err != nil {
			return terror.Error(err)
		}

		err = db.XsynNftMetadataAssignUser(ctx, s.Conn, nft.TokenID, passport.SupremacyZaibatsuUserID)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}

func (s *Seeder) SeedAndAssignRedMountain(ctx context.Context) error {
	newNFT := []*passport.XsynNftMetadata{
		{
			Collection:         "SUPREMACY",
			Name:               "Olympus Mons LY07",
			Description:        "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
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
			},
		},
		{
			Collection:         "SUPREMACY",
			Name:               "Olympus Mons LY07 B",
			Description:        "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
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
			},
		},
	}

	for _, nft := range newNFT {
		err := db.XsynNftMetadataInsert(ctx, s.Conn, nft)
		if err != nil {
			return terror.Error(err)
		}

		err = db.XsynNftMetadataAssignUser(ctx, s.Conn, nft.TokenID, passport.SupremacyRedMountainUserID)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}

func (s *Seeder) SeedAndAssignBoston(ctx context.Context) error {
	newNFT := []*passport.XsynNftMetadata{
		{
			Collection:         "SUPREMACY",
			Name:               "Law Enforcer X-1000",
			Description:        "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
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
			},
		},
		{
			Collection:         "SUPREMACY",
			Name:               "Law Enforcer X-1000 B",
			Description:        "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*passport.Attribute{
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: passport.Number,
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
			},
		},
	}

	for _, nft := range newNFT {
		err := db.XsynNftMetadataInsert(ctx, s.Conn, nft)
		if err != nil {
			return terror.Error(err)
		}

		err = db.XsynNftMetadataAssignUser(ctx, s.Conn, nft.TokenID, passport.SupremacyBostonCyberneticsUserID)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}
