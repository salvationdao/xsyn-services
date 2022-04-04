package api

import (
	"context"
	"encoding/json"
	"xsyn-services/passport/db"
	"xsyn-services/passport/helpers"
	"xsyn-services/types"

	"github.com/ninja-software/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
)

// RoleController holds handlers for roles
type RoleController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewRoleController creates the role hub
func NewRoleController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *RoleController {
	roleHub := &RoleController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "role_hub"),
		API:  api,
	}

	api.SecureCommandWithPerm(HubKeyRoleList, roleHub.ListHandler, types.PermRoleList)
	api.SecureCommandWithPerm(HubKeyRoleGet, roleHub.GetHandler, types.PermRoleRead)
	api.SecureCommandWithPerm(HubKeyRoleCreate, roleHub.CreateHandler, types.PermRoleCreate)
	api.SecureCommandWithPerm(HubKeyRoleUpdate, roleHub.UpdateHandler, types.PermRoleUpdate)
	api.SecureCommandWithPerm(HubKeyRoleArchive, roleHub.ArchiveHandler, types.PermRoleArchive)
	api.SecureCommandWithPerm(HubKeyRoleUnarchive, roleHub.UnarchiveHandler, types.PermRoleUnarchive)

	return roleHub
}

// HubKeyRoleList is a hub key to list roles
const HubKeyRoleList hub.HubCommandKey = "ROLE:LIST"

// RoleListRequest requests holds the filter for role list
type RoleListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   db.RoleColumn         `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Archived bool                  `json:"archived"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// RoleListResponse is the response from the role list request
type RoleListResponse struct {
	Records []*types.Role `json:"records"`
	Total   int           `json:"total"`
}

// ListHandler lists roles with pagination
func (ctrlr *RoleController) ListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	req := &RoleListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	roles := []*types.Role{}
	total, err := db.RoleList(ctx,
		ctrlr.Conn,
		&roles,
		req.Payload.Search,
		req.Payload.Archived,
		req.Payload.Filter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return terror.Error(err, "Could not get list of roles, try again or contact support.")
	}

	resp := &RoleListResponse{
		Total:   total,
		Records: roles,
	}
	reply(resp)

	return nil
}

// HubKeyRoleGet is a hub key to get a role
const HubKeyRoleGet hub.HubCommandKey = "ROLE:GET"

// RoleGetRequest to get a role
type RoleGetRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Name string       `json:"name"`
		ID   types.RoleID `json:"id"`
	} `json:"payload"`
}

// GetHandler to get a role
func (ctrlr *RoleController) GetHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Role not found, check the URL and try again or contact support."
	req := &RoleGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Get role
	var role *types.Role
	if req.Payload.ID.IsNil() {
		role, err = db.RoleByName(ctx, ctrlr.Conn, req.Payload.Name)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	} else {
		role, err = db.RoleGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	reply(role)

	return nil
}

// HubKeyRoleCreate is a hub key to create a role
const HubKeyRoleCreate hub.HubCommandKey = "ROLE:CREATE"

// RolePayload used for create and update requests
type RolePayload struct {
	Name        string    `json:"name"`
	Permissions *[]string `json:"permissions,omitempty"`
}

// RoleCreateRequest to create a role
type RoleCreateRequest struct {
	*hub.HubCommandRequest
	Payload RolePayload `json:"payload"`
}

// CreateHandler to create a role
func (ctrlr *RoleController) CreateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not create role, try again or contact support."

	req := &RoleCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Validation
	if req.Payload.Name == "" {
		return terror.Error(terror.ErrInvalidInput, "Name is required.")
	}

	// Create Role
	role := &types.Role{
		Name: req.Payload.Name,
	}
	if req.Payload.Permissions != nil {
		for _, p := range *req.Payload.Permissions {
			validPerm := false
			for _, vp := range types.AllPerm {
				if p == vp.String() {
					validPerm = true
					break
				}
			}
			if !validPerm {
				return terror.Error(terror.ErrInvalidInput, "Invalid permission to create role.")
			}
		}
		role.Permissions = *req.Payload.Permissions

		// Prevent making roles equal to SUPERADMIN
		if len(role.Permissions) >= len(types.AllPerm) {
			return terror.Error(terror.ErrUnauthorised)
		}
	}

	err = db.RoleCreate(ctx, ctrlr.Conn, role)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(role)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Created Role",
		types.ObjectTypeRole,
		helpers.StringPointer(role.ID.String()),
		helpers.StringPointer(role.Name),
		helpers.StringPointer(role.Name),
		&types.UserActivityChangeData{
			Name: db.TableNames.Roles,
			From: nil,
			To:   role,
		},
	)
	return nil
}

// HubKeyRoleUpdate is a hub key to update a role
const HubKeyRoleUpdate hub.HubCommandKey = "ROLE:UPDATE"

// RoleUpdateRequest to update a role
type RoleUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID types.RoleID `json:"id"`
		RolePayload
	} `json:"payload"`
}

// UpdateHandler to update a role
func (ctrlr *RoleController) UpdateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not update user role, try again or contact support."

	req := &RoleUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Find Role
	role, err := db.RoleGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Setup user activity tracking
	var oldRole types.Role = *role

	// Update Values
	if req.Payload.Name != "" {
		role.Name = req.Payload.Name
	}
	if req.Payload.Permissions != nil {
		for _, p := range *req.Payload.Permissions {
			validPerm := false
			for _, vp := range types.AllPerm {
				if p == vp.String() {
					validPerm = true
					break
				}
			}
			if !validPerm {
				return terror.Error(terror.ErrInvalidInput, "Invalid permission to update role.")
			}
		}
		role.Permissions = *req.Payload.Permissions

		// Prevent making roles equal to SUPERADMIN
		if len(role.Permissions) >= len(types.AllPerm) {
			return terror.Error(terror.ErrUnauthorised)
		}
	}

	// Update Role
	err = db.RoleUpdate(ctx, ctrlr.Conn, role)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(role)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Updated Role",
		types.ObjectTypeRole,
		helpers.StringPointer(role.ID.String()),
		helpers.StringPointer(role.Name),
		helpers.StringPointer(role.Name),
		&types.UserActivityChangeData{
			Name: db.TableNames.Roles,
			From: oldRole,
			To:   role,
		},
	)

	return nil
}

const (
	// HubKeyRoleArchive archives the role
	HubKeyRoleArchive hub.HubCommandKey = hub.HubCommandKey("ROLE:ARCHIVE")

	// HubKeyRoleUnarchive unarchives the role
	HubKeyRoleUnarchive hub.HubCommandKey = hub.HubCommandKey("ROLE:UNARCHIVE")
)

// RoleToggleArchiveRequest requests to archive a role
type RoleToggleArchiveRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID types.RoleID `json:"id"`
	} `json:"payload"`
}

// ArchiveHandler archives a role
func (ctrlr *RoleController) ArchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not archive role, try again or contact support."

	req := &RoleToggleArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Archive
	err = db.RoleArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, true)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	role, err := db.RoleGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(role)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Archived Role",
		types.ObjectTypeRole,
		helpers.StringPointer(role.ID.String()),
		helpers.StringPointer(role.Name),
		helpers.StringPointer(role.Name),
	)

	return nil
}

// UnarchiveHandler unarchives a role
func (ctrlr *RoleController) UnarchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not unarchive role, try again or contact support."

	req := &RoleToggleArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Unarchive
	err = db.RoleArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, false)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	role, err := db.RoleGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(role)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Unarchived Role",
		types.ObjectTypeRole,
		helpers.StringPointer(role.ID.String()),
		helpers.StringPointer(role.Name),
		helpers.StringPointer(role.Name),
	)

	return nil
}
