package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"xsyn-services/types"

	"github.com/ninja-software/terror/v2"
)

// GameserverRequest send gameserver webhook request
func (api *API) GameserverRequest(method string, endpoint string, data interface{}, dist interface{}) error {
	jd, err := json.Marshal(data)
	if err != nil {
		return terror.Error(err, "failed to marshal data into json struct")
	}

	url := fmt.Sprintf("%s/api/%s/Supremacy_game%s", api.GameserverHostUrl, types.SupremacyGameUserID, endpoint)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jd))
	if err != nil {
		return err
	}

	req.Header.Add("Passport-Authorization", api.GameserverWebhookToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return terror.Error(err, "Could not fetch contract reward")
		}
		defer resp.Body.Close()
		return terror.Error(fmt.Errorf("%s", b), "Could not fetch contract reward")
	}

	if dist != nil {
		err = json.NewDecoder(resp.Body).Decode(&dist)
		if err != nil {
			return terror.Error(err, "failed to decode response")
		}
	}

	return nil
}

// SupremacyWorldWebhookSend push a supremacy world webhook
func (api *API) SupremacyWorldWebhookSend[T any](method string, endpoint string, data []T, dist interface{}) error {
	jd, err := json.Marshal(data)
	if err != nil {
		return terror.Error(err, "failed to marshal data into json struct")
	}

	url := fmt.Sprintf("%s/api/%s", api.SupremacyWorldHostUrl, types.SupremacyGameUserID, endpoint)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jd))
	if err != nil {
		return err
	}

	req.Header.Add("Supremacy-World-Authorization", api.SupremacyWorldWebhookToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return terror.Error(err, "Could not fetch contract reward")
		}
		defer resp.Body.Close()
		return terror.Error(fmt.Errorf("%s", b), "Could not fetch contract reward")
	}

	if dist != nil {
		err = json.NewDecoder(resp.Body).Decode(&dist)
		if err != nil {
			return terror.Error(err, "failed to decode response")
		}
	}

	return nil
}
