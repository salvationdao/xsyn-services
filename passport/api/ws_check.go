package api

import (
	"context"

	"github.com/ninja-syndicate/ws"

	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/terror/v2"

	"github.com/rs/zerolog"
)

// CheckControllerWS holds handlers for checking server status
type CheckControllerWS struct {
	Log *zerolog.Logger
	API *API
}

// NewCheckController creates the check hub
func NewCheckController(log *zerolog.Logger, api *API) *CheckControllerWS {
	checkHub := &CheckControllerWS{
		Log: log_helpers.NamedLogger(log, "check_hub"),
		API: api,
	}

	api.Command(HubKeyCheck, checkHub.Handler)

	return checkHub
}

// HubKeyCheck is used to route to the  handler
const HubKeyCheck = "CHECK"

type CheckResponse struct {
	Check string `json:"check"`
}

func (ch *CheckControllerWS) Handler(ctx context.Context, key string, payload []byte, reply ws.ReplyFunc) error {
	response := CheckResponse{Check: "ok"}
	err := check()
	if err != nil {
		return terror.Error(err, "Server check failed, try again or contact support.")
	}

	reply(response)
	return nil
}
