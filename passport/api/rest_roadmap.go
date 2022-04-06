package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"xsyn-services/passport/db"

	"github.com/ethereum/go-ethereum/common"
	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid"
	"github.com/ninja-software/terror/v2"
	cache "github.com/victorspringer/http-cache"
	memory "github.com/victorspringer/http-cache/adapter/memory"
)

func RoadmapRoutes() (chi.Router, error) {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create memory adaptor: %w", err)
	}
	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(10*time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create cache client: %w", err)
	}

	r := chi.NewRouter()
	r.Get("/check", WithError(RoadmapCheck))
	r.Get("/items", cacheClient.Middleware(WithError(Roadmap())).ServeHTTP)
	r.Get("/changelog", cacheClient.Middleware(WithError(Changelog())).ServeHTTP)

	return r, nil
}

type RoadmapRequest struct {
	From           uuid.UUID      `json:"from"`
	To             uuid.UUID      `json:"to"`
	CollectionAddr common.Address `json:"collection_addr"`
	TokenID        int            `json:"token_id"`
}

func RoadmapCheck(w http.ResponseWriter, r *http.Request) (int, error) {
	w.Write([]byte("ok"))
	return http.StatusOK, nil
}

func Roadmap() func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		params := url.Values{}
		apiKey := db.GetStr(db.KeyCannyApiKey)
		if apiKey == "" {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("api key not set in db"), "Missing api key in database.")
		}

		params.Add("apiKey", apiKey)
		params.Add("boardID", `624ab81cc47b9252868928de`)
		params.Add("limit", `500`)

		body := strings.NewReader(params.Encode())

		req, err := http.NewRequest("POST", "https://canny.io/api/v1/posts/list", body)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get roadmap")
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not execute request")
		}

		if resp.StatusCode != 200 {
			return http.StatusInternalServerError, fmt.Errorf("non 200 status code: %d", resp.StatusCode)
		}

		defer resp.Body.Close()

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not copy data")
		}

		return http.StatusOK, nil
	}
	return fn
}

func Changelog() func(w http.ResponseWriter, r *http.Request) (int, error) {
	fn := func(w http.ResponseWriter, r *http.Request) (int, error) {
		params := url.Values{}
		apiKey := db.GetStr(db.KeyCannyApiKey)
		if apiKey == "" {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("api key not set in db"), "Missing api key in database.")
		}

		params.Add("apiKey", apiKey)
		params.Add("limit", `500`)

		body := strings.NewReader(params.Encode())

		req, err := http.NewRequest("POST", "https://canny.io/api/v1/entries/list", body)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not get roadmap")
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not execute request")
		}
		if resp.StatusCode != 200 {
			return http.StatusInternalServerError, terror.Error(fmt.Errorf("non 200 status code: %d", resp.StatusCode))
		}
		defer resp.Body.Close()

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			return http.StatusBadRequest, terror.Error(err, "Could not copy data")
		}

		return http.StatusOK, nil
	}
	return fn
}
