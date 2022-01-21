package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"passport"
	"passport/db"
	"passport/email"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/hub/v3/ext/auth"
	"github.com/ninja-software/terror/v2"
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
	ctx := context.Background()
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
		return terror.Error(fmt.Errorf("unable to get userid from token"))
	}

	userUUID, err := uuid.FromString(userID.(string))
	if err != nil {
		return terror.Error(err, "unable to form UUID from token")
	}

	tokenID, ok := token.Get(openid.JwtIDKey)
	if !ok {
		return terror.Error(fmt.Errorf("unable to get tokenid from token"))
	}

	tokenUUID, err := uuid.FromString(tokenID.(string))
	if err != nil {
		return terror.Error(err, "unable to form UUID from token")
	}

	err = db.AuthSaveToken(ctx, t.Conn, passport.IssueTokenID(tokenUUID), passport.UserID(userUUID))
	if err != nil {
		return terror.Error(err)
	}
	return nil
}

func (t Tokens) Retrieve(uuid uuid.UUID) (auth.Token, auth.SecureUser, error) {
	ctx := context.Background()
	token, err := db.AuthFindToken(ctx, t.Conn, passport.IssueTokenID(uuid))
	if err != nil {
		return nil, nil, terror.Error(err)
	}

	user, err := db.UserGet(ctx, t.Conn, token.UserID)
	if err != nil {
		return nil, nil, terror.Error(err)
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
		return terror.Error(err)
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
