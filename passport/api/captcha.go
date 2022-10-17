package api

import (
	"encoding/json"
	"fmt"
	"github.com/ninja-software/terror/v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"xsyn-services/passport/passlog"
)

type captcha struct {
	secret    string
	siteKey   string
	verifyUrl string
}

func (c *captcha) verify(token string) error {
	return nil

	if token == "" {
		return terror.Error(fmt.Errorf("token is empty"), "Token is empty.")
	}

	resp, err := http.PostForm(c.verifyUrl, url.Values{
		"secret":   {c.secret},
		"response": {token},
	})

	if err != nil {
		return terror.Error(err, "Failed to verify token")
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return terror.Error(err, "Read token")
	}

	type captchaResp struct {
		Success   bool     `json:"success"`
		ErrorCode []string `json:"error-codes"`
	}

	cr := &captchaResp{}
	err = json.Unmarshal(body, cr)
	if err != nil {
		return terror.Error(err, "Failed to read captcha response")
	}

	if len(cr.ErrorCode) > 0 {
		passlog.L.Error().Strs("error codes", cr.ErrorCode).Msg("error validating captcha")
	}

	if !cr.Success {
		return terror.Error(fmt.Errorf("verification failed"), "Failed to verify captcha token")
	}

	return nil
}
