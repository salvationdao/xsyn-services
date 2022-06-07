package api

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/tokens"
	"xsyn-services/types"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofrs/uuid"
	twitch_jwt "github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type WalletLoginRequest struct {
	RedirectURL   *string            `json:"redirectURL"`
	PublicAddress string             `json:"public_address"`
	Signature     string             `json:"signature"`
	SessionID     hub.SessionID      `json:"session_id"`
	Fingerprint   *users.Fingerprint `json:"fingerprint"`
}

// LoginResponse is a response for login
type LoginResponse struct {
	User          *types.User `json:"user"`
	Token         string      `json:"token"`
	IsNew         bool        `json:"is_new"`
	RedirectToken *string     `json:"redirect_token,omitempty"`
}

func (api *API) WriteCookie(w http.ResponseWriter, token string) error {
	b64, err := api.Cookie.EncryptToBase64(token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Encryption error: %v", err), http.StatusBadRequest)
		return err
	}
	cookie := &http.Cookie{
		Name:     "xsyn-token",
		Value:    b64,
		Expires:  time.Now().Add(time.Hour * 24 * 7),
		Path:     "/",
		Secure:   api.IsCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)
	return nil
}

func (api *API) DeleteCookie(w http.ResponseWriter) error {
	cookie := &http.Cookie{
		Name:     "xsyn-token",
		Value:    "",
		Expires:  time.Now().AddDate(0, 0, -1),
		Path:     "/",
		Secure:   api.IsCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, cookie)
	return nil
}

func (api *API) ExternalLoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		passlog.L.Warn().Err(err).Msg("suspicious behaviour on external login form")
		return
	}

	authType := r.Form.Get("authType")
	redir := r.Form.Get("redirectURL")
	switch authType {
	case "wallet":
		req := &WalletLoginRequest{
			PublicAddress: r.Form.Get("public_address"),
			Signature:     r.Form.Get("signature"),
		}
		if redir != "" {
			req.RedirectURL = &redir
		}
		resp, err := api.WalletLogin(req, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = api.WriteCookie(w, resp.Token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if resp.RedirectToken != nil && redir != "" {
			escapedUrl := url.QueryEscape(*resp.RedirectToken)
			http.Redirect(w, r, redir+"?token="+escapedUrl, http.StatusSeeOther)
			return
		}
	case "token":
		req := &TokenLoginRequest{
			Token: r.Form.Get("token"),
		}
		resp, err := api.TokenAuth(req, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = api.WriteCookie(w, req.Token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if resp.RedirectToken != nil && redir != "" {
			escapedUrl := url.QueryEscape(*resp.RedirectToken)
			http.Redirect(w, r, redir+"?token="+escapedUrl, http.StatusSeeOther)
			return
		}
	case "cookie":
		cookie, err := r.Cookie("xsyn-token")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var token string
		if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// check user from token
		resp, err := api.TokenLogin(token, "")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		redirectToken := api.OneTimeToken(resp.User.ID, r.UserAgent())

		if redirectToken != nil && redir != "" {
			escapedUrl := url.QueryEscape(*redirectToken)
			http.Redirect(w, r, redir+"?token="+escapedUrl, http.StatusSeeOther)
			return
		}
	}

}

func (api *API) WalletLoginHandler(w http.ResponseWriter, r *http.Request) {
	req := &WalletLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return
	}

	resp, err := api.WalletLogin(req, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = api.WriteCookie(w, resp.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	b, _ := json.Marshal(resp.User)
	_, _ = w.Write(b)
}

func (api *API) WalletLogin(req *WalletLoginRequest, r *http.Request) (*LoginResponse, error) {
	// Take public address Hex to address(Make it a checksum mixed case address) convert back to Hex for string of checksum
	commonAddr := common.HexToAddress(req.PublicAddress)

	// Check if there are any existing users associated with the public address
	user, err := users.PublicAddress(commonAddr)
	if err != nil {
		return nil, fmt.Errorf("public address fail: %w", err)
	}

	// Fingerprint user
	if req.Fingerprint != nil {
		// todo: include ip in upsert
		err = api.DoFingerprintUpsert(*req.Fingerprint, user.ID)
		if err != nil {
			return nil, fmt.Errorf("browser identification fail: %w", err)
		}
	}

	user, _, token, err := api.IssueToken(&IssueTokenConfig{
		Encrypted: true,
		Key:       api.TokenEncryptionKey,
		Device:    r.UserAgent(),
		Action:    "login",
		User:      user.User,
	})
	if err != nil {
		return nil, fmt.Errorf("There was a problem creating a session for your account, please try again. %w", err)
	}

	err = api.VerifySignature(req.Signature, user.Nonce.String, commonAddr)
	if err != nil {
		return nil, err
	}

	if user.DeletedAt.Valid {
		return nil, fmt.Errorf("user does not exist")
	}

	var otToken *string = nil
	if req.RedirectURL != nil {
		otToken = api.OneTimeToken(user.ID, r.UserAgent())
	}

	return &LoginResponse{user, token, false, otToken}, nil
}

func (api *API) DoFingerprintUpsert(fingerprint users.Fingerprint, userID string) error {
	err := users.FingerprintUpsert(fingerprint, userID)
	if err != nil {
		return terror.Warn(err, fmt.Sprintf("Could not upsert fingerprint for user %s", userID))
	}

	return nil
}

type IssueTokenConfig struct {
	Encrypted bool
	Key       []byte
	Device    string
	Action    string
	Email     string
	Picture   string
	User      *boiler.User
	Mutate    func(jwt.Token) jwt.Token
}

var ErrNoUserInformation = errors.New("no user information provided to IssueToken()")

func (api *API) OneTimeToken(userID string, userAgent string) *string {
	var err error
	tokenID := uuid.Must(uuid.NewV4())

	expires := time.Now().Add(time.Second * 60)

	// save user detail as jwt
	jwt, sign, err := tokens.GenerateOneTimeJWT(
		tokenID,
		userID, expires)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to generate one time token")
		return nil
	}

	jwtSigned, err := sign(jwt, true, api.TokenEncryptionKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to sign jwt")
		return nil
	}

	token := base64.StdEncoding.EncodeToString(jwtSigned)

	it := boiler.IssueToken{
		ID:        tokenID.String(),
		UserID:    userID,
		UserAgent: userAgent,
		ExpiresAt: null.TimeFrom(expires),
	}
	err = it.Insert(passdb.StdConn, boil.Infer())
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to insert one time token")
		return nil
	}

	return &token
}

func (api *API) IssueToken(config *IssueTokenConfig) (*types.User, uuid.UUID, string, error) {
	var err error
	errMsg := "There was a problem with your authentication, please check your details and try again."

	// Get user by email
	if config.Email == "" && config.User == nil {
		return nil, uuid.Nil, "", terror.Error(ErrNoUserInformation, errMsg)
	}
	var user *types.User
	if config.User == nil {
		user, err = users.Email(config.Email)
		if err != nil {
			return nil, uuid.Nil, "", terror.Error(err, errMsg)
		}
	} else {
		user, err = types.UserFromBoil(config.User)
		if err != nil {
			return nil, uuid.Nil, "", terror.Error(err, errMsg)
		}
	}

	tokenID := uuid.Must(uuid.NewV4())
	// save user detail as jwt
	jwt, sign, err := tokens.GenerateJWT(
		tokenID,
		user.User,
		config.Device,
		config.Action,
		api.TokenExpirationDays)
	if err != nil {
		return nil, uuid.Nil, "", terror.Error(err, errMsg)
	}
	// Record token in issued token records
	if config.Mutate != nil {
		jwt = config.Mutate(jwt)
	}
	jwtSigned, err := sign(jwt, config.Encrypted, config.Key)
	if err != nil {
		return nil, uuid.Nil, "", terror.Error(err, "unable to sign jwt")
	}

	token := base64.StdEncoding.EncodeToString(jwtSigned)

	err = tokens.Save(token, api.TokenExpirationDays, api.TokenEncryptionKey)
	if err != nil {
		return nil, uuid.Nil, "", terror.Error(err, "unable to save jwt")
	}

	return user, tokenID, token, nil
}

func (api *API) VerifySignature(signature string, nonce string, publicKey common.Address) error {
	decodedSig, err := hexutil.Decode(signature)
	if err != nil {
		return err
	}

	if decodedSig[64] == 0 || decodedSig[64] == 1 {
		//https://ethereum.stackexchange.com/questions/102190/signature-signed-by-go-code-but-it-cant-verify-on-solidity
		decodedSig[64] += 27
	} else if decodedSig[64] != 27 && decodedSig[64] != 28 {
		return terror.Error(fmt.Errorf("decode sig invalid %v", decodedSig[64]))
	}
	decodedSig[64] -= 27

	msg := []byte(fmt.Sprintf("%s:\n %s", api.Eip712Message, nonce))
	prefixedNonce := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)

	hash := crypto.Keccak256Hash([]byte(prefixedNonce))
	recoveredPublicKey, err := crypto.Ecrecover(hash.Bytes(), decodedSig)
	if err != nil {
		return err
	}

	secp256k1RecoveredPublicKey, err := crypto.UnmarshalPubkey(recoveredPublicKey)
	if err != nil {
		return err
	}

	recoveredAddress := crypto.PubkeyToAddress(*secp256k1RecoveredPublicKey).Hex()
	isClientAddressEqualToRecoveredAddress := strings.ToLower(publicKey.Hex()) == strings.ToLower(recoveredAddress)
	if !isClientAddressEqualToRecoveredAddress {
		return terror.Error(fmt.Errorf("public address does not match recovered address"))
	}
	return nil
}

// TwitchJWTClaims is the payload of a JWT sent by the Twitch extension
type TwitchJWTClaims struct {
	OpaqueUserID    string `json:"opaque_user_id,omitempty"`
	TwitchAccountID string `json:"user_id"`
	ChannelID       string `json:"channel_id,omitempty"`
	Role            string `json:"role"`
	twitch_jwt.StandardClaims
}

// GetClaimsFromTwitchExtensionToken verifies token from Twitch extension
func (api *API) GetClaimsFromTwitchExtensionToken(token string) (*TwitchJWTClaims, error) {
	// Get claims
	claims := &TwitchJWTClaims{}

	_, err := twitch_jwt.ParseWithClaims(token, claims, func(t *twitch_jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*twitch_jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return api.Twitch.ExtensionSecret, nil
	})
	if err != nil {
		return nil, terror.Error(terror.ErrBadClaims, "Invalid token")
	}

	return claims, nil
}

type GetNonceResponse struct {
	Nonce string `json:"nonce"`
}

func (api *API) NewNonce(user *boiler.User) (string, error) {
	newNonce := helpers.RandStringBytes(16)

	user.Nonce = null.StringFrom(newNonce)
	i, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Nonce))
	if err != nil {
		return "", err
	}

	if i == 0 {
		return "", terror.Error(fmt.Errorf("nonce could not be updated"))
	}

	return newNonce, nil
}

func (api *API) GetNonce(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := r.URL.Query().Get("public-address")
	userID := r.URL.Query().Get("user-id")

	if publicAddress == "" && userID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing public address or user id"))
	}
	if publicAddress != "" {
		// Take public address Hex to address(Make it a checksum mixed case address) convert back to Hex for string of checksum
		commonAddr := common.HexToAddress(publicAddress)
		user, err := users.PublicAddress(commonAddr)
		if err != nil && errors.Is(sql.ErrNoRows, err) {
			passlog.L.Info().Err(err).Msg("new user being created")
			username := commonAddr.Hex()[0:10]

			// If user does not exist, create new user with their username set to their MetaMask public address
			user, err = users.UserCreator("", "", helpers.TrimUsername(username), "", "", "", "", "", "", "", commonAddr, "")
			if err != nil {
				passlog.L.Error().Err(err).Msg("user creation failed")
				return http.StatusInternalServerError, err
			}
		}
		passlog.L.Info().Err(err).Msg("doing nonce")

		newNonce, err := api.NewNonce(user.User)
		if err != nil {
			passlog.L.Error().Err(err).Msg("no nonce")
			return http.StatusBadRequest, err
		}

		passlog.L.Info().Err(err).Msg("after doing nonce")
		resp := &GetNonceResponse{
			Nonce: newNonce,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			passlog.L.Error().Err(err).Msg("json failed")
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	}

	user, err := boiler.FindUser(passdb.StdConn, userID)
	if err != nil {
		return http.StatusBadRequest, err
	}

	newNonce, err := api.NewNonce(user)
	if err != nil {
		return http.StatusBadRequest, err
	}

	resp := &GetNonceResponse{
		Nonce: newNonce,
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// TokenLoginRequest is an auth request that uses a JWT
type TokenLoginRequest struct {
	Token              string             `json:"token"`
	SessionID          hub.SessionID      `json:"session_id"`
	TwitchExtensionJWT string             `json:"twitch_extension_jwt"`
	Fingerprint        *users.Fingerprint `json:"fingerprint"`
	RedirectURL        *string            `json:"redirectURL"`
}

// TokenLoginResponse is an auth request that uses a JWT
type TokenLoginResponse struct {
	User *boiler.User `json:"user"`
}

func (api *API) TokenAuth(req *TokenLoginRequest, r *http.Request) (*LoginResponse, error) {

	resp, err := api.TokenLogin(req.Token, req.TwitchExtensionJWT)
	if err != nil {
		return nil, fmt.Errorf("failed to login with token: %w", err)
	}

	// Fingerprint user
	if req.Fingerprint != nil {
		userID := resp.User.ID
		// todo: include ip in upsert
		err = api.DoFingerprintUpsert(*req.Fingerprint, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to identify browser: %w", err)
		}
	}

	if resp.User.DeletedAt.Valid {
		return nil, fmt.Errorf("user does not exist")
	}

	var otToken *string = nil
	if req.RedirectURL != nil {
		otToken = api.OneTimeToken(resp.User.ID, r.UserAgent())
	}

	user, err := types.UserFromBoil(resp.User)
	if err != nil {
		return nil, fmt.Errorf("failed to identify user: %w", err)
	}

	return &LoginResponse{user, req.Token, false, otToken}, nil
}

func (api *API) AuthCheckHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		// check whether token is attached
		token := r.URL.Query().Get("token")
		if token == "" {
			passlog.L.Error().Msg("No token found")
			return http.StatusBadRequest, terror.Warn(fmt.Errorf("no cookie and token are provided"), "User are not signed in.")
		}

		// check user from token
		resp, err := api.TokenLogin(token, "")
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
		}

		// write cookie
		err = api.WriteCookie(w, token)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to write cookie")
		}

		return helpers.EncodeJSON(w, resp.User)
	}

	var token string
	if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to decrypt token")
	}

	// check user from token
	resp, err := api.TokenLogin(token, "")
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
	}

	return helpers.EncodeJSON(w, resp.User)
}

func (api *API) AuthLogoutHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	_, err := r.Cookie("xsyn-token")
	if err != nil {
		// check whether token is attached
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no cookie are provided"), "User is not signed in.")
	}

	// clear and expire cookie and push to browser
	err = api.DeleteCookie(w)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to logout user.")
	}

	return helpers.EncodeJSON(w, true)
}

// TokenLoginHandler lets you log in with just a jwt
func (api *API) TokenLoginHandler(w http.ResponseWriter, r *http.Request) {
	req := &TokenLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "auth fail", http.StatusBadRequest)
		return
	}

	resp, err := api.TokenAuth(req, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = api.WriteCookie(w, req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "failed state", http.StatusBadRequest)
	}
}

// TokenLogin gets a user from the token
func (api *API) TokenLogin(tokenBase64 string, twitchExtensionJWT string) (*TokenLoginResponse, error) {
	tokenStr, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return nil, terror.Error(err, "")
	}

	token, err := tokens.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		if errors.Is(err, tokens.ErrTokenExpired) {
			tknUuid, err := tokens.TokenID(token)
			if err != nil {
				return nil, err
			}
			err = tokens.Remove(tknUuid)
			if err != nil {
				return nil, err
			}
			return nil, terror.Warn(err, "Session has expired, please log in again.")
		}
		return nil, err
	}

	jwtIDI, ok := token.Get(openid.JwtIDKey)

	if !ok {
		return nil, terror.Error(errors.New("unable to get ID from token"), "unable to read token")
	}

	jwtID, err := uuid.FromString(jwtIDI.(string))
	if err != nil {
		return nil, terror.Error(err, "unable to form UUID from token")
	}

	retrievedToken, user, err := tokens.Retrieve(jwtID)
	if err != nil {
		return nil, err
	}

	if !retrievedToken.Whitelisted() {
		return nil, terror.Error(tokens.ErrTokenNotWhitelisted)
	}

	// check twitch extension jwt
	if twitchExtensionJWT != "" {
		claims, err := api.GetClaimsFromTwitchExtensionToken(twitchExtensionJWT)
		if err != nil {
			return nil, terror.Error(err, "failed to parse twitch extension token")
		}

		twitchUser, err := users.TwitchID(claims.TwitchAccountID)
		if err != nil {
			return nil, terror.Error(err, "failed to get twitch user")
		}

		// check twitch user match the token user
		if twitchUser.ID != user.ID {
			return nil, terror.Error(tokens.ErrUserNotMatch, "twitch id does not match")
		}
	}

	return &TokenLoginResponse{user}, nil
}

type BotListResponse struct {
	RedMountainBotIDs []string `json:"red_mountain_bot_ids"`
	BostonBotIDs      []string `json:"boston_bot_ids"`
	ZaibatsuBotIDs    []string `json:"zaibatsu_bot_ids"`
}

type BotListHandlerRequest struct {
	BotSecretKey string `json:"bot_secret_key"`
}

func (api *API) BotListHandler(w http.ResponseWriter, r *http.Request) {
	req := &BotListHandlerRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "auth fail", http.StatusBadRequest)
		return
	}
	// check header
	if req.BotSecretKey == "" || req.BotSecretKey != api.botSecretKey {
		passlog.L.Warn().Str("expected secret key", api.botSecretKey).Str("provided secret key", r.Header.Get("bot_secret_key")).Msg("bot secret key check failed")
		http.Error(w, "auth fail", http.StatusUnauthorized)
		return
	}

	bots, err := boiler.Users(
		qm.Select(
			fmt.Sprintf("%s as id", qm.Rels(boiler.TableNames.Users, boiler.UserColumns.ID)),
			boiler.UserColumns.FactionID,
		),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = %s AND %s = 'Bot'",
				boiler.TableNames.Roles,
				qm.Rels(boiler.TableNames.Roles, boiler.RoleColumns.ID),
				qm.Rels(boiler.TableNames.Users, boiler.UserColumns.RoleID),
				qm.Rels(boiler.TableNames.Roles, boiler.RoleColumns.Name),
			),
		),
	).All(passdb.StdConn)
	if err != nil {
		http.Error(w, "failed to get bot list", http.StatusInternalServerError)
		return
	}

	resp := &BotListResponse{
		RedMountainBotIDs: []string{},
		BostonBotIDs:      []string{},
		ZaibatsuBotIDs:    []string{},
	}

	for _, b := range bots {
		switch b.FactionID.String {
		case types.RedMountainFactionID.String():
			resp.RedMountainBotIDs = append(resp.RedMountainBotIDs, b.ID)
		case types.BostonCyberneticsFactionID.String():
			resp.BostonBotIDs = append(resp.BostonBotIDs, b.ID)
		case types.ZaibatsuFactionID.String():
			resp.ZaibatsuBotIDs = append(resp.ZaibatsuBotIDs, b.ID)
		}
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, "failed state", http.StatusBadRequest)
	}
}

type BotTokenLoginRequest struct {
	BotSecretKey string `json:"bot_secret_key"`
	BotID        string `json:"bot_id"`
}

type BotTokenResponse struct {
	User  *boiler.User `json:"user"`
	Token string       `json:"token"`
}

// BotTokenLoginHandler return a bot user and access token from the given bot token
func (api *API) BotTokenLoginHandler(w http.ResponseWriter, r *http.Request) {
	req := &BotTokenLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "auth fail", http.StatusBadRequest)
		return
	}

	if req.BotSecretKey == "" || req.BotSecretKey != api.botSecretKey {
		passlog.L.Warn().Str("expected secret key", api.botSecretKey).Str("provided secret key", r.Header.Get("bot_secret_key")).Msg("bot secret key check failed")
		http.Error(w, "auth fail", http.StatusBadRequest)
		return
	}

	// return a bot user and generate an access_token
	user, err := boiler.Users(
		boiler.UserWhere.ID.EQ(req.BotID),
		qm.InnerJoin(
			fmt.Sprintf(
				"%s ON %s = %s AND %s = 'Bot'",
				boiler.TableNames.Roles,
				qm.Rels(boiler.TableNames.Roles, boiler.RoleColumns.ID),
				qm.Rels(boiler.TableNames.Users, boiler.UserColumns.RoleID),
				qm.Rels(boiler.TableNames.Roles, boiler.RoleColumns.Name),
			),
		),
		qm.Load(boiler.UserRels.Faction),
	).One(passdb.StdConn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _, token, err := api.IssueToken(&IssueTokenConfig{
		Encrypted: true,
		Key:       api.TokenEncryptionKey,
		Device:    "gamebot",
		Action:    "bot_login",
		User:      user,
	})
	if err != nil {
		http.Error(w, "failed to generate access token", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(&BotTokenResponse{user, token})
	if err != nil {
		http.Error(w, "failed state", http.StatusBadRequest)
	}
}

// UserFingerprintHandler stores a fingerprint entry that may or may not be linked to a user yet
func (api *API) UserFingerprintHandler(w http.ResponseWriter, r *http.Request) error {
	req := &UserFingerprintRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return err
	}

	fingerprintExists, err := boiler.Fingerprints(boiler.FingerprintWhere.VisitorID.EQ(req.Payload.Fingerprint.VisitorID)).Exists(passdb.StdConn)
	if err != nil {
		return err
	}

	if !fingerprintExists {
		newFingerprint := boiler.Fingerprint{
			VisitorID:  req.Payload.Fingerprint.VisitorID,
			OsCPU:      null.StringFrom(req.Payload.Fingerprint.OSCPU),
			Platform:   null.StringFrom(req.Payload.Fingerprint.Platform),
			Timezone:   null.StringFrom(req.Payload.Fingerprint.Timezone),
			Confidence: decimal.NewNullDecimal(decimal.NewFromFloat32(req.Payload.Fingerprint.Confidence)),
			UserAgent:  null.StringFrom(req.Payload.Fingerprint.UserAgent),
		}
		err = newFingerprint.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}

	fingerprint, err := boiler.Fingerprints(boiler.FingerprintWhere.VisitorID.EQ(req.Payload.Fingerprint.VisitorID)).One(passdb.StdConn)
	if err != nil {
		return err
	}

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ipaddr, _, _ := net.SplitHostPort(r.RemoteAddr)
		userIP := net.ParseIP(ipaddr)
		if userIP == nil {
			ip = ipaddr
		} else {
			ip = userIP.String()
		}
	}

	userIPExists, err := boiler.FingerprintIps(boiler.FingerprintIPWhere.IP.EQ(ip), boiler.FingerprintIPWhere.FingerprintID.EQ(fingerprint.ID)).Exists(passdb.StdConn)
	if err != nil {
		return err
	}
	if !userIPExists {
		// IP not logged for this fingerprint yet; create one
		newFingerprintIP := boiler.FingerprintIP{
			IP:            ip,
			FingerprintID: fingerprint.ID,
		}
		err = newFingerprintIP.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			return err
		}
	}

	return nil
}
