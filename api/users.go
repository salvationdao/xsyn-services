package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/auth"
	"github.com/ninja-software/terror/v2"
	"github.com/rs/zerolog"
	"passport"
	"passport/crypto"
	"passport/db"
	"passport/email"
	"passport/helpers"
	"time"
)

type UserGetter struct {
	Log    *zerolog.Logger
	Conn   *pgxpool.Pool
	Mailer *email.Mailer
}

func (ug *UserGetter) UserCreator(firstName, lastName, username, email, number, publicAddress, password string, other ...interface{}) (auth.SecureUser, error) {
	ctx := context.Background()
	if username == "" {
		return nil, terror.Error(fmt.Errorf("username cannot be empty"), "Username cannot be empty.")
	}

	user := &passport.User{
		FirstName:     firstName,
		LastName:      lastName,
		Username:      username,
		Email:         email,
		PublicAddress: &publicAddress,
		RoleID:        passport.UserRoleMemberID,
	}

	if password != "" && email != "" {
		passwordHash := crypto.HashPassword(password)
		err := db.AuthRegister(ctx, ug.Conn, user, passwordHash)
		if err != nil {
			return nil, terror.Error(err)
		}

		return &Secureuser{
			User:   user,
			Conn:   ug.Conn,
			Mailer: ug.Mailer,
		}, nil
	}

	err := db.UserCreate(ctx, ug.Conn, user)
	if err != nil {
		return nil, terror.Error(err)
	}

	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) PublicAddress(s string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByPublicAddress(ctx, ug.Conn, s)
	if err != nil {
		return nil, terror.Error(err)
	}
	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) ID(id uuid.UUID) (auth.SecureUser, error) {
	ctx := context.Background()
	userUUID := passport.UserID(id)
	user, err := db.UserGet(ctx, ug.Conn, userUUID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) Token(id uuid.UUID) (auth.SecureUser, error) {
	ctx := context.Background()
	tokenUUID := passport.IssueTokenID(id)
	result, err := db.AuthFindToken(ctx, ug.Conn, tokenUUID)
	if err != nil {
		return nil, terror.Error(err)
	}

	user, err := db.UserGet(ctx, ug.Conn, result.UserID)
	if err != nil {
		return nil, terror.Error(err)
	}
	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) Email(email string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByEmail(ctx, ug.Conn, email)
	if err != nil {
		return nil, terror.Error(err)
	}
	hash, err := db.HashByUserID(ctx, ug.Conn, user.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, terror.Error(err)
		}
	}
	return &Secureuser{
		User:         user,
		Conn:         ug.Conn,
		Mailer:       ug.Mailer,
		passwordHash: hash,
	}, nil
}

type Secureuser struct {
	*passport.User
	passwordHash string
	Mailer       *email.Mailer
	Conn         *pgxpool.Pool
}

func (user *Secureuser) NewNonce() (string, error) {
	ctx := context.Background()
	newNonce := helpers.RandStringBytes(16)
	err := db.UserUpdateNonce(ctx, user.Conn, user.ID, newNonce)
	if err != nil {
		return "", terror.Error(err)
	}

	return newNonce, nil
}

func (user *Secureuser) SetHash(hash string) {
	user.passwordHash = hash
}

func (user Secureuser) CheckPassword(pw string) bool {
	if user.passwordHash == "" {
		return false
	}
	err := crypto.ComparePassword(user.passwordHash, pw)
	if err != nil {
		return false
	}
	return true
}

func (user Secureuser) SendVerificationEmail(token string, tokenID string, newAccount bool) error {
	err := user.Mailer.SendVerificationEmail(&email.User{
		IsAdmin:   user.IsAdmin(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, token, newAccount)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func (user Secureuser) SendForgotPasswordEmail(token string, tokenID string) error {
	err := user.Mailer.SendForgotPasswordEmail(&email.User{
		IsAdmin:   user.IsAdmin(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, token)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func (user Secureuser) Verify() error {
	ctx := context.Background()
	err := db.UserVerify(ctx, user.Conn, user.ID)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}
func (user Secureuser) UpdatePasswordSetting(oldPasswordRequired bool) error {
	return nil
}

func (user Secureuser) HasPermission(perm string) bool {
	for _, r := range user.Role.Permissions {
		if r == perm {
			return true
		}
	}
	return false
}

func (user Secureuser) UpdateAvatar(url string, fileName string) error {
	// TODO: update avatar
	return nil
}

func (user Secureuser) Fields() hub.UserFields {
	return UserFields{
		Secureuser: user,
	}
}

type UserFields struct {
	Secureuser Secureuser
}

func (userFields UserFields) ID() uuid.UUID {
	return uuid.UUID(userFields.Secureuser.ID)
}
func (userFields UserFields) Email() string {
	return userFields.Secureuser.Email
}
func (userFields UserFields) FirstName() string {
	return userFields.Secureuser.FirstName
}
func (userFields UserFields) LastName() string {
	return userFields.Secureuser.LastName
}
func (userFields UserFields) Verified() bool {
	return userFields.Secureuser.Verified
}
func (userFields UserFields) Deleted() bool {
	return userFields.Secureuser.DeletedAt != nil || !userFields.Secureuser.DeletedAt.IsZero()
}
func (userFields UserFields) AvatarID() *uuid.UUID {
	return (*uuid.UUID)(userFields.Secureuser.AvatarID)
}
func (userFields UserFields) DeletedAt() *time.Time {
	return userFields.Secureuser.DeletedAt
}
func (userFields UserFields) Nonce() string {
	if userFields.Secureuser.Nonce != nil {
		return *userFields.Secureuser.Nonce
	}
	return ""
}
func (userFields UserFields) PublicAddress() string {
	if userFields.Secureuser.PublicAddress != nil {
		return *userFields.Secureuser.PublicAddress
	}
	return ""
}

type UserMetaMaskGetter struct {
	Log    *zerolog.Logger
	Conn   *pgxpool.Pool
	Mailer *email.Mailer
}
