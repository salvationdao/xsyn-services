package api

import (
	"context"
	"encoding/json"
	"fmt"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub"
	"github.com/ninja-syndicate/hub/ext/messagebus"
	"github.com/rs/zerolog"
)

// TransactionController holds handlers for transaction endpoints
type TransactionController struct {
	Conn *pgxpool.Pool
	Log  *zerolog.Logger
	API  *API
}

// NewTransactionController creates the user hub
func NewTransactionController(log *zerolog.Logger, conn *pgxpool.Pool, api *API) *TransactionController {
	transactionHub := &TransactionController{
		Conn: conn,
		Log:  log_helpers.NamedLogger(log, "transaction_hub"),
		API:  api,
	}

	api.SecureCommand(HubKeyTransactionGroups, transactionHub.TransactionGroupsHandler)
	api.SecureCommand(HubKeyTransactionList, transactionHub.TransactionListHandler)
	api.SecureUserSubscribeCommand(HubKeyTransactionSubscribe, transactionHub.TransactionSubscribeHandler) // Auth check inside handler

	return transactionHub
}

const HubKeyTransactionGroups hub.HubCommandKey = "TRANSACTION:GROUPS"

type TransactionGroup struct {
	Group     string   `json:"group"`
	SubGroups []string `json:"sub_groups"`
}

// TransactionGroupsHandler returns a list of group IDs that the user's transactions exist in
func (tc *TransactionController) TransactionGroupsHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not get user's group of transactions, try again or contact support."
	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	userID := types.UserID(uid)

	groups, err := db.UsersTransactionGroups(userID, ctx, tc.Conn)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(groups)
	return nil
}

// TransactionListRequest requests for a transaction list
type TransactionListRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		SortDir  db.SortByDir          `json:"sort_dir"`
		SortBy   db.TransactionColumn  `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter,omitempty"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// TransactionListResponse is the response from get Transaction list
type TransactionListResponse struct {
	Total          int      `json:"total"`
	TransactionIDs []string `json:"transaction_ids"`
}

const HubKeyTransactionList hub.HubCommandKey = "TRANSACTION:LIST"

func (tc *TransactionController) TransactionListHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) error {
	errMsg := "Could not get list of user transactions, try again or contact support."
	req := &TransactionListRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	offset := 0
	if req.Payload.Page > 0 {
		offset = req.Payload.Page * req.Payload.PageSize
	}

	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return terror.Error(err, errMsg)
	}

	userID := types.UserID(uid)

	total, transactions, err := db.TransactionList(
		ctx, tc.Conn,
		&userID,
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

	resultTransactionIDs := make([]string, 0)
	for _, s := range transactions {
		resultTransactionIDs = append(resultTransactionIDs, s.ID)
	}

	resp := &TransactionListResponse{
		total,
		resultTransactionIDs,
	}

	reply(resp)
	return nil
}

const HubKeyTransactionSubscribe hub.HubCommandKey = "TRANSACTION:SUBSCRIBE"

type TransactionSubscribeRequest struct {
	*hub.HubCommandRequest
	Payload struct {
		TransactionID string `json:"transaction_id"`
	} `json:"payload"`
}

func (tc *TransactionController) TransactionSubscribeHandler(ctx context.Context, hubc *hub.Client, payload []byte, reply hub.ReplyFunc) (string, messagebus.BusKey, error) {
	errMsg := "Issue subscribing user to transactions lists, try again or contact support."
	req := &TransactionSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return req.TransactionID, "", terror.Error(err, "Invalid request received.")
	}

	transaction, err := db.TransactionGet(ctx, tc.Conn, req.Payload.TransactionID)
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	// get user
	uid, err := uuid.FromString(hubc.Identifier())
	if err != nil {
		return "", "", terror.Error(err, errMsg)
	}

	userID := types.UserID(uid)
	if transaction.Credit != userID && transaction.Debit != userID {
		return "", "", terror.Error(fmt.Errorf("unauthorized"), "You do not have permission to view this item.")
	}

	reply(transaction)
	return req.TransactionID, messagebus.BusKey(fmt.Sprintf("%s:%s", HubKeyTransactionSubscribe, transaction.ID)), nil
}
