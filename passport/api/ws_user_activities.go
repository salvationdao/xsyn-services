package api

import (
	"context"
	"encoding/json"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/log_helpers"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

// UserActivityController holds handlers for user activity
type UserActivityController struct {
	Log *zerolog.Logger
	API *API
}

// NewUserActivityController creates the userActivity controller
func NewUserActivityController(log *zerolog.Logger, api *API) *UserActivityController {
	uah := &UserActivityController{
		Log: log_helpers.NamedLogger(log, "user_activity_hub"),
		API: api,
	}

	uah.API.SecureCommandWithPerm(HubKeyUserActivityList, uah.UserActivityListHandler, types.PermUserActivityList)
	uah.API.SecureCommandWithPerm(HubKeyUserActivityGet, uah.UserActivityGetHandler, types.PermUserActivityList)
	uah.API.SecureCommand(HubKeyUserActivityCreate, uah.UserActivityCreateHandler)

	return uah
}

// HubKeyUserActivityList is a hub key to list userActivities
const HubKeyUserActivityList = "USER_ACTIVITY:LIST"

// UserActivityListRequest  requests holds the filter for userActivity list
type UserActivityListRequest struct {
	*hub.HubCommandKey
	Payload struct {
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   db.UserActivityColumn `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// UserActivityListResponse is the response from the userActivity update request
type UserActivityListResponse struct {
	Records []*types.UserActivity `json:"records"`
	Total   int                   `json:"total"`
}

// UserActivityListHandler lists userActivities with pagination
func (hub *UserActivityController) UserActivityListHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Could not get user activities list, try again or contact support."

	req := &UserActivityListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	userActivities := []*types.UserActivity{}
	total, err := db.UserActivityList(
		userActivities,
		req.Payload.Search,
		req.Payload.Filter,
		offset,
		req.Payload.PageSize,
		req.Payload.SortBy,
		req.Payload.SortDir,
	)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := &UserActivityListResponse{
		Total:   total,
		Records: userActivities,
	}
	reply(resp)

	return nil
}

// HubKeyUserActivityCreate is a hub key to create userActivities
const HubKeyUserActivityCreate = "USER_ACTIVITY:CREATE"

// UserActivityPayload used for create requests
type UserActivityPayload struct {
	UserID     types.UserID     `json:"user_id"`
	Action     string           `json:"action"`
	ObjectID   *string          `json:"object_id"`
	ObjectSlug *string          `json:"object_slug"`
	ObjectName *string          `json:"object_name"`
	ObjectType types.ObjectType `json:"object_type"`
	OldData    null.JSON        `json:"old_data"`
	NewData    null.JSON        `json:"new_data"`
}

// UserActivityCreateRequest requests a create userActivity
type UserActivityCreateRequest struct {
	Payload UserActivityPayload `json:"payload"`
}

// UserActivityCreateHandler creates userActivities
func (hub *UserActivityController) UserActivityCreateHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue creating user activity, please try again or contact support."

	req := &UserActivityCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Validation
	if req.Payload.UserID.IsNil() {
		return terror.Error(terror.ErrInvalidInput, "UserID is required.")
	}
	if req.Payload.Action == "" {
		return terror.Error(terror.ErrInvalidInput, "Action is required.")
	}

	// Create userActivity
	if err != nil {
		return terror.Error(err, errMsg)
	}

	err = db.UserActivityCreate(
		req.Payload.UserID,
		req.Payload.Action,
		req.Payload.ObjectType,
		req.Payload.ObjectID,
		req.Payload.ObjectSlug,
		req.Payload.ObjectName,
		req.Payload.OldData,
		req.Payload.NewData,
	)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	return nil
}

// HubKeyUserActivityGet is a hub key to grab userActivity
const HubKeyUserActivityGet = "USER_ACTIVITY:GET"

// UserActivityGetRequest requests an create userActivity
type UserActivityGetRequest struct {
	Payload struct {
		ID types.UserActivityID `json:"id"`
	} `json:"payload"`
}

// UserActivityGetHandler get userActivities
func (hub *UserActivityController) UserActivityGetHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "User activity not found, check the URL and try again."
	req := &UserActivityGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Get userActivity
	userActivity := &types.UserActivity{}
	err = db.UserActivityGet(userActivity, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(userActivity)

	return nil
}
