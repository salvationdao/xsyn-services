package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/email"
	"passport/passdb"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/auth"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Tokens struct {
	Conn                *pgxpool.Pool
	Mailer              *email.Mailer
	encryptToken        bool
	tokenExpirationDays int
	encryptTokenKey     string
}

// Save takes a jwt token, pulls out the token uuid and user uuid and saves it the issue_token table
func (t Tokens) Save(tokenEncoded string) error {
	tokenStr, err := base64.StdEncoding.DecodeString(tokenEncoded)
	if err != nil {
		return terror.Error(err)
	}

	token, err := auth.ReadJWT(tokenStr, t.EncryptToken(), t.EncryptTokenKey())
	if err != nil {
		return terror.Error(err)
	}
	userID, ok := token.Get("user-id")
	if !ok {
		return terror.Error(fmt.Errorf("unable to get userid from token"), "Unable to get userid from token")
	}

	userUUID, err := uuid.FromString(userID.(string))
	if err != nil {
		return terror.Error(err, "Unable to form UUID from token")
	}

	tokenID, ok := token.Get(openid.JwtIDKey)
	if !ok {
		return terror.Error(fmt.Errorf("unable to get tokenid from token"), "Unable to get tokenid from token")
	}

	tokenUUID, err := uuid.FromString(tokenID.(string))
	if err != nil {
		return terror.Error(err, "Unable to form UUID from token")
	}

	device := ""
	if userAgent, ok := token.Get("device"); ok {
		ua, ok := userAgent.(string)
		if ok {
			device = ua
		}
	}

	it := boiler.IssueToken{
		ID:        tokenUUID.String(),
		UserID:    userUUID.String(),
		UserAgent: device,
		ExpiresAt: null.TimeFrom(time.Now().AddDate(0, 0, 30)),
	}
	err = it.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert into issue token table")
	}
	return nil
}

func (t Tokens) Retrieve(uuid uuid.UUID) (auth.Token, auth.SecureUser, error) {
	ctx := context.Background()
	token, err := db.AuthFindToken(ctx, t.Conn, passport.IssueTokenID(uuid))
	if err != nil {
		return nil, nil, terror.Error(err, "Failed to find auth token")
	}

	user, err := db.UserGet(ctx, t.Conn, token.UserID)
	if err != nil {
		return nil, nil, terror.Error(err, "Failed to get user from database")
	}

	return token, &Secureuser{
		User:   user,
		Conn:   t.Conn,
		Mailer: t.Mailer,
	}, nil
}

func (t Tokens) Remove(uuid uuid.UUID) error {
	ctx := context.Background()
	err := db.AuthRemoveTokenWithID(ctx, t.Conn, passport.IssueTokenID(uuid))
	if err != nil {
		return terror.Error(err, "Failed to remove token with ID")
	}

	return nil
}

func (t Tokens) TokenExpirationDays() int {
	return t.tokenExpirationDays
}
func (t Tokens) EncryptToken() bool {
	return t.encryptToken
}
func (t Tokens) EncryptTokenKey() []byte {
	return []byte(t.encryptTokenKey)
}
