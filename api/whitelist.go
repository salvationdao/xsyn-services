package api

import (
	"encoding/json"
	"net/http"

	"github.com/ninja-software/terror/v2"
)

func (api *API) WhitelistOnlyWalletCheck(w http.ResponseWriter, r *http.Request) (int, error) {
	err := json.NewEncoder(w).Encode(api.walletOnlyConnect)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err)
	}
	return http.StatusOK, nil
}
