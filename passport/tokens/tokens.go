package tokens

import (
	"encoding/base64"
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/auth"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// Save takes a jwt token, pulls out the token uuid and user uuid and saves it the issue_token table
func Save(tokenEncoded string, tokenExpirationDays int, encryptKey []byte) error {
	tokenStr, err := base64.StdEncoding.DecodeString(tokenEncoded)
	if err != nil {
		return err
	}

	token, err := auth.ReadJWT(tokenStr, true, encryptKey)
	if err != nil {
		return err
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
		ExpiresAt: null.TimeFrom(time.Now().AddDate(0, 0, tokenExpirationDays)),
	}
	err = it.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Failed to insert into issue token table")
	}
	return nil
}

func Retrieve(id uuid.UUID) (auth.Token, *boiler.User, error) {
	token, err := boiler.FindIssueToken(passdb.StdConn, id.String())
	if err != nil {
		return nil, nil, terror.Error(err, "Failed to find auth token")
	}

	user, err := boiler.FindUser(passdb.StdConn, token.UserID)
	if err != nil {
		return nil, nil, terror.Error(err, "Failed to get user from database")
	}

	tk := &types.IssueToken{
		ID:     types.IssueTokenID(uuid.FromStringOrNil(token.ID)),
		UserID: token.UserID,
	}

	return tk, user, nil
}

func Remove(uuid uuid.UUID) error {
	err := db.AuthRemoveTokenWithID(types.IssueTokenID(uuid))
	if err != nil {
		return terror.Error(err, "Failed to remove token with ID")
	}

	return nil
}
