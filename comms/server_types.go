package comms

import (
	"passport"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type AssetRepairStatReq struct {
	ApiKey            string
	AssetRepairRecord *passport.AssetRepairRecord `json:"asset_repair_record"`
}

type AssetRepairStatResp struct{}
type DefaultWarMachinesReq struct {
	ApiKey    string
	FactionID passport.FactionID `json:"faction_id"`
}

type DefaultWarMachinesResp struct {
	WarMachines []*passport.WarMachineMetadata `json:"war_machines"`
}
type WarMachineQueuePositionReq struct {
	ApiKey                  string
	WarMachineQueuePosition []*passport.WarMachineQueueStat `json:"war_machine_queue_position"`
}

type WarMachineQueuePositionResp struct{}

type UserConnectionUpgradeReq struct {
	ApiKey    string
	SessionID hub.SessionID `json:"session_id"`
}

type UserConnectionUpgradeResp struct{}
type FactionAllReq struct {
	ApiKey string
}

type FactionAllResp struct {
	Factions []*passport.Faction `json:"factions"`
}
type SpendSupsReq struct {
	ApiKey               string
	Amount               string                        `json:"amount"`
	FromUserID           uuid.UUID                     `json:"from_user_id"`
	ToUserID             uuid.UUID                     `json:"to_user_id"`
	TransactionReference passport.TransactionReference `json:"transaction_reference"`
	Group                passport.TransactionGroup     `json:"group,omitempty"`
	SubGroup             string                        `json:"sub_group"`   //TODO: send battle id
	Description          string                        `json:"description"` //TODO: send descritpion

	NotSafe bool `json:"not_safe"`
}
type SpendSupsResp struct {
	TXID string `json:"txid"`
}
type ReleaseTransactionsReq struct {
	ApiKey string
	TxIDs  []string `json:"tx_ids"`
}
type ReleaseTransactionsResp struct{}
type TickerTickReq struct {
	ApiKey  string
	UserMap map[int][]passport.UserID `json:"user_map"`
}
type TickerTickResp struct{}

type GetSpoilOfWarReq struct {
	ApiKey string
}
type GetSpoilOfWarResp struct {
	Amount string
}
type UserSupsMultiplierSendReq struct {
	ApiKey                  string
	UserSupsMultiplierSends []*passport.UserSupsMultiplierSend `json:"user_sups_multiplier_sends"`
}

type UserSupsMultiplierSendResp struct{}
type TransferBattleFundToSupPoolReq struct {
	ApiKey string
}
type TransferBattleFundToSupPoolResp struct{}
type TopSupsContributorReq struct {
	ApiKey    string
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type TopSupsContributorResp struct {
	TopSupsContributors       []*passport.User    `json:"top_sups_contributors"`
	TopSupsContributeFactions []*passport.Faction `json:"top_sups_contribute_factions"`
}

type User struct {
	ID uuid.UUID
}
type GetMechOwnerReq struct {
	ApiKey  string
	Payload types.JSON
}
type GetMechOwnerResp struct {
	Payload types.JSON
}
type GetAllMechsReq struct {
	ApiKey string
}

type GetAll struct {
	AssetPayload    types.JSON
	MetadataPayload types.JSON
	StorePayload    types.JSON
	UserPayload     types.JSON
	FactionPayload  types.JSON
}

type GetAllTemplatesReq struct {
	ApiKey string
}
type GetAllTemplatesResp struct {
	Payload types.JSON
}
