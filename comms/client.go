package comms

import (
	"time"
)

func (c *C) SupremacyUserConnectionUpgradeHandler(req UserConnectionUpgradeReq, resp *UserConnectionUpgradeResp) error {
	// register game server client session id to passport
	c.HubSessionIDMap.Store(req.SessionID, time.Now())
	return nil
}
