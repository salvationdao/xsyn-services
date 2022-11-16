package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kevinms/leakybucket-go"
	"github.com/shopspring/decimal"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
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
	api.SecureCommand(HubKeyMakeSupremacyWorldTransaction, transactionHub.TransactSupremacyWorldHandler)

	return transactionHub
}

const HubKeyMakeSupremacyWorldTransaction = "TRANSACT:SUPREMACY_WORLD"

type TransactSupremacyWorldReq struct {
	Payload struct {
		ClaimID string          `json:"claim_id"`
		Amount  decimal.Decimal `json:"amount"`
	} `json:"payload"`
}

var TransactSupremacyWorldBucket = leakybucket.NewCollector(0.5, 1, true)

// TransactSupremacyWorldHandler makes a transaction and then hits the supremacy world webhook to notify them of transaction
func (tc *TransactionController) TransactSupremacyWorldHandler(ctx context.Context, user *types.User, key string, payload []byte, reply ws.ReplyFunc) error {
	l := passlog.L.With().Str("func", "TransactSupremacyWorldHandler").Str("userID", user.ID).Logger()
	req := &TransactSupremacyWorldReq{}
	err := json.Unmarshal(payload, req)
	if err != nil {
		return terror.Error(err, "Invalid request received.")
	}

	b := TransactSupremacyWorldBucket.Add(req.Payload.ClaimID, 1)
	if b == 0 {
		return terror.Warn(fmt.Errorf("too many requests"), "Too many requests.")
	}
	
	l = l.With().Interface("req", req).Logger()

	if req.Payload.ClaimID == "" {
		err = fmt.Errorf("missing claim id")
		l.Error().Err(err).Msg("claim id is empty")
		return terror.Error(err, "Invalid request, please try again or contact support.")
	}
	if req.Payload.Amount.IsZero() {
		err = fmt.Errorf("missing transaction amount")
		l.Error().Err(err).Msg("amount is zero")
		return terror.Error(err, "Invalid request, please try again or contact support.")
	}

	txID, err := tc.API.userCacheMap.Transact(&types.NewTransaction{
		Credit:               types.SupremacyWorldUserID.String(),
		Debit:                user.ID,
		Amount:               req.Payload.Amount,
		TransactionReference: types.TransactionReference(fmt.Sprintf("supremacy_world_transaction|%s|%d", req.Payload.ClaimID, time.Now().UnixNano())),
		Description:          fmt.Sprintf("Supremacy World Purchase - %s", req.Payload.ClaimID),
		Group:                types.TransactionGroupSupremacyWorld,
		SubGroup:             "Purchase",
	})
	if err != nil {
		return terror.Error(err, "Payment failed, please check your balance and try again or contact support.")
	}

	err = tc.API.SupremacyWorldTransactionWebhookSend(&SupremacyWorldTransactionWebhookPayload{
		TransactionID: txID,
		UserID:        user.ID,
		ClaimID:       req.Payload.ClaimID,
	})
	if err != nil {
		l.Error().Err(err).Msg("failed to process claim on supremacy world")
		// refund the payment
		tc.API.userCacheMap.Transact(&types.NewTransaction{
			Credit:               user.ID,
			Debit:                types.SupremacyWorldUserID.String(),
			Amount:               req.Payload.Amount,
			TransactionReference: types.TransactionReference(fmt.Sprintf("supremacy_world_transaction|%s|%d|refund", req.Payload.ClaimID, time.Now().UnixNano())),
			Description:          fmt.Sprintf("Supremacy World Purchase Refund - %s", req.Payload.ClaimID),
			Group:                types.TransactionGroupSupremacyWorld,
			SubGroup:             "Refund",
		})
		return terror.Error(err, "Failed to handle payment on Supremacy World, please try again or contact support.")
	}

	reply(true)
	return nil
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
		SortBy   string                `json:"sort_by"`
		Filter   *db.ListFilterRequest `json:"filter,omitempty"`
		Search   string                `json:"search"`
		PageSize int                   `json:"page_size"`
		Page     int                   `json:"page"`
	} `json:"payload"`
}

// TransactionListResponse is the response from get Transaction list
type TransactionListResponse struct {
	Total        int                       `json:"total"`
	Transactions []*db.TransactionDetailed `json:"transactions"`
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

	total, transactions, err := db.TransactionIDList(
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
		transactions,
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

	if transaction.Credit != user.ID && transaction.Debit != user.ID {
		return terror.Error(fmt.Errorf("unauthorized"), "You do not have permission to view this item.")
	}
	debitOwner, err := tc.API.userCacheMap.GetAccountOwner(transaction.Debit)
	if err != nil {
		return terror.Error(err, "Failed to get debit account owner")
	}
	creditOwner, err := tc.API.userCacheMap.GetAccountOwner(transaction.Credit)
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
	_, err := ucm.Get(accountID)
	if err != nil {
		return nil, err
	}

	user, err := boiler.FindUser(passdb.StdConn, accountID)
	if err != nil {
		return nil, terror.Error(err, "Failed to get user.")
	}

	return &AccountOwner{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}
