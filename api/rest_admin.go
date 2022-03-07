package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"passport"
	"passport/db"
	"passport/db/boiler"
	"passport/passdb"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
)

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
