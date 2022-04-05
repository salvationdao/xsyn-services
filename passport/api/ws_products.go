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

// ProductController holds handlers for products
type ProductController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewProductController creates the product hub
func NewProductController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *ProductController {
	productHub := &ProductController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "product_hub"),
		API:  api,
	}

	api.SecureCommandWithPerm(HubKeyProductList, productHub.ListHandler, types.PermProductList)
	api.SecureCommandWithPerm(HubKeyProductGet, productHub.GetHandler, types.PermProductRead)
	api.SecureCommandWithPerm(HubKeyProductCreate, productHub.CreateHandler, types.PermProductCreate)
	api.SecureCommandWithPerm(HubKeyProductUpdate, productHub.UpdateHandler, types.PermProductUpdate)
	api.SecureCommandWithPerm(HubKeyProductArchive, productHub.ArchiveHandler, types.PermProductArchive)
	api.SecureCommandWithPerm(HubKeyProductUnarchive, productHub.UnarchiveHandler, types.PermProductUnarchive)
	api.SecureCommand(HubKeyImageList, productHub.ImageListHandler)

	return productHub
}

// HubKeyProductList is a hub key to list products
const HubKeyProductList hub.HubCommandKey = "PRODUCT:LIST"

// ProductListRequest requests holds the filter for product list
type ProductListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   db.ProductColumn      `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter"`
		Archived bool                  `json:"archived"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// ProductListResponse is the response from the product list request
type ProductListResponse struct {
	Records []*types.Product `json:"records"`
	Total   int              `json:"total"`
}

// ListHandler lists products with pagination
func (ctrlr *ProductController) ListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not get list of products, try again or contact support."
	req := &ProductListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	products := []*types.Product{}
	total, err := db.ProductList(ctx,
		ctrlr.Conn,
		&products,
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

	resp := &ProductListResponse{
		Total:   total,
		Records: products,
	}
	reply(resp)

	return nil
}

// HubKeyProductGet is a hub key to get a product
const HubKeyProductGet hub.HubCommandKey = "PRODUCT:GET"

// ProductGetRequest to get a product
type ProductGetRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Slug string          `json:"slug"`
		ID   types.ProductID `json:"id"`
	} `json:"payload"`
}

// GetHandler to get a product
func (ctrlr *ProductController) GetHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Product not found, check the URL and try again or contact support."
	req := &ProductGetRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Get product
	var product *types.Product
	if req.Payload.ID.IsNil() {
		product, err = db.ProductGetBySlug(ctx, ctrlr.Conn, req.Payload.Slug)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	} else {
		product, err = db.ProductGet(ctx, ctrlr.Conn, req.Payload.ID)
		if err != nil {
			return terror.Error(err, errMsg)
		}
	}

	reply(product)

	return nil
}

// HubKeyProductCreate is a hub key to create a product
const HubKeyProductCreate hub.HubCommandKey = "PRODUCT:CREATE"

// ProductPayload used for create and update requests
type ProductPayload struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	ImageID     *types.BlobID `json:"imageID"`
}

// ProductCreateRequest to create a product
type ProductCreateRequest struct {
	*hub.HubCommandRequest
	Payload ProductPayload `json:"payload"`
}

// CreateHandler to create a product
func (ctrlr *ProductController) CreateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not create product, try again or contact support."
	req := &ProductCreateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Validation
	if req.Payload.Name == "" {
		return terror.Error(terror.ErrInvalidInput, "Name is required.")
	}

	// Create Product
	product := &types.Product{
		Name:        req.Payload.Name,
		Description: req.Payload.Description,
	}
	if req.Payload.ImageID != nil {
		product.ImageID = req.Payload.ImageID
	}

	err = db.ProductCreate(ctx, ctrlr.Conn, product)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(product)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Created Product",
		types.ObjectTypeProduct,
		helpers.StringPointer(product.ID.String()),
		helpers.StringPointer(product.Slug),
		helpers.StringPointer(product.Name),
		&types.UserActivityChangeData{
			Name: db.TableNames.Products,
			From: nil,
			To:   product,
		},
	)
	return nil
}

// HubKeyProductUpdate is a hub key to update a product
const HubKeyProductUpdate hub.HubCommandKey = "PRODUCT:UPDATE"

// ProductUpdateRequest to update an product
type ProductUpdateRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID types.ProductID `json:"id"`
		ProductPayload
	} `json:"payload"`
}

// UpdateHandler to update a product
func (ctrlr *ProductController) UpdateHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {

	errMsg := "Could not update product, try again or contact support."
	req := &ProductUpdateRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Validation
	if req.Payload.Name == "" {
		return terror.Error(terror.ErrInvalidInput, "Name is required.")
	}

	// Find Product
	product, err := db.ProductGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	// Setup user activity tracking
	var oldProduct types.Product = *product

	// Update Values
	product.Name = req.Payload.Name
	product.Description = req.Payload.Description
	if req.Payload.ImageID != nil {
		product.ImageID = req.Payload.ImageID
	}

	// Update Product
	err = db.ProductUpdate(ctx, ctrlr.Conn, product)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(product)

	//Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Updated Product",
		types.ObjectTypeProduct,
		helpers.StringPointer(product.ID.String()),
		helpers.StringPointer(product.Slug),
		helpers.StringPointer(product.Name),
		&types.UserActivityChangeData{
			Name: db.TableNames.Products,
			From: oldProduct,
			To:   product,
		},
	)

	return nil
}

const (
	// HubKeyProductArchive archives the product
	HubKeyProductArchive = hub.HubCommandKey("PRODUCT:ARCHIVE")

	// HubKeyProductUnarchive unarchives the product
	HubKeyProductUnarchive = hub.HubCommandKey("PRODUCT:UNARCHIVE")
)

// ProductToggleArchiveRequest requests to archive an product
type ProductToggleArchiveRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		ID types.ProductID `json:"id"`
	} `json:"payload"`
}

// ArchiveHandler archives a product
func (ctrlr *ProductController) ArchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not archive product, try again or contact support."
	req := &ProductToggleArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Archive
	err = db.ProductArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, true)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	product, err := db.ProductGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(product)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Archived Product",
		types.ObjectTypeProduct,
		helpers.StringPointer(product.ID.String()),
		helpers.StringPointer(product.Slug),
		helpers.StringPointer(product.Name),
	)

	return nil
}

// UnarchiveHandler unarchives a product
func (ctrlr *ProductController) UnarchiveHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not unarchive product, try again or contact support."
	req := &ProductToggleArchiveRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	// Unarchive
	err = db.ProductArchiveUpdate(ctx, ctrlr.Conn, req.Payload.ID, false)
	if err != nil {
		return terror.Error(err, errMsg)
	}
	product, err := db.ProductGet(ctx, ctrlr.Conn, req.Payload.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(product)

	// Record user activity
	ctrlr.API.RecordUserActivity(ctx,
		hubc.Identifier(),
		"Unarchived Product",
		types.ObjectTypeProduct,
		helpers.StringPointer(product.ID.String()),
		helpers.StringPointer(product.Slug),
		helpers.StringPointer(product.Name),
	)

	return nil
}

// HubKeyImageList is a hub key to get a list of images in the system (excluding avatars)
const HubKeyImageList hub.HubCommandKey = "IMAGE:LIST"

// NudgeImageListRequest requests a list of images in the system (excluding avatars)
type NudgeImageListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		Search   string `json:"search"`
		PageSize int    `json:"page_size"`
		Page     int    `json:"page"`
	} `json:"payload"`
}

// ImageListResponse is the response from the image list request
type ImageListResponse struct {
	Records []*types.Blob `json:"records"`
	Total   int           `json:"total"`
}

// ImageListHandler gets a list of images in the system (excluding avatars)
func (ctrlr *ProductController) ImageListHandler(ctx context.Context, wsc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not get images, please try again or contact support."
	req := &NudgeImageListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	total, images, err := db.BlobList(ctx, ctrlr.Conn, req.Payload.Search, offset, req.Payload.PageSize)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	resp := &ImageListResponse{
		Total:   total,
		Records: *images,
	}
	reply(resp)

	return nil
}
