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
	"xsyn-services/passport/asset"
	"xsyn-services/passport/db"
	"xsyn-services/passport/nft1155"
	"xsyn-services/passport/passdb"
	"xsyn-services/passport/passlog"
	"xsyn-services/passport/payments"
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
	r.Get("/store_items", WithError(WithAdmin(ListStoreItems)))

	r.Post("/purchased_items/register/{template_id}/{owner_id}", WithError(WithAdmin(PurchasedItemRegisterHandler)))
	r.Post("/purchased_items/set_owner/{purchased_item_id}/{owner_id}", WithError(WithAdmin(PurchasedItemSetOwner)))
	r.Post("/purchased_items/register/1155/{public_address}/{collection_slug}/{token_id}/{amount}", WithError(WithAdmin(Register1155Asset)))

	r.Post("/transactions/create", WithError(WithAdmin(CreateTransaction(ucm))))
	r.Post("/transactions/reverse/{transaction_id}", WithError(WithAdmin(ReverseUserTransaction(ucm))))
	r.Get("/transactions/list/user/{public_address}", WithError(WithAdmin(ListUserTransactions)))

	r.Get("/users/unlock_account/{public_address}", WithError(WithAdmin(UnlockAccount)))
	r.Get("/users/unlock_withdraw/{public_address}", WithError(WithAdmin(UnlockWithdraw)))
	r.Get("/users/unlock_mint/{public_address}", WithError(WithAdmin(UnlockMint)))

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

		item, err := boiler.PurchasedItemsOlds(
			boiler.PurchasedItemsOldWhere.ExternalTokenID.EQ(tokenID),
			boiler.PurchasedItemsOldWhere.CollectionID.EQ(c.ID),
		).One(passdb.StdConn)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get purchased item")
		}
		if item.OwnerID != from {
			return http.StatusBadRequest, errors.New("from user does not own the asset")
		}
		item.OwnerID = to
		_, err = item.Update(passdb.StdConn, boil.Whitelist(boiler.PurchasedItemsOldColumns.OwnerID))
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
		_, err = ucm.Transact(newTx)
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
		_, err = ucm.Transact(refundTx)
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

func UserGiveAdminPermission(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	permission := chi.URLParam(r, "ADMIN")
	u, err := boiler.Users(
		boiler.UserWhere.PublicAddress.EQ(null.StringFrom(publicAddress.Hex())),
	).One(passdb.StdConn)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "Could not get user")
	}
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return http.StatusBadRequest, terror.Error(err, "User does not exist")
	}
	u.Permissions = null.StringFrom(permission)
	_, err = u.Update(passdb.StdConn, boil.Whitelist(boiler.UserColumns.Permissions))
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not update user")
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
	_, err := db.UserList(result, "", false, nil, 0, 20000, db.UserColumnUsername, db.SortByDirAsc)
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

func PurchasedItemRegisterHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	templateIdStr := chi.URLParam(r, "template_id")
	templateId, err := uuid.FromString(templateIdStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad template ID")
	}
	ownerId := chi.URLParam(r, "owner_id")
	ownerUUID, err := uuid.FromString(ownerId)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad owner ID")
	}
	result, err := db.PurchasedItemRegister(templateId, ownerUUID)
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
	ownerIDStr := chi.URLParam(r, "owner_id")
	ownerID, err := uuid.FromString(ownerIDStr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad owner ID")
	}

	assetIDstr := chi.URLParam(r, "purchased_item_id")
	assetID, err := uuid.FromString(assetIDstr)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Bad asset ID")
	}

	_, err = asset.TransferAssetADMIN(assetID, ownerID)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to transfer asset.")
	}

	return http.StatusOK, nil
}

func AdminCheck(w http.ResponseWriter, r *http.Request) (int, error) {
	w.Write([]byte("ok"))
	return http.StatusOK, nil
}

func UnlockAccount(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	u, err := boiler.Users(
		boiler.UserWhere.PublicAddress.EQ(null.StringFrom(publicAddress.Hex())),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not get user")
	}

	u.WithdrawLock = false
	u.MintLock = false
	u.TotalLock = false

	_, err = u.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not update user to unlock account.")
	}

	return http.StatusOK, nil
}

func UnlockWithdraw(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	u, err := boiler.Users(
		boiler.UserWhere.PublicAddress.EQ(null.StringFrom(publicAddress.Hex())),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not get user")
	}

	u.WithdrawLock = false

	_, err = u.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not update user to unlock withdrawals.")
	}

	return http.StatusOK, nil
}

func UnlockMint(w http.ResponseWriter, r *http.Request) (int, error) {
	publicAddress := common.HexToAddress(chi.URLParam(r, "public_address"))
	u, err := boiler.Users(
		boiler.UserWhere.PublicAddress.EQ(null.StringFrom(publicAddress.Hex())),
	).One(passdb.StdConn)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not get user")
	}

	u.MintLock = false

	_, err = u.Update(passdb.StdConn, boil.Infer())
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not update user to unlock minting.")
	}

	return http.StatusOK, nil
}

func Register1155Asset(w http.ResponseWriter, r *http.Request) (int, error) {
	address := chi.URLParam(r, "public_address")
	if address == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("no public address provided when registering 1155 asset"), "No public address given")
	}
	collectionSlug := chi.URLParam(r, "collection_slug")
	if collectionSlug == "" {
		return http.StatusBadRequest, terror.Error(fmt.Errorf("no collection slug provided when registering 1155 asset"), "No collection slug given")
	}
	tokenID, err := strconv.Atoi(chi.URLParam(r, "token_id"))
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to read token id")
	}
	amount, err := strconv.Atoi(chi.URLParam(r, "amount"))
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to read amount")
	}

	user, err := payments.CreateOrGetUser(common.HexToAddress(address))
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to get user")
	}

	asset, err := nft1155.CreateOrGet1155Asset(tokenID, user, collectionSlug)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to create or get 1155 asset")
	}

	asset.Count += amount

	_, err = asset.Update(passdb.StdConn, boil.Whitelist(boiler.UserAssets1155Columns.Count))
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Failed to update assets count")
	}

	return http.StatusOK, nil
}
