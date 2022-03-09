package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
)

func AdminRoutes(ucm *UserCacheMap) chi.Router {
	r := chi.NewRouter()
	r.Get("/purchased_items", WithError(WithAdmin(ListPurchasedItems)))
	r.Get("/store_items", WithError(WithAdmin(ListStoreItems)))
	r.Get("/users", WithError(WithAdmin(ListUsers)))
	r.Post("/purchased_items/register/{template_id}/{owner_id}", WithError(WithAdmin(PurchasedItemRegisterHandler)))
	r.Post("/purchased_items/set_owner/{purchased_item_id}/{owner_id}", WithError(WithAdmin(PurchasedItemSetOwner)))
	r.Get("/check", WithError(WithAdmin(AdminCheck)))
	r.Post("/reverse_transaction/{transaction_id}", WithError(WithAdmin(ReverseUserTransaction((ucm)))))
	r.Post("/sync/store_items", WithError(WithAdmin(SyncStoreItems)))
	r.Get("/user_transactions/{public_address}", WithError(WithAdmin(ListUserTransactions)))
	return r
}

// WithAdmin checks that admin key is in the header.
func WithAdmin(next func(w http.ResponseWriter, r *http.Request) (int, error)) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		apiKeyIDStr := r.Header.Get("X-Authorization")
		apiKeyID, err := uuid.FromString(apiKeyIDStr)
		if err != nil {
			return http.StatusUnauthorized, terror.Error(err, "Unauthorized.")
		}
		apiKey, err := db.APIKey(apiKeyID)
		if err != nil {
			return http.StatusUnauthorized, terror.Error(err, "Unauthorized.")
		}
		if apiKey.Type != "ADMIN" {
			return http.StatusUnauthorized, terror.Error(fmt.Errorf("not admin key: %s", apiKey.Type), "Unauthorized.")
		}
		return next(w, r)
	}
	return fn
}
func ReverseUserTransaction(ucm *UserCacheMap) func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		txID := chi.URLParam(r, "transaction_id")
		tx, err := boiler.FindTransaction(passdb.StdConn, txID)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get transaction")
		}
		refundTx := &passport.NewTransaction{
			To:                   passport.UserID(uuid.Must(uuid.FromString(tx.Debit))),
			From:                 passport.UserID(uuid.Must(uuid.FromString(tx.Credit))),
			Amount:               *tx.Amount.BigInt(),
			TransactionReference: passport.TransactionReference(fmt.Sprintf("REFUND - %s", tx.TransactionReference)),
			Description:          "Reverse transaction",
			Group:                passport.TransactionGroupStore,
			SubGroup:             "Refund",
		}
		_, _, _, err = ucm.Process(refundTx)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get transaction")
		}
		return http.StatusOK, nil
	}
	return fn
}
func SyncStoreItems(w http.ResponseWriter, r *http.Request) (int, error) {
	err := db.SyncStoreItems()
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not sync store items")
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
	result := []*passport.User{}
	_, err := db.UserList(r.Context(), passdb.Conn, &result, "", false, nil, 0, 1000, db.UserColumnUsername, db.SortByDirAsc)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not list users")
	}
	if len(result) == 0 {
		result = []*passport.User{}
	}
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		return http.StatusBadRequest, terror.Error(err, "Could not encode JSON")
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
