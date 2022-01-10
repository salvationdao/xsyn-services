package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"passport"
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
		MaxMembersPerOrganisation,
		passport.UserRoleMemberID,
		nil,
		userSuperAdminID, userAdminID, userMemberID,
	)
	if err != nil {
		return terror.Error(err, "generate random users")
	}
	if len(randomUsers) == 0 {
		return terror.Error(terror.ErrWrongLength, "Random Users return wrong length")
	}
	userIndex := 0

	passwordHash := crypto.HashPassword("NinjaDojo_!")

	fmt.Println(" - set superadmin user")
	user := randomUsers[userIndex]
	user.Email = passport.NewString("superadmin@example.com")
	user.RoleID = passport.UserRoleSuperAdminID
	err = db.UserUpdate(ctx, s.Conn, user)
	if err != nil {
		return terror.Error(err)
	}
	err = db.AuthSetPasswordHash(ctx, s.Conn, user.ID, passwordHash)
	if err != nil {
		return terror.Error(err)
	}

	userIndex++

	fmt.Println(" - set admin user")
	user = randomUsers[userIndex]
	user.Email = passport.NewString("admin@example.com")
	user.RoleID = passport.UserRoleAdminID
	err = db.UserUpdate(ctx, s.Conn, user)
	if err != nil {
		return terror.Error(err)
	}
	err = db.AuthSetPasswordHash(ctx, s.Conn, user.ID, passwordHash)
	if err != nil {
		return terror.Error(err)
	}
	userIndex++

	fmt.Println(" - set member user")
	user = randomUsers[userIndex]
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
	userIndex++

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
