package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"
	"xsyn-services/passport/crypto"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/shopspring/decimal"

	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"syreclabs.com/go/faker"
)

// Users for database spinup
func (s *Seeder) Users(ctx context.Context, organisations []*types.Organisation) error {
	// Seed Random Users (use constant ids for first 4 users)
	randomUsers, err := s.RandomUsers(
		ctx,
		10,
		types.UserRoleMemberID,
		nil)
	if err != nil {
		return terror.Error(err, "generate random users")
	}
	if len(randomUsers) == 0 {
		return terror.Error(terror.ErrWrongLength, "Random Users return wrong length")
	}

	passwordHash := crypto.HashPassword("NinjaDojo_!")

	user := randomUsers[0]
	user.Email = types.NewString("member@example.com")
	user.RoleID = types.UserRoleMemberID
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
	roleID types.RoleID,
	org *types.Organisation,
	// optional ids to use for the seeded users (will use until ran out if less than `amount`)
	ids ...types.UserID,
) ([]*types.User, error) {
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
	users := []*types.User{}
	for i, result := range result.Results {
		// Get avatar
		avatar, err := types.BlobFromURL(result.Picture.Large, uuid.Must(uuid.NewV4()).String()+".jpg")
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
		u := &types.User{
			FirstName: result.Name.First,
			LastName:  result.Name.Last,
			Email:     types.NewString(result.Email),
			Verified:  true,
			AvatarID:  &avatar.ID,
			RoleID:    roleID,
		}

		u.Username = fmt.Sprintf("%s%s", u.FirstName, u.LastName)

		if len(ids) > i {
			u.ID = ids[i]
		}

		// Insert
		err = db.UserCreateNoRPC(ctx, s.Conn, u)
		if err != nil {
			return nil, terror.Error(err)
		}

		passwordHash := crypto.HashPassword(faker.Internet().Password(8, 20))
		err = db.AuthSetPasswordHash(ctx, s.Conn, u.ID, passwordHash)
		if err != nil {
			return nil, terror.Error(err)
		}

		users = append(users, u)
	}

	return users, nil
}

func (s *Seeder) ETHChainUser(ctx context.Context) (*types.User, error) {
	// Create user
	u := &types.User{
		ID:       types.OnChainUserID,
		Username: types.OnChainUsername,
		RoleID:   types.UserRoleOffChain,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) XsynTreasuryUser(ctx context.Context) (*types.User, error) {
	// Create user
	u := &types.User{
		ID:       types.XsynTreasuryUserID,
		Username: types.XsynTreasuryUsername,
		RoleID:   types.UserRoleXsynTreasury,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	// add 300mil
	amount, ok := big.NewInt(0).SetString("300000000000000000000000000", 0)
	if !ok {
		return nil, terror.Error(fmt.Errorf("invalid string for big int"))

	}

	txObj := &types.NewTransaction{
		ID:                   fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond()),
		From:                 types.OnChainUserID,
		To:                   u.ID,
		Amount:               decimal.NewFromBigInt(amount, 0),
		Description:          "Initial supply seed.",
		TransactionReference: types.TransactionReference("Initial supply seed."),
	}

	q := `INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit)
    				VALUES((SELECT count(*) from transactions), $1, $2, $3, $4, $5);`

	_, err = s.Conn.Exec(ctx, q, txObj.Description, txObj.TransactionReference, txObj.Amount.String(), txObj.To, txObj.From)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SupremacyUser(ctx context.Context) (*types.User, error) {
	// Create user
	u := &types.User{
		ID:       types.SupremacyGameUserID,
		Username: types.SupremacyGameUsername,
		RoleID:   types.UserRoleGameAccount,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	// add 12mil
	amount, ok := big.NewInt(0).SetString("12000000000000000000000000", 0)
	if !ok {
		return nil, terror.Error(fmt.Errorf("invalid string for big int"))

	}
	id := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	txObj := &types.NewTransaction{
		ID:                   id,
		From:                 types.XsynTreasuryUserID,
		To:                   u.ID,
		Amount:               decimal.NewFromBigInt(amount, 0),
		Description:          "",
		TransactionReference: types.TransactionReference(id),
	}

	q := `INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit)
        				VALUES((SELECT count(*) from transactions), $1, $2, $3, $4, $5);`

	_, err = s.Conn.Exec(ctx, q, txObj.Description, txObj.TransactionReference, txObj.Amount.String(), txObj.To, txObj.From)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) XsynSaleUser(ctx context.Context) (*types.User, error) {
	// Create user
	u := &types.User{
		ID:       types.XsynSaleUserID,
		Username: types.XsynSaleUsername,
		RoleID:   types.UserRoleXsynSaleTreasury,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	amount, ok := big.NewInt(0).SetString("217000000000000000000000000", 0)
	if !ok {
		return nil, terror.Error(fmt.Errorf("invalid string for big int"))

	}

	// create xsynSaleUser balance of 217M from the xsynTreasuryUser
	id := fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond())
	txObj := &types.NewTransaction{
		ID:                   id,
		From:                 types.XsynTreasuryUserID,
		To:                   u.ID,
		Amount:               decimal.NewFromBigInt(amount, 0),
		Description:          "",
		TransactionReference: types.TransactionReference(id),
	}

	q := `INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit)
        				VALUES((SELECT count(*) from transactions), $1, $2, $3, $4, $5);`

	_, err = s.Conn.Exec(ctx, q, txObj.Description, txObj.TransactionReference, txObj.Amount.String(), txObj.To, txObj.From)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SupremacyBattleUser(ctx context.Context) (*types.User, error) {
	// Create user
	u := &types.User{
		ID:       types.SupremacyBattleUserID,
		Username: types.SupremacyBattleUsername,
		RoleID:   types.UserRoleGameAccount,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SupremacySupPoolUser(ctx context.Context) (*types.User, error) {
	// Create user
	u := &types.User{
		ID:       types.SupremacySupPoolUserID,
		Username: types.SupremacySupPoolUsername,
		RoleID:   types.UserRoleGameAccount,
		Verified: true,
	}

	// Insert
	err := db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SupremacyFactionUsers(ctx context.Context) (*types.User, error) {
	// add 1mil
	amount, ok := big.NewInt(0).SetString("1000000000000000000000000", 0)
	if !ok {
		return nil, terror.Error(fmt.Errorf("invalid string for big int"))

	}

	//supremacyCollection, err := db.CollectionGet(ctx, s.Conn, "supremacy")
	//if err != nil {
	//	return nil, terror.Error(err)
	//}

	supremacyCollectionID := types.CollectionID{}
	q := `select id from collections where name ILIKE $1`
	err := pgxscan.Get(ctx, s.Conn, &supremacyCollectionID, q, "supremacy genesis")
	if err != nil {
		return nil, terror.Error(err)
	}

	// Create user
	u := &types.User{
		ID:        types.SupremacyZaibatsuUserID,
		Username:  types.SupremacyZaibatsuUsername,
		RoleID:    types.UserRoleGameAccount,
		Verified:  true,
		FactionID: &types.ZaibatsuFactionID,
	}

	// Insert
	err = db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	txObj := &types.NewTransaction{
		ID:                   fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond()),
		From:                 types.XsynTreasuryUserID,
		To:                   u.ID,
		Amount:               decimal.NewFromBigInt(amount, 0),
		Description:          "Initial supremacy Zaibatsu supply seed.",
		TransactionReference: types.TransactionReference("Initial supremacy Zaibatsu supply seed."),
	}

	q = `INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit)
        				VALUES((SELECT count(*) from transactions), $1, $2, $3, $4, $5);`

	_, err = s.Conn.Exec(ctx, q, txObj.Description, txObj.TransactionReference, txObj.Amount.String(), txObj.To, txObj.From)
	if err != nil {
		return nil, terror.Error(err)
	}

	// Create user
	u = &types.User{
		ID:        types.SupremacyBostonCyberneticsUserID,
		Username:  types.SupremacyBostonCyberneticsUsername,
		RoleID:    types.UserRoleGameAccount,
		Verified:  true,
		FactionID: &types.BostonCyberneticsFactionID,
	}

	// Insert
	err = db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	txObj = &types.NewTransaction{
		ID:                   fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond()),
		From:                 types.XsynTreasuryUserID,
		To:                   u.ID,
		Amount:               decimal.NewFromBigInt(amount, 0),
		Description:          "Initial supremacy BostonCybernetics supply seed.",
		TransactionReference: types.TransactionReference("Initial supremacy BostonCybernetics supply seed."),
	}
	q = `INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit)
        				VALUES((SELECT count(*) from transactions), $1, $2, $3, $4, $5);`

	_, err = s.Conn.Exec(ctx, q, txObj.Description, txObj.TransactionReference, txObj.Amount.String(), txObj.To, txObj.From)
	if err != nil {
		return nil, terror.Error(err)
	}

	// Create user
	u = &types.User{
		ID:        types.SupremacyRedMountainUserID,
		Username:  types.SupremacyRedMountainUsername,
		RoleID:    types.UserRoleGameAccount,
		Verified:  true,
		FactionID: &types.RedMountainFactionID,
	}

	// Insert
	err = db.InsertSystemUser(ctx, s.Conn, u)
	if err != nil {
		return nil, terror.Error(err)
	}

	txObj = &types.NewTransaction{
		ID:                   fmt.Sprintf("%s|%d", uuid.Must(uuid.NewV4()), time.Now().Nanosecond()),
		From:                 types.XsynTreasuryUserID,
		Amount:               decimal.NewFromBigInt(amount, 0),
		To:                   u.ID,
		Description:          "Initial supremacy RedMountain supply seed.",
		TransactionReference: types.TransactionReference("Initial supremacy RedMountain supply seed."),
	}

	q = `INSERT INTO transactions(id, description, transaction_reference, amount, credit, debit)
        				VALUES((SELECT count(*) from transactions), $1, $2, $3, $4, $5);`

	_, err = s.Conn.Exec(ctx, q, txObj.Description, txObj.TransactionReference, txObj.Amount.String(), txObj.To, txObj.From)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = s.SeedAndAssignZaibatsu(ctx, supremacyCollectionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = s.SeedAndAssignRedMountain(ctx, supremacyCollectionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	err = s.SeedAndAssignBoston(ctx, supremacyCollectionID)
	if err != nil {
		return nil, terror.Error(err)
	}

	return u, nil
}

func (s *Seeder) SeedAndAssignZaibatsu(ctx context.Context, collectionID types.CollectionID) error {
	newNFT := []*types.XsynMetadata{
		{
			CollectionID:       collectionID,
			Name:               "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Zaibatsu",
				},
				{
					TraitType: "Model",
					Value:     "WREX",
				},
				{
					TraitType: "Skin",
					Value:     "Black",
				},
				{
					TraitType: "Name",
					Value:     "Alex",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: types.Number,
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
			CollectionID:       collectionID,
			Name:               "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Zaibatsu",
				},
				{
					TraitType: "Model",
					Value:     "WREX",
				},
				{
					TraitType: "Skin",
					Value:     "Black",
				},
				{
					TraitType: "Name",
					Value:     "John",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: types.Number,
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
			CollectionID:       collectionID,
			Name:               "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Zaibatsu",
				},
				{
					TraitType: "Model",
					Value:     "WREX",
				},
				{
					TraitType: "Skin",
					Value:     "Black",
				},
				{
					TraitType: "Name",
					Value:     "Mac",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Speed",
					Value:       2500,
					DisplayType: types.Number,
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
	}

	for _, nft := range newNFT {
		q := `	INSERT INTO xsyn_metadata (token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata)
    			VALUES((SELECT nextval('token_id_seq')),$1, $2, $3, $4, $5, $6, $7, $8)
    			RETURNING token_id as external_token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata`

		err := pgxscan.Get(ctx, s.Conn, nft, q, nft.Name, nft.CollectionID, nft.GameObject, nft.Description, nft.Image, nft.AnimationURL, nft.Attributes, nft.AdditionalMetadata)
		if err != nil {
			return terror.Error(err)
		}

		q = `INSERT INTO 
        			xsyn_assets (token_id, user_id)
        		VALUES
        			($1, $2);`

		_, err = s.Conn.Exec(ctx, q, nft.ExternalTokenID, types.SupremacyZaibatsuUserID)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}

func (s *Seeder) SeedAndAssignRedMountain(ctx context.Context, collectionID types.CollectionID) error {
	newNFT := []*types.XsynMetadata{
		{
			CollectionID:       collectionID,
			Name:               "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Red Mountain",
				},
				{
					TraitType: "Model",
					Value:     "BXSD",
				},
				{
					TraitType: "Skin",
					Value:     "Red_Steel",
				},
				{
					TraitType: "Name",
					Value:     "Vinnie",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
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
			CollectionID:       collectionID,
			Name:               "Olympus Mons LY07 B",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Red Mountain",
				},
				{
					TraitType: "Model",
					Value:     "BXSD",
				},
				{
					TraitType: "Skin",
					Value:     "Pink",
				},
				{
					TraitType: "Name",
					Value:     "Owen",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
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
			CollectionID:       collectionID,
			Name:               "Olympus Mons LY07 B",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Red Mountain",
				},
				{
					TraitType: "Model",
					Value:     "BXSD",
				},
				{
					TraitType: "Skin",
					Value:     "Pink",
				},
				{
					TraitType: "Name",
					Value:     "James",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1500,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
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
	}

	for _, nft := range newNFT {
		q := `	INSERT INTO xsyn_metadata (token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata)
    			VALUES((SELECT nextval('token_id_seq')),$1, $2, $3, $4, $5, $6, $7, $8)
    			RETURNING token_id as external_token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata`

		err := pgxscan.Get(ctx, s.Conn, nft, q, nft.Name, nft.CollectionID, nft.GameObject, nft.Description, nft.Image, nft.AnimationURL, nft.Attributes, nft.AdditionalMetadata)
		if err != nil {
			return terror.Error(err)
		}

		q = `INSERT INTO 
        			xsyn_assets (token_id, user_id)
        		VALUES
        			($1, $2);`

		_, err = s.Conn.Exec(ctx, q, nft.ExternalTokenID, types.SupremacyZaibatsuUserID)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}

func (s *Seeder) SeedAndAssignBoston(ctx context.Context, collectionID types.CollectionID) error {
	metadata := []*types.XsynMetadata{
		{
			CollectionID:       collectionID,
			Name:               "",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Boston Cybernetics",
				},
				{
					TraitType: "Model",
					Value:     "XFVS",
				},
				{
					TraitType: "Skin",
					Value:     "BlueWhite",
				},
				{
					TraitType: "Name",
					Value:     "Darren",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
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
			CollectionID:       collectionID,
			Name:               "Law Enforcer X-1000 B",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Boston Cybernetics",
				},
				{
					TraitType: "Model",
					Value:     "XFVS",
				},
				{
					TraitType: "Skin",
					Value:     "Police_DarkBlue",
				},
				{
					TraitType: "Name",
					Value:     "Yong",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
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
			CollectionID:       collectionID,
			Name:               "Law Enforcer X-1000 B",
			ExternalUrl:        "",
			Image:              "",
			AdditionalMetadata: nil,
			Attributes: []*types.Attribute{
				{
					TraitType: "Brand",
					Value:     "Boston Cybernetics",
				},
				{
					TraitType: "Model",
					Value:     "XFVS",
				},
				{
					TraitType: "Skin",
					Value:     "Police_DarkBlue",
				},
				{
					TraitType: "Name",
					Value:     "Corey",
				},
				//{
				//	TraitType: "SubModel",
				//	Value:     "Neon",
				//},
				{
					TraitType: "Asset Type",
					Value:     "War Machine",
				},
				{
					TraitType:   "Max Structure Hit Points",
					Value:       1000,
					DisplayType: types.Number,
				},
				{
					TraitType:   "Max Shield Hit Points",
					Value:       1000,
					DisplayType: types.Number,
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
	}

	for _, nft := range metadata {
		q := `	INSERT INTO xsyn_metadata (token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata)
    			VALUES((SELECT nextval('token_id_seq')),$1, $2, $3, $4, $5, $6, $7, $8)
    			RETURNING token_id as external_token_id, name, collection_id, game_object, description, image, animation_url,  attributes, additional_metadata`

		err := pgxscan.Get(ctx, s.Conn, nft, q, nft.Name, nft.CollectionID, nft.GameObject, nft.Description, nft.Image, nft.AnimationURL, nft.Attributes, nft.AdditionalMetadata)
		if err != nil {
			return terror.Error(err)
		}

		q = `INSERT INTO 
        			xsyn_assets (token_id, user_id)
        		VALUES
        			($1, $2);`

		_, err = s.Conn.Exec(ctx, q, nft.ExternalTokenID, types.SupremacyZaibatsuUserID)
		if err != nil {
			return terror.Error(err)
		}
	}

	return nil
}
