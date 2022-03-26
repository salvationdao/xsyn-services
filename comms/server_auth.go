package comms

import (
	"fmt"
	"passport"
	"passport/db/boiler"
	"passport/passdb"
	"passport/passlog"

	"github.com/ninja-software/terror/v2"
)

// IsServerClient checks the given api key is a server client
func IsServerClient(apikey string) (string, error) {
	if apikey == "" {
		passlog.L.Err(fmt.Errorf("missing api key")).Msg("api key empty")
		return "", terror.Error(fmt.Errorf("missing api key"))
	}

	apiKeyEntry, err := boiler.FindAPIKey(passdb.StdConn, apikey)
	if err != nil {
		passlog.L.Err(err).Str("api_key", apikey).Msg("error finding api key")
		return "", terror.Error(err)
	}

	if apiKeyEntry.Type != "SERVER_CLIENT" {
		passlog.L.Err(err).Str("api_key", apikey).Str("key_type", apiKeyEntry.Type).Str("required_type", "SERVER_CLIENT").Msg("does not have permission")
		return "", terror.Error(fmt.Errorf("api key is missing SERVER_CLIENT permission"))
	}

	return apiKeyEntry.UserID, nil
}

// IsSupremacyClient checks if the given api key belongs to the supremacy user
func IsSupremacyClient(apikey string) (string, error) {
	if apikey == "" {
		passlog.L.Err(fmt.Errorf("missing api key")).Msg("api key empty")
		return "", terror.Error(fmt.Errorf("missing api key"))
	}

	apiKeyEntry, err := boiler.FindAPIKey(passdb.StdConn, apikey)
	if err != nil {
		passlog.L.Err(err).Str("api_key", apikey).Msg("error finding api key")
		return "", terror.Error(err)
	}

	if apiKeyEntry.Type != "SERVER_CLIENT" {
		passlog.L.Err(err).Str("api_key", apikey).Str("key_type", apiKeyEntry.Type).Str("required_type", "SERVER_CLIENT").Msg("does not have permission")
		return "", terror.Error(fmt.Errorf("api key is missing SERVER_CLIENT permission"))
	}

	user, err := boiler.FindUser(passdb.StdConn, apiKeyEntry.UserID)
	if err != nil {
		passlog.L.Err(err).Str("api_key", apikey).Str("user_id", apiKeyEntry.UserID).Msg("error finding user from api key")
		return "", terror.Error(err)
	}

	if user.Username != passport.SupremacyGameUsername {
		passlog.L.Err(err).Str("api_key", apikey).Str("key_username", user.Username).Str("expect_username", passport.SupremacyGameUsername).Msg("username mismatch")
		return "", terror.Error(fmt.Errorf("api key owner username mismatch"))
	}
	return user.ID, nil
}
