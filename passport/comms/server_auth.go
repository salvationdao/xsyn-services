package comms

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/benchmark"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/tokens"
	"xsyn-services/types"

	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/jwt/openid"
)

// IsServerClient checks the given api key is a server client
func IsServerClient(apikey string) (string, error) {
	if apikey == "" {
		passlog.L.Err(fmt.Errorf("missing api key")).Msg("api key empty")
		return "", terror.Error(fmt.Errorf("missing api key"))
	}

	apiKeyEntry, err := boiler.FindAPIKey(passdb.StdConn, apikey)
	if err != nil {
		passlog.L.Err(err).Str("api_key", apikey).Msg("error finding api key")
		return "", err
	}

	if apiKeyEntry.Type != "SERVER_CLIENT" {
		passlog.L.Err(err).Str("api_key", apikey).Str("key_type", apiKeyEntry.Type).Str("required_type", "SERVER_CLIENT").Msg("does not have permission")
		return "", terror.Error(fmt.Errorf("api key is missing SERVER_CLIENT permission"))
	}

	return apiKeyEntry.UserID, nil
}

// IsSupremacyClient checks if the given api key belongs to the supremacy user
func IsSupremacyClient(apikey string) (string, error) {
	bm := benchmark.New()
	if apikey == "" {
		passlog.L.Err(fmt.Errorf("missing api key")).Msg("api key empty")
		return "", terror.Error(fmt.Errorf("missing api key"))
	}

	bm.Start("find_api_key")
	apiKeyEntry, err := boiler.FindAPIKey(passdb.StdConn, apikey)
	bm.End("find_api_key")
	if err != nil {
		passlog.L.Err(err).Str("api_key", apikey).Msg("error finding api key")
		return "", err
	}

	if apiKeyEntry.Type != "SERVER_CLIENT" {
		passlog.L.Err(err).Str("api_key", apikey).Str("key_type", apiKeyEntry.Type).Str("required_type", "SERVER_CLIENT").Msg("does not have permission")
		return "", terror.Error(fmt.Errorf("api key is missing SERVER_CLIENT permission"))
	}

	bm.Start("find_user")
	user, err := boiler.FindUser(passdb.StdConn, apiKeyEntry.UserID)
	bm.End("find_user")
	if err != nil {
		passlog.L.Err(err).Str("api_key", apikey).Str("user_id", apiKeyEntry.UserID).Msg("error finding user from api key")
		return "", err
	}

	if user.Username != types.SupremacyGameUsername {
		passlog.L.Err(err).Str("api_key", apikey).Str("key_username", user.Username).Str("expect_username", types.SupremacyGameUsername).Msg("username mismatch")
		return "", terror.Error(fmt.Errorf("api key owner username mismatch"))
	}

	bm.Alert(50)
	return user.ID, nil
}

type TokenReq struct {
	ApiKey      string
	TokenBase64 string
	Device      string
	Action      string
}

type TokenResp struct {
	*UserResp
	Token     string
	ExpiredAt time.Time
}

func (s *S) OneTimeTokenLogin(req TokenReq, resp *TokenResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	tokenStr, err := base64.StdEncoding.DecodeString(req.TokenBase64)
	if err != nil {
		fmt.Println("error", err, tokenStr)
		return terror.Error(err, "token is fail")
	}

	token, err := tokens.ReadJWT(tokenStr, true, s.TokenEncryptionKey)
	if err != nil {
		if errors.Is(err, tokens.ErrTokenExpired) {
			tknUuid, err := tokens.TokenID(token)
			if err != nil {
				return err
			}
			err = tokens.Remove(tknUuid)
			if err != nil {
				return err
			}
			return terror.Warn(err, "Session has expired, please log in again.")
		}
		return err
	}

	jwtIDI, ok := token.Get(openid.JwtIDKey)

	if !ok {
		return terror.Error(errors.New("unable to get ID from token"), "unable to read token")
	}

	jwtID, err := uuid.FromString(jwtIDI.(string))
	if err != nil {
		return terror.Error(err, "unable to form UUID from token")
	}

	retrievedToken, user, err := tokens.Retrieve(jwtID)
	if err != nil {
		return err
	}

	if !retrievedToken.Whitelisted() {
		return tokens.ErrTokenNotWhitelisted
	}

	resp.UserResp = &UserResp{}
	resp.FactionID = user.FactionID
	resp.ID = user.ID
	resp.PublicAddress = user.PublicAddress
	resp.Username = user.Username

	tokenID := uuid.Must(uuid.NewV4())

	// save user detail as jwt
	jwt, sign, err := tokens.GenerateJWT(
		tokenID,
		user,
		req.Device,
		req.Action,
		s.TokenExpirationDays)
	if err != nil {
		return err
	}
	jwtSigned, err := sign(jwt, true, s.TokenEncryptionKey)
	if err != nil {
		return err
	}

	resp.Token = base64.StdEncoding.EncodeToString(jwtSigned)
	resp.ExpiredAt = time.Now().AddDate(0, 0, s.TokenExpirationDays)

	err = tokens.Save(resp.Token, s.TokenExpirationDays, s.TokenEncryptionKey)
	if err != nil {
		return terror.Error(err, "unable to save jwt")
	}

	return nil
}

type GenTokenReq struct {
	ApiKey string
	UserID string
	Device string
	Action string
}

func (s *S) TokenLogin(req TokenReq, resp *UserResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}
	tokenStr, err := base64.StdEncoding.DecodeString(req.TokenBase64)
	if err != nil {
		return terror.Error(err, "token is fail")
	}

	token, err := tokens.ReadJWT(tokenStr, true, s.TokenEncryptionKey)
	if err != nil {
		if errors.Is(err, tokens.ErrTokenExpired) {
			tknUuid, err := tokens.TokenID(token)
			if err != nil {
				return err
			}
			err = tokens.Remove(tknUuid)
			if err != nil {
				return err
			}
			return terror.Error(fmt.Errorf("session is expired"), "Session has expired, please log in again.")
		}
		return err
	}

	jwtIDI, ok := token.Get(openid.JwtIDKey)

	if !ok {
		return terror.Error(errors.New("unable to get ID from token"), "unable to read token")
	}

	jwtID, err := uuid.FromString(jwtIDI.(string))
	if err != nil {
		return terror.Error(err, "unable to form UUID from token")
	}

	retrievedToken, user, err := tokens.Retrieve(jwtID)
	if err != nil {
		return err
	}

	if !retrievedToken.Whitelisted() {
		return tokens.ErrTokenNotWhitelisted
	}

	resp.FactionID = user.FactionID
	resp.ID = user.ID
	resp.PublicAddress = user.PublicAddress
	resp.Username = user.Username
	resp.IsAdmin = user.RoleID == null.StringFrom(types.UserRoleAdminID.String())

	return nil
}

type GenOneTimeTokenReq struct {
	ApiKey string
	UserID string
}

type GenOneTimeTokenResp struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}

// GenOneTimeToken generate one time token for device login
func (s *S) GenOneTimeToken(req GenOneTimeTokenReq, resp *GenOneTimeTokenResp) error {
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		return err
	}

	user, err := boiler.FindUser(passdb.StdConn, req.UserID)
	if err != nil {
		return terror.Error(err, "Failed to get user")
	}

	tokenID := uuid.Must(uuid.NewV4())

	// Token expires in 5 minutes from now
	expires := time.Now().Add(time.Second * (60 * 5))

	// save user detail as jwt
	jwt, sign, err := tokens.GenerateOneTimeJWT(
		tokenID,
		expires,
		user.ID)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to generate one time token")
		return err
	}

	jwtSigned, err := sign(jwt, true, s.TokenEncryptionKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to sign jwt")
		return err
	}

	token := base64.StdEncoding.EncodeToString(jwtSigned)

	it := boiler.IssueToken{
		ID:        tokenID.String(),
		UserID:    user.ID,
		UserAgent: "qr-code-login",
		ExpiresAt: null.TimeFrom(expires),
	}
	err = it.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to insert one time token")
		return err
	}

	resp.Token = token
	resp.ExpiredAt = it.ExpiresAt.Time

	return nil
}

type LogoutResp struct {
	LogoutSuccess bool `json:"logout_success"`
}

func (s *S) TokenLogout(req TokenReq, resp *LogoutResp) error {
	resp.LogoutSuccess = false
	_, err := IsServerClient(req.ApiKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("is not a server client")
		return err
	}
	err = s.API.AuthLogout(req.TokenBase64, s.TokenEncryptionKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to delete all issued token to logout")
		return terror.Error(err, "Unable to delete all current sessions")
	}
	resp.LogoutSuccess = true
	return nil
}
