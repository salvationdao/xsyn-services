package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"passport"

	"github.com/ninja-software/terror/v2"
)

// GameserverRequest set gameserver webhook request
func (api *API) GameserverRequest(method string, endpoint string, data interface{}, dist interface{}) error {
	jd, err := json.Marshal(data)
	if err != nil {
		return terror.Error(err, "failed to marshal data into json struct")
	}

	url := fmt.Sprintf("%s/api/%s/Supremacy_game%s", api.GameserverHostUrl, passport.SupremacyGameUserID, endpoint)

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jd))
	if err != nil {
		return terror.Error(err)
	}

	req.Header.Add("Passport-Authorization", api.WebhookToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return terror.Error(err)
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
