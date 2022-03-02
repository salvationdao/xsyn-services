package comms

import (
	"context"
	"errors"
	"fmt"
	"passport"
	"passport/api"
	"passport/db"

	"github.com/ninja-software/terror/v2"
	"github.com/ninja-syndicate/hub/ext/messagebus"
)

var ErrNoRows = errors.New("no rows in result set")

type DefaultWarMachinesReq struct {
	FactionID passport.FactionID `json:"factionID"`
}

type DefaultWarMachinesResp struct {
	WarMachines []*passport.WarMachineMetadata `json:"warMachines"`
}

func (c *C) SupremacyDefaultWarMachinesHandler(req DefaultWarMachinesReq, resp *DefaultWarMachinesResp) error {

	ctx := context.Background()
	var warMachines []*passport.WarMachineMetadata
	// check user own this asset, and it has not joined the queue yet
	switch req.FactionID {
	case passport.RedMountainFactionID:
		faction, err := db.FactionGet(ctx, c.Conn, passport.RedMountainFactionID)
		if err != nil {
			return terror.Error(err)
		}
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, c.Conn, passport.SupremacyRedMountainUserID)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineMetadata := &passport.WarMachineMetadata{}
			// parse metadata
			passport.ParseWarMachineMetadata(wmmd, warMachineMetadata)
			warMachineMetadata.OwnedByID = passport.SupremacyRedMountainUserID
			warMachineMetadata.FactionID = passport.RedMountainFactionID
			warMachineMetadata.Faction = faction

			// TODO: set war machine ability, when ability is available

			warMachines = append(warMachines, warMachineMetadata)
		}
	case passport.BostonCyberneticsFactionID:
		faction, err := db.FactionGet(ctx, c.Conn, passport.BostonCyberneticsFactionID)
		if err != nil {
			return terror.Error(err)
		}
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, c.Conn, passport.SupremacyBostonCyberneticsUserID)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineMetadata := &passport.WarMachineMetadata{}
			// parse metadata
			passport.ParseWarMachineMetadata(wmmd, warMachineMetadata)
			warMachineMetadata.OwnedByID = passport.SupremacyBostonCyberneticsUserID
			warMachineMetadata.FactionID = passport.BostonCyberneticsFactionID
			warMachineMetadata.Faction = faction

			// TODO: set war machine ability, when ability is available

			warMachines = append(warMachines, warMachineMetadata)
		}
	case passport.ZaibatsuFactionID:
		faction, err := db.FactionGet(ctx, c.Conn, passport.ZaibatsuFactionID)
		if err != nil {
			return terror.Error(err)
		}
		warMachinesMetaData, err := db.DefaultWarMachineGet(ctx, c.Conn, passport.SupremacyZaibatsuUserID)
		if err != nil {
			return terror.Error(err)
		}
		for _, wmmd := range warMachinesMetaData {
			warMachineMetadata := &passport.WarMachineMetadata{}
			// parse metadata
			passport.ParseWarMachineMetadata(wmmd, warMachineMetadata)
			warMachineMetadata.OwnedByID = passport.SupremacyZaibatsuUserID
			warMachineMetadata.FactionID = passport.ZaibatsuFactionID
			warMachineMetadata.Faction = faction

			// TODO: set war machine ability, when ability is available

			warMachines = append(warMachines, warMachineMetadata)
		}
	}
	c.Log.Info().Str("fn", "SupremacyDefaultWarMachinesHandler").Str("faction_id", req.FactionID.String()).Msg("rpc")
	resp.WarMachines = warMachines
	return nil
}

type WarMachineQueuePositionReq struct {
	WarMachineQueuePosition []*passport.WarMachineQueueStat `json:"warMachineQueuePosition"`
}

type WarMachineQueuePositionResp struct{}

// SupremacyWarMachineQueuePositionHandler broadcast the updated battle queue position detail
func (c *C) SupremacyWarMachineQueuePositionHandler(req WarMachineQueuePositionReq, resp *WarMachineQueuePositionResp) error {
	ctx := context.Background()
	// broadcast war machine position to all user client
	for _, uwm := range req.WarMachineQueuePosition {
		c.MessageBus.Send(ctx, messagebus.BusKey(fmt.Sprintf("%s:%s", api.HubKeyWarMachineQueueStatSubscribe, uwm.Hash)), uwm)
	}

	return nil
}
