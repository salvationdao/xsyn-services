package api

import (
	"context"
	"encoding/json"
	"errors"
	"passport"
	"passport/db"

	"github.com/ninja-software/log_helpers"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
)

// UserActivityController holds handlers for user activity
type UserActivityController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewUserActivityController creates the userActivity controller
func NewUserActivityController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *UserActivityController {
	uah := &UserActivityController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "user_activity_hub"),
		API:  api,
	}

	uah.API.SecureCommandWithPerm(HubKeyUserActivityList, uah.UserActivityListHandler, passport.PermUserActivityList)
	uah.API.SecureCommandWithPerm(HubKeyUserActivityGet, uah.UserActivityGetHandler, passport.PermUserActivityList)
	uah.API.SecureCommand(HubKeyUserActivityCreate, uah.UserActivityCreateHandler)

	return uah
}

// HubKeyUserActivityList is a hub key to list userActivities
const HubKeyUserActivityList hub.HubCommandKey = "USER_ACTIVITY:LIST"

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
	Records []*passport.UserActivity `json:"records"`
	Total   int                      `json:"total"`
}

// UserActivityListHandler lists userActivities with pagination
func (hub *UserActivityController) UserActivityListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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

	userActivities := []*passport.UserActivity{}
	total, err := db.UserActivityList(
		ctx,
		hub.Conn,
		&userActivities,
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
const HubKeyUserActivityCreate hub.HubCommandKey = "USER_ACTIVITY:CREATE"

// UserActivityPayload used for create requests
type UserActivityPayload struct {
	UserID     passport.UserID     `json:"user_id"`
	Action     string              `json:"action"`
	ObjectID   *string             `json:"object_id"`
	ObjectSlug *string             `json:"object_slug"`
	ObjectName *string             `json:"object_name"`
	ObjectType passport.ObjectType `json:"object_type"`
	OldData    null.JSON           `json:"old_data"`
	NewData    null.JSON           `json:"new_data"`
}

// UserActivityCreateRequest requests a create userActivity
type UserActivityCreateRequest struct {
	*hub.HubCommandRequest
	Payload UserActivityPayload `json:"payload"`
}

// UserActivityCreateHandler creates userActivities
func (hub *UserActivityController) UserActivityCreateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
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
	tx, err := hub.Conn.Begin(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			hub.Log.Err(err).Msg("error rolling back.")
		}
	}(tx, ctx)

	err = db.UserActivityCreate(
		ctx,
		tx,
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

	err = tx.Commit(ctx)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	return nil
}

// HubKeyUserActivityGet is a hub key to grab userActivity
const HubKeyUserActivityGet hub.HubCommandKey = "USER_ACTIVITY:GET"

// UserActivityGetRequest requests an create userActivity
type UserActivityGetRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID passport.UserActivityID `json:"id"`
	} `json:"payload"`
}

// UserActivityGetHandler get userActivities
func (hub *UserActivityController) UserActivityGetHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "User activity not found, check the URL and try again."
	req := &UserActivityGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Get userActivity
	userActivity := &passport.UserActivity{}
	err = db.UserActivityGet(ctx, hub.Conn, userActivity, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(userActivity)

	return nil
}
