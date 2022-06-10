package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/auth"
	DatadogTracer "github.com/ninja-syndicate/hub/ext/datadog"
	"io/ioutil"
	"net/http"
	"xsyn-services/boiler"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
)

type ErrorMessage string

const (
	Unauthorised          ErrorMessage = "Unauthorised - Please log in or contact System Administrator"
	Forbidden             ErrorMessage = "Forbidden - You do not have permissions for this, please contact System Administrator"
	InternalErrorTryAgain ErrorMessage = "Internal Error - Please try again in a few minutes or Contact Support"
	InputError            ErrorMessage = "Input Error - Please try again"
)

func (errMsg ErrorMessage) String() string {
	return string(errMsg)
}

// ErrorObject is used by the front end react-fetching-library
type ErrorObject struct {
	Message   string `json:"message"`
	ErrorCode string `json:"error_code"`
}

// WithError handles error responses.
func WithError(next func(w http.ResponseWriter, r *http.Request) (int, error)) http.HandlerFunc {
	// TODO: Ask about sentry ideas?
	fn := func(w http.ResponseWriter, r *http.Request) {
		contents, err := ioutil.ReadAll(r.Body)

		r.Body = ioutil.NopCloser(bytes.NewReader(contents))
		defer r.Body.Close()
		code, err := next(w, r)
		if err != nil {
			terror.Echo(err)
			errObj := &ErrorObject{
				Message:   err.Error(),
				ErrorCode: fmt.Sprintf("%d", code),
			}
			var bErr *terror.TError
			if errors.As(err, &bErr) {
				errObj.Message = bErr.Message

				switch bErr.Level {
				case terror.ErrLevelWarn:
					passlog.L.Warn().Err(err).Str("stack trace", terror.Echo(bErr, false)).Msg("rest error")
				default:
					passlog.L.Err(err).Str("stack trace", terror.Echo(bErr, false)).Msg("rest error")
				}

				// set generic messages if friendly message not set making genric messages overrideable
				if bErr.Error() == bErr.Message {

					// if internal server error set as genric internal error message
					if code == 500 {
						errObj.Message = InternalErrorTryAgain.String()
					}

					// if forbidden error set genric forbidden error
					if code == 403 {
						errObj.Message = Forbidden.String()
					}

					// if unauthed error set genric unauthed error
					if code == 401 {
						errObj.Message = Unauthorised.String()
					}

					// if badstatus request
					if code == 400 {
						errObj.Message = InputError.String()
					}
				}
			} else {
				passlog.L.Err(err).Str("r.URL.Path",r.URL.Path).Msg("rest error")
			}

			jsonErr, err := json.Marshal(errObj)
			if err != nil {
				terror.Echo(err)
				DatadogTracer.HttpFinishSpan(r.Context(), code, err)
				http.Error(w, `{"message":"JSON failed, please contact IT.","error_code":"00001"}`, code)
				return
			}

			DatadogTracer.HttpFinishSpan(r.Context(), code, bErr)
			http.Error(w, string(jsonErr), code)
			return
		}
		DatadogTracer.HttpFinishSpan(r.Context(), code, nil)
	}
	return fn
}

// WithUser checks the hub for authenticated user.
func WithUser(api *API, next func(w http.ResponseWriter, r *http.Request, user *boiler.User) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		user, code, err := GetUserFromToken(api, r)
		if err != nil {
			return code, err
		}
		if user != nil {
			return next(w, r, user)
		}
		return http.StatusUnauthorized, terror.Error(err, "Error - Please log in")
	}
	return fn
}

func GetUserFromToken(api *API, r *http.Request) (*boiler.User, int, error) {
	tokenB64 := r.URL.Query().Get("token")

	tokenStr, err := base64.StdEncoding.DecodeString(tokenB64)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	jwt, err := auth.ReadJWT(tokenStr, true, api.TokenEncryptionKey)
	if err != nil {
		return nil, http.StatusUnauthorized, terror.Warn(err)
	}

	jwtIDI, ok := jwt.Get(openid.JwtIDKey)

	if !ok {
		return nil, http.StatusUnauthorized, err
	}

	jwtID, err := uuid.FromString(jwtIDI.(string))
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	token, err := boiler.FindIssueToken(passdb.StdConn, jwtID.String())
	if err != nil {
		return nil, http.StatusUnauthorized, terror.Error(err, "Failed to secure user")
	}

	user, err := boiler.FindUser(passdb.StdConn, token.UserID)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}
	return user, http.StatusOK, nil
}
