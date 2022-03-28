package comms

import (
	"fmt"
	"passport/api"

	"github.com/ninja-syndicate/hub/ext/messagebus"
)

// SupremacyWarMachineQueuePositionHandler broadcast the updated battle queue position detail
func (c *S) SupremacyWarMachineQueuePositionHandler(req WarMachineQueuePositionReq, resp *WarMachineQueuePositionResp) error {
	// broadcast war machine position to all user client
	for _, uwm := range req.WarMachineQueuePosition {
		c.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyWarMachineQueueStatSubscribe, uwm.Hash)), uwm)
	}

	return nil
}
