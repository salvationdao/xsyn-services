package types

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
	"github.com/shopspring/decimal"
	"github.com/volatiletech/null/v8"
)

const (
	XsynTreasuryUsername               string = "Xsyn"
	SupremacyGameUsername              string = "Supremacy"
	SupremacyBattleUsername            string = "Supremacy-Battle-Arena"
	SupremacySupPoolUsername           string = "Supremacy-Sup-Pool"
	SupremacyZaibatsuUsername          string = "Zaibatsu"
	SupremacyRedMountainUsername       string = "RedMountain"
	SupremacyBostonCyberneticsUsername string = "BostonCybernetics"
	OnChainUsername                    string = "OnChain"
	XsynSaleUsername                   string = "XsynSale"
)

var (
	XsynTreasuryUserID               = UserID(uuid.Must(uuid.FromString("ebf30ca0-875b-4e84-9a78-0b3fa36a1f87")))
	SupremacyGameUserID              = UserID(uuid.Must(uuid.FromString("4fae8fdf-584f-46bb-9cb9-bb32ae20177e")))
	SupremacyBattleUserID            = UserID(uuid.Must(uuid.FromString("87c60803-b051-4abb-aa60-487104946bd7")))
	SupremacySupPoolUserID           = UserID(uuid.Must(uuid.FromString("c579bb47-7efb-4286-a5cc-e5edbb54626d")))
	SupremacyZaibatsuUserID          = UserID(uuid.Must(uuid.FromString("1a657a32-778e-4612-8cc1-14e360665f2b")))
	SupremacyRedMountainUserID       = UserID(uuid.Must(uuid.FromString("305da475-53dc-4973-8d78-a30d390d3de5")))
	SupremacyBostonCyberneticsUserID = UserID(uuid.Must(uuid.FromString("15f29ee9-e834-4f76-aff8-31e39faabe2d")))
	OnChainUserID                    = UserID(uuid.Must(uuid.FromString("2fa1a63e-a4fa-4618-921f-4b4d28132069")))
	XsynSaleUserID                   = UserID(uuid.Must(uuid.FromString("1429a004-84a1-11ec-a8a3-0242ac120002")))
)

func (e UserID) IsSystemUser() bool {
	switch e {
	case XsynTreasuryUserID,
		SupremacyGameUserID,
		OnChainUserID,
		SupremacyBattleUserID,
		SupremacySupPoolUserID,
		SupremacyZaibatsuUserID,
		SupremacyRedMountainUserID,
		SupremacyBostonCyberneticsUserID:
		return true
	}
	return false
}

const ServerClientLevel = 5

// User is a single user on the platform
type User struct {
	ID                               UserID        `json:"id" db:"id"`
	FirstName                        string        `json:"first_name" db:"first_name"`
	LastName                         string        `json:"last_name" db:"last_name"`
	Email                            null.String   `json:"email" db:"email"`
	FacebookID                       null.String   `json:"facebook_id" db:"facebook_id"`
	GoogleID                         null.String   `json:"google_id" db:"google_id"`
	TwitchID                         null.String   `json:"twitch_id" db:"twitch_id"`
	TwitterID                        null.String   `json:"twitter_id" db:"twitter_id"`
	DiscordID                        null.String   `json:"discord_id" db:"discord_id"`
	FactionID                        *FactionID    `json:"faction_id" db:"faction_id"`
	MobileNumber                     null.String   `json:"mobile_number" db:"mobile_number"`
	Faction                          *Faction      `json:"faction"`
	Username                         string        `json:"username" db:"username"`
	Verified                         bool          `json:"verified" db:"verified"`
	OldPasswordRequired              bool          `json:"old_password_required" db:"old_password_required"`
	RoleID                           RoleID        `json:"role_id" db:"role_id"`
	Role                             Role          `json:"role" db:"role"`
	Organisation                     *Organisation `json:"organisation" db:"organisation"`
	AvatarID                         *BlobID       `json:"avatar_id" db:"avatar_id"`
	Sups                             BigInt
	Online                           bool         `json:"online"`
	TwoFactorAuthenticationActivated bool         `json:"two_factor_authentication_activated" db:"two_factor_authentication_activated"`
	TwoFactorAuthenticationSecret    string       `json:"two_factor_authentication_secret" db:"two_factor_authentication_secret"`
	TwoFactorAuthenticationIsSet     bool         `json:"two_factor_authentication_is_set" db:"two_factor_authentication_is_set"`
	HasRecoveryCode                  bool         `json:"has_recovery_code" db:"has_recovery_code"`
	Pass2FA                          bool         `json:"pass_2_fa"`
	Nonce                            null.String  `json:"-" db:"nonce"`
	PublicAddress                    null.String  `json:"public_address,omitempty" db:"public_address"`
	CreatedAt                        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt                        time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt                        *time.Time   `json:"deleted_at" db:"deleted_at"`
	Metadata                         UserMetadata `json:"metadata" db:"metadata"`
	WithdrawLock                     bool         `json:"withdraw_lock" db:"withdraw_lock"`
	MintLock                         bool         `json:"mint_lock" db:"mint_lock"`
	TotalLock                        bool         `json:"total_lock" db:"total_lock"`
}

type UserBrief struct {
	ID       UserID  `json:"id" db:"id"`
	Username string  `json:"username" db:"username"`
	AvatarID *BlobID `json:"avatar_id" db:"avatar_id"`
}

type UserMetadata struct {
	BoughtStarterWarmachines int  `json:"bought_starter_warmachines"`
	BoughtLootboxes          int  `json:"bought_lootboxes"`
	WatchedVideo             bool `json:"watched_video"`
}

// IsAdmin is needed for the hub interface, no admins here!
func (user *User) IsAdmin() bool {
	return false
}

// IsMember returns true if user is a member
func (user *User) IsMember() bool {
	return user.RoleID == UserRoleMemberID
}

func (user *User) CheckUserIsLocked(level string) bool {
	if level == "withdrawals" && user.WithdrawLock {
		return true
	}

	if level == "minting" && user.MintLock {
		return true
	}

	if level == "account" && user.TotalLock {
		return true
	}

	return false
}

// IssueToken contains token information used for login and verifying accounts
type IssueToken struct {
	ID     IssueTokenID `json:"id" db:"id"`
	UserID UserID       `json:"user_id" db:"user_id"`
}

func (i IssueToken) Whitelisted() bool {
	return !i.ID.IsNil()
}

func (i IssueToken) TokenID() uuid.UUID {
	return uuid.UUID(i.ID)
}

//
//func (i IssueToken) Token() jwt.Token {
//	data, err := base64.StdEncoding.DecodeString(string(text))
//
//	panic("implement me")
//}

// UserOnlineStatusChange is the payload sent to when a user online status changes
type UserOnlineStatusChange struct {
	ID     UserID `json:"id" db:"id"`
	Online bool   `json:"online"`
}

// from game server
type UserStat struct {
	ID                    UserID `json:"id"`
	ViewBattleCount       int64  `json:"view_battle_count"`
	TotalVoteCount        int64  `json:"total_vote_count"`
	TotalAbilityTriggered int64  `json:"total_ability_triggered"`
	KillCount             int64  `json:"kill_count"`
}

type UserSupsMultiplierSend struct {
	ToUserID        UserID            `json:"to_user_id"`
	ToUserSessionID *hub.SessionID    `json:"to_user_session_id,omitempty"`
	SupsMultipliers []*SupsMultiplier `json:"sups_multiplier"`
}

type SupsMultiplier struct {
	Key       string    `json:"key"`
	Value     int       `json:"value"`
	ExpiredAt time.Time `json:"expired_at"`
}

type WarMachineQueueStat struct {
	Hash           string          `json:"hash"`
	Position       *int            `json:"position,omitempty"`
	ContractReward decimal.Decimal `json:"contract_reward,omitempty"`
}
