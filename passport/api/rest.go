package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/jwt/openid"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/auth"
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
		contents, _ := ioutil.ReadAll(r.Body)

		r.Body = ioutil.NopCloser(bytes.NewReader(contents))
		defer r.Body.Close()
		code, err := next(w, r)
		if err != nil && errors.Is(err, sql.ErrNoRows) {
			errStr := terror.Echo(err, true)
			passlog.L.Warn().Err(err).Msg("rest record not found. " + errStr)

			errObj := &ErrorObject{
				Message:   "record not found",
				ErrorCode: fmt.Sprintf("%d", http.StatusNotFound),
			}
			jsonErr, err := json.Marshal(errObj)
			if err != nil {
				terror.Echo(err)
				http.Error(w, `{"message":"JSON failed, please contact IT.","error_code":"00001"}`, http.StatusInternalServerError)
				return
			}
			http.Error(w, string(jsonErr), http.StatusNotFound)
			return
		}
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
					passlog.L.Warn().Err(err).Msg("rest error")
				default:
					passlog.L.Err(err).Msg("rest error")
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
				passlog.L.Err(err).Msg("rest error")
			}

			jsonErr, err := json.Marshal(errObj)
			if err != nil {
				terror.Echo(err)
				http.Error(w, `{"message":"JSON failed, please contact IT.","error_code":"00001"}`, code)
				return
			}

			http.Error(w, string(jsonErr), code)
			return
		}
	}
	return fn
}

// WithUser checks the hub for authenticated user.
func WithUser(api *API, next func(w http.ResponseWriter, r *http.Request, user *types.User) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		user, code, err := GetUserFromToken(api, r)
		if err != nil {
			return code, terror.Error(err)
		}
		if user != nil {
			return next(w, r, user)
		}
		return http.StatusUnauthorized, terror.Error(err, "Error - Please log in")
	}
	return fn
}

func GetUserFromToken(api *API, r *http.Request) (*types.User, int, error) {
	ctx := context.Background()
	tokenB64 := r.URL.Query().Get("token")

	tokenStr, err := base64.StdEncoding.DecodeString(tokenB64)
	if err != nil {
		return nil, http.StatusUnauthorized, terror.Error(err)
	}

	jwt, err := auth.ReadJWT(tokenStr, api.Tokens.EncryptToken(), api.Tokens.EncryptTokenKey())
	if err != nil {
		return nil, http.StatusUnauthorized, terror.Warn(err)
	}

	jwtIDI, ok := jwt.Get(openid.JwtIDKey)

	if !ok {
		return nil, http.StatusUnauthorized, terror.Error(err)
	}

	jwtID, err := uuid.FromString(jwtIDI.(string))
	if err != nil {
		return nil, http.StatusUnauthorized, terror.Error(err)
	}

	token, err := db.AuthFindToken(ctx, api.Conn, types.IssueTokenID(jwtID))
	if err != nil {
		return nil, http.StatusUnauthorized, terror.Error(err)
	}

	user, err := db.UserGet(ctx, api.Conn, token.UserID)
	if err != nil {
		return nil, http.StatusUnauthorized, terror.Error(err)
	}
	return user, http.StatusOK, nil
}
