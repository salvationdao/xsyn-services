package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/types"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func AdminRoutes(ucm *Transactor) chi.Router {
	r := chi.NewRouter()
	r.Get("/check", WithError(WithAdmin(AdminCheck)))
	r.Get("/users", WithError(WithAdmin(ListUsers)))
	r.Get("/users/{public_address}", WithError(WithAdmin(UserHandler)))
	r.Get("/find_by_userid/{userID}", WithError(WithAdmin(GetUserByUserID)))
	r.Get("/find_by_username/{username}", WithError(WithAdmin(GetUserByUsername)))
	r.Get("/chat_timeout_username/{username}/{minutes}", WithError(WithAdmin(ChatTimeoutUsername)))
	r.Get("/chat_timeout_userid/{userID}/{minutes}", WithError(WithAdmin(ChatTimeoutUserID)))
	r.Get("/rename_ban_username/{username}/{banned}", WithError(WithAdmin(RenameBanUsername)))
	r.Get("/rename_ban_userID/{userID}/{banned}", WithError(WithAdmin(RenameBanUserID)))
	r.Get("/rename_asset/{hash}/{newName}", WithError(WithAdmin(RenameAsset)))
	r.Get("/purchased_items", WithError(WithAdmin(ListPurchasedItems)))
	r.Get("/store_items", WithError(WithAdmin(ListStoreItems)))

	r.Post("/purchased_items/register/{template_id}/{owner_id}", WithError(WithAdmin(PurchasedItemRegisterHandler)))
	r.Post("/purchased_items/set_owner/{purchased_item_id}/{owner_id}", WithError(WithAdmin(PurchasedItemSetOwner)))
	r.Post("/purchased_items/transfer/from/{from}/to/{to}/collection_id/{collection_id}/token_id/{token_id}", WithError(WithAdmin(TransferAsset())))

	r.Post("/transactions/create", WithError(WithAdmin(CreateTransaction(ucm))))
	r.Post("/transactions/reverse/{transaction_id}", WithError(WithAdmin(ReverseUserTransaction(ucm))))
	r.Get("/transactions/list/user/{public_address}", WithError(WithAdmin(ListUserTransactions)))

	r.Post("/sync/store_items", WithError(WithAdmin(SyncStoreItems)))
	r.Post("/sync/purchased_items", WithError(WithAdmin(SyncPurchasedItems)))
	return r
}

// WithAdmin checks that admin key is in the header.
func WithAdmin(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		apiKeyIDStr := r.Header.Get("X-Authorization")
		apiKeyID, err := uuid.FromString(apiKeyIDStr)
		if err != nil {
			passlog.L.Warn().Err(err).Str("apiKeyID", apiKeyIDStr).Msg("unauthed attempted at mod rest end point")
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}
		apiKey, err := db.APIKey(apiKeyID)
		if err != nil {
			passlog.L.Warn().Err(err).Str("apiKeyID", apiKeyIDStr).Msg("unauthed attempted at mod rest end point")
			return http.StatusUnauthorized, terror.Error(terror.ErrUnauthorised, "Unauthorized.")
		}
		if apiKey.Type != "ADMIN" {
			return http.StatusUnauthorized, terror.Error(fmt.Errorf("not admin key: %s", apiKey.Type), "Unauthorized.")
		}
		return next(w, r)
	}
	return fn
}

type TransferAssetRequest struct {
	From           uuid.UUID      `json:"from"`
	To             uuid.UUID      `json:"to"`
	CollectionAddr common.Address `json:"collection_addr"`
	TokenID        int            `json:"token_id"`
}

func TransferAsset() func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {

		from := chi.URLParam(r, "from")
		to := chi.URLParam(r, "to")
		collectionAddr := common.HexToAddress(chi.URLParam(r, "collection_addr"))
		tokenIDStr := chi.URLParam(r, "token_id")
		tokenID, err := strconv.Atoi(tokenIDStr)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not convert tokenID to int")
		}

		c, err := db.CollectionByMintAddress(collectionAddr)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get collection")
		}

		item, err := boiler.PurchasedItems(
			boiler.PurchasedItemWhere.ExternalTokenID.EQ(tokenID),
			boiler.PurchasedItemWhere.CollectionID.EQ(c.ID),
		).One(passdb.StdConn)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get purchased item")
		}
		if item.OwnerID != from {
			return http.StatusBadRequest, errors.New("from user does not own the asset")
		}
		item.OwnerID = to
		_, err = item.Update(passdb.StdConn, boil.Whitelist(boiler.PurchasedItemColumns.OwnerID))
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not update purchased item")
		}
		return http.StatusOK, nil
	}
	return fn
}

type CreateTransactionRequest struct {
	Amount decimal.Decimal `json:"amount"`
	Credit uuid.UUID       `json:"credit"`
	Debit  uuid.UUID       `json:"debit"`
}

func CreateTransaction(ucm *Transactor) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		req := &CreateTransactionRequest{}
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not decode json")
		}

		ref := fmt.Sprintf("TRANSFER - %d", time.Now().UnixNano())
		newTx := &types.NewTransaction{
			To:                   types.UserID(req.Credit),
			From:                 types.UserID(req.Debit),
			Amount:               req.Amount,
			TransactionReference: types.TransactionReference(ref),
			Description:          ref,
			Group:                types.TransactionGroupStore,
			SubGroup:             "Transfer",
		}
		_, _, _, err = ucm.Transact(newTx)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get transaction")
		}
		return http.StatusOK, nil
	}
	return fn
}

func ReverseUserTransaction(ucm *Transactor) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		txID := chi.URLParam(r, "transaction_id")
		tx, err := boiler.FindTransaction(passdb.StdConn, txID)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get transaction")
		}
		refundTx := &types.NewTransaction{
			To:                   types.UserID(uuid.Must(uuid.FromString(tx.Debit))),
			From:                 types.UserID(uuid.Must(uuid.FromString(tx.Credit))),
			Amount:               tx.Amount,
			TransactionReference: types.TransactionReference(fmt.Sprintf("REFUND - %s", tx.TransactionReference)),
			Description:          "Reverse transaction",
			Group:                types.TransactionGroupStore,
			SubGroup:             "Refund",
			RelatedTransactionID: null.StringFrom(tx.ID),
		}
		_, _, _, err = ucm.Transact(refundTx)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get transaction")
		}
		return http.StatusOK, nil
	}
	return fn
}
func UserHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	u, err := boiler.Users(
		boiler.UserWhere.PublicAddress.EQ(null.StringFrom(publicAddress.Hex())),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "Could not get user")
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "User does not exist")
	}
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not marshal user")
	}
	return http.StatusOK, nil
}

func GetUserByUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("userID cannot be empty"), "Unable to find userID, userID empty.")
	}
	u, err := boiler.Users(
		boiler.UserWhere.ID.EQ(userID),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "Could not get user")
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "User does not exist")
	}
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not marshal user")
	}
	return http.StatusOK, nil
}

func GetUserByUsername(w http.ResponseWriter, r *http.Request) (int, error) {
	username := chi.URLParam(r, "username")
	if username == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("username cannot be empty"), "Unable to find username, username empty.")
	}
	u, err := boiler.Users(
		boiler.UserWhere.Username.EQ(username),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "Could not get user")
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "User does not exist")
	}
	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not marshal user")
	}
	return http.StatusOK, nil
}

func SyncStoreItems(w http.ResponseWriter, r *http.Request) (int, error) {
	err := db.SyncStoreItems()
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not sync store items")
	}
	return http.StatusOK, nil
}
func SyncPurchasedItems(w http.ResponseWriter, r *http.Request) (int, error) {
	err := db.SyncPurchasedItems()
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not sync purchased items")
	}
	return http.StatusOK, nil
}
func ListUserTransactions(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	u, err := boiler.Users(boiler.UserWhere.PublicAddress.EQ(null.StringFrom(publicAddress.Hex()))).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not get users")
	}
	txes, err := boiler.Transactions(qm.Where("credit = ? OR debit = ?", u.ID, u.ID)).All(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not list txes")
	}
	err = json.NewEncoder(w).Encode(txes)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not encode JSON")
	}
	return http.StatusOK, nil
}
func ListUsers(w http.ResponseWriter, r *http.Request) (int, error) {
	result := []*types.User{}
	_, err := db.UserList(r.Context(), passdb.Conn, &result, "", false, nil, 0, 20000, db.UserColumnUsername, db.SortByDirAsc)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not list users")
	}
	if len(result) == 0 {
		result = []*types.User{}
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not encode JSON")
	}
	return http.StatusOK, nil
}

func RenameBanUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("userID cannot be empty"), "Unable to find userID, userID empty.")
	}
	banned := chi.URLParam(r, "banned")
	if banned == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("banned status cannot be empty"), "Unable to find banned status, banned status empty.")
	}

	user, err := boiler.FindUser(passdb.StdConn, userID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to find user")
	}

	user.RenameBanned = null.BoolFrom(strings.ToLower(banned) == "true")

	_, err = user.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update user renamed banned status")
	}

	return http.StatusOK, nil
}

func RenameAsset(w http.ResponseWriter, r *http.Request) (int, error) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("hash cannot be empty"), "Unable to find hash, hash empty.")
	}
	newName := chi.URLParam(r, "newName")
	if newName == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("newName cannot be empty"), "Unable to find newName, newName empty.")
	}

	item, err := db.PurchasedItemByHash(hash)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to find asset.")
	}
	if item == nil {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("asset is nil"), "Unable to find asset, asset nil.")
	}

	// update asset name
	item, err = db.PurchasedItemSetName(uuid.Must(uuid.FromString(item.ID)), newName)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Unable to update asset name.")
	}

	return http.StatusOK, nil
}

func RenameBanUsername(w http.ResponseWriter, r *http.Request) (int, error) {
	username := chi.URLParam(r, "username")
	if username == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("username cannot be empty"), "Unable to find username, userID empty.")
	}
	banned := chi.URLParam(r, "banned")
	if banned == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("banned status cannot be empty"), "Unable to find banned status, banned status empty.")
	}

	user, err := boiler.Users(boiler.UserWhere.Username.EQ(username)).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to find user")
	}

	user.RenameBanned = null.BoolFrom(strings.ToLower(banned) == "true")

	_, err = user.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update user renamed banned status")
	}

	return http.StatusOK, nil
}

func ChatTimeoutUserID(w http.ResponseWriter, r *http.Request) (int, error) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("userID cannot be empty"), "Unable to find userID, userID empty.")
	}
	minutes := chi.URLParam(r, "minutes")
	minutesInt, err := strconv.Atoi(minutes)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to create int from minutes")
	}

	user, err := boiler.FindUser(passdb.StdConn, userID)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to find user")
	}

	user.ChatBannedUntil = null.TimeFrom(time.Now().Add(time.Minute * time.Duration(minutesInt)))

	_, err = user.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update user chat banned time")
	}

	return http.StatusOK, nil
}

func ChatTimeoutUsername(w http.ResponseWriter, r *http.Request) (int, error) {
	username := chi.URLParam(r, "username")
	if username == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("username cannot be empty"), "Unable to find username, username empty.")
	}
	minutes := chi.URLParam(r, "minutes")
	minutesInt, err := strconv.Atoi(minutes)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to create int from minutes")
	}

	user, err := boiler.Users(boiler.UserWhere.Username.EQ(username)).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Unable to find user")
	}

	user.ChatBannedUntil = null.TimeFrom(time.Now().Add(time.Minute * time.Duration(minutesInt)))

	_, err = user.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to update user chat banned time")
	}

	return http.StatusOK, nil
}

func ListStoreItems(w http.ResponseWriter, r *http.Request) (int, error) {
	storeItems, err := db.StoreItems()
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not list store items")
	}
	if len(storeItems) == 0 {
		storeItems = []*boiler.StoreItem{}
	}
	err = json.NewEncoder(w).Encode(storeItems)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not encode JSON")
	}
	return http.StatusOK, nil
}

func ListPurchasedItems(w http.ResponseWriter, r *http.Request) (int, error) {
	purchasedItems, err := db.PurchasedItems()
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not list store items")
	}
	if len(purchasedItems) == 0 {
		purchasedItems = []*boiler.PurchasedItem{}
	}
	err = json.NewEncoder(w).Encode(purchasedItems)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not encode JSON")
	}
	return http.StatusOK, nil
}

func PurchasedItemRegisterHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	templateIdStr := chi.URLParam(r, "template_id")
	templateId, err := uuid.FromString(templateIdStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad template ID")
	}
	ownerIdStr := chi.URLParam(r, "owner_id")
	ownerId, err := uuid.FromString(ownerIdStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad owner ID")
	}
	result, err := db.PurchasedItemRegister(templateId, ownerId)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not register new purchased item")
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not encode JSON")
	}
	return http.StatusOK, nil
}

func PurchasedItemSetOwner(w http.ResponseWriter, r *http.Request) (int, error) {
	purchasedItemIdStr := chi.URLParam(r, "purchased_item_id")
	purchasedItemId, err := uuid.FromString(purchasedItemIdStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad purchasedItem ID")
	}
	ownerIdStr := chi.URLParam(r, "owner_id")
	ownerId, err := uuid.FromString(ownerIdStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad owner ID")
	}
	result, err := db.PurchasedItemSetOwner(purchasedItemId, ownerId)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not change owner")
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not encode JSON")
	}
	return http.StatusOK, nil

}

func AdminCheck(w http.ResponseWriter, r *http.Request) (int, error) {
	w.Write([]byte("ok"))
	return http.StatusOK, nil
}
