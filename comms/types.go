package comms

import (
	"passport"
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/volatiletech/null/v8"
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
	GroupID              passport.TransactionGroup     `json:"groupID,omitempty"`
	NotSafe              bool                          `json:"notSafe"`
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
	*User
}
type GetMechOwnerResp struct {
}
type Mech struct {
	ID           uuid.UUID
	Hash         string
	CollectionID uuid.UUID

	TemplateID uuid.UUID
	OwnerByID  uuid.UUID

	Brand                 string
	Model                 string
	Skin                  string
	Name                  string
	AssetType             string
	MaxStructureHitPoints string
	MaxShieldHitPoints    string
	Speed                 string
	WeaponHardpoints      string
	TurretHardpoints      string
	UtilitySlots          string
	WeaponOne             string
	WeaponTwo             string
	TurretOne             string
	TurretTwo             string
	UtilityOne            string
	ShieldRechargeRate    string

	CreatedAt time.Time
	DeletedAt null.Time
	UpdatedAt time.Time
}

type MigrateOnlyGetAllMechsReq struct {
}

type MigrateOnlyGetAllMechsResp struct {
	Mechs []*Mech
}

type MechTemplate struct {
	ID           uuid.UUID
	CollectionID uuid.UUID

	Hash                  string
	Brand                 string
	Model                 string
	Skin                  string
	Name                  string
	AssetType             string
	MaxStructureHitPoints string
	MaxShieldHitPoints    string
	Speed                 string
	WeaponHardpoints      string
	TurretHardpoints      string
	UtilitySlots          string
	WeaponOne             string
	WeaponTwo             string
	TurretOne             string
	TurretTwo             string
	UtilityOne            string
	ShieldRechargeRate    string

	CreatedAt time.Time
	DeletedAt null.Time
	UpdatedAt time.Time
}
type MigrateOnlyGetAllTemplatesReq struct {
}
type MigrateOnlyGetAllTemplatesResp struct {
	MechTemplates []*MechTemplate
}
