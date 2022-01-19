package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"passport"
	"passport/crypto"
	"passport/db"
	"passport/helpers"
	"passport/log_helpers"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"github.com/jackc/pgx/v4"

	"github.com/volatiletech/null/v8"
	"google.golang.org/api/idtoken"

	"github.com/ninja-software/hub/v2"
	"github.com/ninja-software/hub/v2/ext/auth"
	"github.com/ninja-software/hub/v2/ext/messagebus"
	"github.com/ninja-software/terror/v2"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

// UserController holds handlers for authentication
type UserController struct {
	Conn   *pgxpool.Pool
	Log    *zerolog.Logger
	API    *API
	Google *auth.GoogleConfig
	Twitch *auth.TwitchConfig
}

// NewUserController creates the user hub
func NewUserController(log *zerolog.Logger, conn *pgxpool.Pool, api *API, googleConfig *auth.GoogleConfig, twitchConfig *auth.TwitchConfig) *UserController {
	userHub := &UserController{
		Conn:   conn,
		Log:    log_helpers.NamedLogger(log, "user_hub"),
		API:    api,
		Google: googleConfig,
		Twitch: twitchConfig,
	}
	api.Command(HubKeyUserGet, userHub.GetHandler) // Perm check inside handler (users can get themselves; need UserRead permission to get other users)
	api.SecureCommand(HubKeyUserUpdate, userHub.UpdateHandler)
	api.SecureCommand(HubKeyUserFactionUpdate, userHub.UpdateUserFactionHandler) // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveFacebook, userHub.RemoveFacebookHandler)   // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddFacebook, userHub.AddFacebookHandler)         // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveGoogle, userHub.RemoveGoogleHandler)       // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddGoogle, userHub.AddGoogleHandler)             // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveTwitch, userHub.RemoveTwitchHandler)       // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddTwitch, userHub.AddTwitchHandler)             // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserRemoveWallet, userHub.RemoveWalletHandler)       // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserAddWallet, userHub.AddWalletHandler)             // Perm check inside handler (handler used to update self or for user w/ permission to update another user)
	api.SecureCommand(HubKeyUserCreate, userHub.CreateHandler)
	api.SecureCommandWithPerm(HubKeyUserList, userHub.ListHandler, passport.PermUserList)
	api.SecureCommandWithPerm(HubKeyUserArchive, userHub.ArchiveHandler, passport.PermUserArchive)
	api.SecureCommandWithPerm(HubKeyUserUnarchive, userHub.UnarchiveHandler, passport.PermUserUnarchive)
	api.SecureCommandWithPerm(HubKeyUserChangePassword, userHub.ChangePasswordHandler, passport.PermUserUpdate)
	api.SecureCommandWithPerm(HubKeyUserForceDisconnect, userHub.ForceDisconnectHandler, passport.PermUserForceDisconnect)

	api.SubscribeCommand(HubKeyUserForceDisconnected, userHub.ForceDisconnectedHandler)
	api.SubscribeCommand(HubKeyUserSubscribe, userHub.UpdatedSubscribeHandler)
	api.SubscribeCommand(HubKeyUserOnlineStatus, userHub.OnlineStatusSubscribeHandler)

	// listen on queuing war machine
	api.SecureUserSubscribeCommand(HubKeyUserWarMachineQueuePositionSubscribe, userHub.WarMachineQueuePositionUpdatedSubscribeHandler)

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
func (ctrlr *UserController) GetHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &GetUserRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID.IsNil() && req.Payload.Username == "" {
		return terror.Error(terror.ErrInvalidInput, "User ID or username is required")
	}

	if !req.Payload.ID.IsNil() {
		user, err := db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
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

	user, err := db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
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
func (ctrlr *UserController) UpdateUserFactionHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

	user, err := db.UserGet(ctx, ctrlr.Conn, req.Payload.UserID)
	if err != nil {
		return terror.Error(err, "Unable to load current user")
	}

	user.FactionID = &req.Payload.FactionID

	err = db.UserFactionEnlist(ctx, ctrlr.Conn, user)
	if err != nil {
		return terror.Error(err, "Unable to update user faction")
	}

	faction, err := db.FactionGet(ctx, ctrlr.Conn, req.Payload.FactionID)
	if err != nil {
		return terror.Error(err)
	}
	user.Faction = faction

	// send user changes to connected clients
	ctrlr.API.SendToAllServerClient(&ServerClientMessage{
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
func (ctrlr *UserController) UpdateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		user, err = db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
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
	if req.Payload.NewPassword != nil && *req.Payload.NewPassword != "" {
		err = helpers.IsValidPassword(*req.Payload.NewPassword)
		if err != nil {
			passwordErr := err.Error()
			var bErr *terror.TError
			if errors.As(err, &bErr) {
				passwordErr = bErr.Message
			}
			return terror.Error(err, passwordErr)
		}

		confirmPassword = req.Payload.ID.String() == hubc.Identifier() && user.OldPasswordRequired
	}

	if confirmPassword {
		if req.Payload.CurrentPassword == nil {
			return terror.Error(terror.ErrInvalidInput, "Current Password is required")
		}
		hashB64, err := db.HashByUserID(ctx, ctrlr.Conn, req.Payload.ID)
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
	tx, err := ctrlr.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ctrlr.Log.Err(err).Msg("error rolling back")
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
		return terror.Error(err, errMsg)
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
	user, err = db.UserGet(ctx, ctrlr.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if user.FactionID != nil && !user.FactionID.IsNil() {
		faction, err := db.FactionGet(ctx, ctrlr.Conn, *user.FactionID)
		if err != nil {
			return terror.Error(err)
		}
		user.Faction = faction
	}

	reply(user)
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	// send user changes to connected clients
	ctrlr.API.SendToAllServerClient(&ServerClientMessage{
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
func (ctrlr *UserController) CreateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
	tx, err := ctrlr.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ctrlr.Log.Err(err).Msg("error rolling back")
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

	err = db.UserCreate(ctx, ctrlr.Conn, user)
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
	user, err = db.UserGet(ctx, ctrlr.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

// // UserListResponse is the response from get user list
type UserListResponse struct {
	Records []*passport.User `json:"records"`
	Total   int              `json:"total"`
}

// ListHandler lists users with pagination
func (ctrlr *UserController) ListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		ctx, ctrlr.Conn, &users,
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
func (ctrlr *UserController) ArchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UserArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Failed to unmarshal data")
	}
	err = db.UserArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, true)
	if err != nil {
		return terror.Error(err, "Issue while updating User, please try again.")
	}

	// Return user
	user, err := db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err)
	}
	reply(user)

	// Record user activity
	if err == nil {
		ctrlr.API.RecordUserActivity(ctx,
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
func (ctrlr *UserController) UnarchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UserArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Failed to unmarshal data")
	}
	err = db.UserArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, false)
	if err != nil {
		return terror.Error(err, "Issue while updating User, please try again.")
	}

	// Return user
	user, err := db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err)
	}
	reply(user)

	//// Record user activity
	if err == nil {
		ctrlr.API.RecordUserActivity(ctx,
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
func (ctrlr *UserController) ChangePasswordHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &UserChangePasswordRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}
	if req.Payload.ID == passport.UserID(uuid.Nil) {
		return terror.Error(terror.ErrInvalidInput, "User ID is required")
	}

	user, err := db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
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

	tx, err := ctrlr.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ctrlr.Log.Err(err).Msg("error rolling back")
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
	ctrlr.API.RecordUserActivity(ctx,
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
func (ctrlr *UserController) ForceDisconnectHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

	user, err := db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, "Unable to load current user")
	}

	// Trying to disconnect user w/ higher role than you?
	if user.ID.String() != hubc.Identifier() && hubc.IsHigherOrSameLevel(user.Role.Tier) {
		return terror.Error(terror.ErrUnauthorised, "You do not have permission to force disconnect this user")
	}

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserForceDisconnected, user.ID.String())), nil)
	reply(true)

	// Delete issue tokens
	err = db.AuthRemoveTokensFromUserID(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err)
	}

	//Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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
func (ctrlr *UserController) ForceDisconnectedHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
func (ctrlr *UserController) OnlineStatusSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
		id, err := db.UserIDFromUsername(ctx, ctrlr.Conn, req.Payload.Username)
		if err != nil {
			return req.TransactionID, "", terror.Error(err, "Unable to load current user")
		}
		userID = *id
	}

	if userID.IsNil() {
		return req.TransactionID, "", terror.Error(fmt.Errorf("userId is still nil for %s %s", req.Payload.ID, req.Payload.Username), "Unable to load current user")
	}

	// get current online status
	online := false
	ctrlr.API.Hub.Clients(func(clients hub.ClientsList) {
		for cl := range clients {
			if cl.Identifier() == userID.String() {
				online = true
				break
			}
		}
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

func (ctrlr *UserController) RemoveFacebookHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		user, err = db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
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

	// Update user
	err = db.UserRemoveFacebook(ctx, ctrlr.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, ctrlr.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserAddFacebook removes a linked Facebook account
const HubKeyUserAddFacebook hub.HubCommandKey = "USER:ADD_FACEBOOK"

func (ctrlr *UserController) AddFacebookHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
	user, err := db.UserGet(ctx, ctrlr.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "failed to query user")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Update user's Facebook ID
	err = db.UserAddFacebook(ctx, ctrlr.Conn, user, resp.ID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, ctrlr.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	return nil
}

// HubKeyUserRemoveGoogle removes a linked Google account
const HubKeyUserRemoveGoogle hub.HubCommandKey = "USER:REMOVE_GOOGLE"

func (ctrlr *UserController) RemoveGoogleHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		user, err = db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
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

	// Update user
	err = db.UserRemoveGoogle(ctx, ctrlr.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, ctrlr.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserRemoveTwitch adds a linked Twitch account
const HubKeyUserAddGoogle hub.HubCommandKey = "USER:ADD_GOOGLE"

func (ctrlr *UserController) AddGoogleHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
	resp, err := idtoken.Validate(ctx, req.Payload.Token, ctrlr.Google.ClientID)
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
	user, err := db.UserGet(ctx, ctrlr.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "failed to query user")
	}

	// Setup user activity tracking
	var oldUser passport.User = *user

	// Update user's Google ID
	err = db.UserAddGoogle(ctx, ctrlr.Conn, user, googleID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, ctrlr.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

	return nil
}

// HubKeyUserRemoveTwitch removes a linked Twitch account
const HubKeyUserRemoveTwitch hub.HubCommandKey = "USER:REMOVE_TWITCH"

func (ctrlr *UserController) RemoveTwitchHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		user, err = db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
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

	// Update user
	err = db.UserRemoveTwitch(ctx, ctrlr.Conn, user)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Get user
	user, err = db.UserGet(ctx, ctrlr.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
	return nil
}

// HubKeyUserRemoveTwitch adds a linked Twitch account
const HubKeyUserAddTwitch hub.HubCommandKey = "USER:ADD_TWITCH"

type AddTwitchRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Token       string `json:"token"`
		RedirectURI string `json:"redirectURI"`
	} `json:"payload"`
}

func (ctrlr *UserController) AddTwitchHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	req := &AddTwitchRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received")
	}

	if req.Payload.Token == "" {
		return terror.Error(terror.ErrInvalidInput, "Twitch JWT is empty")
	}

	keySet := oidc.NewRemoteKeySet(ctx, "https://id.twitch.tv/oauth2/keys")
	oidcVerifier := oidc.NewVerifier("https://id.twitch.tv/oauth2", keySet, &oidc.Config{
		ClientID: ctrlr.Twitch.ClientID,
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

	twitchID := claims.Sub
	if twitchID == "" {
		return terror.Error(terror.ErrInvalidInput, "No Twitch account ID is provided")
	}

	userID, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, "Could not convert user ID to UUID")
	}

	// Get user
	user, err := db.UserGet(ctx, ctrlr.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	// Activity tracking
	var oldUser passport.User = *user

	// Update user's Twitch ID
	err = db.UserAddTwitch(ctx, ctrlr.Conn, user, twitchID)
	if err != nil {
		return terror.Error(err)
	}

	// Get user
	user, err = db.UserGet(ctx, ctrlr.Conn, passport.UserID(userID))
	if err != nil {
		return terror.Error(err, "Failed to query user")
	}

	reply(user)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)

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
func (ctrlr *UserController) RemoveWalletHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		user, err = db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
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
	tx, err := ctrlr.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ctrlr.Log.Err(err).Msg("error rolling back")
		}
	}(tx, ctx)

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
	user, err = db.UserGet(ctx, ctrlr.Conn, user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(user)

	//// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
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
func (ctrlr *UserController) AddWalletHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
		user, err = db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, "Failed to get user")
		}
	} else {
		user, err = db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
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
	err = ctrlr.API.Auth.VerifySignature(req.Payload.Signature, user.Nonce.String, req.Payload.PublicAddress)
	if err != nil {
		return terror.Error(err)
	}

	// Start transaction
	tx, err := ctrlr.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			ctrlr.Log.Err(err).Msg("error rolling back")
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
	user, err = db.UserGet(ctx, ctrlr.Conn, user.ID)
	if err != nil {
		return terror.Error(err)
	}

	reply(user)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
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

	ctrlr.API.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), user)
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

func (ctrlr *UserController) UpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	req := &UpdatedSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received")
	}

	var user *passport.User

	if !req.Payload.ID.IsNil() {
		user, err = db.UserGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return req.TransactionID, "", terror.Error(err)
		}
	} else if req.Payload.Username != "" {
		user, err = db.UserByUsername(ctx, ctrlr.Conn, req.Payload.Username)
		if err != nil {
			return req.TransactionID, "", terror.Error(err)
		}
	}

	if user == nil {
		return req.TransactionID, "", terror.Error(fmt.Errorf("unable to get user"))
	}

	//// Permission check
	if user.ID.String() != client.Identifier() && !client.HasPermission(passport.PermUserRead.String()) {
		return req.TransactionID, "", terror.Error(terror.ErrUnauthorised, "You do not have permission to look at other users.")
	}

	reply(user)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserSubscribe, user.ID.String())), nil
}

const HubKeyUserWarMachineQueuePositionSubscribe hub.HubCommandKey = "USER:WAR:MACHINE:QUEUE:POSITION:SUBSCRIBE"

// WarMachineQueuePositionUpdatedSubscribeHandler
func (ctrlr *UserController) WarMachineQueuePositionUpdatedSubscribeHandler(ctx context.Context, client *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
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
	user, err := db.UserGet(ctx, ctrlr.Conn, userID)
	if err != nil {
		return "", "", terror.Error(err)
	}

	busKey := messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyUserWarMachineQueuePositionSubscribe, userID))

	// TODO: run a request to gameserver to get the war machine list
	ctrlr.API.SendToAllServerClient(&ServerClientMessage{
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
