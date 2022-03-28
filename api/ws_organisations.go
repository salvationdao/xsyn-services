package api

import (
	"context"
	"encoding/json"
	"passport"
	"passport/db"
	"passport/helpers"

	"github.com/ninja-software/log_helpers"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/rs/zerolog"
)

// OrganisationController holds handlers for organisations
type OrganisationController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewOrganisationController creates the organisation hub
func NewOrganisationController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *OrganisationController {
	organisationHub := &OrganisationController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "organisation_hub"),
		API:  api,
	}

	api.SecureCommandWithPerm(HubKeyOrganisationList, organisationHub.ListHandler, passport.PermOrganisationList)
	api.SecureCommandWithPerm(HubKeyOrganisationGet, organisationHub.GetHandler, passport.PermOrganisationRead)
	api.SecureCommandWithPerm(HubKeyOrganisationCreate, organisationHub.CreateHandler, passport.PermOrganisationCreate)
	api.SecureCommandWithPerm(HubKeyOrganisationUpdate, organisationHub.UpdateHandler, passport.PermOrganisationUpdate)
	api.SecureCommandWithPerm(HubKeyOrganisationArchive, organisationHub.ArchiveHandler, passport.PermOrganisationArchive)
	api.SecureCommandWithPerm(HubKeyOrganisationUnarchive, organisationHub.UnarchiveHandler, passport.PermOrganisationUnarchive)

	return organisationHub
}

// HubKeyOrganisationList is a hub key to list organisations
const HubKeyOrganisationList hub.HubCommandKey = "ORGANISATION:LIST"

// OrganisationListRequest requests holds the filter for organisation list
type OrganisationListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   db.OrganisationColumn `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Archived bool                  `json:"archived"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// OrganisationListResponse is the response from the organisation list request
type OrganisationListResponse struct {
	Records []*passport.Organisation `json:"records"`
	Total   int                      `json:"total"`
}

// ListHandler lists organisations with pagination
func (ctrlr *OrganisationController) ListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not get list of organisations, try again or contact support."
	req := &OrganisationListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	organisations := []*passport.Organisation{}
	total, err := db.OrganisationList(ctx,
		ctrlr.Conn,
		&organisations,
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

	resp := &OrganisationListResponse{
		Total:   total,
		Records: organisations,
	}
	reply(resp)

	return nil
}

// HubKeyOrganisationGet is a hub key to get an organisation
const HubKeyOrganisationGet hub.HubCommandKey = "ORGANISATION:GET"

// OrganisationGetRequest to get an organisation
type OrganisationGetRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Slug string                  `json:"slug"`
		ID   passport.OrganisationID `json:"id"`
	} `json:"payload"`
}

// GetHandler to get an organisation
func (ctrlr *OrganisationController) GetHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Organisation not found, check the URL and try again or contact support."
	req := &OrganisationGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Get organisation
	var organisation *passport.Organisation
	if req.Payload.ID.IsNil() {
		organisation, err = db.OrganisationGetBySlug(ctx, ctrlr.Conn, req.Payload.Slug)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	} else {
		organisation, err = db.OrganisationGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	reply(organisation)

	return nil
}

// HubKeyOrganisationCreate is a hub key to create an organisation
const HubKeyOrganisationCreate hub.HubCommandKey = "ORGANISATION:CREATE"

// OrganisationPayload used for create and update requests
type OrganisationPayload struct {
	Name string `json:"name"`
}

// OrganisationCreateRequest to create an organisation
type OrganisationCreateRequest struct {
	*hub.HubCommandRequest
	Payload OrganisationPayload `json:"payload"`
}

// CreateHandler to create an organisation
func (ctrlr *OrganisationController) CreateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not create organisation, try again or contact support."
	req := &OrganisationCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Validation
	if req.Payload.Name == "" {
		return terror.Error(terror.ErrInvalidInput, "Name is required.")
	}

	// Create Organisation
	organisation := &passport.Organisation{
		Name: req.Payload.Name,
	}
	err = db.OrganisationCreate(ctx, ctrlr.Conn, organisation)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(organisation)

	//Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Created Organisation",
		passport.ObjectTypeOrganisation,
		helpers.StringPointer(organisation.ID.String()),
		helpers.StringPointer(organisation.Slug),
		helpers.StringPointer(organisation.Name),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Organisations,
			From: nil,
			To:   organisation,
		},
	)
	return nil
}

// HubKeyOrganisationUpdate is a hub key to update an organisation
const HubKeyOrganisationUpdate hub.HubCommandKey = "ORGANISATION:UPDATE"

// OrganisationUpdateRequest to update an organisation
type OrganisationUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID passport.OrganisationID `json:"id"`
		OrganisationPayload
	} `json:"payload"`
}

// UpdateHandler to update an organisation
func (ctrlr *OrganisationController) UpdateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not update organisation, try again or contact support."
	req := &OrganisationUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Validation
	if req.Payload.Name == "" {
		return terror.Error(terror.ErrInvalidInput, "Name is required.")
	}

	// Find Organisation
	organisation, err := db.OrganisationGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Setup user activity tracking
	var oldOrganisation passport.Organisation = *organisation

	// Update Values
	organisation.Name = req.Payload.Name

	// Update Organisation
	err = db.OrganisationUpdate(ctx, ctrlr.Conn, organisation)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(organisation)

	//Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Updated Organisation",
		passport.ObjectTypeOrganisation,
		helpers.StringPointer(organisation.ID.String()),
		helpers.StringPointer(organisation.Slug),
		helpers.StringPointer(organisation.Name),
		&passport.UserActivityChangeData{
			Name: db.TableNames.Organisations,
			From: oldOrganisation,
			To:   organisation,
		},
	)

	return nil
}

const (
	// HubKeyOrganisationArchive archives the organisation
	HubKeyOrganisationArchive = hub.HubCommandKey("ORGANISATION:ARCHIVE")

	// HubKeyOrganisationUnarchive unarchives the organisation
	HubKeyOrganisationUnarchive = hub.HubCommandKey("ORGANISATION:UNARCHIVE")
)

// OrganisationToggleArchiveRequest requests to archive an organisation
type OrganisationToggleArchiveRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID passport.OrganisationID `json:"id"`
	} `json:"payload"`
}

// ArchiveHandler archives an organisation
func (ctrlr *OrganisationController) ArchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not archive organisation, try again or contact support."
	req := &OrganisationToggleArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Unarchive
	err = db.OrganisationArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, true)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	organisation, err := db.OrganisationGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(organisation)

	//Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Archived Organisation",
		passport.ObjectTypeOrganisation,
		helpers.StringPointer(organisation.ID.String()),
		helpers.StringPointer(organisation.Slug),
		helpers.StringPointer(organisation.Name),
	)

	return nil
}

// UnarchiveHandler unarchives an organisation
func (ctrlr *OrganisationController) UnarchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not unarchive organisation, try again or contact support."
	req := &OrganisationToggleArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Archive
	err = db.OrganisationArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, false)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	organisation, err := db.OrganisationGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(organisation)

	//Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Unarchived Organisation",
		passport.ObjectTypeOrganisation,
		helpers.StringPointer(organisation.ID.String()),
		helpers.StringPointer(organisation.Slug),
		helpers.StringPointer(organisation.Name),
	)

	return nil
}
