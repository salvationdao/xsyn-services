package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/api/users"
	"xsyn-services/passport/crypto"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	petname "github.com/dustinkirkland/golang-petname"

	"github.com/ninja-syndicate/ws"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	oidc "github.com/coreos/go-oidc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/microcosm-cc/bluemonday"
	"github.com/ninja-software/log_helpers"
	btypes "github.com/volatiletech/sqlboiler/v4/types"

	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/auth"
	"github.com/rs/zerolog"

	"github.com/pquerna/otp/totp"
)

// UserController holds handlers for authentication
type UserController struct {
	Log     *zerolog.Logger
	API     *API
	Google  *auth.GoogleConfig
	Twitch  *auth.TwitchConfig
	Discord *auth.DiscordConfig
}

// NewUserController creates the user hub
func NewUserController(log *zerolog.Logger, api *API, googleConfig *auth.GoogleConfig, twitchConfig *auth.TwitchConfig, discordConfig *auth.DiscordConfig) *UserController {
	userHub := &UserController{
		Log:     log_helpers.NamedLogger(log, "user_hub"),
		API:     api,
		Google:  googleConfig,
		Twitch:  twitchConfig,
		Discord: discordConfig,
	}

	api.SecureCommand(HubKeyUserGet, userHub.GetHandler) // Perm check inside handler (users can get themselves; need UserRead permission to get other users)
	api.SecureCommand(HubKeyUserUpdate, userHub.UpdateHandler)
	api.SecureCommand(HubKeyUserUsernameUpdate, userHub.UpdateUserUsernameHandler)
	api.SecureCommand(HubKeyUserRemoveFacebook, userHub.RemoveFacebookHandler) // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddFacebook, userHub.AddFacebookHandler)       // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveGoogle, userHub.RemoveGoogleHandler)     // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddGoogle, userHub.AddGoogleHandler)           // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveTwitch, userHub.RemoveTwitchHandler)     // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddTwitch, userHub.AddTwitchHandler)           // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveTwitter, userHub.RemoveTwitterHandler)   // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddTwitter, userHub.AddTwitterHandler)         // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveDiscord, userHub.RemoveDiscordHandler)   // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddDiscord, userHub.AddDiscordHandler)         // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveWallet, userHub.RemoveWalletHandler)     // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddWallet, userHub.AddWalletHandler)           // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserCreate, userHub.CreateHandler)
	api.SecureCommand(HubKeyUserLock, userHub.LockHandler)

	api.SecureCommand(HubKeyUser, userHub.UpdatedSubscribeHandler)
	api.SecureCommand(HubKeyUserInit, userHub.InitHandler)
	api.SecureCommand(HubKeyUserSendVerify, userHub.SendVerifyHandler)

	// TFA
	api.SecureCommand(HubKeyGenerateTFASecret, userHub.GenerateTFAHandler)
	api.SecureCommand(HubKeyTFACancel, userHub.CancelTFAHandler)
	api.SecureCommand(HubKeyTFAVerification, userHub.TFAVerificationHandler)
	api.SecureCommand(HubKeyTFARecoveryGet, userHub.TFARecoveryCodeGetHandler)
	api.SecureCommand(HubKeyTFARecoveryVerify, userHub.TFARecoveryVerifyHandler)

	return userHub
}

// GetUserRequest requests an update for an existing user
type GetUserRequest struct {
	Payload struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyUserGet, UserController.GetHandler)
const HubKeyUserGet = "USER"

// GetHandler gets the details for a user
func (uc *UserController) GetHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &GetUserRequest{}
	_ = json.Unmarshal(payload, req)

	if req.Payload.ID == "" && req.Payload.Username == "" {
		reply(user)
		return nil
	}

	// if hub user isn't requested user, clear private data
	if user.ID != req.Payload.ID {
		buser, err := boiler.Users(
			boiler.UserWhere.ID.EQ(req.Payload.ID),
			qm.Load(qm.Rels(boiler.UserRels.Faction)),
		).One(passdb.StdConn)
		if err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		b := &types.UserBrief{
			ID:       buser.ID,
			Username: buser.Username,
		}

		if buser.AvatarID.Valid {
			*b.AvatarID = buser.AvatarID.String
		}

		if buser.FactionID.Valid {
			*b.FactionID = buser.FactionID.String
		}

		b.Faction = buser.R.Faction

		reply(b)
		return nil
	}
	return nil

}

const HubKeyUserInit = "USER:INIT"

// Tracks if user should log out from change password
func (uc *UserController) InitHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	reply(user)
	return nil
}

const HubKeyUserSendVerify = "USER:VERIFY:SEND"

type SendVerifyRequest struct {
	Payload struct {
		Email     string `json:"email"`
		UserAgent string `json:"user_agent"`
	} `json:"payload"`
}

// Sends another email to user to verify email
//
// OR sends email to new email address for user to verify
func (uc *UserController) SendVerifyHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Failed to send a verification email. Please try again or contact support"
	req := &SendVerifyRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	lowerEmail := strings.ToLower(req.Payload.Email)

	// Check if email is valid
	_, err = mail.ParseAddress(lowerEmail)
	if err != nil {
		return terror.Error(err, "Invalid email provided.")
	}

	// If a new email address
	// Check if email address is already taken
	if lowerEmail != user.Email.String {
		u, _ := users.Email(lowerEmail)
		if u != nil {
			err = fmt.Errorf("email address is already taken by another user")
			return terror.Error(err, "Email address is already used. Please use a different email.")
		}
	}

	token, err := uc.API.OneTimeVerification()
	if err != nil || token == "" {
		err := fmt.Errorf("fail to generate verification code")
		passlog.L.Error().Err(err).Msg(err.Error())
		return terror.Error(err, errMsg)
	}
	code, err := uc.API.ReadCodeJWT(token)
	if err != nil {
		passlog.L.Error().Err(err).Msg("unable to get verify code from token")
		return terror.Error(err, errMsg)
	}

	err = uc.API.Mailer.SendVerificationEmail(context.Background(), user, code, lowerEmail)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	user.VerifyToken = token
	reply(user)
	return nil
}

type UpdateUserUsernameRequest struct {
	Payload struct {
		Username string `json:"username"`
	} `json:"payload"`
}

const HubKeyUserUsernameUpdate = "USER:USERNAME:UPDATE"

func (uc *UserController) UpdateUserUsernameHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating username, try again or contact support."
	req := &UpdateUserUsernameRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.Username == "" {
		return terror.Error(terror.ErrInvalidInput, "Username cannot be empty.")
	}

	bm := bluemonday.StrictPolicy()
	username := bm.Sanitize(strings.TrimSpace(req.Payload.Username))

	// Validate username
	err = helpers.IsValidUsername(username)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Check availability of username
	if user.Username == username {
		return terror.Error(fmt.Errorf("username must be different"))
	}

	isAvailable, err := users.UsernameExist(username)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	if !isAvailable {
		return terror.Error(fmt.Errorf("A user with that username already exists."), "A user with that username already exists.")
	}

	oldUserName := user.Username

	// Update username
	user.Username = username
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Username))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Log username change
	if oldUserName != user.Username {
		uh := boiler.UsernameHistory{
			UserID:      user.ID,
			OldUsername: oldUserName,
			NewUsername: user.Username,
		}
		err := uh.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			passlog.L.Warn().Err(err).Str("old username", oldUserName).Str("new username", user.Username).Msg("Failed to log username change in db")
			return terror.Error(err, errMsg)
		}
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Updated user's username",
		types.ObjectTypeUser,
		helpers.StringPointer(user.ID),
		&user.Username,
		helpers.StringPointer(user.FirstName.String+" "+user.LastName.String),
		&types.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: struct {
				Username     string `json:"username"`
				PrevUsername string `json:"previous_username"`
			}{
				Username:     user.Username,
				PrevUsername: oldUserName,
			},
			To: user,
		},
	)

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)
	return nil
}

// UpdateUserFactionRequest requests update user faction
type UpdateUserFactionRequest struct {
	Payload struct {
		UserID    types.UserID    `json:"user_id"`
		FactionID types.FactionID `json:"faction_id"`
	} `json:"payload"`
}

// HubKeyUserUpdate updates a user
const HubKeyUserUpdate = "USER:UPDATE"

// UpdateUserRequest requests an update for an existing user
type UpdateUserRequest struct {
	Payload struct {
		Username                         string      `json:"username"`
		NewUsername                      *string     `json:"new_username"`
		FirstName                        string      `json:"first_name"`
		LastName                         string      `json:"last_name"`
		MobileNumber                     string      `json:"mobile_number"`
		Email                            null.String `json:"email"`
		AvatarID                         *string     `json:"avatar_id"`
		CurrentPassword                  *string     `json:"current_password"`
		NewPassword                      *string     `json:"new_password"`
		TwoFactorAuthenticationActivated bool        `json:"two_factor_authentication_activated"`
		Code                             string      `json:"code"`
		UserAgent                        string      `json:"user_agent"`
	} `json:"payload"`
}

// UpdateHandler updates a user
func (uc *UserController) UpdateHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating user details, try again or contact support."
	req := &UpdateUserRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Setup user activity tracking
	oldUser := *user

	if req.Payload.Email.Valid && req.Payload.Email.String != "" {
		lowerEmail := strings.ToLower(req.Payload.Email.String)
		_, err := mail.ParseAddress(lowerEmail)
		if err != nil {
			return terror.Error(err, "Invalid email address.")
		}

		// Check if email address is already taken
		if null.StringFrom(lowerEmail) != user.Email {
			u, _ := users.Email(lowerEmail)
			if u != nil {
				err = fmt.Errorf("email address is already taken by another user.")
				return terror.Error(err, "Email address is already taken by another user.")
			}
			user.Email = null.StringFrom(lowerEmail)
		}
		user.Verified = true
	}
	if req.Payload.NewUsername != nil && req.Payload.Username != *req.Payload.NewUsername {
		// Validate username
		err = helpers.IsValidUsername(*req.Payload.NewUsername)
		if err != nil {
			return terror.Error(err, errMsg)
		}

		bm := bluemonday.StrictPolicy()
		sanitizedUsername := html.UnescapeString(bm.Sanitize(strings.TrimSpace(*req.Payload.NewUsername)))

		isTaken, err := users.UsernameExist(sanitizedUsername)
		if err != nil {
			return terror.Error(err, "A user with that username already exists.")
		}
		if isTaken {
			return terror.Error(fmt.Errorf("Username already taken."), "A user with that username already exists.")
		}
		user.Username = sanitizedUsername
	}

	user.FirstName = null.StringFrom(req.Payload.FirstName)
	user.LastName = null.StringFrom(req.Payload.LastName)

	if req.Payload.MobileNumber != "" && req.Payload.MobileNumber != user.MobileNumber.String {
		number, err := uc.API.SMS.Lookup(req.Payload.MobileNumber)
		if err != nil {
			return terror.Warn(fmt.Errorf("invalid mobile number %s", req.Payload.MobileNumber), "Invalid mobile number, please insure correct country code.")
		}

		user.MobileNumber = null.StringFrom(number)
	}
	if req.Payload.MobileNumber == "" {
		user.MobileNumber = null.NewString("", false)
	}

	if req.Payload.AvatarID != nil {
		user.AvatarID = null.StringFrom(*req.Payload.AvatarID)
	}

	// Start transaction
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	defer tx.Rollback()

	// Update user
	_, err = user.Update(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Log username change
	if oldUser.Username != user.Username {
		uh := boiler.UsernameHistory{
			UserID:      user.ID,
			OldUsername: oldUser.Username,
			NewUsername: user.Username,
		}
		err := uh.Insert(passdb.StdConn, boil.Infer())
		if err != nil {
			passlog.L.Warn().Err(err).Str("old username", oldUser.Username).Str("new username", user.Username).Msg("Failed to log username change in db")
			return terror.Error(err, errMsg)
		}
	}

	reply(user)
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Updated User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	var resp struct {
		IsSuccess bool `json:"is_success"`
	}
	// update game client server
	err = uc.API.GameserverRequest(http.MethodPost, "/user_update", struct {
		User *types.User `json:"user"`
	}{
		User: user,
	}, &resp)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	return nil
}

// UpdateUserSupsRequest requests an update for an existing user
type UpdateUserSupsRequest struct {
	Payload struct {
		UserID     types.UserID `json:"user_id"`
		SupsChange int64        `json:"sups_change"`
	} `json:"payload"`
}

// HubKeyUserCreate creates a user
const HubKeyUserCreate = "USER:CREATE"

// CreateUserRequest requests an create for an existing user
type CreateUserRequest struct {
	Payload struct {
		FirstName   string      `json:"first_name"`
		LastName    string      `json:"last_name"`
		Email       null.String `json:"email"`
		AvatarID    string      `json:"avatar_id"`
		NewPassword *string     `json:"new_password"`
		RoleID      string      `json:"role_id"`
	} `json:"payload"`
}

// CreateHandler creates a user
func (uc *UserController) CreateHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &CreateUserRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	email := strings.TrimSpace(req.Payload.Email.String)
	email = strings.ToLower(email)

	// Validation
	if req.Payload.FirstName == "" {
		return terror.Error(terror.ErrInvalidInput, "First Name is required.")
	}
	if req.Payload.LastName == "" {
		return terror.Error(terror.ErrInvalidInput, "Last Name is required.")
	}
	if !helpers.IsValidEmail(email) {
		return terror.Error(terror.ErrInvalidInput, "Email is required.")
	}
	if req.Payload.RoleID == "" {
		return terror.Error(terror.ErrInvalidInput, "Role is required.")
	}
	if req.Payload.NewPassword == nil {
		return terror.Error(terror.ErrInvalidInput, "Password is required.")
	}
	err = helpers.IsValidPassword(*req.Payload.NewPassword)
	if err != nil {
		return terror.Error(err, "Password is invalid.")
	}

	// Start transaction
	errMsg := "Unable to create user, please try again."
	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	defer tx.Rollback()

	// Create user
	newUser := &boiler.User{
		FirstName: null.StringFrom(req.Payload.FirstName),
		LastName:  null.StringFrom(req.Payload.LastName),
		Email:     req.Payload.Email,
		RoleID:    null.StringFrom(req.Payload.RoleID),
	}
	if req.Payload.AvatarID != "" {
		newUser.AvatarID = null.StringFrom(req.Payload.AvatarID)
	}

	err = newUser.Insert(tx, boil.Infer())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Set password

	userPassword, err := boiler.FindPasswordHash(passdb.StdConn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	userPassword.PasswordHash = crypto.HashPassword(*req.Payload.NewPassword)
	_, err = userPassword.Update(passdb.StdConn, boil.Whitelist(boiler.PasswordHashColumns.PasswordHash))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(newUser)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Created User",
		types.ObjectTypeUser,
		helpers.StringPointer(newUser.ID),
		&user.Username,
		helpers.StringPointer(user.FirstName.String+" "+user.LastName.String),
		&types.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: nil,
			To:   user,
		},
	)

	return nil
}

// HubKeyIntakeList is a hub key to run list user intake
const HubKeyUserList = "USER:LIST"

// ListHandlerRequest requests holds the filter for user list
type ListHandlerRequest struct {
	Payload struct {
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   db.UserColumn         `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Archived bool                  `json:"archived"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// UserListResponse is the response from get user list
type UserListResponse struct {
	Records []*types.User `json:"records"`
	Total   int           `json:"total"`
}

// ListHandler lists users with pagination
func (uc *UserController) ListHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Could not get users, try again or contact support."

	req := &ListHandlerRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	users := []*types.User{}
	total, err := db.UserList(
		users,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.Filter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)

	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := &UserListResponse{
		Total:   total,
		Records: users,
	}

	reply(resp)

	return nil
}

type RemoveServiceRequest struct {
	Payload struct {
		ID       types.UserID `json:"id"`
		Username string       `json:"username"`
	} `json:"payload"`
}

type AddFacebookRequest struct {
	Payload struct {
		FacebookToken string `json:"facebook_token"`
	} `json:"payload"`
}

// HubKeyUserRemoveFacebook removes a linked Facebook account
const HubKeyUserRemoveFacebook = "USER:FACEBOOK:REMOVE"

func (uc *UserController) RemoveFacebookHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	// Setup user activity tracking
	var oldUser types.User = *user

	// Check if user can remove service
	serviceCount := getUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(fmt.Errorf("You cannot unlink your only connection to this account"))
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	user.FacebookID = null.NewString("", false)
	_, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.FacebookID))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to remove user's facebook")
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Removed Facebook account from User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)
	return nil
}

// HubKeyUserAddFacebook removes a linked Facebook account
const HubKeyUserAddFacebook = "USER:FACEBOOK:ADD"

func (uc *UserController) AddFacebookHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {

	req := &AddFacebookRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check if Facebook ID is already taken
	facebookDetails, err := uc.API.FacebookToken(req.Payload.FacebookToken)
	if err != nil {
		passlog.L.Error().Err(err).Msg("user provided invalid facebook token")
		return terror.Error(err, "Invalid google token received")
	}
	u, _ := users.FacebookID(facebookDetails.FacebookID)
	if u != nil {
		return terror.Error(fmt.Errorf("This facebook account is already registered to a different user."))
	}

	// Setup user activity tracking
	var oldUser types.User = *user

	// Update user's Facebook ID
	user.FacebookID = null.StringFrom(facebookDetails.FacebookID)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.FacebookID))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to add user's facebook")
		return terror.Error(err, "Unable to update user details.")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Added Facebook account to User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)
	return nil
}

// HubKeyUserRemoveGoogle removes a linked Google account
const HubKeyUserRemoveGoogle = "USER:GOOGLE:REMOVE"

func (uc *UserController) RemoveGoogleHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {

	// Setup user activity tracking
	var oldUser types.User = *user

	// Check if user can remove service
	serviceCount := getUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(fmt.Errorf("You cannot unlink your only connection to this account"))
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	user.GoogleID = null.NewString("", false)
	_, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.GoogleID))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to remove user's google")
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Removed Google account from User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

type AddGoogleRequest struct {
	Payload struct {
		GoogleToken string `json:"google_token"`
	} `json:"payload"`
}

// HubKeyUserAddGoogle adds a linked Google account
const HubKeyUserAddGoogle = "USER:GOOGLE:ADD"

func (uc *UserController) AddGoogleHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &AddGoogleRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Check if Google ID is already taken
	googleDetails, err := uc.API.GoogleToken(req.Payload.GoogleToken)
	if err != nil {
		passlog.L.Error().Err(err).Msg("user provided invalid facebook token")
		return terror.Error(err, "Invalid google token received")
	}
	u, _ := users.GoogleID(googleDetails.GoogleID)
	if u != nil {
		return terror.Error(fmt.Errorf("This google account is already registered to a different user."))
	}
	// Setup user activity tracking
	var oldUser types.User = *user

	// Update user's Google ID
	user.GoogleID = null.StringFrom(googleDetails.GoogleID)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.GoogleID))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to add user's google")
		return terror.Error(err, "Unable to update user details.")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Added Google account to User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

// HubKeyUserRemoveTwitch removes a linked Twitch account
const HubKeyUserRemoveTwitch = "USER:TWITCH:REMOVE"

func (uc *UserController) RemoveTwitchHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &RemoveServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	// Setup user activity tracking
	var oldUser types.User = *user

	// Check if user can remove service
	serviceCount := getUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(fmt.Errorf("You cannot unlink your only connection to this account"))
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	user.TwitchID = null.NewString("", false)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwitchID))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Removed Twitch account from User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

// HubKeyUserRemoveTwitch adds a linked Twitch account
const HubKeyUserAddTwitch = "USER:TWITCH:ADD"

type AddTwitchRequest struct {
	Payload struct {
		Token   string `json:"token"`
		Website bool   `json:"website"`
	} `json:"payload"`
}

func (uc *UserController) AddTwitchHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating user's Twitch ID, try again or contact support."
	req := &AddTwitchRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.Token == "" {
		return terror.Error(terror.ErrInvalidInput, errMsg)
	}

	twitchID := ""
	if req.Payload.Website {
		keySet := oidc.NewRemoteKeySet(ctx, "https://id.twitch.tv/oauth2/keys")
		oidcVerifier := oidc.NewVerifier("https://id.twitch.tv/oauth2", keySet, &oidc.Config{
			ClientID: uc.Twitch.ClientID,
		})

		token, err := oidcVerifier.Verify(ctx, req.Payload.Token)
		if err != nil {
			return terror.Error(err, errMsg)
		}

		var claims struct {
			Iss   string `json:"iss"`
			Sub   string `json:"sub"`
			Aud   string `json:"aud"`
			Exp   int32  `json:"exp"`
			Iat   int32  `json:"iat"`
			Nonce string `json:"nonce"`
			Email string `json:"email"`
		}
		if err := token.Claims(&claims); err != nil {
			return terror.Error(err, errMsg)
		}

		twitchID = claims.Sub

	} else {
		claims, err := uc.API.GetClaimsFromTwitchExtensionToken(req.Payload.Token)
		if err != nil {
			return terror.Error(err, errMsg)
		}

		if !strings.HasPrefix(claims.OpaqueUserID, "U") {
			return terror.Error(terror.ErrInvalidInput, "Twitch user is not logged in, log in and try again.")
		}

		twitchID = claims.TwitchAccountID
	}

	if twitchID == "" {
		return terror.Error(terror.ErrInvalidInput, "No Twitch account ID is provided")
	}

	// Activity tracking
	var oldUser types.User = *user

	user.TwitchID = null.StringFrom(twitchID)
	// Update user's Twitch ID
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwitchID))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Added Twitch account to User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

// HubKeyUserRemoveTwitter removes a linked Twitter account
const HubKeyUserRemoveTwitter = "USER:TWITTER:REMOVE"

func (uc *UserController) RemoveTwitterHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue removing user's twitter account, try again or contact support."
	// Setup user activity tracking
	var oldUser types.User = *user

	// Check if user can remove service
	serviceCount := getUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(fmt.Errorf("You cannot unlink your only connection to this account"))
	}

	// Update user
	user.TwitterID = null.NewString("", false)
	_, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwitterID))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to remove user's twitter")
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Removed Twitter account from User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

type AddTwitterRequest struct {
	Payload struct {
		OAuthToken    string `json:"oauth_token"`
		OAuthVerifier string `json:"oauth_verifier"`
	} `json:"payload"`
}

// HubKeyUserRemoveTwitter adds a linked Twitter account
const HubKeyUserAddTwitter = "USER:TWITTER:ADD"

func (uc *UserController) AddTwitterHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating user's twitter account, try again or contact support."
	req := &AddTwitterRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.OAuthToken == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitter OAuth token is empty.")
	}
	if req.Payload.OAuthVerifier == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitter OAuth verifier is empty.")
	}

	params := url.Values{}
	params.Set("oauth_token", req.Payload.OAuthToken)
	params.Set("oauth_verifier", req.Payload.OAuthVerifier)
	twitterDetails, err := uc.API.TwitterToken(params.Encode())
	if err != nil {
		return terror.Error(err, "Invalid twitter token provided.")
	}

	twitterID := twitterDetails.TwitterID
	if twitterID == "" {
		return terror.Error(terror.ErrInvalidInput, "No Twitter account ID is provided.")
	}

	// Activity tracking
	var oldUser types.User = *user

	// Update user's Twitter ID
	user.TwitterID = null.StringFrom(twitterID)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwitterID))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to add user's twitter")
		return terror.Error(err, errMsg)
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

// HubKeyUserRemoveDiscord removes a linked Discord account
const HubKeyUserRemoveDiscord = "USER:DISCORD:REMOVE"

func (uc *UserController) RemoveDiscordHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue removing user's discord account, try again or contact support."
	req := &RemoveServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required.")
	}

	// Setup user activity tracking
	var oldUser types.User = *user

	// Check if user can remove service
	serviceCount := getUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(fmt.Errorf("You cannot unlink your only connection to this account"))
	}

	// Update user
	user.DiscordID = null.NewString("", false)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.DiscordID))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Removed Discord account from User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

// HubKeyUserRemoveDiscord adds a linked Discord account
const HubKeyUserAddDiscord = "USER:DISCORD:ADD"

type AddDiscordRequest struct {
	Payload struct {
		Code        string `json:"code"`
		RedirectURI string `json:"redirect_uri"`
	} `json:"payload"`
}

func (uc *UserController) AddDiscordHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue adding user's discord account, try again or contact support."
	req := &AddDiscordRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.Code == "" {
		return terror.Error(terror.ErrInvalidInput, errMsg)
	}

	// Validate Discord code and get access token
	data := url.Values{}
	data.Set("code", req.Payload.Code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", req.Payload.RedirectURI)

	client := &http.Client{}
	req1, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	req1.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(uc.Discord.ClientID+":"+uc.Discord.ClientSecret)))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return terror.Error(err, errMsg)
	}
	res, err := client.Do(req1)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer res.Body.Close()

	resp := &struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}{}
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return terror.Error(err, "Failed to authenticate user with Discord.")
	}

	// Get Discord user using access token
	client = &http.Client{}
	req2, err := http.NewRequest("GET", "https://discord.com/api/oauth2/@me", nil)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	req2.Header.Set("Authorization", "Bearer "+resp.AccessToken)
	res2, err := client.Do(req2)
	if err != nil {
		return terror.Error(err, "Failed to get user with access token.")
	}
	defer res2.Body.Close()

	resp2 := &struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}{}
	err = json.NewDecoder(res2.Body).Decode(resp2)
	if err != nil {
		return terror.Error(err, "Failed to authenticate user with Discord.")
	}

	discordID := resp2.User.ID
	if discordID == "" {
		return terror.Error(terror.ErrInvalidInput, errMsg)
	}

	// Activity tracking
	var oldUser types.User = *user

	// Update user's Discord ID
	user.DiscordID = null.StringFrom(discordID)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.DiscordID))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Added Discord account to User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

// HubKeyUserRemoveWallet removes a linked wallet address
const HubKeyUserRemoveWallet = "USER:WALLLET:REMOVE"

// RemoveWalletRequest requests an update for an existing user
type RemoveWalletRequest struct {
	Payload struct {
		ID       types.UserID `json:"id"`
		Username string       `json:"username"`
	} `json:"payload"`
}

// RemoveWalletHandler removes a linked wallet address
func (uc *UserController) RemoveWalletHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue removing user's wallet address, try again or contact support."
	req := &RemoveWalletRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required.")
	}

	// Setup user activity tracking
	var oldUser types.User = *user

	// Check if user can remove service
	serviceCount := getUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(fmt.Errorf("You cannot unlink your only connection to this account"))
	}

	// Update user
	user.PublicAddress = null.NewString("", false)
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.PublicAddress))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Updated User",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

// HubKeyUserAddWallet links a wallet to an account
const HubKeyUserAddWallet = "USER:WALLET:ADD"

type AddWalletRequest struct {
	Payload struct {
		ID            types.UserID `json:"id"`
		Username      string       `json:"username"`
		PublicAddress string       `json:"public_address"`
		Signature     string       `json:"signature"`
		Nonce         string       `json:"nonce"`
	} `json:"payload"`
}

// AddWalletHandler links a wallet address to a user
func (uc *UserController) AddWalletHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue adding user's wallet address, try again or contact support."
	req := &AddWalletRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}
	if req.Payload.ID.IsNil() && req.Payload.Username == "" {
		return terror.Error(terror.ErrInvalidInput, "User ID or Username is required.")
	}

	if req.Payload.PublicAddress == "" {
		return terror.Error(terror.ErrInvalidInput, "Public Address is required.")
	}
	if req.Payload.Signature == "" {
		return terror.Error(terror.ErrInvalidInput, "Signature is required.")
	}

	// Setup user activity tracking
	var oldUser = *user

	publicAddr := common.HexToAddress(req.Payload.PublicAddress)

	// verify they signed it
	if user.Nonce.String == "" {
		user.Nonce = null.StringFrom(req.Payload.Nonce)
		_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Nonce))
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}
	err = uc.API.VerifySignature(req.Payload.Signature, user.Nonce.String, publicAddr)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Update user
	hexPublicAddress := publicAddr.Hex()
	user.PublicAddress = null.StringFrom(hexPublicAddress)

	// Check public address is hex address
	if !common.IsHexAddress(hexPublicAddress) {
		passlog.L.Error().Err(err).Msg("Public address provided is not a hex address")
		return terror.Error(err, "failed to provide a valid wallet address")
	}

	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.PublicAddress))
	if err != nil {
		return terror.Error(err, "Wallet is already connected to another user.")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Add User Wallet",
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

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

const HubKeyUser = "USER"

// UpdatedSubscribeRequest to subscribe to user updates
type UpdatedSubscribeRequest struct {
	Payload struct {
		ID       types.UserID `json:"id"`
		Username string       `json:"username"`
	} `json:"payload"`
}

func (uc *UserController) UpdatedSubscribeHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &UpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	reply(user)
	return nil
}

const HubKeyUserSupsSubscribe = "USER:SUPS:SUBSCRIBE"

func (api *API) UserSupsUpdatedSubscribeHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	sups, _, err := api.userCacheMap.Get(user.AccountID)
	// get current on world sups
	if err != nil {
		return terror.Error(err, "Issue subscribing to user SUPs updates, try again or contact support.")
	}

	reply(sups.StringFixed(0))
	return nil
}

type UserFactionDetail struct {
	RecruitID      string          `json:"recruit_id"`
	SupsEarned     decimal.Decimal `json:"sups_earned"`
	Rank           string          `json:"rank"`
	SpectatedCount int64           `json:"spectated_count"`

	// faction detail
	FactionID        string      `json:"faction_id"`
	LogoBlobID       string      `json:"logo_blob_id" db:"logo_blob_id"`
	BackgroundBlobID string      `json:"background_blob_id" db:"background_blob_id"`
	Theme            btypes.JSON `json:"theme,omitempty"`
}

const HubKeyUserFactionSubscribe = "USER:FACTION:SUBSCRIBE"

func (uc *UserController) UserFactionUpdatedSubscribeHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue subscribing to user faction updates, try again or contact support."
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// get user faction
	faction, err := boiler.FindFaction(passdb.StdConn, user.FactionID.String)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return terror.Error(err, errMsg)
	}

	f := &UserFactionDetail{
		RecruitID:        "3000",
		SupsEarned:       decimal.Zero,
		Rank:             "100",
		SpectatedCount:   100,
		FactionID:        faction.ID,
		Theme:            faction.Theme,
		LogoBlobID:       faction.LogoBlobID,
		BackgroundBlobID: faction.BackgroundBlobID,
	}

	if faction != nil {
		reply(f)
	}
	return nil
}

type WarMachineQueuePositionRequest struct {
	Payload struct {
		AssetHash string `json:"asset_hash"`
	} `json:"payload"`
}

// getUserServiceCount returns the amount of services (email, facebook, google, discord etc.) the user is currently connected to
func getUserServiceCount(user *types.User) int {
	count := 0
	if user.Email.Valid {
		hasPassword, _ := boiler.PasswordHashExists(passdb.StdConn, user.ID)
		if hasPassword {
			count++
		}
	}
	if user.FacebookID.Valid {
		count++
	}
	if user.GoogleID.Valid {
		count++
	}
	if user.TwitchID.Valid {
		count++
	}
	if user.TwitterID.Valid {
		count++
	}
	if user.DiscordID.Valid {
		count++
	}
	if user.PublicAddress.Valid {
		count++
	}

	return count
}

const HubKeySUPSRemainingSubscribe = "SUPS:TREASURY"

func (uc *UserController) TotalSupRemainingHandler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	usr, err := boiler.FindUser(passdb.StdConn, types.XsynSaleUserID.String())
	if err != nil {
		terror.Error(err, "Issue getting total SUPs remaining handler, try again or contact support.")
	}

	sups, _, err := uc.API.userCacheMap.Get(usr.AccountID)
	if err != nil {
		return terror.Error(err, "Issue getting total SUPs remaining handler, try again or contact support.")
	}
	reply(sups.StringFixed(0))
	return nil
}

const HubKeyUserTransactionsSubscribe = "USER:SUPS:TRANSACTIONS:SUBSCRIBE"

func (api *API) UserTransactionsSubscribeHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	// get users transactions
	list, err := db.UserTransactionGetList(user.AccountID, 5)
	if err != nil {
		return terror.Error(err, "Failed to get transactions, try again or contact support.")
	}
	reply(list)
	return nil
}

type UserFingerprintRequest struct {
	Payload struct {
		Fingerprint auth.Fingerprint `json:"fingerprint"`
	} `json:"payload"`
}

const HubKeyUserLock = "USER:LOCK"

type UserLockRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Type string `json:"type"`
	} `json:"payload"`
}

// LockHandler return updates user table to lock account according to requested level
func (uc *UserController) LockHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	req := &UserLockRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	if req.Payload.Type == "account" {
		if user.TotalLock {
			return terror.Error(fmt.Errorf("user: %s: has already locked account", user.ID), "Account is already locked.")
		}

		user.TotalLock = true
		user.WithdrawLock = true
		user.MintLock = true
	}

	if req.Payload.Type == "minting" {
		if user.MintLock {
			return terror.Error(fmt.Errorf("user: %s: has already locked minting", user.ID), "Minting is already locked.")
		}

		user.MintLock = true
	}

	if req.Payload.Type == "withdrawals" {
		if user.WithdrawLock {
			return terror.Error(fmt.Errorf("user: %s: has already locked withdrawals", user.ID), "Withdrawals is already locked.")
		}

		user.WithdrawLock = true
	}

	columns, err := user.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return terror.Error(err, "Could not update account lock settings.")
	}
	if columns < 1 {
		return terror.Error(fmt.Errorf("Did not update user columns"), "Could not update account lock settings.")
	}

	reply(true)

	return nil
}

const HubKeyGenerateTFASecret = "USER:TFA:GENERATE"

type GenerateTFAResponse struct {
	Secret    string `json:"secret"`
	QRCodeStr string `json:"qr_code_str"`
}

// TFAVerificationHandler generate QR code and secret of user
func (uc *UserController) GenerateTFAHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {

	if user.TwoFactorAuthenticationIsSet {
		return terror.Error(fmt.Errorf("Two-Factor Authentication already set"))
	}

	otpKey, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "XSYN",
		AccountName: user.Username,
	})

	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to generate 2fa secret")
		return terror.Error(err, "Failed to generate two factor authentication secret")
	}

	user.TwoFactorAuthenticationSecret = otpKey.Secret()

	oldUser := *user

	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwoFactorAuthenticationSecret))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to update 2fa secret")
		return terror.Error(err, "Failed to update user 2fa secret")
	}

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Generate TFA secret",
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

	reply(&GenerateTFAResponse{
		Secret:    otpKey.Secret(),
		QRCodeStr: otpKey.String(),
	})

	return nil
}

const HubKeyTFACancel = "USER:TFA:CANCEL"

// TFAVerificationHandler cancles TFA of user
func (uc *UserController) CancelTFAHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue updating user details, try again or contact support."
	oldUser := *user

	// reset 2fa flag
	user.TwoFactorAuthenticationIsSet = false
	_, err := user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwoFactorAuthenticationIsSet))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to unset 2fa for user")
		return terror.Error(err, errMsg)
	}

	// clear 2fa secret
	user.TwoFactorAuthenticationSecret = ""
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwoFactorAuthenticationSecret))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to clear user 2fa secret")
		return terror.Error(err, errMsg)
	}

	// delete recovery code
	_, err = boiler.UserRecoveryCodes(boiler.UserRecoveryCodeWhere.UserID.EQ(user.ID)).UpdateAll(passdb.StdConn, boiler.M{
		boiler.IssueTokenColumns.DeletedAt: time.Now(),
	})

	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to delete user recovery codes")
		return terror.Error(err, errMsg)
	}

	// update 2fa flag
	user.TwoFactorAuthenticationActivated = false
	_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwoFactorAuthenticationActivated))
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to remove user 2fa activated")
		return terror.Error(err, errMsg)
	}

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Remove User 2FA",
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

	reply(user)

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

const HubKeyTFAVerification = "USER:TFA:VERIFICATION"

type TFAVerificationRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Passcode string `json:"passcode"`
	} `json:"payload"`
}

// TFAVerificationHandler handles user tfa verification
func (uc *UserController) TFAVerificationHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Unable to verify two-factor authentication, try again or contact support."
	req := &TFAVerificationRequest{}
	err := json.Unmarshal(payload, req)

	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to parse 2fa verify request")
		return terror.Error(err, errMsg)
	}

	oldUser := *user

	err = users.VerifyTFA(user.TwoFactorAuthenticationSecret, req.Payload.Passcode)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// set user's tfa status
	if !user.TwoFactorAuthenticationIsSet {
		// Get recovery code
		nRecoverCodes, _ := boiler.UserRecoveryCodes(boiler.UserRecoveryCodeWhere.UserID.EQ(user.ID)).Count(passdb.StdConn)
		if nRecoverCodes > 0 {
			return terror.Error(err, "User already has recovery codes")
		}

		// Set TFA
		user.TwoFactorAuthenticationIsSet = true
		_, err = user.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.TwoFactorAuthenticationIsSet))
		if err != nil {
			passlog.L.Error().Err(err).Msg("failed to update user 2fa is set")
			return terror.Error(err, errMsg)
		}

		// generate recovery code
		for i := 0; i < 16; i++ {
			petname.NonDeterministicMode()
			code := petname.Generate(2, "-")

			recoveryCode := &boiler.UserRecoveryCode{
				RecoveryCode: code,
				UserID:       user.ID,
			}
			err = recoveryCode.Insert(passdb.StdConn, boil.Infer())
			if err != nil {
				return terror.Error(err, errMsg)
			}
		}

		// Record user activity
		uc.API.RecordUserActivity(ctx,
			user.ID,
			"Set User TFA and recovery code",
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
	}

	// send back user
	reply(user)

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}

const HubKeyTFARecoveryGet = "USER:TFA:RECOVERY:GET"

// Get recover code
func (uc *UserController) TFARecoveryCodeGetHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	// Get recovery code
	userRecoveryCode, err := users.GetTFARecovery(user.ID)
	if err != nil {
		return terror.Error(err, "User has no recovery codes.")
	}

	// send back user
	reply(userRecoveryCode)

	return nil
}

// HubKeyTFARecovery is the key used to run the AuthLogin handler
const HubKeyTFARecoveryVerify = "USER:TFA:RECOVERY:VERIFY"

// TFARecoveryResquest is a request to recover users' own 2fa
type TFARecoveryRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		RecoveryCode string `json:"recoveryCode"`
	} `json:"payload"`
}

// TFARecoveryVerifyHandler will verify recovery code and give user access once and use up the recovery code
func (uc *UserController) TFARecoveryVerifyHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue verifying TFA, try again or contact support."
	req := &TFARecoveryRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		passlog.L.Error().Err(err).Msg("failed to parse 2fa recovery code verify request")
		return terror.Error(err, errMsg)
	}

	oldUser := *user

	tx, err := passdb.StdConn.Begin()
	if err != nil {
		return terror.Error(err, errMsg)
	}

	defer tx.Rollback()

	// Check if code matches
	err = users.VerifyTFARecovery(user.ID, req.Payload.RecoveryCode)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	err = tx.Commit()
	if err != nil {
		return terror.Error(err, errMsg)
	}
	// Record user activity
	uc.API.RecordUserActivity(ctx,
		user.ID,
		"Remove User 2FA",
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
	reply(user)

	ws.PublishMessage("/user/"+user.ID, HubKeyUser, user)

	return nil
}
