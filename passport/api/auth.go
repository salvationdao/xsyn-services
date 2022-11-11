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
	"google.golang.org/api/idtoken"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofrs/uuid"
	twitch_jwt "github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
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
	CaptchaToken    string               `json:"captcha_token"`
}

type WalletLoginRequest struct {
	RedirectURL   string             `json:"redirect_url"`
	Tenant        string             `json:"tenant"`
	PublicAddress string             `json:"public_address"`
	Signature     string             `json:"signature"`
	Nonce         string             `json:"nonce"`
	SessionID     hub.SessionID      `json:"session_id"`
	Fingerprint   *users.Fingerprint `json:"fingerprint"`
	AuthType      string             `json:"auth_type"`
	Username      string             `json:"username"`
	CaptchaToken  string             `json:"captcha_token"`
}

type EmailSignupVerifyRequest struct {
	RedirectURL  string `json:"redirect_url"`
	Tenant       string `json:"tenant"`
	Email        string `json:"email"`
	CaptchaToken string `json:"captcha_token"`
}

type EmailLoginRequest struct {
	RedirectURL      string             `json:"redirect_url"`
	Tenant           string             `json:"tenant"`
	Email            string             `json:"email"`
	Password         string             `json:"password"`
	SessionID        hub.SessionID      `json:"session_id"`
	Fingerprint      *users.Fingerprint `json:"fingerprint"`
	AuthType         string             `json:"auth_type"`
	Username         string             `json:"username"`
	Token            string             `json:"token"`
	AcceptsMarketing string             `json:"accepts_marketing"`
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
	Password    string             `json:"password"`
	TokenID     string             `json:"id"`
	Token       string             `json:"token"`
	NewPassword string             `json:"new_password"`
	SessionID   hub.SessionID      `json:"session_id"`
	Fingerprint *users.Fingerprint `json:"fingerprint"`
}

type GoogleLoginRequest struct {
	RedirectURL  string             `json:"redirect_url"`
	Tenant       string             `json:"tenant"`
	GoogleToken  string             `json:"google_token"`
	SessionID    hub.SessionID      `json:"session_id"`
	Fingerprint  *users.Fingerprint `json:"fingerprint"`
	AuthType     string             `json:"auth_type"`
	Username     string             `json:"username"`
	CaptchaToken string             `json:"captcha_token"`
}

type FacebookLoginRequest struct {
	RedirectURL   string             `json:"redirect_url"`
	Tenant        string             `json:"tenant"`
	FacebookToken string             `json:"facebook_token"`
	SessionID     hub.SessionID      `json:"session_id"`
	Fingerprint   *users.Fingerprint `json:"fingerprint"`
	AuthType      string             `json:"auth_type"`
	Username      string             `json:"username"`
	CaptchaToken  string             `json:"captcha_token"`
}

type TwitterSignupRequest struct {
	RedirectURL  string             `json:"redirect_url"`
	Tenant       string             `json:"tenant"`
	TwitterToken string             `json:"twitter_token"`
	SessionID    hub.SessionID      `json:"session_id"`
	Fingerprint  *users.Fingerprint `json:"fingerprint"`
	Username     string             `json:"username"`
	CaptchaToken string             `json:"captcha_token"`
}

type TFAVerifyRequest struct {
	RedirectURL  string             `json:"redirect_url"`
	Tenant       string             `json:"tenant"`
	Token        string             `json:"token"`
	Passcode     string             `json:"passcode"`
	RecoveryCode string             `json:"recovery_code"`
	SessionID    hub.SessionID      `json:"session_id"`
	Fingerprint  *users.Fingerprint `json:"fingerprint"`
	IsVerified   bool               `json:"is_verified"`
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

// WriteCookie writes cookie on host domain
func (api *API) WriteCookie(w http.ResponseWriter, r *http.Request, token string) error {
	b64, err := api.Cookie.EncryptToBase64(token)
	if err != nil {
		passlog.L.Error().Msg("invalid token when writing cookie, unable to encrypt to base64")
		return terror.Error(err, "Invalid token, unable to encrypt to base64.")
	}

	// get domain
	d := domain(r.Host)
	if d == "" {
		passlog.L.Warn().Msg("Cookie's domain not found")
		return terror.Error(fmt.Errorf("domain not found"), "failed to write cookie")
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

// Gets domain from subdomain if host is a subdomain
func domain(host string) string {
	parts := strings.Split(host, ".")

	if len(parts) < 2 {
		return ""
	}
	//this is rigid as fuck
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

// remove cookie on domain
func (api *API) DeleteCookie(w http.ResponseWriter, r *http.Request) error {
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
	passlog.L.Info().Msg("deleting cookie from given domain")
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
	passlog.L.Info().Msg("deleting cookie from current site")
	http.SetCookie(w, cookie)
	return nil
}

// Handles login from non-passport sites, use HTML form.submit() to send request.
//
// Writes the external cookie by being on the same host.
func (api *API) ExternalLoginHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		passlog.L.Warn().Err(err).Msg("suspicious behaviour on external login form")
		errDataDog := DatadogTracer.HttpFinishSpan(r.Context(), http.StatusBadRequest, terror.Error(err, "suspicious behaviour on external login form"))
		if errDataDog != nil {
			passlog.L.Error().Err(errDataDog).Msg("data dog failed")
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authType := r.Form.Get("auth_type")
	redir := r.Form.Get("redirect_url")

	if redir == "" {
		passlog.L.Warn().Msg("no redirect url provided in external login")
		errDataDog := DatadogTracer.HttpFinishSpan(r.Context(), http.StatusBadRequest, fmt.Errorf("Missing redirect url on external login"))
		if errDataDog != nil {
			passlog.L.Error().Err(errDataDog).Msg("data dog failed")
		}
		http.Error(w, "Missing redirect url on external login", http.StatusBadRequest)
		return
	}

	switch authType {
	case "wallet":
		req := &WalletLoginRequest{
			RedirectURL:   redir,
			Tenant:        r.Form.Get("tenant"),
			PublicAddress: r.Form.Get("public_address"),
			Signature:     r.Form.Get("signature"),
		}
		commonAddr := common.HexToAddress(req.PublicAddress)
		user, err := users.PublicAddress(commonAddr)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "User does not exist")
			return
		}

		err = api.VerifySignature(req.Signature, user.Nonce.String, commonAddr)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Invalid signature provided from wallet")
			return
		}

		// Login user
		loginReq := &FingerprintTokenRequest{
			User:        &user.User,
			Fingerprint: req.Fingerprint,
			RedirectURL: redir,
			Tenant:      req.Tenant,
		}

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Unable to issue token")
			return
		}

	case "email":
		req := &EmailLoginRequest{
			RedirectURL: redir,
			Tenant:      r.Form.Get("tenant"),
			Email:       r.Form.Get("email"),
			Password:    r.Form.Get("password"),
			Token:       r.Form.Get("token"),
		}
		//Check if email exist with password
		user, err := users.EmailPassword(req.Email, req.Password)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Invalid email or password")
			return
		}

		// Check if logging in from password reset
		// If so, pass 2fa
		var pass2fa bool
		if req.Token != "" {
			userID, err := api.ReadUserIDJWT(req.Token)
			if err != nil || userID != user.ID {
				passlog.L.Error().Err(err).Msg("Invalid token to login from external.")
				externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Invalid email or password")
				return
			}
			pass2fa = true
		}

		loginReq := &FingerprintTokenRequest{
			User:        &user.User,
			Fingerprint: req.Fingerprint,
			RedirectURL: redir,
			Tenant:      req.Tenant,
			Pass2FA:     pass2fa,
		}

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Unable to issue token")
			return
		}
	case "facebook":
		req := &FacebookLoginRequest{
			RedirectURL:   redir,
			Tenant:        r.Form.Get("tenant"),
			FacebookToken: r.Form.Get("facebook_token"),
		}
		facebookDetails, err := api.FacebookToken(req.FacebookToken)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Invalid facebook token")
			return
		}
		user, err := users.FacebookID(facebookDetails.FacebookID)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Unable to find facebook user")
			return
		}

		// Login user after register
		loginReq := &FingerprintTokenRequest{
			User:        user,
			Fingerprint: req.Fingerprint,
			RedirectURL: redir,
			Tenant:      req.Tenant,
		}

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Unable to issue token")
			return
		}

		http.Redirect(w, r, redir, http.StatusSeeOther)

	case "google":
		req := &GoogleLoginRequest{
			RedirectURL: redir,
			Tenant:      r.Form.Get("tenant"),
			GoogleToken: r.Form.Get("google_token"),
			Username:    r.Form.Get("username"),
		}

		googleDetails, err := api.GoogleToken(req.GoogleToken)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Invalid google token provided")
			return
		}
		user, err := users.GoogleID(googleDetails.GoogleID)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Unable to find google user")
			return
		}

		loginReq := &FingerprintTokenRequest{
			User:        user,
			Fingerprint: req.Fingerprint,
			RedirectURL: redir,
			Tenant:      req.Tenant,
		}

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Unable to issue token")
			return
		}

		http.Redirect(w, r, redir, http.StatusSeeOther)
	case "cookie":
		req := &struct {
			Tenant string
		}{
			Tenant: r.Form.Get("tenant"),
		}

		err := externalLoginCheck(api, w, r)
		if err != nil {
			externalErrorHandler(w, r, err, "/external/login", req.Tenant, redir, "Unable to login user from cookie")
			return
		}

	case "twitter":
		req := &TwitterSignupRequest{
			RedirectURL:  redir,
			Tenant:       r.Form.Get("tenant"),
			TwitterToken: r.Form.Get("twitter_token"),
		}

		id, err := api.ReadUserIDJWT(req.TwitterToken)
		if err != nil {
			externalErrorHandler(w, r, err, "/signup", req.Tenant, redir, "Unable to read user from token")
			return
		}
		userID := id

		var isSignup bool
		signupCheckVal, err := api.ReadKeyJWT(req.TwitterToken, "twitter-signup")
		if err != nil && !errors.Is(err, ErrJWTKeyNotFound) {
			externalErrorHandler(w, r, err, "/signup", req.Tenant, redir, "Unable to read user from token")
			return
		}
		isSignup = signupCheckVal == "true"
		if isSignup {
			twitterUser, err := users.TwitterID(id)
			if err != nil {
				externalErrorHandler(w, r, err, "/signup", req.Tenant, redir, "Unable to locate user.")
				return
			}
			userID = twitterUser.ID
		}
		user, err := users.ID(userID)
		if err != nil {
			externalErrorHandler(w, r, err, "/signup", req.Tenant, redir, "Unable to locate user.")
			return
		}
		loginReq := &FingerprintTokenRequest{
			User:        &user.User,
			Fingerprint: req.Fingerprint,
			RedirectURL: redir,
			Tenant:      req.Tenant,
		}

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			externalErrorHandler(w, r, err, "/signup", req.Tenant, redir, "Unable to issue token")
			return
		}

		http.Redirect(w, r, redir, http.StatusSeeOther)

	case "tfa":
		req := &TFAVerifyRequest{
			RedirectURL:  redir,
			Tenant:       r.Form.Get("tenant"),
			Token:        r.Form.Get("token"),
			Passcode:     r.Form.Get("passcode"),
			RecoveryCode: r.Form.Get("recovery_code"),
		}

		user, err := api.TFAVerify(req, w, r)
		if err != nil {
			externalErrorHandler(w, r, err, "/tfa/check", req.Tenant, redir, "Two factor authentication failed.")
			return
		}

		loginReq := &FingerprintTokenRequest{
			User:        &user.User,
			Fingerprint: req.Fingerprint,
			RedirectURL: redir,
			Tenant:      req.Tenant,
			Pass2FA:     true,
		}

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			externalErrorHandler(w, r, err, "/tfa/check", req.Tenant, redir, "Unable to issue token")
			return
		}

	}

	http.Redirect(w, r, redir, http.StatusSeeOther)

}

func externalErrorHandler(w http.ResponseWriter, r *http.Request, err error, page string, tenant string, redir string, msg string) {
	passlog.L.Error().Err(err).Str("From", "External Login").Msg(msg)
	errDataDog := DatadogTracer.HttpFinishSpan(r.Context(), http.StatusInternalServerError, err)
	if errDataDog != nil {
		passlog.L.Error().Err(errDataDog).Msg("data dog failed")
	}
	http.Redirect(w, r, fmt.Sprintf("%s%s?tenant=%s&redirectURL=%s&err=%s", r.Header.Get("origin"), page, tenant, redir, terror.Error(err, msg)), http.StatusSeeOther)
}

func externalLoginCheck(api *API, w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		passlog.L.Error().Err(err).Str("From", "External Login").Msg("Unable to read cookie")
		return err
	}

	var token string
	if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
		passlog.L.Error().Err(err).Str("From", "External Login").Msg("Unable to decrypt token from cookie")
		return err
	}

	// check user from token
	_, err = api.TokenLogin(token, "")
	if err != nil {
		passlog.L.Error().Err(err).Str("From", "External Login").Msg("Unable to login from token")
		return err
	}

	// write cookie on domain
	err = api.WriteCookie(w, r, token)
	if err != nil {
		passlog.L.Error().Err(err).Str("From", "External Login").Msg("Unable to wite cookie on domain")
		return err
	}

	return nil

}

// Sends a code to user email to be verified.
func (api *API) EmailSignupVerifyHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &EmailSignupVerifyRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode request.")
	}
	err = api.EmailSignupVerify(req, w, r)

	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}

	return http.StatusCreated, nil
}

// Generate one time code and send to user's email
func (api *API) EmailSignupVerify(req *EmailSignupVerifyRequest, w http.ResponseWriter, r *http.Request) error {
	lowerEmail := strings.ToLower(req.Email)

	// Verify user passed captcha test
	if req.CaptchaToken == "" {
		return terror.Error(errors.New("captcha token missing"), "Failed to complete captcha verification.")
	}
	err := api.captcha.verify(req.CaptchaToken)
	if err != nil {
		return terror.Error(err, "Failed to complete captcha verification.")
	}

	// Check if there are any existing users associated with the email address
	user, _ := users.Email(lowerEmail)

	if user != nil {
		return terror.Error(errors.New("email already used"), "Email is already used by a different user.")
	}

	token, err := api.OneTimeVerification()
	if err != nil || token == "" {
		err := fmt.Errorf("unable to generate verification code")
		passlog.L.Error().Err(err).Msg(err.Error())
		return terror.Error(err, "Failed to generate verification code.")
	}
	code, err := api.ReadCodeJWT(token)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to get verify code from token")
		return terror.Error(err, "Unable to get verify code from token")
	}

	err = api.Mailer.SendSignupEmail(context.Background(), lowerEmail, code)
	if err != nil {
		return terror.Error(err, "Unable to send signup email")
	}
	resp := &struct {
		Token string `json:"token"`
	}{
		Token: token,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to encode response to json")
		return terror.Error(err, "Unable to generate response token")
	}
	_, err = w.Write(b)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to write response to user")
		return terror.Error(err, "Unable to write response token")
	}
	return nil
}

// Update username of user except for emails which create account after password is set.
func (api *API) SignupHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &SignupRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode request")
	}
	username := req.Username
	// Check if username is valid
	err = helpers.IsValidUsername(username)
	if err != nil {
		return http.StatusInternalServerError, err // returns terror error already
	}

	usernameTaken, err := users.UsernameExist(username)
	redirectURL := ""
	if err != nil || usernameTaken {
		err := fmt.Errorf("Username is already taken")
		return http.StatusInternalServerError, terror.Error(err, "Username is already taken, please try a different username.")
	}

	authType := req.AuthType
	if authType == "" {
		passlog.L.Error().Err(err).Msg("auth type is missing in user signup")
		return http.StatusInternalServerError, terror.Error(err, "Invalid signup request provided.")
	}
	u := &types.User{}

	switch authType {
	case "wallet":
		// Check user exist
		commonAddr := common.HexToAddress(req.WalletRequest.PublicAddress)
		user, err := users.PublicAddress(commonAddr)
		if user != nil {
			passlog.L.Error().Err(err).Msg("user already exist on signup")
			return http.StatusBadRequest, terror.Error(err, "User exist already.")
		}
		if errors.Is(err, sql.ErrNoRows) {
			if req.CaptchaToken == "" {
				err := fmt.Errorf("captcha token missing")
				return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
			}
			err := api.captcha.verify(req.CaptchaToken)
			if err != nil {
				return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
			}

			// Signup user but dont log them before username is provided
			// If user does not exist, create new user with their username set to their MetaMask public address
			user, err = users.UserCreator("",
				"",
				username,
				"",
				"",
				"",
				"",
				"",
				"",
				"",
				commonAddr,
				"",
				false,
				api.Environment,
			)
			if err != nil {
				return http.StatusInternalServerError, terror.Error(err, "Unable to create user with wallet.")
			}

			// update nonce value
			user.Nonce = null.StringFrom(req.WalletRequest.Nonce)
			_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Nonce))
			if err != nil {
				return http.StatusInternalServerError, terror.Error(err, "Unable to update user nonce.")
			}
		} else if err != nil {
			err := fmt.Errorf("Failed to signup user.")
			return http.StatusBadRequest, terror.Error(err, "User with wallet address does not exist.")
		}

		// Redeclare u variable
		u = user

		redirectURL = req.WalletRequest.RedirectURL
	case "email":
		// Check no user with email exist
		user, err := users.Email(req.EmailRequest.Email)
		if user != nil {
			passlog.L.Error().Err(err).Msg("user already exist on signup")
			return http.StatusBadRequest, terror.Error(err, "User exist already.")
		}
		if err != nil && errors.Is(sql.ErrNoRows, err) {
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator(
				"",
				"",
				username,
				req.EmailRequest.Email,
				"",
				"",
				"",
				"",
				"",
				"",
				commonAddress,
				req.EmailRequest.Password,
				req.EmailRequest.AcceptsMarketing == "true",
				api.Environment,
			)
			if err != nil {
				if err.Error() != "password does not meet requirements" {
					passlog.L.Error().Err(err).Msg("unable to create user with email and password")
				}
				return http.StatusInternalServerError, terror.Error(err)
			}
		} else if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to signup with email.")
		}
		redirectURL = req.EmailRequest.RedirectURL
	case "facebook":
		// Check user facebook exist
		facebookDetails, err := api.FacebookToken(req.FacebookRequest.FacebookToken)
		if err != nil {
			passlog.L.Error().Err(err).Msg("user provided invalid facebook token")
			return http.StatusBadRequest, terror.Error(err, "Invalid facebook token provided.")
		}

		user, err := users.FacebookID(facebookDetails.FacebookID)
		if user != nil {
			passlog.L.Error().Err(err).Msg("user already exist on signup")
			return http.StatusBadRequest, terror.Error(err, "User exist already.")
		}
		if errors.Is(err, sql.ErrNoRows) {
			if req.CaptchaToken == "" {
				err := fmt.Errorf("captcha token missing")
				return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
			}
			err := api.captcha.verify(req.CaptchaToken)
			if err != nil {
				return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
			}

			// Create user with default username
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator("",
				"",
				username,
				"",
				facebookDetails.FacebookID,
				"",
				"",
				"",
				"",
				"",
				commonAddress,
				"",
				false,
				api.Environment,
			)
			if err != nil {
				return http.StatusInternalServerError, terror.Error(err, "Failed to create new user with facebook.")
			}
		} else if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to get user with facebook account during signup.")
		}

		redirectURL = req.FacebookRequest.RedirectURL
	case "google":
		googleDetails, err := api.GoogleToken(req.GoogleRequest.GoogleToken)
		if err != nil {
			passlog.L.Error().Err(err).Msg("user provided invalid google token")
			return http.StatusBadRequest, terror.Error(err, "Invalid google token provided.")
		}
		// Check google id exist
		user, err := users.GoogleID(googleDetails.GoogleID)
		if user != nil {
			passlog.L.Error().Err(err).Msg("user already exist on signup")
			return http.StatusBadRequest, terror.Error(err, "User exist already.")
		}
		if errors.Is(err, sql.ErrNoRows) {
			if req.CaptchaToken == "" {
				passlog.L.Error().Err(err).Msg("Captcha token not provided")
				return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
			}
			err := api.captcha.verify(req.CaptchaToken)
			if err != nil {
				return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
			}
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator("",
				"",
				username,
				googleDetails.Email,
				"",
				googleDetails.GoogleID,
				"",
				"",
				"",
				"",
				commonAddress,
				"",
				false,
				api.Environment,
			)
			if err != nil {
				return http.StatusInternalServerError, terror.Error(err, "Failed to create new user with google account")
			}
		} else if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to get user with google account during signup.")
		}

		redirectURL = req.GoogleRequest.RedirectURL
	case "twitter":
		if req.CaptchaToken == "" {
			err := fmt.Errorf("captcha token missing")
			return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
		}
		err := api.captcha.verify(req.CaptchaToken)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Failed to complete captcha verification.")
		}

		// Twitter Token will BE JWT
		twitterID, err := api.ReadUserIDJWT(req.TwitterRequest.TwitterToken)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to read user jwt")
			return http.StatusBadRequest, terror.Error(err, "Invalid twitter token provided.")
		}

		// Check if user exist already
		user, err := users.TwitterID(twitterID)
		if user != nil {
			passlog.L.Error().Err(err).Msg("user already exist on signup")
			return http.StatusBadRequest, terror.Error(err, "User exist already.")
		}

		// Create user
		if errors.Is(err, sql.ErrNoRows) {
			commonAddress := common.HexToAddress("")
			u, err = users.UserCreator("",
				"",
				username,
				"",
				"",
				"",
				"",
				twitterID,
				"",
				"",
				commonAddress,
				"",
				false,
				api.Environment,
			)
			if err != nil {
				return http.StatusInternalServerError, terror.Error(err, "Failed to create user with twitter.")
			}
		} else if err != nil {
			return http.StatusInternalServerError, terror.Error(err, "Failed to create user with twitter.")
		}

		redirectURL = req.TwitterRequest.RedirectURL
	}

	if u == nil {
		passlog.L.Error().Err(err).Msg("unable to create user, no user passed through")
		return http.StatusInternalServerError, terror.Error(err, "Unable to update username during signup.")
	}

	// Login
	// Write cookie for passport
	loginReq := &FingerprintTokenRequest{
		User:        &u.User,
		Fingerprint: req.Fingerprint,
		RedirectURL: redirectURL,
	}

	err = api.FingerprintAndIssueToken(w, r, loginReq)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to issue a token for login.")
	}

	if redirectURL != "" {
		b, err := json.Marshal(u)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return http.StatusInternalServerError, terror.Error(err, "Unable to decode response to user.")
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return http.StatusInternalServerError, terror.Error(err, "Unable to write response to user.")
		}
	}

	return http.StatusCreated, nil

}

// Get user from jwt token
func (api *API) UserEmailFromToken(w http.ResponseWriter, r *http.Request, tokenBase64 string) (string, string, error) {
	errMsg := "Unable to create token for user request."
	tokenStr, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	// Decode token user with new email
	token, err := tokens.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	uID, ok := token.Get("user-id")
	if !ok {
		passlog.L.Error().Err(err).Msg("unable to get user id from token")
		return "", "", terror.Error(fmt.Errorf("Invalid token found"), errMsg)
	}
	email, ok := token.Get(openid.EmailKey)
	if !ok {
		passlog.L.Error().Err(err).Msg("unable to get email from token")
		return "", "", terror.Error(fmt.Errorf("Invalid token provided"), errMsg)
	}

	newEmail := email.(string)
	userID := uID.(string)

	return userID, newEmail, nil
}

// Get user from jwt token
func (api *API) UserFromToken(tokenBase64 string) (*types.User, error) {
	errMsg := "Unable to process token."
	tokenStr, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return nil, terror.Error(err, errMsg)
	}

	// Decode token user with new email
	token, err := tokens.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		return nil, terror.Error(err, errMsg)
	}

	uID, ok := token.Get("user-id")
	if !ok {
		passlog.L.Error().Err(err).Msg("unable to get user id from token")
		return nil, terror.Error(fmt.Errorf("Invalid token found"), errMsg)
	}

	userID := uID.(string)
	user, err := users.ID(userID)
	if err != nil {
		return nil, terror.Error(err, errMsg)
	}

	return user, nil
}

// Handler for EmailLogin
func (api *API) EmailLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &EmailLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode user request.")
	}

	err = api.EmailLogin(req, w, r)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Invalid user email or password. Please try again.")
	}

	return http.StatusOK, nil
}

// Handles email login with password
func (api *API) EmailLogin(req *EmailLoginRequest, w http.ResponseWriter, r *http.Request) error {
	user, err := users.EmailPassword(req.Email, req.Password)
	if err != nil {
		return terror.Error(err, "Invalid user email or password.")
	}

	loginReq := &FingerprintTokenRequest{
		User:        &user.User,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
	}
	err = api.FingerprintAndIssueToken(w, r, loginReq)

	// If external or new user signup
	if req.RedirectURL != "" && !user.TwoFactorAuthenticationIsSet {
		b, err := json.Marshal(req)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return terror.Error(err, "Unable to encode response.")
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return terror.Error(err, "Unable to write response back to user.")
		}
		return nil
	}

	if err != nil {
		return terror.Error(err, "Unable to issue a login token to user.")
	}
	return nil
}

type VerifyCodeRequest struct {
	Token string `json:"token"`
	Code  string `json:"code"`
}

// Handles code to be verified from jwt token
func (api *API) VerifyCodeHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &VerifyCodeRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode request.")
	}
	code, err := api.ReadCodeJWT(req.Token)
	if err != nil {
		err := fmt.Errorf("fail to read verification code")
		passlog.L.Error().Err(err).Msg(err.Error())
		return http.StatusBadRequest, terror.Error(err, "Invalid token provided.")
	}
	success := false
	if code == req.Code {
		success = true
	} else {
		err := fmt.Errorf("verify code does not match")
		passlog.L.Error().Err(err).Msg(err.Error())
		return http.StatusBadRequest, terror.Error(err, "Code does not match.")
	}
	resp := &struct {
		Success bool `json:"success"`
	}{
		Success: success,
	}
	b, err := json.Marshal(resp)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode response.")
	}
	httpStatus, err := w.Write(b)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to write response back to user.")
	}
	return httpStatus, nil

}

// Handles forgot password and verify email provided exist
func (api *API) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &ForgotPasswordRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode request.")
	}

	user, err := users.Email(req.Email)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "User with email does not exist.")
	}
	token, err := api.OneTimeToken(user.ID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to generate a forgot password link.")
	}

	err = api.Mailer.SendForgotPasswordEmail(context.Background(), user, token)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to send email to user.")

	}
	resp := &ForgotPasswordResponse{Message: "Success! An email has been sent to recover your account."}

	b, err := json.Marshal(resp.Message)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode response back to user.")
	}
	httpStatus, err := w.Write(b)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to write response back to user.")
	}
	return httpStatus, nil

}

// Handles resetting password from jwt token
func (api *API) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &PasswordUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode user request.")
	}

	userID, err := api.ReadUserIDJWT(req.Token)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to read user jwt")
		return http.StatusBadRequest, terror.Error(err, "Link has expired.")
	}
	user, err := users.ID(userID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to find user.")
	}

	statusCode, err := passwordReset(api, w, r, req, &user.User)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}
	// Send back login request to login external also
	resp := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		AuthType string `json:"auth_type"`
	}{
		Email:    user.Email.String,
		Password: req.NewPassword,
		AuthType: "email",
	}
	b, err := json.Marshal(resp)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to encode response to json")
		return http.StatusInternalServerError, terror.Error(err, "Unable to decode response to user.")
	}
	_, err = w.Write(b)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to write response to user")
		return http.StatusInternalServerError, terror.Error(err, "Unable to write response to user.")
	}

	return statusCode, nil
}

// Handles changes password with current password
func (api *API) ChangePasswordHandler(w http.ResponseWriter, r *http.Request, user *boiler.User) (int, error) {
	req := &PasswordUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode user request.")
	}

	userPassword, err := boiler.FindPasswordHash(passdb.StdConn, user.ID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "User does not have any existing passwords.")
	}

	// Check if current password is correct
	err = pCrypto.ComparePassword(userPassword.PasswordHash, req.Password)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Password does not match current password.")
	}

	return passwordReset(api, w, r, req, user)
}

// Setup new password for user thats logged in
func (api *API) NewPasswordHandler(w http.ResponseWriter, r *http.Request, user *boiler.User) (int, error) {
	req := &PasswordUpdateRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to decode request from user.")
	}

	// Check if user has password already
	passwordExist, err := boiler.PasswordHashExists(passdb.StdConn, user.ID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed create new password.")
	}

	if passwordExist {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("User already has a password"), "Failed create new password. User already has a password.")
	}

	return passwordReset(api, w, r, req, user)
}

// Handles password reset/update
func passwordReset(api *API, w http.ResponseWriter, r *http.Request, req *PasswordUpdateRequest, user *boiler.User) (int, error) {
	if user == nil {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("no user provided"), "Unable to process user request.")
	}
	// Check if new password is valid
	err := helpers.IsValidPassword(req.NewPassword)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Invalid password provided. Please try again.")
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
			return http.StatusInternalServerError, terror.Error(err, "Unable to register a password to user.")
		}
	} else {
		// Update password
		userPassword.PasswordHash = newPasswordHash
		_, err = userPassword.Update(passdb.StdConn, boil.Whitelist(boiler.PasswordHashColumns.PasswordHash))
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to update user password")
			return http.StatusInternalServerError, terror.Error(err, "Unable to user password.")
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
		return http.StatusInternalServerError, terror.Error(err, "Unable to delete all current sessions")
	}

	// Generate new token and login
	loginReq := &FingerprintTokenRequest{
		User:        user,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
		Pass2FA:     true,
	}
	err = api.FingerprintAndIssueToken(w, r, loginReq)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to issue a new login token.")
	}

	// Send message to users to logout
	URI := fmt.Sprintf("/user/%s", user.ID)
	passlog.L.Info().Str("Password update", "Logging out all issue token from user")
	ws.PublishMessage(URI, HubKeyUserInit, nil)

	return http.StatusCreated, nil
}

// Handles wallet login
func (api *API) WalletLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	errMsg := "Failed to authenticate user with wallet."
	req := &WalletLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode wallet login request")
		return http.StatusBadRequest, terror.Error(err, errMsg)
	}
	err = api.WalletLogin(req, w, r)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, errMsg)
	}

	return http.StatusCreated, nil
}

// Check if new user, then get nonce from request otherwise get from user nonce in db.
func (api *API) WalletLogin(req *WalletLoginRequest, w http.ResponseWriter, r *http.Request) error {
	// Take public address Hex to address(Make it a checksum mixed case address) convert back to Hex for string of checksum
	commonAddr := common.HexToAddress(req.PublicAddress)

	// Check if there are any existing users associated with the public address
	user, err := users.PublicAddress(commonAddr)

	// Check if its a new user signup
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// Verify new user signature nonce from request
		err = api.VerifySignature(req.Signature, req.Nonce, commonAddr)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to verify signature")
			return terror.Error(err, "Invalid signature provided.")
		}
	} else if err != nil {
		return terror.Error(err, "Failed to check if user with wallet address exists.")
	}

	if user != nil {
		// Verify wallet signature of current user using nonce from db
		err = api.VerifySignature(req.Signature, user.Nonce.String, commonAddr)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to verify signature")
			return terror.Error(err, "Invalid signature provided.")
		}

		// Write cookie and login user for passport if user already exist
		loginReq := &FingerprintTokenRequest{
			User:        &user.User,
			RedirectURL: req.RedirectURL,
			Tenant:      req.Tenant,
			Fingerprint: req.Fingerprint,
		}

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			return terror.Error(err, "Unable to issue a login token to user")
		}

		if user.DeletedAt.Valid {
			return fmt.Errorf("User does not exist")
		}

		// Dont send back request as response if 2FA
		if user.TwoFactorAuthenticationIsSet {
			return nil
		}
	}

	// Send response back to client
	resp := struct {
		WalletLoginRequest
		NewUser         bool `json:"new_user"`
		CaptchaRequired bool `json:"captcha_required"`
	}{
		WalletLoginRequest: *req,
		NewUser:            user == nil,
		CaptchaRequired:    user == nil, // ALL user signup through wallet requires captcha
	}
	b, err := json.Marshal(resp)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to encode response to json")
		return terror.Error(err, "Unable to decode response to user.")
	}
	_, err = w.Write(b)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to write response to user")
		return terror.Error(err, "Unable to write response to user.")
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

// Handler for GoogleLogin
func (api *API) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &GoogleLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode google login request")
		return http.StatusBadRequest, terror.Error(err, "Unable to decode user request.")
	}
	err = api.GoogleLogin(req, w, r)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to google login")
		return http.StatusBadRequest, terror.Error(err, "Failed to authenticate login with Google.")
	}
	return http.StatusCreated, nil
}

// Handles google login and signup
func (api *API) GoogleLogin(req *GoogleLoginRequest, w http.ResponseWriter, r *http.Request) error {
	// Check if there are any existing users associated with the email address
	googleDetails, err := api.GoogleToken(req.GoogleToken)
	if err != nil {
		passlog.L.Error().Err(err).Msg("user provided invalid google token")
		return terror.Error(err, "Invalid google token provided.")
	}
	user, err := users.GoogleID(googleDetails.GoogleID)

	loginReq := &FingerprintTokenRequest{
		User:        user,
		RedirectURL: req.RedirectURL,
		Tenant:      req.Tenant,
		Fingerprint: req.Fingerprint,
	}

	// If new user signup
	if err != nil && errors.Is(sql.ErrNoRows, err) {
		// Check if user gmail already exist
		if googleDetails.Email == "" {
			noEmailErr := fmt.Errorf("Missing google email.")
			passlog.L.Error().Err(noEmailErr).Msg("no email provided for google auth")
			return terror.Error(noEmailErr)
		}

		user, _ = boiler.Users(boiler.UserWhere.Email.EQ(null.StringFrom(googleDetails.Email))).One(passdb.StdConn)
		// If user exist already
		if user != nil {
			user.GoogleID = null.StringFrom(googleDetails.GoogleID)
			user.Verified = true
			loginReq.User = user

			_, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.GoogleID, boiler.UserColumns.Verified))
			if err != nil {
				passlog.L.Error().Err(err).Msg("unable to add google id to user with existing gmail")
				return terror.Error(err, "Failed to merge google account with existing user with same gmail.")
			}

		}

	} else if err != nil {
		return terror.Error(err, "Unable to update or create user with google account")
	}

	// Write cookie for passport if user already exist
	if user != nil {

		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			return terror.Error(err, "Unable to issue a login token to user")
		}

		if user.DeletedAt.Valid {
			return fmt.Errorf("User does not exist")
		}
		// Dont send back request as response if 2FA
		if user.TwoFactorAuthenticationIsSet {
			return nil
		}
	}

	// Write response back client
	resp := struct {
		GoogleLoginRequest
		NewUser         bool `json:"new_user"`
		CaptchaRequired bool `json:"captcha_required"`
	}{
		GoogleLoginRequest: *req,
		NewUser:            user == nil,
		CaptchaRequired:    user == nil, // ALL user signup through google requires captcha
	}

	b, err := json.Marshal(resp)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to encode response to json")
		return terror.Error(err, "Unable to decode response to user.")
	}
	_, err = w.Write(b)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to write response to user")
		return terror.Error(err, "Unable to write response to user.")
	}

	return nil
}

// Handles two factor authentication
func (api *API) TFAVerifyHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &TFAVerifyRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode 2fa verify request")
		return http.StatusBadRequest, terror.Error(err, "Unable to decode user request.")
	}

	// Get user from token
	user, err := api.TFAVerify(req, w, r)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
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
			return http.StatusBadRequest, terror.Error(err, "Unable to issue login token to user.")
		}
	} else {
		// If user already logged in then no need for token
		b, err := json.Marshal(user)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to convert user to json 2fa response")
			return http.StatusBadRequest, terror.Error(err, "Unable to decode response to user.")
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write 2fa response to user")
			return http.StatusBadRequest, terror.Error(err, "Unable to write response to user.")
		}
	}

	if req.IsVerified {
		return http.StatusOK, nil
	}

	// If external forward request to external handler
	if req.RedirectURL != "" {
		b, err := json.Marshal(req)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return http.StatusBadRequest, terror.Error(err, "Unable to decode response to user.")
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return http.StatusBadRequest, terror.Error(err, "Unable to write response to user.")
		}
	}

	return http.StatusOK, nil
}

// Check two factor authentication pass code or recovery code
func (api *API) TFAVerify(req *TFAVerifyRequest, w http.ResponseWriter, r *http.Request) (*types.User, error) {
	// Get user from token
	// If user is logged in, user is from cookie or query
	// If user is not logged in, token is passed from request

	user, _, _ := GetUserFromToken(api, r)

	if user == nil && req.Token != "" {
		u, err := api.UserFromToken(req.Token)
		if err != nil || u == nil {
			return nil, terror.Error(err, "Invalid token unable to get user.")
		}
		user = &u.User
	}

	if user == nil {
		err := fmt.Errorf("no token or cookie provided to verify tfa")
		passlog.L.Error().Err(err).Msg("unable to write response to user")
		return nil, terror.Error(err, "User is missing from request.")
	}

	// Check if there is a passcode and verify it
	if req.Passcode != "" {
		err := users.VerifyTFA(user.TwoFactorAuthenticationSecret, req.Passcode)
		if err != nil {
			return nil, terror.Error(err, "Incorrect passcode provided. Please try again.")
		}
	} else if req.RecoveryCode != "" {
		// Check if there is a recovery code and verify it
		err := users.VerifyTFARecovery(user.ID, req.RecoveryCode)
		if err != nil {
			errMsg := "Invalid recovery code provided."
			passlog.L.Error().Err(err).Msg(errMsg)
			return nil, terror.Error(err, errMsg)
		}
	} else {
		return nil, terror.Error(fmt.Errorf("passcode or verify code are missing"), "Passcode or verify code are missing.")
	}

	u, err := types.UserFromBoil(user)
	if err != nil {
		return nil, terror.Error(err, "Unable to get user type as response.")
	}
	return u, nil

}

// Handles facebook login and signup using facebooklogin
func (api *API) FacebookLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	req := &FacebookLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to decode facebook login request")
		return http.StatusBadRequest, terror.Error(err, "Unable to decode request.")
	}

	err = api.FacebookLogin(req, w, r)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err)
	}
	return http.StatusOK, nil
}

// Login user using facebook access token
func (api *API) FacebookLogin(req *FacebookLoginRequest, w http.ResponseWriter, r *http.Request) error {
	facebookDetails, err := api.FacebookToken(req.FacebookToken)
	if err != nil {
		return terror.Error(err, "Invalid facebook token provided.")
	}
	// Check if there are any existing users with facebookid
	user, err := users.FacebookID(facebookDetails.FacebookID)
	if err != nil && !errors.Is(sql.ErrNoRows, err) {
		return terror.Error(err, "Unable to authenticate user with Facebook.")
	}

	// Write cookie for passport if user already exist
	if user != nil {
		loginReq := &FingerprintTokenRequest{
			User:        user,
			RedirectURL: req.RedirectURL,
			Tenant:      req.Tenant,
			Fingerprint: req.Fingerprint,
		}
		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			return terror.Error(err, "Unable to issue a login token to user")
		}

		if user.DeletedAt.Valid {
			return fmt.Errorf("User does not exist")
		}
		// Dont send back request as response if 2FA
		if user.TwoFactorAuthenticationIsSet {
			return nil
		}

	}

	// Write response back client
	resp := struct {
		FacebookLoginRequest
		NewUser         bool `json:"new_user"`
		CaptchaRequired bool `json:"captcha_required"`
	}{
		FacebookLoginRequest: *req,
		NewUser:              user == nil,
		CaptchaRequired:      user == nil, // ALL user signup through facebook requires captcha
	}

	b, err := json.Marshal(resp)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to encode response to json")
		return terror.Error(err, "Unable to decode response to user.")
	}
	_, err = w.Write(b)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to write response to user")
		return terror.Error(err, "Unable to write response to user.")
	}
	if err != nil {
		passlog.L.Error().Err(err).Msg("Failed to verify captcha.")
		return terror.Error(err, "Failed to complete captcha verification.")
	}

	return nil
}

type TwitterAuthResponse struct {
	UserID *string `json:"user_id"`
}

type AddTwitterResponse struct {
	Error string      `json:"error"`
	User  *types.User `json:"user"`
}

// The TwitterAuth endpoint kicks off the OAuth 1.0a flow with signup/login/add connection
func (api *API) TwitterAuth(w http.ResponseWriter, r *http.Request) (int, error) {
	errMsg := "Failed to authenticate user with twitter."
	oauthVerifier := r.URL.Query().Get("oauth_verifier")
	oauthCallback := r.URL.Query().Get("oauth_callback")
	oauthToken := r.URL.Query().Get("oauth_token")
	redirect := r.URL.Query().Get("redirect")
	redirectURL := r.URL.Query().Get("redirectURL")
	addTwitter := r.URL.Query().Get("add")
	tenant := r.URL.Query().Get("tenant")
	var jwtToken string

	if redirect == "" && oauthVerifier != "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("Missing redirect and verifier."), errMsg)
	}

	if oauthVerifier != "" {
		params := url.Values{}
		params.Set("oauth_token", oauthToken)
		params.Set("oauth_verifier", oauthVerifier)
		twitterDetails, err := api.TwitterToken(params.Encode())
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, errMsg)
		}

		// Generate jwt token
		jwtToken, err = api.OneTimeTwitterJWT(twitterDetails.TwitterID, twitterDetails.ScreenName)
		if err != nil {
			return http.StatusInternalServerError, terror.Error(err, errMsg)
		}

		// Check if user exist
		user, err := users.TwitterID(twitterDetails.TwitterID)
		// Add twitter user handler
		if addTwitter != "" {
			return api.AddTwitterUser(w, r, redirect, user, twitterDetails.TwitterID, addTwitter)
		}

		if err != nil && errors.Is(sql.ErrNoRows, err) {
			// Redirect to signup page using jwt token with twitter id
			http.Redirect(w, r, fmt.Sprintf("%s?token=%s&redirectURL=%s", redirect, jwtToken, redirectURL), http.StatusSeeOther)
			return http.StatusSeeOther, nil
		}
		loginReq := &FingerprintTokenRequest{
			User:        user,
			RedirectURL: redirect,
			Tenant:      tenant,
			IsTwitter:   true,
		}
		err = api.FingerprintAndIssueToken(w, r, loginReq)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, errMsg)
		}

		http.Redirect(w, r, fmt.Sprintf("%s?login=ok&token=%s", redirect, jwtToken), http.StatusSeeOther)

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
		return http.StatusInternalServerError, terror.Error(err, errMsg)
	}

	http.Redirect(w, r, fmt.Sprintf("https://api.twitter.com/oauth/authorize?oauth_token=%s", requestToken), http.StatusSeeOther)
	return http.StatusOK, nil
}

type AuthTwitterResponse struct {
	Error *string
}

func (api *API) AddTwitterUser(w http.ResponseWriter, r *http.Request, redirect string, userWithTwitterID *boiler.User, twitterID string, userID string) (int, error) {
	payload := &AddTwitterResponse{}
	URI := fmt.Sprintf("/user/%s", userID)

	// Redirect to loading page
	http.Redirect(w, r, redirect, http.StatusSeeOther)

	if userWithTwitterID != nil {
		payload.Error = "Twitter account already registered with a different user"
		ws.PublishMessage(URI, HubKeyUserAddTwitter, payload)
		return http.StatusOK, nil

	} else {
		// Check if user exist
		user, err := users.ID(userID)
		if err != nil {
			payload.Error = "User ID does not exist"
			ws.PublishMessage(URI, HubKeyUserAddTwitter, payload)
			return http.StatusOK, nil
		}
		// Activity tracking
		var oldUser types.User = *user

		// Update user's Twitter ID
		user.TwitterID = null.StringFrom(twitterID)
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
	if req.User == nil {
		err := fmt.Errorf("user does not exist in issuing token")
		passlog.L.Error().Err(err).Msg(err.Error())
		return terror.Error(err, "Invalid request, unable to get user details.")
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
			return terror.Error(err, "Unable to generate a two-factor authentication token.")
		}

		origin := r.Header.Get("origin")

		// IF redirect is from Twitter Auth origin will be passport.xsyn.io/twitter-redirect
		if origin == "" {
			origin = strings.ReplaceAll(req.RedirectURL, "/twitter-redirect", "")
		}

		// Redirect to 2fa
		if req.IsTwitter {
			// add query tfa=ok for twitter message
			rURL := fmt.Sprintf("%s/tfa/check?token=%s&redirectURL=%s?tfa=ok&tenant=%s", origin, token, req.RedirectURL, req.Tenant)
			http.Redirect(w, r, rURL, http.StatusSeeOther)
			return nil
		}

		tokenResp := struct {
			TFAToken string `json:"tfa_token"`
		}{
			TFAToken: token,
		}
		// Send response to user and pass token to redirect to 2fa
		b, err := json.Marshal(tokenResp)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return terror.Error(err, "Failed to decode token to user.")
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return terror.Error(err, "Failed to write token to user.")
		}

		return nil
	}

	// Fingerprint user
	if req.Fingerprint != nil {
		err := api.DoFingerprintUpsert(*req.Fingerprint, req.User.ID)
		if err != nil {
			return terror.Error(err, "Failed to process user login.")
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
		return terror.Error(err, "Failed to process user login.")
	}
	if req.User.DeletedAt.Valid {
		return terror.Error(fmt.Errorf("User was deleted but try to login."), "Failed to process user login.")
	}
	err = api.WriteCookie(w, r, token)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to write a cookie")
		return terror.Error(err, "Failed to process user login.")
	}

	if req.RedirectURL == "" {
		b, err := json.Marshal(u)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to encode response to json")
			return terror.Error(err, "Failed to process user login.")
		}
		_, err = w.Write(b)
		if err != nil {
			passlog.L.Error().Err(err).Msg("unable to write response to user")
			return terror.Error(err, "Failed to process user login.")
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

// For forget password
func (api *API) OneTimeToken(userID string) (string, error) {
	var err error
	tokenID := uuid.Must(uuid.NewV4())

	expires := time.Now().Add(time.Minute * 10)

	// save user detail as jwt
	jwt, sign, err := tokens.GenerateOneTimeJWT(
		tokenID,
		expires,
		userID)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to generate one time token")
		return "", terror.Error(err, "Failed to create token.")
	}

	jwtSigned, err := sign(jwt, true, api.TokenEncryptionKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to sign jwt")
		return "", terror.Error(err, "Failed to process token.")
	}

	token := base64.StdEncoding.EncodeToString(jwtSigned)
	return token, nil
}

// For twitter signup flow
func (api *API) OneTimeTwitterJWT(twitterID string, screenName string) (string, error) {
	var err error
	tokenID := uuid.Must(uuid.NewV4())

	expires := time.Now().Add(time.Minute * 10)

	// save user detail as jwt
	jwt, sign, err := tokens.GenerateOneTimeJWT(
		tokenID,
		expires,
		twitterID,
		"twitter-screenname", screenName,
		"twitter-signup", "true",
	)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to generate one time signup token")
		return "", terror.Error(err, "Failed to create token.")
	}

	jwtSigned, err := sign(jwt, true, api.TokenEncryptionKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to sign jwt for signup")
		return "", terror.Error(err, "Failed to process token.")
	}

	token := base64.StdEncoding.EncodeToString(jwtSigned)
	return token, nil
}

// For user verification
func (api *API) OneTimeVerification() (string, error) {
	var err error
	tokenID := uuid.Must(uuid.NewV4())

	expires := time.Now().Add(time.Minute * 10)

	// save user detail as jwt
	jwt, sign, err := tokens.GenerateVerifyCodeJWT(
		tokenID, expires)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to generate one time token")
		return "", terror.Error(err, "Failed to create token.")
	}

	jwtSigned, err := sign(jwt, true, api.TokenEncryptionKey)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to sign jwt")
		return "", terror.Error(err, "Failed to process token.")
	}

	token := base64.StdEncoding.EncodeToString(jwtSigned)
	return token, nil
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

func (api *API) IssueToken(config *TokenConfig) (*types.User, uuid.UUID, string, error) {
	return token(api, config, true, api.TokenExpirationDays)
}

func (api *API) VerifySignature(signature string, nonce string, publicKey common.Address) error {
	errMsg := "Unable to verify signature."
	decodedSig, err := hexutil.Decode(signature)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if decodedSig[64] == 0 || decodedSig[64] == 1 {
		//https://ethereum.stackexchange.com/questions/102190/signature-signed-by-go-code-but-it-cant-verify-on-solidity
		decodedSig[64] += 27
	} else if decodedSig[64] != 27 && decodedSig[64] != 28 {
		return terror.Error(fmt.Errorf("decode sig invalid %v", decodedSig[64]), errMsg)
	}
	decodedSig[64] -= 27

	msg := []byte(fmt.Sprintf("%s:\n %s", api.Eip712Message, nonce))
	prefixedNonce := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(msg), msg)

	hash := crypto.Keccak256Hash([]byte(prefixedNonce))
	recoveredPublicKey, err := crypto.Ecrecover(hash.Bytes(), decodedSig)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	secp256k1RecoveredPublicKey, err := crypto.UnmarshalPubkey(recoveredPublicKey)
	if err != nil {
		return terror.Error(err, errMsg)
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

	if user != nil {
		user.Nonce = null.StringFrom(newNonce)
		i, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Nonce))
		if err != nil {
			return "", terror.Error(fmt.Errorf("nonce could not be updated"), "Unable to process wallet login.")
		}

		if i == 0 {
			return "", terror.Error(fmt.Errorf("nonce could not be updated"), "Unable to process wallet login.")
		}
	}

	return newNonce, nil
}

func (api *API) GetNonce(w http.ResponseWriter, r *http.Request) (int, error) {
	errMsg := "Unable to process wallet connection."
	publicAddress := r.URL.Query().Get("public-address")
	userID := r.URL.Query().Get("user-id")

	L := passlog.L.With().Str("publicAddress", publicAddress).Str("userID", userID).Logger()

	if publicAddress == "" && userID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("missing public address or user id"))
	}
	if publicAddress != "" && common.IsHexAddress(publicAddress) {
		// Take public address Hex to address(Make it a checksum mixed case address) convert back to Hex for string of checksum
		commonAddr := common.HexToAddress(publicAddress)
		user, _ := users.PublicAddress(commonAddr)

		u := &boiler.User{}
		if user == nil {
			u = nil
		} else {
			u = &user.User
		}
		newNonce, err := api.NewNonce(u)
		if err != nil {
			L.Error().Err(err).Msg("no nonce")
			return http.StatusBadRequest, terror.Error(err, errMsg)
		}

		resp := &GetNonceResponse{
			Nonce: newNonce,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			L.Error().Err(err).Msg("json failed")
			return http.StatusInternalServerError, terror.Error(err, errMsg)
		}
		return http.StatusOK, nil
	}

	user, err := boiler.FindUser(passdb.StdConn, userID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, errMsg)
	}

	newNonce, err := api.NewNonce(user)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, errMsg)
	}

	resp := &GetNonceResponse{
		Nonce: newNonce,
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, errMsg)
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
	errMsg := "Failed to authenticate with token."

	resp, err := api.TokenLogin(req.Token, req.TwitchExtensionJWT)
	if err != nil {
		return nil, terror.Error(err, "Failed to authenticate with token.")
	}

	// Fingerprint user
	if req.Fingerprint != nil {
		userID := resp.User.ID
		// todo: include ip in upsert
		err = api.DoFingerprintUpsert(*req.Fingerprint, userID)
		if err != nil {
			return nil, terror.Error(fmt.Errorf("failed to identify browser: %w", err), errMsg)
		}
	}

	if resp.User.DeletedAt.Valid {
		return nil, terror.Error(fmt.Errorf("user does not exist"), errMsg)
	}

	user, err := types.UserFromBoil(resp.User)
	if err != nil {
		return nil, terror.Error(fmt.Errorf("failed to identify user: %w", err), errMsg)
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
		return http.StatusBadRequest, terror.Error(err, "Failed to process token")
	}

	// check user from token
	resp, err := api.TokenLogin(token, "")
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to authenticate user.")
	}

	return helpers.EncodeJSON(w, resp.User)
}

func (api *API) AuthLogoutHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	cookie, err := r.Cookie("xsyn-token")
	if err != nil {
		// check whether token is attached
		return http.StatusBadRequest, terror.Warn(fmt.Errorf("no cookie are provided"), "User is not signed in.")
	}

	// Find user from cookie
	var token string
	if err = api.Cookie.DecryptBase64(cookie.Value, &token); err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to process token")
	}

	// check user from token
	resp, err := api.UserFromToken(token)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to find user.")
	}

	// Delete all issued token
	_, err = resp.User.IssueTokens().UpdateAll(passdb.StdConn, boiler.M{
		boiler.IssueTokenColumns.DeletedAt: time.Now(),
	})
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to delete all issued token to logout")
		return http.StatusInternalServerError, terror.Error(err, "Unable to delete all current sessions")
	}

	// clear and expire cookie and push to browser
	err = api.DeleteCookie(w, r)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to logout user.")
	}

	return helpers.EncodeJSON(w, true)
}

// TokenLoginHandler lets you log in with just a jwt
func (api *API) TokenLoginHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	errMsg := "Failed to login user."
	req := &TokenLoginRequest{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, errMsg)
	}

	resp, err := api.TokenAuth(req, r)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, errMsg)
	}

	err = api.WriteCookie(w, r, req.Token)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, errMsg)
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, errMsg)
	}

	return http.StatusOK, nil
}

// TokenLogin gets a user from the token
func (api *API) TokenLogin(tokenBase64 string, twitchExtensionJWT string) (*TokenLoginResponse, error) {
	errMsg := "Failed to login user with token."
	tokenStr, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return nil, terror.Error(err, errMsg)
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
		return nil, terror.Error(errors.New("unable to get ID from token"), errMsg)
	}

	jwtID, err := uuid.FromString(jwtIDI.(string))
	if err != nil {
		return nil, terror.Error(err, errMsg)
	}

	retrievedToken, user, err := tokens.Retrieve(jwtID)
	if err != nil {
		return nil, terror.Error(err, errMsg)
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

var ErrJWTKeyNotFound = fmt.Errorf("failed to read data from token")

func (api *API) ReadUserIDJWT(tokenBase64 string) (string, error) {
	userID, err := api.ReadKeyJWT(tokenBase64, "user-id")
	if err != nil {
		return "", terror.Error(err)
	}
	return userID, nil
}

func (api *API) ReadKeyJWT(tokenBase64 string, key string) (string, error) {
	errMsg := "Failed to read token."
	tokenStr, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		passlog.L.Err(err).Msg("Failed to decode token.")
		return "", terror.Error(err, errMsg)
	}
	// Decode token user with new email
	token, err := tokens.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		return "", terror.Error(err, errMsg)
	}

	val, ok := token.Get(key)
	if !ok {
		return "", terror.Error(ErrJWTKeyNotFound, errMsg)
	}

	valStr, ok := val.(string)
	if !ok {
		return "", terror.Error(ErrJWTKeyNotFound, errMsg)
	}
	return valStr, nil
}

func (api *API) ReadCodeJWT(tokenBase64 string) (string, error) {
	errMsg := "Failed to read token."
	tokenStr, err := base64.StdEncoding.DecodeString(tokenBase64)
	if err != nil {
		return "", terror.Error(err, errMsg)
	}
	// Decode token user with new email
	token, err := tokens.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		return "", terror.Error(err, errMsg)
	}

	c, _ := token.Get("code")
	code, ok := c.(string)

	if !ok {
		return "", terror.Error(errors.New("failed to read code from token"), errMsg)
	}
	return code, nil
}

type GoogleValidateResponse struct {
	GoogleID string
	Email    string
	Username string
}

func (api *API) GoogleToken(token string) (*GoogleValidateResponse, error) {
	errMsg := "There was a problem finding a user associated with the provided Google account, please check your details and try again."
	payload, err := idtoken.Validate(context.Background(), token, api.Google.ClientID)
	if err != nil {
		return nil, terror.Error(err, errMsg)
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return nil, terror.Error(err, errMsg)
	}
	googleID, ok := payload.Claims["sub"].(string)
	if !ok {
		return nil, terror.Error(err, errMsg)
	}
	username, ok := payload.Claims["given_name"].(string)
	if !ok {
		return nil, terror.Error(err, errMsg)
	}

	resp := &GoogleValidateResponse{
		Email:    email,
		GoogleID: googleID,
		Username: username,
	}
	return resp, nil
}

type FacebookValidateResponse struct {
	FacebookID string `json:"id"`
	Name       string `json:"name"`
}

func (api *API) FacebookToken(token string) (*FacebookValidateResponse, error) {
	errMsg := "There was a problem finding a user associated with the provided Facebook account, please check your details and try again."
	r, err := http.Get("https://graph.facebook.com/me?field=name&access_token=" + url.QueryEscape(token))
	if err != nil {
		return nil, terror.Error(err)
	}
	defer r.Body.Close()
	resp := &FacebookValidateResponse{}
	err = json.NewDecoder(r.Body).Decode(resp)
	if err != nil {
		return nil, terror.Error(err, errMsg)
	}

	return resp, nil
}

type TwitterValidateResponse struct {
	OauthToken       string
	OauthTokenSecret string
	TwitterID        string
	ScreenName       string
}

func (api *API) TwitterToken(token string) (*TwitterValidateResponse, error) {
	errMsg := "There was a problem finding a user associated with the provided Twitter account, please check your details and try again."
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/oauth/access_token?%s", token), nil)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to get twitter access token")
		return nil, terror.Error(err, errMsg)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to request for twitter access token")
		return nil, terror.Error(err, errMsg)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to get body from twitter response")
		return nil, terror.Error(err, errMsg)
	}
	resp := &TwitterValidateResponse{}
	values := strings.Split(string(body), "&")
	for _, v := range values {
		pair := strings.Split(v, "=")
		switch pair[0] {
		case "oauth_token":
			resp.OauthToken = pair[1]
		case "oauth_token_secret":
			resp.OauthTokenSecret = pair[1]
		case "user_id":
			resp.TwitterID = pair[1]
		case "screen_name":
			resp.ScreenName = pair[1]
		}
	}

	return resp, nil
}
