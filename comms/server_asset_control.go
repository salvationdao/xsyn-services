package comms

import (
	"fmt"
	"passport/api"

	"github.com/ninja-syndicate/hub/ext/messagebus"
)

// AssetContractRewardRedeem redeem faction contract reward
func (c *S) SupremacyAssetRepairStatUpdateHandler(req AssetRepairStatReq, resp *AssetRepairStatResp) error {

	// if repair complete, send nil
	if req.AssetRepairRecord.CompletedAt != nil {
		c.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyAssetRepairStatUpdate, req.AssetRepairRecord.Hash)), nil)
		return nil
	}

	// if repair not complete, send current record
	c.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyAssetRepairStatUpdate, req.AssetRepairRecord.Hash)), req.AssetRepairRecord)
	return nil
}

type SupremacyQueueUpdateResp struct {
}
type SupremacyQueueUpdateReq struct {
	Hash           string  `json:"hash"`
	Position       *int    `json:"position"`
	ContractReward *string `json:"contractReward"`
}

func (c *S) SupremacyQueueUpdateHandler(req SupremacyQueueUpdateReq, resp *SupremacyQueueUpdateResp) error {

	// if repair not complete, send current record
	go c.MessageBus.Send(messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyWarMachineQueueStatSubscribe, req.Hash)), req)
	return nil
}
