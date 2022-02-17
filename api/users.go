package api

import (
	"context"
	"errors"
	"fmt"
	"passport"
	"passport/crypto"
	"passport/db"
	"passport/email"
	"passport/helpers"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/auth"
	"github.com/rs/zerolog"
)

type UserGetter struct {
	Log    *zerolog.Logger
	Conn   *pgxpool.Pool
	Mailer *email.Mailer
}

func (ug *UserGetter) FacebookID(s string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByFacebookID(ctx, ug.Conn, s)
	if err != nil {
		return nil, terror.Error(err)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
	}

	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) GoogleID(s string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByGoogleID(ctx, ug.Conn, s)
	if err != nil {
		return nil, terror.Error(err)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
	}

	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) TwitchID(s string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByTwitchID(ctx, ug.Conn, s)
	if err != nil {
		return nil, terror.Error(err)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
	}

	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) TwitterID(s string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByTwitterID(ctx, ug.Conn, s)
	if err != nil {
		return nil, terror.Error(err)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
	}

	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) DiscordID(s string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByDiscordID(ctx, ug.Conn, s)
	if err != nil {
		return nil, terror.Error(err)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
	}

	return &Secureuser{
		User:   user,
		Conn:   ug.Conn,
		Mailer: ug.Mailer,
	}, nil
}

func (ug *UserGetter) UserCreator(firstName, lastName, username, email, facebookID, googleID, twitchID, twitterID, discordID, number, publicAddress, password string, other ...interface{}) (auth.SecureUser, error) {
	ctx := context.Background()
	throughOauth := true
	if facebookID == "" && googleID == "" && publicAddress == "" && twitchID == "" && twitterID == "" && discordID == "" {
		if email == "" {
			return nil, terror.Error(fmt.Errorf("email empty"), "Email cannot be empty")
		}

		throughOauth = false

		err := helpers.IsValidPassword(password)
		if err != nil {
			return nil, terror.Error(err)
		}

		emailAvailable, err := db.EmailAvailable(ctx, ug.Conn, email, nil)
		if err != nil {
			return nil, terror.Error(err, "Something went wrong. Please try again.")
		}
		if !emailAvailable {
			return nil, terror.Error(fmt.Errorf("user already exists"), "A user with that email already exists. Perhaps you'd like to login instead?")
		}
	}

	trimmedUsername := strings.TrimSpace(username)
	err := helpers.IsValidUsername(trimmedUsername)
	if err != nil {
		return nil, terror.Error(err)
	}

	usernameAvailable, err := db.UsernameAvailable(ctx, ug.Conn, trimmedUsername, nil)
	if err != nil {
		return nil, terror.Error(err, "Something went wrong. Please try again.")
	}
	if !usernameAvailable {
		return nil, terror.Error(fmt.Errorf("user already exists"), "A user with that username already exists. Perhaps you'd like to login instead?")
	}

	user := &passport.User{
		FirstName:     firstName,
		LastName:      lastName,
		Username:      trimmedUsername,
		FacebookID:    passport.NewString(facebookID),
		GoogleID:      passport.NewString(googleID),
		TwitchID:      passport.NewString(twitchID),
		TwitterID:     passport.NewString(twitterID),
		DiscordID:     passport.NewString(discordID),
		Email:         passport.NewString(email),
		PublicAddress: passport.NewString(publicAddress),
		RoleID:        passport.UserRoleMemberID,
		Verified:      throughOauth, // verify users directly if they go through Oauth
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

	err = db.UserCreate(ctx, ug.Conn, user)
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

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
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

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
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

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
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

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
	}

	return &Secureuser{
		User:         user,
		Conn:         ug.Conn,
		Mailer:       ug.Mailer,
		passwordHash: hash,
	}, nil
}

func (ug *UserGetter) Username(email string) (auth.SecureUser, error) {
	ctx := context.Background()
	user, err := db.UserByUsername(ctx, ug.Conn, email)
	if err != nil {
		return nil, terror.Error(err)
	}

	hash, err := db.HashByUserID(ctx, ug.Conn, user.ID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, terror.Error(err)
		}
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ug.Conn, *user.FactionID)
		if err != nil {
			return nil, terror.Error(err)
		}
		user.Faction = faction
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

	return err == nil
}

func (user Secureuser) SendVerificationEmail(token string, tokenID string, newAccount bool) error {

	err := user.Mailer.SendVerificationEmail(context.Background(), &email.User{
		IsAdmin:   user.IsAdmin(),
		Email:     user.Email.String,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, token, newAccount)
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func (user Secureuser) SendForgotPasswordEmail(token string, tokenID string) error {
	if !user.Email.Valid {
		return terror.Error(fmt.Errorf("user missing email"))
	}

	err := user.Mailer.SendForgotPasswordEmail(context.Background(), &email.User{
		IsAdmin:   user.IsAdmin(),
		Email:     user.Email.String,
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
	return userFields.Secureuser.Email.String
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
func (userFields UserFields) FactionID() *uuid.UUID {
	return (*uuid.UUID)(userFields.Secureuser.FactionID)
}
func (userFields UserFields) DeletedAt() *time.Time {
	return userFields.Secureuser.DeletedAt
}
func (userFields UserFields) Nonce() string {
	if userFields.Secureuser.Nonce.Valid {
		return userFields.Secureuser.Nonce.String
	}
	return ""
}
func (userFields UserFields) PublicAddress() string {
	if userFields.Secureuser.PublicAddress.Valid {
		return userFields.Secureuser.PublicAddress.String
	}
	return ""
}

type UserMetaMaskGetter struct {
	Log    *zerolog.Logger
	Conn   *pgxpool.Pool
	Mailer *email.Mailer
}
