package api

import (
	"encoding/json"
	"net/http"
)

func (api *API) WhitelistOnlyWalletCheck(w http.ResponseWriter, r *http.Request) (int, error) {
	err := json.NewEncoder(w).Encode(api.walletOnlyConnect)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
