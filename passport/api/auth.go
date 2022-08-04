package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	pCrypto "xsyn-services/passport/crypto"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/tokens"
	"xsyn-services/types"

	"github.com/dghubble/oauth1"
	"github.com/dghubble/oauth1/twitter"
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
	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type SignupRequest struct {
	Username        string               `json:"username"`
	Fingerprint     *users.Fingerprint   `json:"fingerprint"`
	AuthType        string               `json:"auth_type"`
	WalletRequest   WalletLoginRequest   `json:"wallet_request"`
	GoogleRequest   GoogleLoginRequest   `json:"google_request"`
	FacebookRequest FacebookLoginRequest `json:"facebook_request"`
	EmailRequest    EmailLoginRequest    `json:"email_request"`
	TwitterRequest  TwitterSignupRequest `json:"twitter_request"`
}

type WalletLoginRequest struct {
	RedirectURL   string             `json:"redirect_url"`
	Tenant        string             `json:"tenant"`
	PublicAddress string             `json:"public_address"`
	Signature     string             `json:"signature"`
	SessionID     hub.SessionID      `json:"session_id"`
	Fingerprint   *users.Fingerprint `json:"fingerprint"`
	AuthType      string             `json:"auth_type"`
}

type EmailSignupVerifyRequest struct {
	RedirectURL string `json:"redirect_url"`
	Tenant      string `json:"tenant"`
	Email       string `json:"email"`
	Code        string `json:"code"`
}

type EmailLoginRequest struct {
	RedirectURL string             `json:"redirect_url"`
	Tenant      string             `json:"tenant"`
	Email       string             `json:"email"`
	Password    string             `json:"password"`
	SessionID   hub.SessionID      `json:"session_id"`
	Fingerprint *users.Fingerprint `json:"fingerprint"`
}
type ForgotPasswordRequest struct {
	Tenant      string             `json:"tenant"`
	Email       string             `json:"email"`
	SessionID   hub.SessionID      `json:"session_id"`
	Fingerprint *users.Fingerprint `json:"fingerprint"`
}

type PasswordUpdateRequest struct {
	RedirectURL string             `json:"redirect_url"`
	Tenant      string             `json:"tenant"`
	UserID      string             `json:"user_id"`
	Password    string             `json:"password"`
	TokenID     string             `json:"id"`
	Token       string             `json:"token"`
	NewPassword string             `json:"new_password"`
	SessionID   hub.SessionID      `json:"session_id"`
	Fingerprint *users.Fingerprint `json:"fingerprint"`
}

type GoogleLoginRequest struct {
	RedirectURL string             `json:"redirect_url"`
	Tenant      string             `json:"tenant"`
	GoogleID    string             `json:"google_id"`
	Email       string             `json:"email"`
	SessionID   hub.SessionID      `json:"session_id"`
	Fingerprint *users.Fingerprint `json:"fingerprint"`
	AuthType    string             `json:"auth_type"`
}

type FacebookLoginRequest struct {
	RedirectURL string             `json:"redirect_url"`
	Tenant      string             `json:"tenant"`
	FacebookID  string             `json:"facebook_id"`
	SessionID   hub.SessionID      `json:"session_id"`
	Fingerprint *users.Fingerprint `json:"fingerprint"`
	AuthType    string             `json:"auth_type"`
}

type TwitterSignupRequest struct {
	RedirectURL string             `json:"redirect_url"`
	Tenant      string             `json:"tenant"`
	TwitterID   string             `json:"twitter_id"`
	SessionID   hub.SessionID      `json:"session_id"`
	Fingerprint *users.Fingerprint `json:"fingerprint"`
}

type TFAVerifyRequest struct {
	RedirectURL  string             `json:"redirect_url"`
	Tenant       string             `json:"tenant"`
	UserID       string             `json:"user_id"`
	Token        string             `json:"token"`
	Passcode     string             `json:"passcode"`
	RecoveryCode string             `json:"recovery_code"`
	SessionID    hub.SessionID      `json:"session_id"`
	Fingerprint  *users.Fingerprint `json:"fingerprint"`
}

// LoginResponse is a response for login
type LoginResponse struct {
	User  *types.User `json:"user"`
	Token string      `json:"token"`
	IsNew bool        `json:"is_new"`
}

type ForgotPasswordResponse struct {
	Message string `json:"message"`
}

func (api *API) WriteCookie(w http.ResponseWriter, r *http.Request, token string) error {
	b64, err := api.Cookie.EncryptToBase64(token)
	if err != nil {
		return err
	}

	// get domain

	d := domain(r.Host)
	if d == "" {
		passlog.L.Warn().Msg("Cookie's domain not found")
		return fmt.Errorf("failed to write cookie")
	}

	cookie := &http.Cookie{
		Name:     "xsyn-token",
		Value:    b64,
		Expires:  time.Now().AddDate(0, 0, api.TokenExpirationDays), // sync with token
		Path:     "/",
		Secure:   api.IsCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Domain:   d,
	}
	http.SetCookie(w, cookie)
	return nil
}

func domain(host string) string {
	parts := strings.Split(host, ".")

	if len(parts) < 2 {
		return ""
	}
	//this is rigid as fuck
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

func (api *API) DeleteCookie(w http.ResponseWriter, r *http.Request) error {
	// remove cookie on domain
	cookie := &http.Cookie{
		Name:     "xsyn-token",
		Value:    "",
		Expires:  time.Now().AddDate(0, 0, -1),
		Path:     "/",
		Secure:   api.IsCookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Domain:   domain(r.Host),
	}
	http.SetCookie(w, cookie)

	// remove cookie on the site, just in case there is one
	cookie = &http.Cookie{
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

func (api *API) ExternalLoginCheckHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	err := r.ParseForm()
	if err != nil {
		passlog.L.Warn().Err(err).Msg("suspicious behaviour on external cookie check")
		return http.StatusBadRequest, err
	}

	redirectURL := r.Form.Get("redirect_url")
	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		return http.StatusBadRequest, err
	}

	var token string
	if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
		return http.StatusBadRequest, err
	}

	// check user from token
	_, err = api.TokenLogin(token, "")
	if err != nil {
		return http.StatusBadRequest, err
	}

	// write cookie on domain
	err = api.WriteCookie(w, r, token)
	if err != nil {
		return http.StatusBadRequest, err
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	return http.StatusOK, nil

}

func (api *API) EmailSignupVerifyHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &EmailSignupVerifyRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}
	err = api.EmailSignupVerify(req, w, r)

	if err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusCreated, nil
}

func (api *API) SignupHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &SignupRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}
	username := req.Username
	authType := req.AuthType
	if authType == "" {
		passlog.L.Error().Err(err).Msg("auth type is missing in user signup")
		return http.StatusBadRequest, err
	}
	u := &types.User{}

	switch authType {
	case "wallet":
		// Take public address Hex to address(Make it a checksum mixed case address) convert back to Hex for string of checksum
		commonAddr := common.HexToAddress(req.WalletRequest.PublicAddress)
		user, err := users.PublicAddress(commonAddr)
		if err != nil && errors.Is(sql.ErrNoRows, err) {
			// If user does not exist, create new user with their username set to their MetaMask public address
			u, err = users.UserCreator("", "", username, "", "", "", "", "", "", "", commonAddr, "")
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to create user with wallet")
				return http.StatusInternalServerError, err
			}
		} else if user != nil {
			err := fmt.Errorf("User already exist")
			return http.StatusInternalServerError, err
		}
	case "email":
		// Check no user with email exist
		user, err := users.Email(req.EmailRequest.Email)
		if err != nil && errors.Is(sql.ErrNoRows, err) {
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator("", "", username, req.EmailRequest.Email, "", "", "", "", "", "", commonAddress, req.EmailRequest.Password)
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to create user with email and password")
				return http.StatusInternalServerError, err
			}
		} else if user != nil {
			err := fmt.Errorf("User already exist")
			return http.StatusInternalServerError, err
		}
	case "facebook":
		// Check no user with email exist
		user, err := users.FacebookID(req.FacebookRequest.FacebookID)
		if err != nil && errors.Is(sql.ErrNoRows, err) {
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator("", "", username, "", req.FacebookRequest.FacebookID, "", "", "", "", "", commonAddress, "")
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to create user with email and password")
				return http.StatusInternalServerError, err
			}

		} else if user != nil {
			err := fmt.Errorf("User already exist")
			return http.StatusInternalServerError, err
		}
	case "google":
		// Check no user with email exist
		user, err := users.GoogleID(req.GoogleRequest.GoogleID)
		if err != nil && errors.Is(sql.ErrNoRows, err) {
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator("", "", username, req.GoogleRequest.Email, "", req.GoogleRequest.GoogleID, "", "", "", "", commonAddress, "")
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to create user with email and password")
				return http.StatusInternalServerError, err
			}
		} else if user != nil {
			err := fmt.Errorf("User already exist")
			return http.StatusInternalServerError, err
		}
	case "twitter":
		// Check no user with email exist
		user, err := users.TwitterID(req.TwitterRequest.TwitterID)
		if err != nil && errors.Is(sql.ErrNoRows, err) {
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator("", "", username, "", "", "", "", req.TwitterRequest.TwitterID, "", "", commonAddress, "")
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to create user with email and password")
				return http.StatusInternalServerError, err
			}
		} else if user != nil {
			err := fmt.Errorf("User already exist")
			return http.StatusInternalServerError, err
		}
	}

	if u.ID == "" {
		passlog.L.Error().Err(err).Msg("unable to create user, no user passed through")
		return http.StatusInternalServerError, err
	}
	// Login
	loginReq := &FingerprintTokenRequest{
		User:        &u.User,
		Fingerprint: req.Fingerprint,
	}

	err = api.FingerprintAndIssueToken(w, r, loginReq)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil

}

func (api *API) EmailVerifyHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	tokenBase64 := r.URL.Query().Get("token")
	userID, newEmail, err := api.UserFromToken(w, r, tokenBase64)
	if err != nil {
		return http.StatusBadRequest, err
	}

	user, err := users.ID(userID)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Check if user is new or just updating email
	if user.User.Verified {
		_, err = mail.ParseAddress(newEmail)
		if err != nil {
			return http.StatusBadRequest, err
		}
		user.Email = null.NewString(newEmail, true)
		_, err := user.User.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Email))
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to update user's email")
			return http.StatusBadRequest, err
		}

	} else {
		// New user
		user.User.Verified = true
		_, err := user.User.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Verified))
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to set user as verified")
			return http.StatusBadRequest, err
		}
		ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)
	}

	b, _ := json.Marshal(user)
	_, _ = w.Write(b)

	return http.StatusOK, nil
}

func (api *API) EmailSignupVerify(req *EmailSignupVerifyRequest, w http.ResponseWriter, r *http.Request) error {
	// Check if there are any existing users associated with the email address
	user, _ := users.Email(req.Email)

	if user != nil {
		return fmt.Errorf("email is already used by a different user")
	}

	err := api.Mailer.SendSignupEmail(context.Background(), req.Email, req.Code)
	if err != nil {
		return err
	}
	return nil
}

func (api *API) UserFromToken(w http.ResponseWriter, r *http.Request, tokenBase64 string) (string, string, error) {
	fmt.Println(tokenBase64)
	tokenStr, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return "", "", err
	}

	// Decode token user with new email
	token, err := tokens.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		return "", "", err
	}

	uID, ok := token.Get("user-id")
	if !ok {
		passlog.L.Error().Err(err).Msg("unable to get user id from token")
		return "", "", fmt.Errorf("invalid token found")
	}
	email, ok := token.Get(openid.EmailKey)
	if !ok {
		passlog.L.Error().Err(err).Msg("unable to get email from token")
		return "", "", fmt.Errorf("invalid token provided")
	}

	newEmail := email.(string)
	userID := uID.(string)

	return userID, newEmail, nil
}

func (api *API) EmailLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &EmailLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	err = api.EmailLogin(req, w, r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}
func (api *API) EmailLogin(req *EmailLoginRequest, w http.ResponseWriter, r *http.Request) error {
	user, err := users.EmailPassword(req.Email, req.Password)

	if err != nil {
		return err
	}
	loginReq := &FingerprintTokenRequest{
		User:        &user.User,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
	}
	err = api.FingerprintAndIssueToken(w, r, loginReq)
	if err != nil {
		return err
	}
	return nil
}

func (api *API) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &ForgotPasswordRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	user, err := users.Email(req.Email)
	if err != nil {
		return http.StatusBadRequest, fmt.Errorf("no user found with email: %q", req.Email)
	}
	user, tokenID, token, err := api.ResetPasswordToken(&TokenConfig{
		Encrypted: true,
		Key:       api.TokenEncryptionKey,
		Device:    r.UserAgent(),
		Action:    "forgot",
		User:      &user.User,
	})
	if err != nil {
		return http.StatusBadRequest, err
	}

	err = api.Mailer.SendForgotPasswordEmail(context.Background(), user, token, tokenID)
	if err != nil {
		return http.StatusBadRequest, err

	}
	resp := &ForgotPasswordResponse{Message: "Success! An email has been sent to recover your account."}

	b, err := json.Marshal(resp.Message)
	if err != nil {
		return http.StatusBadRequest, err
	}
	httpStatus, err := w.Write(b)
	if err != nil {
		return http.StatusBadRequest, err
	}
	return httpStatus, nil

}

func (api *API) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &PasswordUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	tokenStr, err := base64.StdEncoding.DecodeString(req.Token)
	if err != nil {
		return http.StatusBadRequest, err
	}
	// Decode token user with new email
	token, err := tokens.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		return http.StatusBadRequest, err
	}

	uID, _ := token.Get("user-id")
	userID, ok := uID.(string)

	if !ok {
		return http.StatusBadRequest, fmt.Errorf("Invalid token provided")
	}
	user, err := users.ID(userID)
	if err != nil {
		return http.StatusBadRequest, err
	}
	if err != nil {
		return http.StatusBadRequest, err
	}
	return passwordReset(api, w, r, req, &user.User)

}

func (api *API) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &PasswordUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	userPassword, err := boiler.FindPasswordHash(passdb.StdConn, req.UserID)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Check if current password is correct
	err = pCrypto.ComparePassword(userPassword.PasswordHash, req.Password)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Get user
	user, err := boiler.Users(
		boiler.UserWhere.ID.EQ(req.UserID),
		qm.Load(qm.Rels(boiler.UserRels.Faction)),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return passwordReset(api, w, r, req, user)

}

func (api *API) NewPasswordHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &PasswordUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Find user by id
	user, err := boiler.FindUser(passdb.StdConn, req.UserID)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Check if user has password already
	passwordExist, err := boiler.PasswordHashExists(passdb.StdConn, req.UserID)
	if err != nil {
		return http.StatusBadRequest, err
	}
	if passwordExist {
		return http.StatusBadRequest, fmt.Errorf("user already has a password")
	}

	return passwordReset(api, w, r, req, user)
}

func passwordReset(api *API, w http.ResponseWriter, r *http.Request, req *PasswordUpdateRequest, user *boiler.User) (int, error) {
	// Check if new password is valid
	err := helpers.IsValidPassword(req.NewPassword)
	if err != nil {
		return http.StatusBadRequest, err
	}
	// Setup user activity tracking
	oldUser := *user

	newPasswordHash := pCrypto.HashPassword(req.NewPassword)

	// Find password
	userPassword, _ := boiler.FindPasswordHash(passdb.StdConn, user.ID)
	if userPassword == nil {
		newPassword := &boiler.PasswordHash{
			UserID:       user.ID,
			PasswordHash: newPasswordHash,
		}
		err = newPassword.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to set new password for user")
			return http.StatusBadRequest, err
		}
	} else {
		// Update password
		userPassword.PasswordHash = newPasswordHash
		_, err = userPassword.Update(passdb.StdConn, boil.Whitelist(boiler.PasswordHashColumns.PasswordHash))
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to update user password")
			return http.StatusBadRequest, err
		}
	}
	api.RecordUserActivity(context.Background(),
		user.ID,
		"Change password",
		types.ObjectTypeUser,
		helpers.StringPointer(user.ID),
		&user.Username,
		helpers.StringPointer(user.FirstName.String+" "+user.LastName.String),
		&types.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	// Delete all issued token
	_, err = user.IssueTokens().UpdateAll(passdb.StdConn, boiler.M{
		boiler.IssueTokenColumns.DeletedAt: time.Now(),
	})
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to delete all issued token for password reset")
		return http.StatusBadRequest, err
	}

	// Generate new token and login
	loginReq := &FingerprintTokenRequest{
		User:        user,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
	}
	err = api.FingerprintAndIssueToken(w, r, loginReq)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Send message to users
	URI := fmt.Sprintf("/user/%s", user.ID)
	ws.PublishMessage(URI, HubKeyUserInit, nil)

	return http.StatusCreated, nil
}

func (api *API) WalletLoginHandler(w http.ResponseWriter, r *http.Request) {
	req := &WalletLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode wallet login request")
		return
	}
	err = api.WalletLogin(req, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func (api *API) WalletLogin(req *WalletLoginRequest, w http.ResponseWriter, r *http.Request) error {
	// Take public address Hex to address(Make it a checksum mixed case address) convert back to Hex for string of checksum
	commonAddr := common.HexToAddress(req.PublicAddress)

	// Check if there are any existing users associated with the public address
	user, err := users.PublicAddress(commonAddr)

	if err != nil && errors.Is(sql.ErrNoRows, err) {
		// Return request back to user for signup
		b, err := json.Marshal(req)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return err
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("public address fail: %w", err)
	}

	err = api.VerifySignature(req.Signature, user.Nonce.String, commonAddr)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to verify signature")
		return err
	}

	loginReq := &FingerprintTokenRequest{
		User:        &user.User,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
	}
	err = api.FingerprintAndIssueToken(w, r, loginReq)
	if err != nil {
		return err
	}

	if user.DeletedAt.Valid {
		return fmt.Errorf("user does not exist")
	}

	return nil
}

func (api *API) DoFingerprintUpsert(fingerprint users.Fingerprint, userID string) error {
	err := users.FingerprintUpsert(fingerprint, userID)
	if err != nil {
		return terror.Warn(err, fmt.Sprintf("Could not upsert fingerprint for user %s", userID))
	}

	return nil
}

func (api *API) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &GoogleLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode google login request")
		return http.StatusBadRequest, err
	}
	err = api.GoogleLogin(req, w, r)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to google login")
		return http.StatusBadRequest, terror.Error(err, "unable to google auth")
	}
	return http.StatusCreated, nil
}

func (api *API) GoogleLogin(req *GoogleLoginRequest, w http.ResponseWriter, r *http.Request) error {
	L := passlog.L.With().Str("func", "GoogleLogin").Logger()
	L.Info().Msg("hit handler")
	// Check if there are any existing users associated with the email address
	user, err := users.GoogleID(req.GoogleID)

	loginReq := &FingerprintTokenRequest{
		User:        user,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
	}

	if err != nil && errors.Is(sql.ErrNoRows, err) {
		// Check if user gmail already exist
		if req.Email == "" {
			noEmailErr := fmt.Errorf("no email provided for google auth")
			passlog.L.Error().Err(noEmailErr).Msg("no email provided for google auth")
			return noEmailErr
		}

		user, _ = boiler.Users(boiler.UserWhere.Email.EQ(null.StringFrom(req.Email))).One(passdb.StdConn)
		if user != nil {
			user.GoogleID = null.StringFrom(req.GoogleID)
			user.Verified = true
			loginReq.User = user

			_, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.GoogleID, boiler.UserColumns.Verified))
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to add google id to user with existing gmail")
				return err
			}
		} else {
			// Return request back to user for signup
			newUserResponse := struct {
				GoogleLoginRequest
				NewUser bool `json:"new_user"`
			}{
				GoogleLoginRequest: *req,
				NewUser:            true,
			}
			b, err := json.Marshal(newUserResponse)
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to encode response to json")
				return err
			}
			_, err = w.Write(b)
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to write response to user")
				return err
			}
			return nil

		}

	}

	if loginReq.User != nil {
		return api.FingerprintAndIssueToken(w, r, loginReq)
	}
	L.Error().Err(err).Msg("invalid google credentials provided")
	return err
}

func (api *API) TFAVerifyHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &TFAVerifyRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode 2fa verify request")
		return http.StatusBadRequest, err
	}

	// Get user from token
	err = api.TFAVerify(req, w, r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusOK, nil
}

func (api *API) TFAVerify(req *TFAVerifyRequest, w http.ResponseWriter, r *http.Request) error {
	// Get user from token
	// OR verify passcode from user id
	// If user is logged in, user id is passed from request
	// If user is not logged in, token is passed from request
	userID := req.UserID
	if userID == "" {
		uid, _, err := api.UserFromToken(w, r, req.Token)
		if err != nil {
			return err
		}
		userID = uid
	}

	user, err := users.ID(userID)
	if err != nil {
		return err
	}
	// Check if there is a passcode and verify it
	if req.Passcode != "" {
		err := users.VerifyTFA(user.TwoFactorAuthenticationSecret, req.Passcode)
		if err != nil {
			return err
		}
	} else if req.RecoveryCode != "" {
		// Check if there is a recovery code and verify it
		err := users.VerifyTFARecovery(req.RecoveryCode)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("code is missing")
	}

	// Issue login token to user
	// Only if jwt token was provided
	if req.Token != "" {
		loginReq := &FingerprintTokenRequest{
			User:        &user.User,
			RedirectURL: req.RedirectURL,
			Tenant:      req.Tenant,
			Fingerprint: req.Fingerprint,
			Pass2FA:     true,
		}
		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			return err
		}
	} else {
		b, err := json.Marshal(user)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to convert user to json 2fa response")
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write 2fa response to user")
		}
	}

	return nil
}

func (api *API) FacebookLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FacebookLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode facebook login request")
		return http.StatusBadRequest, err
	}

	err = api.FacebookLogin(req, w, r)
	if err != nil {
		return http.StatusBadRequest, err
	}
	return http.StatusOK, nil
}

func (api *API) FacebookLogin(req *FacebookLoginRequest, w http.ResponseWriter, r *http.Request) error {
	// Check if there are any existing users associated with the email address
	user, err := users.FacebookID(req.FacebookID)
	if err != nil && errors.Is(sql.ErrNoRows, err) {
		// Return request back to user for signup
		b, err := json.Marshal(req)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return err
		}
		return nil
	} else if err != nil {
		passlog.L.Error().Err(err).Msg("No user found with given facebook id")
		return err
	}
	loginReq := &FingerprintTokenRequest{
		User:        user,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
	}
	err = api.FingerprintAndIssueToken(w, r, loginReq)

	return err
}

type TwitterAuthResponse struct {
	UserID *string `json:"user_id"`
}

type AddTwitterResponse struct {
	Error string      `json:"error"`
	User  *types.User `json:"user"`
}

// The TwitterAuth endpoint kicks off the OAuth 1.0a flow
func (api *API) TwitterAuth(w http.ResponseWriter, r *http.Request) (int, error) {
	oauthVerifier := r.URL.Query().Get("oauth_verifier")
	oauthCallback := r.URL.Query().Get("oauth_callback")
	oauthToken := r.URL.Query().Get("oauth_token")
	redirect := r.URL.Query().Get("redirect")
	addTwitter := r.URL.Query().Get("add")
	tenant := r.URL.Query().Get("tenant")

	if redirect == "" && oauthVerifier != "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("no redirect provided"))
	}

	if oauthVerifier != "" {
		params := url.Values{}
		params.Set("oauth_token", oauthToken)
		params.Set("oauth_verifier", oauthVerifier)
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/oauth/access_token?%s", params.Encode()), nil)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to get twitter access token")
			return http.StatusInternalServerError, terror.Error(err)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to request for twitter access token")
			return http.StatusInternalServerError, terror.Error(err)
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to get body from twitter response")
			return http.StatusInternalServerError, terror.Error(err)
		}
		resp := &AuthTwitterResponse{}
		values := strings.Split(string(body), "&")
		for _, v := range values {
			pair := strings.Split(v, "=")
			switch pair[0] {
			case "oauth_token":
				resp.OauthToken = pair[1]
			case "oauth_token_secret":
				resp.OauthTokenSecret = pair[1]
			case "user_id":
				resp.UserID = pair[1]
			}
		}

		// Check if user exist
		user, err := users.TwitterID(resp.UserID)

		// Add twitter user handler
		if addTwitter != "" {
			return api.AddTwitterUser(w, r, redirect, user, resp, addTwitter)
		}

		if err != nil && errors.Is(sql.ErrNoRows, err) {
			http.Redirect(w, r, fmt.Sprintf("%s?id=%s", redirect, resp.UserID), http.StatusSeeOther)
			return http.StatusOK, nil
		}
		loginReq := &FingerprintTokenRequest{
			User:        user,
			RedirectURL: redirect,
			Tenant:      tenant,
			IsTwitter:   true,
		}
		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			return http.StatusBadRequest, err
		}
		http.Redirect(w, r, fmt.Sprintf("%s?login=ok", redirect), http.StatusSeeOther)

		return http.StatusOK, nil
	}

	oauthConfig := oauth1.Config{
		ConsumerKey:    api.Twitter.APIKey,
		ConsumerSecret: api.Twitter.APISecret,
		CallbackURL:    oauthCallback,
		Endpoint:       twitter.AuthorizeEndpoint,
	}

	requestToken, _, err := oauthConfig.RequestToken()
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to get oauth token from")
		return http.StatusInternalServerError, terror.Error(err)
	}

	http.Redirect(w, r, fmt.Sprintf("https://api.twitter.com/oauth/authorize?oauth_token=%s", requestToken), http.StatusSeeOther)
	return http.StatusOK, nil
}

type AuthTwitterResponse struct {
	OauthToken       string
	OauthTokenSecret string
	UserID           string
}

func (api *API) AddTwitterUser(w http.ResponseWriter, r *http.Request, redirect string, user *boiler.User, resp *AuthTwitterResponse, addTwitter string) (int, error) {
	payload := &AddTwitterResponse{}
	URI := fmt.Sprintf("/user/%s", addTwitter)

	// Redirect to loading page
	http.Redirect(w, r, redirect, http.StatusSeeOther)

	if user != nil {
		payload.Error = "Twitter account already registered with a different user"
		ws.PublishMessage(URI, HubKeyUserAddTwitter, payload)
		return http.StatusOK, nil

	} else {
		// Check if user exist
		user, err := users.ID(addTwitter)
		if err != nil {
			payload.Error = "User ID does not exist"
			ws.PublishMessage(URI, HubKeyUserAddTwitter, payload)
			return http.StatusOK, nil
		}
		// Activity tracking
		var oldUser types.User = *user

		// Update user's Twitter ID
		user.TwitterID = null.StringFrom(resp.UserID)
		_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwitterID))
		if err != nil {
			payload.Error = "Unable to update user"
			ws.PublishMessage(URI, HubKeyUserAddTwitter, payload)
			passlog.L.Error().Err(err).Msg("unable to add user's twitter id")
			return http.StatusInternalServerError, terror.Error(err)
		}

		// Record user activity
		api.RecordUserActivity(context.Background(),
			user.ID,
			"Added Twitter account to User",
			types.ObjectTypeUser,
			helpers.StringPointer(user.ID),
			&user.Username,
			helpers.StringPointer(user.FirstName.String+" "+user.LastName.String),
			&types.UserActivityChangeData{
				Name: db.TableNames.Users,
				From: oldUser,
				To:   user,
			},
		)

		payload.User = user
		ws.PublishMessage(URI, HubKeyUserAddTwitter, payload)
		return http.StatusOK, nil
	}
}

type FingerprintTokenRequest struct {
	User        *boiler.User
	Pass2FA     bool
	RedirectURL string
	Tenant      string
	Fingerprint *users.Fingerprint
	IsTwitter   bool
}

func (api *API) FingerprintAndIssueToken(w http.ResponseWriter, r *http.Request, req *FingerprintTokenRequest) error {
	// Check if tenant is provided for external 2FA logins except for twitter since redirect is always provided
	if req.User.TwoFactorAuthenticationIsSet && req.RedirectURL != "" && !req.Pass2FA && req.Tenant == "" && !req.IsTwitter {
		err := fmt.Errorf("tenant missing for external 2fa login")
		passlog.L.Error().Err(err).Msg(err.Error())
		return err
	}
	if req.User == nil {
		err := fmt.Errorf("user does not exist in issuing token")
		passlog.L.Error().Err(err).Msg(err.Error())
		return err
	}
	// Dont create issue token and tell front-end to start 2FA verification with JWT
	if req.User.TwoFactorAuthenticationIsSet && !req.Pass2FA {
		// Generate jwt with user id
		config := &TokenConfig{
			Encrypted: true,
			Key:       api.TokenEncryptionKey,
			Device:    r.UserAgent(),
			Action:    "verify 2fa",
			User:      req.User,
		}

		_, _, token, err := token(api, config, false, api.TokenExpirationDays)
		if err != nil {
			return err
		}

		origin := r.Header.Get("origin")

		// IF redirect from Twitter Auth
		if origin == "" {
			origin = strings.ReplaceAll(req.RedirectURL, "/twitter-redirect", "")
		}

		// Redirect to 2fa
		if req.RedirectURL != "" {
			http.Redirect(w, r, fmt.Sprintf("%s/tfa/check?token=%s&redirectURL=%s&tenant=%s", origin, token, req.RedirectURL, req.Tenant), http.StatusSeeOther)
			return nil
		}

		// Send response to user and pass token to redirect to 2fa

		b, err := json.Marshal(token)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return err
		}

		return nil
	}

	// Fingerprint user
	if req.Fingerprint != nil {
		err := api.DoFingerprintUpsert(*req.Fingerprint, req.User.ID)
		if err != nil {
			return err
		}
	}
	u, _, token, err := api.IssueToken(&TokenConfig{
		Encrypted: true,
		Key:       api.TokenEncryptionKey,
		Device:    r.UserAgent(),
		Action:    "login",
		User:      req.User,
	})
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to issue token")
		return err
	}
	if req.User.DeletedAt.Valid {
		return err
	}
	err = api.WriteCookie(w, r, token)
	if err != nil {
		return err
	}

	if !req.IsTwitter {

		b, err := json.Marshal(u)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return err
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return err
		}
	}

	return nil
}

type TokenConfig struct {
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
func token(api *API, config *TokenConfig, isIssueToken bool, expireInDays int) (*types.User, uuid.UUID, string, error) {
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
		&user.User,
		config.Device,
		config.Action,
		expireInDays)

	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to generate jwt token")
		return nil, uuid.Nil, "", terror.Error(err, errMsg)
	}
	jwtSigned, err := sign(jwt, config.Encrypted, config.Key)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to sign jwt")
		return nil, uuid.Nil, "", terror.Error(err, "unable to sign jwt")
	}
	token := base64.StdEncoding.EncodeToString(jwtSigned)

	if isIssueToken {
		err = tokens.Save(token, api.TokenExpirationDays, api.TokenEncryptionKey)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to save issue token")
			return nil, uuid.Nil, "", terror.Error(err, "unable to save jwt")
		}
	}
	return user, tokenID, token, nil
}

func (api *API) ResetPasswordToken(config *TokenConfig) (*types.User, uuid.UUID, string, error) {
	return token(api, config, false, api.TokenExpirationDays)
}

func (api *API) VerifyEmailToken(config *TokenConfig) (*types.User, uuid.UUID, string, error) {
	return token(api, config, false, api.TokenExpirationDays)
}

func (api *API) IssueToken(config *TokenConfig) (*types.User, uuid.UUID, string, error) {
	return token(api, config, true, api.TokenExpirationDays)
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
	if user != nil {
		i, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Nonce))
		if err != nil {
			return "", err
		}

		if i == 0 {
			return "", terror.Error(fmt.Errorf("nonce could not be updated"))
		}
	}

	return newNonce, nil
}

func (api *API) GetNonce(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := r.URL.Query().Get("public-address")
	userID := r.URL.Query().Get("user-id")

	L := passlog.L.With().Str("publicAddress", publicAddress).Str("userID", userID).Logger()

	if publicAddress == "" && userID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing public address or user id"))
	}
	if publicAddress != "" && common.IsHexAddress(publicAddress) {
		// Take public address Hex to address(Make it a checksum mixed case address) convert back to Hex for string of checksum
		commonAddr := common.HexToAddress(publicAddress)
		user, err := users.PublicAddress(commonAddr)

		if err != nil && !errors.Is(sql.ErrNoRows, err) {
			L.Error().Err(err).Msg("Error searching for user with public address when getting nonce")
			return http.StatusInternalServerError, err
		}

		// user can be nil
		newNonce, err := api.NewNonce(&user.User)
		if err != nil {
			L.Error().Err(err).Msg("no nonce")
			return http.StatusBadRequest, err
		}

		resp := &GetNonceResponse{
			Nonce: newNonce,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			L.Error().Err(err).Msg("json failed")
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

	user, err := types.UserFromBoil(resp.User)
	if err != nil {
		return nil, fmt.Errorf("failed to identify user: %w", err)
	}

	return &LoginResponse{user, req.Token, false}, nil
}

func (api *API) AuthCheckHandler(w http.ResponseWriter, r *http.Request) (int, error) {

	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		// check whether token is attached
		token := r.URL.Query().Get("token")
		if token == "" {
			passlog.L.Warn().Msg("No token found")
			return http.StatusBadRequest, terror.Warn(fmt.Errorf("no cookie and token are provided"), "User are not signed in.")
		}

		// check user from token
		resp, err := api.TokenLogin(token, "")
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to authentication")
		}

		// write cookie
		err = api.WriteCookie(w, r, token)
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
	err = api.DeleteCookie(w, r)
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

	err = api.WriteCookie(w, r, req.Token)
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
		readErr := err
		if errors.Is(err, tokens.ErrTokenExpired) {
			tknUuid, err := tokens.TokenID(token)
			if err != nil {
				return nil, err
			}
			err = tokens.Remove(tknUuid)
			if err != nil {
				return nil, err
			}
			return nil, terror.Warn(readErr, "Session has expired, please log in again.")
		}
		return nil, readErr
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

	// shuffle bot id
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(resp.RedMountainBotIDs), func(i, j int) {
		resp.RedMountainBotIDs[i], resp.RedMountainBotIDs[j] = resp.RedMountainBotIDs[j], resp.RedMountainBotIDs[i]
	})
	rand.Shuffle(len(resp.BostonBotIDs), func(i, j int) {
		resp.BostonBotIDs[i], resp.BostonBotIDs[j] = resp.BostonBotIDs[j], resp.BostonBotIDs[i]
	})
	rand.Shuffle(len(resp.ZaibatsuBotIDs), func(i, j int) {
		resp.ZaibatsuBotIDs[i], resp.ZaibatsuBotIDs[j] = resp.ZaibatsuBotIDs[j], resp.ZaibatsuBotIDs[i]
	})

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
		http.Error(w, "auth fail", http.StatusUnauthorized)
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

	_, _, token, err := api.IssueToken(&TokenConfig{
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
