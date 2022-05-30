package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/ninja-software/terror/v2"
	"github.com/volatiletech/null/v8"
	"io/ioutil"
	"net/http"
	"xsyn-services/boiler"
	"xsyn-services/passport/db"
	"xsyn-services/passport/passdb"
)

type Collections1155Resp struct {
	Name          string             `json:"name"`
	Description   null.String        `json:"description"`
	Slug          string             `json:"slug"`
	MintContract  null.String        `json:"mint_contract"`
	LogoURL       null.String        `json:"logo_url"`
	BackgroundURL null.String        `json:"background_url"`
	Balances      *AvantUserBalances `json:"balances,omitempty"`
}

func (api *API) Get1155Collections(w http.ResponseWriter, r *http.Request) (int, error) {
	collections, err := boiler.Collections(
		boiler.CollectionWhere.ContractType.EQ(null.StringFrom("EIP-1155")),
		boiler.CollectionWhere.IsVisible.EQ(null.BoolFrom(true)),
	).All(passdb.StdConn)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get 1155 collections")
	}

	var collectionResp []Collections1155Resp
	for _, collection := range collections {
		collectionResp = append(collectionResp, Collections1155Resp{
			Name:          collection.Name,
			Description:   collection.Description,
			Slug:          collection.Slug,
			MintContract:  collection.MintContract,
			LogoURL:       collection.LogoURL,
			BackgroundURL: collection.BackgroundURL,
		})
	}

	err = json.NewEncoder(w).Encode(struct {
		Collections []Collections1155Resp `json:"collections"`
	}{
		Collections: collectionResp,
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get 1155 collections")
	}

	return http.StatusOK, nil
}

type AvantUserBalances struct {
	OwnerAddress string             `json:"owner_address"`
	Balances     []AvantUserBalance `json:"balances"`
}

type AvantUserBalance struct {
	TokenID       int    `json:"token_id"`
	Value         string `json:"value"`
	ValueInt      string `json:"value_int"`
	ValueDecimals int    `json:"value_decimals"`
}

func (api *API) Get1155Collection(w http.ResponseWriter, r *http.Request) (int, error) {
	collectionSlug := chi.URLParam(r, "collection_slug")
	if collectionSlug == "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to provide collection slug"), "Please provide collection slug")
	}
	address := chi.URLParam(r, "public_address")
	if collectionSlug == "" {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("failed to provide collection slug"), "Please provide collection slug")
	}

	collection, err := db.CollectionBySlug(collectionSlug)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection details")
	}
	url := fmt.Sprintf("http://v3-staging.supremacy-api.avantdata.com:3001/api/multi_token_balance?contract_address=%s&owner_address=%s&is_testnet=%t", collection.MintContract.String, address, true)
	req, err := http.NewRequest("GET", url, nil)
	fmt.Println(url)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get collection details")
	}
	if resp.StatusCode != http.StatusOK {
		return http.StatusInternalServerError, terror.Error(fmt.Errorf("response code"), "Failed to get balances")
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	var result []*AvantUserBalances
	err = json.Unmarshal(b, &result)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = json.NewEncoder(w).Encode(Collections1155Resp{
		Name:          collection.Name,
		Description:   collection.Description,
		Slug:          collection.Slug,
		MintContract:  collection.MintContract,
		LogoURL:       collection.LogoURL,
		BackgroundURL: collection.BackgroundURL,
		Balances:      result[0],
	})
	if err != nil {
		return http.StatusInternalServerError, terror.Error(err, "Failed to get 1155 collections")
	}

	return http.StatusOK, nil
}
