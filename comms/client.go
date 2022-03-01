package comms

import (
	"time"

	"github.com/ninja-syndicate/hub"
)

type UserConnectionUpgradeReq struct {
	SessionID hub.SessionID `json:"sessionID"`
}

type UserConnectionUpgradeResp struct{}

func (c *C) SupremacyUserConnectionUpgradeHandler(req UserConnectionUpgradeReq, resp *UserConnectionUpgradeResp) error {
	// register game server client session id to passport
	c.HubSessionIDMap.Store(req.SessionID, time.Now())
	return nil
}
