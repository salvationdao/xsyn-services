package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/types"

	"github.com/friendsofgo/errors"

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

	total, txIDs, err := db.TransactionIDList(
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

	resp := &TransactionListResponse{
		total,
		txIDs,
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

type TransactionResponse struct {
	*boiler.Transaction
	CreditOwner *AccountOwner `json:"to"`
	DebitOwner  *AccountOwner `json:"from"`
}

type AccountOwner struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Username string `json:"username"`
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

	if transaction.CreditAccountID != user.AccountID && transaction.DebitAccountID != user.AccountID {
		return terror.Error(fmt.Errorf("unauthorized"), "You do not have permission to view this item.")
	}
	debitOwner, err := tc.API.userCacheMap.GetAccountOwner(transaction.DebitAccountID)
	if err != nil {
		return terror.Error(err, "Failed to get debit account owner")
	}
	creditOwner, err := tc.API.userCacheMap.GetAccountOwner(transaction.CreditAccountID)
	if err != nil {
		return terror.Error(err, "Failed to get credit account owner")
	}

	reply(&TransactionResponse{
		Transaction: transaction,
		CreditOwner: creditOwner,
		DebitOwner:  debitOwner,
	})
	return err
}

func (ucm *Transactor) GetAccountOwner(accountID string) (*AccountOwner, error) {
	_, accType, err := ucm.Get(accountID)
	if err != nil {
		return nil, err
	}

	if accType == boiler.AccountTypeUSER {
		user, err := boiler.FindUser(passdb.StdConn, accountID)
		if err != nil {
			return nil, terror.Error(err, "Failed to get user.")
		}

		return &AccountOwner{
			ID:       user.ID,
			Type:     accType,
			Username: user.Username,
		}, nil

	}

	syndicate, err := boiler.FindSyndicate(passdb.StdConn, accountID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, terror.Error(err, "Failed to get syndicate.")
	}

	return &AccountOwner{
		ID:       syndicate.ID,
		Type:     accType,
		Username: syndicate.Name,
	}, nil
}
