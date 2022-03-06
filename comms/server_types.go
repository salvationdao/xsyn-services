package comms

import (
	"passport"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/sqlboiler/v4/types"
)

type AssetRepairStatReq struct {
	AssetRepairRecord *passport.AssetRepairRecord `json:"assetRepairRecord"`
}

type AssetRepairStatResp struct{}
type DefaultWarMachinesReq struct {
	FactionID passport.FactionID `json:"factionID"`
}

type DefaultWarMachinesResp struct {
	WarMachines []*passport.WarMachineMetadata `json:"warMachines"`
}
type WarMachineQueuePositionReq struct {
	WarMachineQueuePosition []*passport.WarMachineQueueStat `json:"warMachineQueuePosition"`
}

type WarMachineQueuePositionResp struct{}

type UserConnectionUpgradeReq struct {
	SessionID hub.SessionID `json:"sessionID"`
}

type UserConnectionUpgradeResp struct{}
type FactionAllReq struct{}

type FactionAllResp struct {
	Factions []*passport.Faction `json:"factions"`
}
type SpendSupsReq struct {
	Amount               string                        `json:"amount"`
	FromUserID           passport.UserID               `json:"fromUserID"`
	ToUserID             *passport.UserID              `json:"toUserID,omitempty"`
	TransactionReference passport.TransactionReference `json:"transactionReference"`
	Group                passport.TransactionGroup     `json:"group,omitempty"`
	SubGroup             string                        `json:"subGroup"`    //TODO: send battle id
	Description          string                        `json:"description"` //TODO: send descritpion

	NotSafe bool `json:"notSafe"`
}
type SpendSupsResp struct {
	TXID string `json:"txid"`
}
type ReleaseTransactionsReq struct {
	TxIDs []string `json:"txIDs"`
}
type ReleaseTransactionsResp struct{}
type TickerTickReq struct {
	UserMap map[int][]passport.UserID `json:"userMap"`
}
type TickerTickResp struct{}

type GetSpoilOfWarReq struct{}
type GetSpoilOfWarResp struct {
	Amount string
}
type UserSupsMultiplierSendReq struct {
	UserSupsMultiplierSends []*passport.UserSupsMultiplierSend `json:"userSupsMultiplierSends"`
}

type UserSupsMultiplierSendResp struct{}
type TransferBattleFundToSupPoolReq struct{}
type TransferBattleFundToSupPoolResp struct{}
type TopSupsContributorReq struct {
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type TopSupsContributorResp struct {
	TopSupsContributors       []*passport.User    `json:"topSupsContributors"`
	TopSupsContributeFactions []*passport.Faction `json:"topSupsContributeFactions"`
}

type User struct {
	ID uuid.UUID
}
type GetMechOwnerReq struct {
	Payload types.JSON
}
type GetMechOwnerResp struct {
	Payload types.JSON
}
type GetAllMechsReq struct {
}

type GetAll struct {
	AssetPayload    types.JSON
	MetadataPayload types.JSON
	StorePayload    types.JSON
	UserPayload     types.JSON
	FactionPayload  types.JSON
}

type GetAllTemplatesReq struct {
}
type GetAllTemplatesResp struct {
	Payload types.JSON
}
