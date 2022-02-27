package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/microcosm-cc/bluemonday"
	"io/ioutil"
	"net/http"
	"net/url"
	"passport"
	"passport/crypto"
	"passport/db"
	"passport/helpers"
	"strings"
	"time"

	// "github.com/apex/log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/sale/dispersions"

	oidc "github.com/coreos/go-oidc"
	"github.com/jackc/pgx/v4"

	"github.com/volatiletech/null/v8"
	"google.golang.org/api/idtoken"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/auth"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/ninja-syndicate/supremacy-bridge/bridge"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

// UserController holds handlers for authentication
type UserController struct {
	Conn    *pgxpool.Pool
	Log     *zerolog.Logger
	API     *API
	Google  *auth.GoogleConfig
	Twitch  *auth.TwitchConfig
	Discord *auth.DiscordConfig
}

// NewUserController creates the user hub
func NewUserController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, googleConfig *auth.GoogleConfig, twitchConfig *auth.TwitchConfig, discordConfig *auth.DiscordConfig) *UserController {
	userHub := &UserController{
		Conn:    conn,
		Log:     log_helpers.NamedLogger(log, "user_hub"),
		API:     api,
		Google:  googleConfig,
		Twitch:  twitchConfig,
		Discord: discordConfig,
	}

	api.Command(HubKeyUserGet, userHub.GetHandler) // Perm check inside handler (users can get themselves; need UserRead permission to get other users)
	api.SecureCommand(HubKeyUserUpdate, userHub.UpdateHandler)
	api.SecureCommand(HubKeyUserUsernameUpdate, userHub.UpdateUserUsernameHandler)
	api.SecureCommand(HubKeyUserFactionUpdate, userHub.UpdateUserFactionHandler) // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveFacebook, userHub.RemoveFacebookHandler)   // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddFacebook, userHub.AddFacebookHandler)         // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveGoogle, userHub.RemoveGoogleHandler)       // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddGoogle, userHub.AddGoogleHandler)             // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveTwitch, userHub.RemoveTwitchHandler)       // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddTwitch, userHub.AddTwitchHandler)             // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveTwitter, userHub.RemoveTwitterHandler)     // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddTwitter, userHub.AddTwitterHandler)           // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveDiscord, userHub.RemoveDiscordHandler)     // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddDiscord, userHub.AddDiscordHandler)           // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveWallet, userHub.RemoveWalletHandler)       // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddWallet, userHub.AddWalletHandler)             // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserCreate, userHub.CreateHandler)
	api.SecureCommandWithPerm(HubKeyUserList, userHub.ListHandler, passport.PermUserList)
	api.SecureCommandWithPerm(HubKeyUserArchive, userHub.ArchiveHandler, passport.PermUserArchive)
	api.SecureCommandWithPerm(HubKeyUserUnarchive, userHub.UnarchiveHandler, passport.PermUserUnarchive)
	api.SecureCommandWithPerm(HubKeyUserChangePassword, userHub.ChangePasswordHandler, passport.PermUserUpdate)
	api.SecureCommandWithPerm(HubKeyUserForceDisconnect, userHub.ForceDisconnectHandler, passport.PermUserForceDisconnect)

	api.Command(HubKeyCheckCanAccessStore, userHub.CheckCanAccessStore)

	api.SubscribeCommand(HubKeyUserForceDisconnected, userHub.ForceDisconnectedHandler)
	api.SubscribeCommand(HubKeyUserSubscribe, userHub.UpdatedSubscribeHandler)
	api.SubscribeCommand(HubKeyUserOnlineStatus, userHub.OnlineStatusSubscribeHandler)
	api.SubscribeCommand(HubKeySUPSRemainingSubscribe, userHub.TotalSupRemainingHandler)
	api.SubscribeCommand(HubKeySUPSExchangeRates, userHub.ExchangeRatesHandler)

	// listen on queuing war machine
	api.SecureUserSubscribeCommand(HubKeyUserWarMachineQueuePositionSubscribe, userHub.WarMachineQueuePositionUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyUserSupsSubscribe, userHub.UserSupsUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyUserFactionSubscribe, userHub.UserFactionUpdatedSubscribeHandler)

	api.SecureUserSubscribeCommand(HubKeyBlockConfirmation, userHub.BlockConfirmationHandler)

	// sups multiplier
	api.SecureUserSubscribeCommand(HubKeyUserSupsMultiplierSubscribe, userHub.UserSupsMultiplierUpdatedSubscribeHandler)
	api.SecureUserSubscribeCommand(HubKeyUserStatSubscribe, userHub.UserStatUpdatedSubscribeHandler)

	return userHub
}

// GetUserRequest requests an update for an existing user
type GetUserRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID       passport.UserID `json:"id"`
		Username string          `json:"username"`
	} `json:"payload"`
}

// 	rootHub.SecureCommand(HubKeyUserGet, UserController.GetHandler)
const HubKeyUserGet hub.HubCommandKey = "USER:GET"

// GetHandler gets the details for a user
func (uc *UserController) GetHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &GetUserRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() && req.Payload.Username == "" {
		return terror.Error(terror.ErrInvalidInput, "User ID or username is required")
	}

	if !req.Payload.ID.IsNil() {
		user, err := db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Unable to load current user")
		}

		////// Permission check
		//if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserRead.String()) {
		//	return terror.Error(terror.ErrUnauthorised, "You do not have permission to look at other users.")
		//}

		reply(user)
		return nil
	}

	user, err := db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
	if err != nil {
		return terror.Error(err, "Unable to load current user")
	}

	//// Permission check
	//if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserRead.String()) {
	//	return terror.Error(terror.ErrUnauthorised, "You do not have permission to look at other users.")
	//}

	reply(user)
	return nil

}

type UpdateUserUsernameRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Username string `json:"username"`
	} `json:"payload"`
}

const HubKeyUserUsernameUpdate hub.HubCommandKey = "USER:USERNAME:UPDATE"

func (uc *UserController) UpdateUserUsernameHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		return terror.Error(err)
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Something went wrong. Please try again.")
	}

	// Get user
	user, err := db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "User does not exist.")
	}

	// Activity tracking
	var oldUser passport.User = *user

	// Check availability of username
	if user.Username == username {
		return terror.Error(fmt.Errorf("username cannot be same as current"), "New username cannot be the same as current username.")
	}

	isAvailable, err := db.UsernameAvailable(ctx, uc.Conn, username, &user.ID)
	if err != nil {
		return terror.Error(err, "Something went wrong. Please try again.")
	}
	if !isAvailable {
		return terror.Error(fmt.Errorf("A user with that username already exists."))
	}

	// Update username
	user.Username = username

	// Update user
	err = db.UserUpdate(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, "User does not exist.")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Updated user's username",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// UpdateUserFactionRequest requests update user faction
type UpdateUserFactionRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID    passport.UserID    `json:"userID"`
		FactionID passport.FactionID `json:"factionID"`
	} `json:"payload"`
}

// 	api.SecureCommand(HubKeyUserFactionUpdate, userHub.UpdateUserFactionHandler)
const HubKeyUserFactionUpdate hub.HubCommandKey = "USER:FACTION:UPDATE"

// GetHandler gets the details for a user
func (uc *UserController) UpdateUserFactionHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UpdateUserFactionRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.UserID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	if req.Payload.FactionID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "Faction ID is required")
	}

	user, err := db.UserGet(ctx, uc.Conn, req.Payload.UserID)
	if err != nil {
		return terror.Error(err, "Unable to load current user")
	}

	user.FactionID = &req.Payload.FactionID

	err = db.UserFactionEnlist(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err, "Unable to update user faction")
	}

	faction, err := db.FactionGet(ctx, uc.Conn, req.Payload.FactionID)
	if err != nil {
		return terror.Error(err)
	}
	user.Faction = faction

	// send user changes to connected clients
	uc.API.SendToAllServerClient(ctx, &ServerClientMessage{
		Key: UserUpdated,
		Payload: struct {
			User *passport.User `json:"user"`
		}{
			User: user,
		},
	})

	return nil
}

// HubKeyUserUpdate updates a user
const HubKeyUserUpdate hub.HubCommandKey = "USER:UPDATE"

// UpdateUserRequest requests an update for an existing user
type UpdateUserRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID                               passport.UserID          `json:"id"`
		Username                         string                   `json:"username"`
		NewUsername                      *string                  `json:"newUsername"`
		FirstName                        string                   `json:"firstName"`
		LastName                         string                   `json:"lastName"`
		Email                            null.String              `json:"email"`
		AvatarID                         *passport.BlobID         `json:"avatarID"`
		CurrentPassword                  *string                  `json:"currentPassword"`
		NewPassword                      *string                  `json:"newPassword"`
		OrganisationID                   *passport.OrganisationID `json:"organisationID"`
		TwoFactorAuthenticationActivated bool                     `json:"twoFactorAuthenticationActivated"`
	} `json:"payload"`
}

// UpdateHandler updates a user
func (uc *UserController) UpdateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UpdateUserRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}

	// Trying to update user w/ higher role than you?
	if user.ID.String() != hubc.Identifier() && (hubc.IsHigherOrSameLevel(user.Role.Tier) || !hubc.HasPermission(passport.PermUserUpdate.String())) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update this user")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Update Values
	confirmPassword := false
	if req.Payload.Email.Valid {
		email := strings.TrimSpace(req.Payload.Email.String)
		email = strings.ToLower(email)

		if user.Email.String != email {
			user.Email = null.StringFrom(email)
		}
	}
	if req.Payload.NewUsername != nil && req.Payload.Username != *req.Payload.NewUsername {
		user.Username = *req.Payload.NewUsername
	}
	if req.Payload.NewPassword != nil && *req.Payload.NewPassword != "" {
		if user.Email.String == "" && req.Payload.Email.String == "" {
			return terror.Error(terror.ErrInvalidInput, "Email is required when assigning a new password to this user")
		}

		err = helpers.IsValidPassword(*req.Payload.NewPassword)
		if err != nil {
			passwordErr := err.Error()
			var bErr *terror.TError
			if errors.As(err, &bErr) {
				passwordErr = bErr.Message
			}
			return terror.Error(err, passwordErr)
		}

		hasPassword, err := db.UserHasPassword(ctx, uc.Conn, user)
		if err != nil {
			return terror.Error(err, "Something went wrong. Please try again.")
		}
		confirmPassword = req.Payload.ID.String() == hubc.Identifier() && user.OldPasswordRequired && *hasPassword
	}

	if confirmPassword {
		if req.Payload.CurrentPassword == nil {
			return terror.Error(terror.ErrInvalidInput, "Current Password is required")
		}
		hashB64, err := db.HashByUserID(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Current password is incorrect")
		}
		err = crypto.ComparePassword(hashB64, *req.Payload.CurrentPassword)
		if err != nil {
			return terror.Error(err, "Current password is incorrect")
		}
	}

	if req.Payload.FirstName != "" {
		user.FirstName = req.Payload.FirstName
	}
	if req.Payload.LastName != "" {
		user.LastName = req.Payload.LastName
	}
	user.AvatarID = req.Payload.AvatarID

	// Start transaction
	errMsg := "Unable to update user, please try again."
	tx, err := uc.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			uc.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	if req.Payload.ID.String() == hubc.Identifier() && user.TwoFactorAuthenticationActivated != req.Payload.TwoFactorAuthenticationActivated {
		// if turn off 2fa
		if !req.Payload.TwoFactorAuthenticationActivated {
			userUUID, err := uuid.FromString(hubc.Identifier())
			if err != nil {
				return terror.Error(err, errMsg)
			}
			userID := passport.UserID(userUUID)
			// reset 2fa flag
			err = db.UserUpdate2FAIsSet(ctx, tx, userID, false)
			if err != nil {
				return terror.Error(err)
			}

			// clear 2fa secret
			err = db.User2FASecretSet(ctx, tx, userID, "")
			if err != nil {
				return terror.Error(err)
			}

			// delete recovery code
			err = db.UserDeleteRecoveryCode(ctx, tx, userID)
			if err != nil {
				return terror.Error(err)
			}
		}

		user.TwoFactorAuthenticationActivated = req.Payload.TwoFactorAuthenticationActivated
	}

	// Update user
	err = db.UserUpdate(ctx, tx, user)
	if err != nil {
		return terror.Error(err)
	}

	// Update password?
	if req.Payload.NewPassword != nil {
		err = db.AuthSetPasswordHash(ctx, tx, req.Payload.ID, crypto.HashPassword(*req.Payload.NewPassword))
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	// Set Organisation
	if req.Payload.OrganisationID != nil {
		err = db.UserSetOrganisations(ctx, tx, user.ID, *req.Payload.OrganisationID)
		if err != nil {
			return terror.Error(err)
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, uc.Conn, *user.FactionID)
		if err != nil {
			return terror.Error(err)
		}
		user.Faction = faction
	}

	reply(user)
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Updated User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	// send user changes to connected clients
	uc.API.SendToAllServerClient(ctx, &ServerClientMessage{
		Key: UserUpdated,
		Payload: struct {
			User *passport.User `json:"user"`
		}{
			User: user,
		},
	})

	return nil
}

// UpdateUserSupsRequest requests an update for an existing user
type UpdateUserSupsRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		UserID     passport.UserID `json:"userID"`
		SupsChange int64           `json:"supsChange"`
	} `json:"payload"`
}

// HubKeyUserCreate creates a user
const HubKeyUserCreate hub.HubCommandKey = "USER:CREATE"

// CreateUserRequest requests an create for an existing user
type CreateUserRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		FirstName      string                   `json:"firstName"`
		LastName       string                   `json:"lastName"`
		Email          null.String              `json:"email"`
		AvatarID       *passport.BlobID         `json:"avatarID"`
		NewPassword    *string                  `json:"newPassword"`
		RoleID         passport.RoleID          `json:"roleID"`
		OrganisationID *passport.OrganisationID `json:"organisationID"`
	} `json:"payload"`
}

// CreateHandler creates a user
func (uc *UserController) CreateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &CreateUserRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
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
	if req.Payload.RoleID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "Role is required.")
	}
	if req.Payload.NewPassword == nil {
		return terror.Error(terror.ErrInvalidInput, "Password is required.")
	}
	err = helpers.IsValidPassword(*req.Payload.NewPassword)
	if err != nil {
		return terror.Error(err)
	}

	// Start transaction
	errMsg := "Unable to create user, please try again."
	tx, err := uc.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			uc.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	// Create user
	user := &passport.User{
		FirstName: req.Payload.FirstName,
		LastName:  req.Payload.LastName,
		Email:     req.Payload.Email,
		RoleID:    req.Payload.RoleID,
	}
	if req.Payload.AvatarID != nil {
		user.AvatarID = req.Payload.AvatarID
	}

	err = db.UserCreate(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Set password
	err = db.AuthSetPasswordHash(ctx, tx, user.ID, crypto.HashPassword(*req.Payload.NewPassword))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Set Organisation
	if req.Payload.OrganisationID != nil {
		err = db.UserSetOrganisations(ctx, tx, user.ID, *req.Payload.OrganisationID)
		if err != nil {
			return terror.Error(err)
		}
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Created User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: nil,
			To:   user,
		},
	)

	return nil
}

// HubKeyIntakeList is a hub key to run list user intake
const HubKeyUserList hub.HubCommandKey = "USER:LIST"

// ListHandlerRequest requests holds the filter for user list
type ListHandlerRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir          `json:"sortDir"`
		SortBy   db.UserColumn         `json:"sortBy"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Archived bool                  `json:"archived"`
		Search   string                `json:"search"`
		PageSize int                   `json:"pageSize"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// UserListResponse is the response from get user list
type UserListResponse struct {
	Records []*passport.User `json:"records"`
	Total   int              `json:"total"`
}

// ListHandler lists users with pagination
func (uc *UserController) ListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Something went wrong, please try again."

	req := &ListHandlerRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	users := []*passport.User{}
	total, err := db.UserList(
		ctx, uc.Conn, &users,
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

type UserArchiveRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID passport.UserID `json:"id"`
	} `json:"payload"`
}

const (
	// HubKeyUserArchive archives the user
	HubKeyUserArchive hub.HubCommandKey = hub.HubCommandKey("USER:ARCHIVE")

	// HubKeyUserUnarchive unarchives the user
	HubKeyUserUnarchive hub.HubCommandKey = hub.HubCommandKey("USER:UNARCHIVE")
)

// ArchiveHandler archives a user
func (uc *UserController) ArchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UserArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Failed to unmarshal data")
	}
	err = db.UserArchiveUpdate(ctx, uc.Conn, req.Payload.ID, true)
	if err != nil {
		return terror.Error(err, "Issue while updating User, please try again.")
	}

	// Return user
	user, err := db.UserGet(ctx, uc.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err)
	}
	reply(user)

	// Record user activity
	if err == nil {
		uc.API.RecordUserActivity(ctx,
			hubc.Identifier(),
			"Archived User",
			passport.ObjectTypeUser,
			helpers.StringPointer(user.ID.String()),
			&user.Username,
			helpers.StringPointer(user.FirstName+" "+user.LastName),
		)
	}

	return nil
}

// UnarchiveHandler unarchives a user
func (uc *UserController) UnarchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UserArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Failed to unmarshal data")
	}
	err = db.UserArchiveUpdate(ctx, uc.Conn, req.Payload.ID, false)
	if err != nil {
		return terror.Error(err, "Issue while updating User, please try again.")
	}

	// Return user
	user, err := db.UserGet(ctx, uc.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err)
	}
	reply(user)

	//// Record user activity
	if err == nil {
		uc.API.RecordUserActivity(ctx,
			hubc.Identifier(),
			"Unarchived User",
			passport.ObjectTypeUser,
			helpers.StringPointer(user.ID.String()),
			&user.Username,
			helpers.StringPointer(user.FirstName+" "+user.LastName),
		)
	}

	return nil
}

// HubKeyUserChangePassword updates a user
const HubKeyUserChangePassword hub.HubCommandKey = "USER:CHANGE_PASSWORD"

// UserChangePasswordRequest requests an update for an existing user
type UserChangePasswordRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID          passport.UserID `json:"id"`
		NewPassword string          `json:"newPassword"`
	} `json:"payload"`
}

// ChangePasswordHandler
func (uc *UserController) ChangePasswordHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UserChangePasswordRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID == passport.UserID(uuid.Nil) {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	user, err := db.UserGet(ctx, uc.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, "Unable to load current user")
	}

	// Trying to update user w/ higher role than you?
	if user.ID.String() != hubc.Identifier() && hubc.IsHigherOrSameLevel(user.Role.Tier) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update this user")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Validate
	if req.Payload.NewPassword == "" {
		return terror.Error(terror.ErrInvalidInput, "New Password is required")
	}
	err = helpers.IsValidPassword(req.Payload.NewPassword)
	if err != nil {
		passwordErr := err.Error()
		var bErr *terror.TError
		if errors.As(err, &bErr) {
			passwordErr = bErr.Message
		}
		return terror.Error(err, passwordErr)
	}

	// Update Password
	errMsg := "Unable to update user, please try again."

	tx, err := uc.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			uc.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	err = db.AuthSetPasswordHash(ctx, tx, req.Payload.ID, crypto.HashPassword(req.Payload.NewPassword))
	if err != nil {
		return terror.Error(err, errMsg)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(true)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Changed User Password",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)
	return nil
}

// HubKeyUserForceDisconnect to force disconnect a user and invalidate their tokens
const HubKeyUserForceDisconnect hub.HubCommandKey = "USER:FORCE_DISCONNECT"

// UserForceDisconnectRequest requests to force disconnect a user and invalidate their tokens
type UserForceDisconnectRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID passport.UserID `json:"id"`
	} `json:"payload"`
}

// ForceDisconnectHandler to force disconnect a user and invalidate their tokens
func (uc *UserController) ForceDisconnectHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UserForceDisconnectRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID == passport.UserID(uuid.Nil) {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	if req.Payload.ID.String() == hubc.Identifier() {
		return terror.Error(terror.ErrForbidden, "You cannot force disconnect yourself")
	}

	user, err := db.UserGet(ctx, uc.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, "Unable to load current user")
	}

	// Trying to disconnect user w/ higher role than you?
	if user.ID.String() != hubc.Identifier() && hubc.IsHigherOrSameLevel(user.Role.Tier) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to force disconnect this user")
	}

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserForceDisconnected, user.ID.String())), nil)
	reply(true)

	// Delete issue tokens
	err = db.AuthRemoveTokensFromUserID(ctx, uc.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err)
	}

	//Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Force Disconnected User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
	)
	return nil
}

type ForceDisconnectRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID passport.UserID `json:"id"`
	} `json:"payload"`
}

const HubKeyUserForceDisconnected hub.HubCommandKey = "USER:FORCE_DISCONNECTED"

// ForceDisconnectedHandler subscribes a user to force disconnected messages
func (uc *UserController) ForceDisconnectedHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &ForceDisconnectRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return "", "", terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID == passport.UserID(uuid.Nil) {
		return "", "", terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserForceDisconnected, req.Payload.ID.String())), nil
}

// HubKeyUserOnlineStatus subscribes to a user's online status (returns boolean)
const HubKeyUserOnlineStatus hub.HubCommandKey = "USER:ONLINE_STATUS"

// HubKeyUserOnlineStatusRequest to subscribe to user online status changes
type HubKeyUserOnlineStatusRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID       passport.UserID `json:"id"`
		Username string          `json:"username"` // Optional username instead of id
	} `json:"payload"`
}

// OnlineStatusSubscribeHandler to subscribe to user online status changes
func (uc *UserController) OnlineStatusSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &HubKeyUserOnlineStatusRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	userID := req.Payload.ID
	if userID.IsNil() && req.Payload.Username == "" {
		return req.TransactionID, "", terror.Error(terror.ErrInvalidInput, "User ID or username is required")
	}
	if userID.IsNil() {
		id, err := db.UserIDFromUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return req.TransactionID, "", terror.Error(err, "Unable to load current user")
		}
		userID = *id
	}

	if userID.IsNil() {
		return req.TransactionID, "", terror.Error(fmt.Errorf("userID is still nil for %s %s", req.Payload.ID, req.Payload.Username), "Unable to load current user")
	}

	// get current online status
	online := false
	uc.API.Hub.Clients(func(sessionID hub.SessionID, cl *hub.Client) bool {
		if cl.Identifier() == userID.String() {
			online = true
			return false
		}
		return true
	})

	reply(online)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserOnlineStatus, userID.String())), nil
}

type RemoveServiceRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID       passport.UserID `json:"id"`
		Username string          `json:"username"`
	} `json:"payload"`
}

type AddServiceRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Token string `json:"token"`
	} `json:"payload"`
}

// HubKeyUserRemoveFacebook removes a linked Facebook account
const HubKeyUserRemoveFacebook hub.HubCommandKey = "USER:REMOVE_FACEBOOK"

func (uc *UserController) RemoveFacebookHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &RemoveServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}

	//// Permission check
	if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserUpdate.String()) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update other users.")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Check if user can remove service
	serviceCount := GetUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(terror.ErrForbidden, "You cannot unlink your only connection to this account.")
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	err = db.UserRemoveFacebook(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Removed Facebook account from User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserAddFacebook removes a linked Facebook account
const HubKeyUserAddFacebook hub.HubCommandKey = "USER:ADD_FACEBOOK"

func (uc *UserController) AddFacebookHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AddServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Token == "" {
		return terror.Error(terror.ErrInvalidInput, "Facebook token is empty")
	}

	// Validate Facebook token
	errMsg := "There was a problem finding a user associated with the provided Facebook account, please check your details and try again."
	r, err := http.Get("https://graph.facebook.com/me?&access_token=" + url.QueryEscape(req.Payload.Token))
	if err != nil {
		return terror.Error(err)
	}
	defer r.Body.Close()
	resp := &struct {
		ID string `json:"id"`
	}{}
	err = json.NewDecoder(r.Body).Decode(resp)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not convert user ID to UUID")
	}

	// Get user
	user, err := db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "failed to query user")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Update user's Facebook ID
	err = db.UserAddFacebook(ctx, uc.Conn, user, resp.ID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Added Facebook account to User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	return nil
}

// HubKeyUserRemoveGoogle removes a linked Google account
const HubKeyUserRemoveGoogle hub.HubCommandKey = "USER:REMOVE_GOOGLE"

func (uc *UserController) RemoveGoogleHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &RemoveServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}

	//// Permission check
	if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserUpdate.String()) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update other users.")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Check if user can remove service
	serviceCount := GetUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(terror.ErrForbidden, "You cannot unlink your only connection to this account.")
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	err = db.UserRemoveGoogle(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Removed Google account from User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserAddGoogle adds a linked Google account
const HubKeyUserAddGoogle hub.HubCommandKey = "USER:ADD_GOOGLE"

func (uc *UserController) AddGoogleHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AddServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Token == "" {
		return terror.Error(terror.ErrInvalidInput, "Google token is empty")
	}

	// Validate Google token
	errMsg := "There was a problem finding a user associated with the provided Google account, please check your details and try again."
	resp, err := idtoken.Validate(ctx, req.Payload.Token, uc.Google.ClientID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	googleID, ok := resp.Claims["sub"].(string)
	if !ok {
		return terror.Error(err, errMsg)
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not convert user ID to UUID")
	}

	// Get user
	user, err := db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "failed to query user")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Update user's Google ID
	err = db.UserAddGoogle(ctx, uc.Conn, user, googleID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Added Google account to User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	return nil
}

// HubKeyUserRemoveTwitch removes a linked Twitch account
const HubKeyUserRemoveTwitch hub.HubCommandKey = "USER:REMOVE_TWITCH"

func (uc *UserController) RemoveTwitchHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &RemoveServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}

	//// Permission check
	if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserUpdate.String()) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update other users.")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Check if user can remove service
	serviceCount := GetUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(terror.ErrForbidden, "You cannot unlink your only connection to this account.")
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	err = db.UserRemoveTwitch(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Removed Twitch account from User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserRemoveTwitch adds a linked Twitch account
const HubKeyUserAddTwitch hub.HubCommandKey = "USER:ADD_TWITCH"

type AddTwitchRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Token   string `json:"token"`
		Website bool   `json:"website"`
	} `json:"payload"`
}

func (uc *UserController) AddTwitchHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AddTwitchRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Token == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitch JWT is empty")
	}

	twitchID := ""
	if req.Payload.Website {
		keySet := oidc.NewRemoteKeySet(ctx, "https://id.twitch.tv/oauth2/keys")
		oidcVerifier := oidc.NewVerifier("https://id.twitch.tv/oauth2", keySet, &oidc.Config{
			ClientID: uc.Twitch.ClientID,
		})

		token, err := oidcVerifier.Verify(ctx, req.Payload.Token)
		if err != nil {
			return terror.Error(err, "Failed to verify Twitch JWT")
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
			return terror.Error(err, "Failed to get claims from token")
		}

		twitchID = claims.Sub

	} else {
		claims, err := uc.API.Auth.GetClaimsFromTwitchExtensionToken(req.Payload.Token)
		if err != nil {
			return terror.Error(err, "Failed to parse twitch extension token")
		}

		if !strings.HasPrefix(claims.OpaqueUserID, "U") {
			return terror.Error(terror.ErrInvalidInput, "Twitch user is not login")
		}

		twitchID = claims.TwitchAccountID
	}

	if twitchID == "" {
		return terror.Error(terror.ErrInvalidInput, "No Twitch account ID is provided")
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not convert user ID to UUID")
	}

	// Get user
	user, err := db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	// Activity tracking
	var oldUser passport.User = *user

	// Update user's Twitch ID
	err = db.UserAddTwitch(ctx, uc.Conn, user, twitchID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Added Twitch account to User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	return nil
}

// HubKeyUserRemoveTwitter removes a linked Twitter account
const HubKeyUserRemoveTwitter hub.HubCommandKey = "USER:REMOVE_TWITTER"

func (uc *UserController) RemoveTwitterHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &RemoveServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}

	//// Permission check
	if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserUpdate.String()) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update other users.")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Check if user can remove service
	serviceCount := GetUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(terror.ErrForbidden, "You cannot unlink your only connection to this account.")
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	err = db.UserRemoveTwitter(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Removed Twitter account from User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

type AddTwitterRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		OAuthToken    string `json:"oauthToken"`
		OAuthVerifier string `json:"oauthVerifier"`
	} `json:"payload"`
}

// HubKeyUserRemoveTwitter adds a linked Twitter account
const HubKeyUserAddTwitter hub.HubCommandKey = "USER:ADD_TWITTER"

func (uc *UserController) AddTwitterHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AddTwitterRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.OAuthToken == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitter OAuth token is empty")
	}
	if req.Payload.OAuthVerifier == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitter OAuth verifier is empty")
	}

	params := url.Values{}
	params.Set("oauth_token", req.Payload.OAuthToken)
	params.Set("oauth_verifier", req.Payload.OAuthVerifier)
	client := &http.Client{}
	r, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitter.com/oauth/access_token?%s", params.Encode()), nil)
	if err != nil {
		return terror.Error(err)
	}
	res, err := client.Do(r)
	if err != nil {
		return terror.Error(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return terror.Error(err)
	}

	resp := &struct {
		OauthToken       string
		OauthTokenSecret string
		UserID           string
	}{}
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

	twitterID := resp.UserID
	if twitterID == "" {
		return terror.Error(terror.ErrInvalidInput, "No Twitter account ID is provided")
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not convert user ID to UUID")
	}

	// Get user
	user, err := db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	// Activity tracking
	var oldUser passport.User = *user

	// Update user's Twitter ID
	err = db.UserAddTwitter(ctx, uc.Conn, user, twitterID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Added Twitter account to User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	return nil
}

// HubKeyUserRemoveDiscord removes a linked Discord account
const HubKeyUserRemoveDiscord hub.HubCommandKey = "USER:REMOVE_DISCORD"

func (uc *UserController) RemoveDiscordHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &RemoveServiceRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}

	//// Permission check
	if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserUpdate.String()) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update other users.")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Check if user can remove service
	serviceCount := GetUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(terror.ErrForbidden, "You cannot unlink your only connection to this account.")
	}

	// Update user
	errMsg := "Unable to update user, please try again."
	err = db.UserRemoveDiscord(ctx, uc.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Removed Discord account from User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserRemoveDiscord adds a linked Discord account
const HubKeyUserAddDiscord hub.HubCommandKey = "USER:ADD_DISCORD"

type AddDiscordRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Code        string `json:"code"`
		RedirectURI string `json:"redirectURI"`
	} `json:"payload"`
}

func (uc *UserController) AddDiscordHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AddDiscordRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Code == "" {
		return terror.Error(terror.ErrInvalidInput, "Discord code is empty")
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
		return terror.Error(err)
	}
	res, err := client.Do(req1)
	if err != nil {
		return terror.Error(err, "Failed to verify token")
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
		return terror.Error(err)
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
		return terror.Error(terror.ErrInvalidInput, "No Discord account ID is provided")
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not convert user ID to UUID")
	}

	// Get user
	user, err := db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	// Activity tracking
	var oldUser passport.User = *user

	// Update user's Discord ID
	err = db.UserAddDiscord(ctx, uc.Conn, user, discordID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Added Discord account to User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	return nil
}

// HubKeyUserRemoveWallet removes a linked wallet address
const HubKeyUserRemoveWallet hub.HubCommandKey = "USER:REMOVE_WALLET"

// RemoveWalletRequest requests an update for an existing user
type RemoveWalletRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID       passport.UserID `json:"id"`
		Username string          `json:"username"`
	} `json:"payload"`
}

// RemoveWalletHandler removes a linked wallet address
func (uc *UserController) RemoveWalletHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &RemoveWalletRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}
	//// Permission check
	if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserUpdate.String()) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update other users.")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Start transaction
	errMsg := "Unable to update user, please try again."
	tx, err := uc.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			uc.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	// Check if user can remove service
	serviceCount := GetUserServiceCount(user)
	if serviceCount < 2 {
		return terror.Error(terror.ErrForbidden, "You cannot unlink your only connection to this account.")
	}

	// Update user
	err = db.UserRemoveWallet(ctx, tx, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Updated User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserAddWallet links a wallet to an account
const HubKeyUserAddWallet hub.HubCommandKey = "USER:ADD_WALLET"

type AddWalletRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID            passport.UserID `json:"id"`
		Username      string          `json:"username"`
		PublicAddress string          `json:"publicAddress"`
		Signature     string          `json:"signature"`
	} `json:"payload"`
}

// AddWalletHandler links a wallet address to a user
func (uc *UserController) AddWalletHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AddWalletRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() && req.Payload.Username == "" {
		return terror.Error(terror.ErrInvalidInput, "User ID or Username is required")
	}

	if req.Payload.PublicAddress == "" {
		return terror.Error(terror.ErrInvalidInput, "Public Address is required")
	}
	if req.Payload.Signature == "" {
		return terror.Error(terror.ErrInvalidInput, "Signature is required")
	}

	var user *passport.User
	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	}

	//// Permission check
	if user.ID.String() != hubc.Identifier() && !hubc.HasPermission(passport.PermUserUpdate.String()) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to update other users.")
	}

	// Setup user activity tracking
	var oldUser = *user

	// verify they signed it
	err = uc.API.Auth.VerifySignature(req.Payload.Signature, user.Nonce.String, req.Payload.PublicAddress)
	if err != nil {
		return terror.Error(err)
	}

	// Start transaction
	tx, err := uc.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			uc.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

	// Update user
	err = db.UserAddWallet(ctx, tx, user, req.Payload.PublicAddress)
	if err != nil {
		return terror.Error(err)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, uc.Conn, user.ID)
	if err != nil {
		return terror.Error(err)
	}

	reply(user)

	// Record user activity
	uc.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Updated User",
		passport.ObjectTypeUser,
		helpers.StringPointer(user.ID.String()),
		&user.Username,
		helpers.StringPointer(user.FirstName+" "+user.LastName),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Users,
			From: oldUser,
			To:   user,
		},
	)

	go uc.API.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

const HubKeyUserSubscribe hub.HubCommandKey = "USER:SUBSCRIBE"

// UpdatedSubscribeRequest to subscribe to user updates
type UpdatedSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID       passport.UserID `json:"id"`
		Username string          `json:"username"`
	} `json:"payload"`
}

func (uc *UserController) UpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &UpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	var user *passport.User

	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
		if err != nil {
			return req.TransactionID, "", terror.Error(err)
		}
	} else if req.Payload.Username != "" {
		user, err = db.UserByUsername(ctx, uc.Conn, req.Payload.Username)
		if err != nil {
			return req.TransactionID, "", terror.Error(err)
		}
	}

	if user == nil {
		return req.TransactionID, "", terror.Error(fmt.Errorf("unable to get user"))
	}

	// Permission check
	if user.ID.String() != client.Identifier() && !client.HasPermission(passport.PermUserRead.String()) {
		return req.TransactionID, "", terror.Error(terror.ErrUnauthorised, "You do not have permission to look at other users.")
	}

	reply(user)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), nil
}

const HubKeyUserSupsSubscribe hub.HubCommandKey = "USER:SUPS:SUBSCRIBE"

func (uc *UserController) UserSupsUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	sups, err := uc.API.userCacheMap.Get(client.Identifier())
	// get current on world sups
	if err != nil {
		return "", "", terror.Error(err)
	}

	reply(sups.String())
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsSubscribe, userID)), nil
}

const HubKeyUserSupsMultiplierSubscribe hub.HubCommandKey = "USER:SUPS:MULTIPLIER:SUBSCRIBE"

func (uc *UserController) UserSupsMultiplierUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	// send faction stat request to game server
	uc.API.SendToServerClient(ctx,
		SupremacyGameServer,
		&ServerClientMessage{
			Key: UserSupsMultiplierGet,
			Payload: struct {
				UserID    passport.UserID `json:"userID"`
				SessionID hub.SessionID   `json:"sessionID"`
			}{
				UserID:    passport.UserID(uuid.FromStringOrNil(client.Identifier())),
				SessionID: client.SessionID,
			},
		},
	)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSupsMultiplierSubscribe, client.Identifier())), nil
}

const HubKeyUserStatSubscribe hub.HubCommandKey = "USER:STAT:SUBSCRIBE"

func (uc *UserController) UserStatUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	// send faction stat request to game server
	uc.API.SendToServerClient(
		ctx,
		SupremacyGameServer,
		&ServerClientMessage{
			Key: UserStatGet,
			Payload: struct {
				UserID    passport.UserID `json:"userID"`
				SessionID hub.SessionID   `json:"sessionID"`
			}{
				UserID:    passport.UserID(uuid.FromStringOrNil(client.Identifier())),
				SessionID: client.SessionID,
			},
		},
	)

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserStatSubscribe, client.Identifier())), nil
}

type UserFactionDetail struct {
	RecruitID      string          `json:"recruitID"`
	SupsEarned     passport.BigInt `json:"supsEarned"`
	Rank           string          `json:"rank"`
	SpectatedCount int64           `json:"spectatedCount"`

	// faction detail
	FactionID        string                 `json:"factionID"`
	LogoBlobID       passport.BlobID        `json:"logoBlobID" db:"logo_blob_id"`
	BackgroundBlobID passport.BlobID        `json:"backgroundBlobID" db:"background_blob_id"`
	Theme            *passport.FactionTheme `json:"theme"`
}

const HubKeyUserFactionSubscribe hub.HubCommandKey = "USER:FACTION:SUBSCRIBE"

func (uc *UserController) UserFactionUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	// get user faction
	faction, err := db.FactionGetByUserID(ctx, uc.Conn, userID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", "", terror.Error(err)
	}

	if faction != nil {
		reply(&UserFactionDetail{
			RecruitID:        "3000",
			SupsEarned:       passport.BigInt{},
			Rank:             "100",
			SpectatedCount:   100,
			FactionID:        faction.ID.String(),
			Theme:            faction.Theme,
			LogoBlobID:       faction.LogoBlobID,
			BackgroundBlobID: faction.BackgroundBlobID,
		})
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserFactionSubscribe, userID)), nil
}

const HubKeyUserWarMachineQueuePositionSubscribe hub.HubCommandKey = "USER:WAR:MACHINE:QUEUE:POSITION:SUBSCRIBE"

func (uc *UserController) WarMachineQueuePositionUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &hub.HubCommandRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	userID := passport.UserID(uuid.FromStringOrNil(client.Identifier()))
	if userID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden)
	}

	// get user
	user, err := db.UserGet(ctx, uc.Conn, userID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	if user.FactionID == nil || user.FactionID.IsNil() {
		return "", "", terror.Error(terror.ErrForbidden, "User needs to join a faction to deploy War Machine")
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueuePositionSubscribe, userID))

	// TODO: run a request to gameserver to get the war machine list
	uc.API.SendToAllServerClient(ctx, &ServerClientMessage{
		Key: WarMachineQueuePositionGet,
		Payload: struct {
			UserID    passport.UserID    `json:"userID"`
			FactionID passport.FactionID `json:"factionID"`
		}{
			UserID:    userID,
			FactionID: *user.FactionID,
		},
	})

	return req.TransactionID, busKey, nil
}

// GetUserServiceCount returns the amount of services (email, facebook, google, discord etc.) the user is currently connected to
func GetUserServiceCount(user *passport.User) int {
	count := 0
	if user.Email.String != "" {
		count++
	}
	if user.FacebookID.String != "" {
		count++
	}
	if user.GoogleID.String != "" {
		count++
	}
	if user.TwitchID.String != "" {
		count++
	}
	if user.TwitterID.String != "" {
		count++
	}
	if user.DiscordID.String != "" {
		count++
	}

	return count
}

const HubKeySUPSRemainingSubscribe hub.HubCommandKey = "SUPS:TREASURY"

func (uc *UserController) TotalSupRemainingHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &UpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	sups, err := uc.API.userCacheMap.Get(passport.XsynSaleUserID.String())
	if err != nil {
		return "", "", terror.Error(err)
	}

	reply(sups.String())
	return req.TransactionID, messagebus.BusKey(HubKeySUPSRemainingSubscribe), nil
}

const HubKeySUPSExchangeRates hub.HubCommandKey = "SUPS:EXCHANGE"

func (uc *UserController) ExchangeRatesHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &UpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	reply(uc.API.State)
	return req.TransactionID, messagebus.BusKey(HubKeySUPSExchangeRates), nil
}

//key and handler- payload userid- check they are user return transaction key and error: SecureUserSubscribeCommand

type BlockConfirmationRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID             passport.UserID `json:"id"`
		GetInitialData bool            `json:"getInitialData"`
	} `json:"payload"`
}

const HubKeyBlockConfirmation hub.HubCommandKey = "BLOCK:CONFIRM"

func (uc *UserController) BlockConfirmationHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &BlockConfirmationRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	var user *passport.User

	user, err = db.UserGet(ctx, uc.Conn, req.Payload.ID)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Failed to get user")
	}

	if req.Payload.GetInitialData {
		// db func to get a list of users transaction on the comfirm transaction table
		// reply(their confirm objects)
		return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), nil
	}

	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyBlockConfirmation, user.ID.String())), nil
}

type CheckAllowedStoreAccess struct {
	*hub.HubCommandRequest
	Payload struct {
		WalletAddress string `json:"walletAddress"`
	} `json:"payload"`
}

type CheckAllowedStoreAccessResponse struct {
	IsAllowed bool   `json:"isAllowed"`
	Message   string `json:"message"`
}

const HubKeyCheckCanAccessStore hub.HubCommandKey = "USER:CHECK:CAN_ACCESS_STORE"

func (uc *UserController) CheckCanAccessStore(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	var WinTokens = []int{1, 2, 3, 4, 5, 6}

	req := &CheckAllowedStoreAccess{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.WalletAddress == "" {
		return terror.Error(terror.ErrInvalidInput, "Wallet address is required")
	}

	loc, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	// alpha
	PHASE_ONE := time.Date(2022, time.February, 24, 0, 0, 0, 0, loc)

	PHASE_TWO := time.Date(2022, time.February, 27, 0, 0, 0, 0, loc)

	PHASE_THREE := time.Date(2022, time.February, 27, 12, 0, 0, 0, loc)

	isWhitelisted, err := db.IsUserWhitelisted(ctx, uc.Conn, req.Payload.WalletAddress)
	if err != nil {
		return terror.Error(err, "whitelisted check error")
	}

	isDeathlisted, err := db.IsUserDeathlisted(ctx, uc.Conn, req.Payload.WalletAddress)
	if err != nil {
		return terror.Error(err, "deathlisted check error")
	}

	client, err := ethclient.Dial("wss://speedy-nodes-nyc.moralis.io/1375aa321ac8ac6cfba6aa9c/eth/mainnet/ws")
	if err != nil {
		return terror.Error(terror.ErrInvalidInput, "eth client dial error")
	}

	e, err := bridge.NewERC1155(common.HexToAddress("0x17F5655c7D834e4772171F30E7315bbc3221F1eE"), client)
	if err != nil {
		return terror.Error(terror.ErrInvalidInput, "bridge error")
	}

	isWinHolder, err := e.OwnsAny(common.HexToAddress(req.Payload.WalletAddress), WinTokens)
	if err != nil {
		return terror.Error(terror.ErrInvalidInput, "secondary holder check error")
	}

	isEarly := false
	now := time.Now().In(loc)
	dispersionMap := dispersions.All()
	for k := range dispersionMap {
		if strings.EqualFold(common.HexToAddress(req.Payload.WalletAddress).Hex(), k.Hex()) {
			if now.After(PHASE_ONE) {
				isEarly = true
			} else {
				isEarly = false
			}
		}
	}

	// if between 26th 12am - 27 12am only whitelisted and win holders and early contributors
	if now.After(PHASE_ONE) && now.Before(PHASE_TWO) && !(isWhitelisted || isWinHolder || isEarly) {
		resp := &CheckAllowedStoreAccessResponse{
			IsAllowed: false,
			Message:   "You must be Whitelisted to access the store",
		}
		reply(resp)
		return nil
	}

	// if after 27 12am included deathlisted
	if now.After(PHASE_ONE) && now.Before(PHASE_THREE) && !(isWhitelisted || isWinHolder || isEarly || isDeathlisted) {
		resp := &CheckAllowedStoreAccessResponse{
			IsAllowed: false,
			Message:   "You must be Whitelisted to access the store",
		}
		reply(resp)
		return nil
	}

	resp := &CheckAllowedStoreAccessResponse{
		IsAllowed: true,
		Message:   "",
	}
	reply(resp)
	return nil
}
