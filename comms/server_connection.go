package comms

import (
	"time"
)

func (s *S) SupremacyUserConnectionUpgradeHandler(req UserConnectionUpgradeReq, resp *UserConnectionUpgradeResp) error {
	// register game server client session id to passport
	s.HubSessionIDMap.Store(req.SessionID, time.Now())
	return nil
}
