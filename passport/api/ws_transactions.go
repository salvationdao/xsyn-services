package api

import (
	"context"
	"encoding/json"
	"fmt"
	"xsyn-services/passport/db"
	"xsyn-services/types"

	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/ws"
	"github.com/rs/zerolog"
)

// TransactionController holds handlers for transaction endpoints
type TransactionController struct {
	Log *zerolog.Logger
	API *API
}

// NewTransactionController creates the user hub
func NewTransactionController(log *zerolog.Logger, api *API) *TransactionController {
	transactionHub := &TransactionController{
		Log: log_helpers.NamedLogger(log, "transaction_hub"),
		API: api,
	}

	api.SecureCommand(HubKeyTransactionGroups, transactionHub.TransactionGroupsHandler)
	api.SecureCommand(HubKeyTransactionList, transactionHub.TransactionListHandler)
	api.SecureCommand(HubKeyTransactionSubscribe, transactionHub.TransactionSubscribeHandler) // Auth check inside handler

	return transactionHub
}

const HubKeyTransactionGroups = "TRANSACTION:GROUPS"

type TransactionGroup struct {
	Group     string   `json:"group"`
	SubGroups []string `json:"sub_groups"`
}

// TransactionGroupsHandler returns a list of group IDs that the user's transactions exist in
func (tc *TransactionController) TransactionGroupsHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Could not get user's group of transactions, try again or contact support."

	groups, err := db.UsersTransactionGroups(user.ID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	reply(groups)
	return nil
}

// TransactionListRequest requests for a transaction list
type TransactionListRequest struct {
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

const HubKeyTransactionList = "TRANSACTION:LIST"

func (tc *TransactionController) TransactionListHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
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

	total, transactions, err := db.TransactionList(
		&user.ID,
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

const HubKeyTransactionSubscribe = "TRANSACTION:SUBSCRIBE"

type TransactionSubscribeRequest struct {
	Payload struct {
		TransactionID string `json:"transaction_id"`
	} `json:"payload"`
}

func (tc *TransactionController) TransactionSubscribeHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	errMsg := "Issue subscribing user to transactions lists, try again or contact support."
	req := &TransactionSubscribeRequest{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	transaction, err := db.TransactionGet(req.Payload.TransactionID)
	if err != nil {
		return terror.Error(err, errMsg)
	}

	if transaction.Credit != user.ID && transaction.Debit != user.ID {
		return terror.Error(fmt.Errorf("unauthorized"), "You do not have permission to view this item.")
	}

	reply(transaction)
	return err
}
