package passport

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/ninja-syndicate/hub"
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
	FirstName                        string        `json:"firstName" db:"first_name"`
	LastName                         string        `json:"lastName" db:"last_name"`
	Email                            null.String   `json:"email" db:"email"`
	FacebookID                       null.String   `json:"facebookID" db:"facebook_id"`
	GoogleID                         null.String   `json:"googleID" db:"google_id"`
	TwitchID                         null.String   `json:"twitchID" db:"twitch_id"`
	TwitterID                        null.String   `json:"twitterID" db:"twitter_id"`
	DiscordID                        null.String   `json:"discordID" db:"discord_id"`
	FactionID                        *FactionID    `json:"factionID" db:"faction_id"`
	Faction                          *Faction      `json:"faction"`
	Username                         string        `json:"username" db:"username"`
	Verified                         bool          `json:"verified" db:"verified"`
	OldPasswordRequired              bool          `json:"oldPasswordRequired" db:"old_password_required"`
	RoleID                           RoleID        `json:"roleID" db:"role_id"`
	Role                             Role          `json:"role" db:"role"`
	Organisation                     *Organisation `json:"organisation" db:"organisation"`
	AvatarID                         *BlobID       `json:"avatarID" db:"avatar_id"`
	Sups                             BigInt
	Online                           bool         `json:"online"`
	TwoFactorAuthenticationActivated bool         `json:"twoFactorAuthenticationActivated" db:"two_factor_authentication_activated"`
	TwoFactorAuthenticationSecret    string       `json:"twoFactorAuthenticationSecret" db:"two_factor_authentication_secret"`
	TwoFactorAuthenticationIsSet     bool         `json:"twoFactorAuthenticationIsSet" db:"two_factor_authentication_is_set"`
	HasRecoveryCode                  bool         `json:"hasRecoveryCode" db:"has_recovery_code"`
	Pass2FA                          bool         `json:"pass2FA"`
	Nonce                            null.String  `json:"-" db:"nonce"`
	PublicAddress                    null.String  `json:"publicAddress,omitempty" db:"public_address"`
	CreatedAt                        time.Time    `json:"createdAt" db:"created_at"`
	UpdatedAt                        time.Time    `json:"updatedAt" db:"updated_at"`
	DeletedAt                        *time.Time   `json:"deletedAt" db:"deleted_at"`
	Metadata                         UserMetadata `json:"metadata" db:"metadata"`
}

type UserBrief struct {
	ID       UserID  `json:"id" db:"id"`
	Username string  `json:"username" db:"username"`
	AvatarID *BlobID `json:"avatarID" db:"avatar_id"`
}

type UserMetadata struct {
	BoughtStarterWarmachines int  `json:"boughtStarterWarmachines"`
	BoughtLootboxes          int  `json:"boughtLootboxes"`
	WatchedVideo             bool `json:"watchedVideo"`
}

// IsAdmin is needed for the hub interface, no admins here!
func (user *User) IsAdmin() bool {
	return false
}

// IsMember returns true if user is a member
func (user *User) IsMember() bool {
	return user.RoleID == UserRoleMemberID
}

// IssueToken contains token information used for login and verifying accounts
type IssueToken struct {
	ID     IssueTokenID `json:"id" db:"id"`
	UserID UserID       `json:"userID" db:"user_id"`
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
	ViewBattleCount       int64  `json:"viewBattleCount"`
	TotalVoteCount        int64  `json:"totalVoteCount"`
	TotalAbilityTriggered int64  `json:"totalAbilityTriggered"`
	KillCount             int64  `json:"killCount"`
}

type UserSupsMultiplierSend struct {
	ToUserID        UserID            `json:"toUserID"`
	ToUserSessionID *hub.SessionID    `json:"toUserSessionID,omitempty"`
	SupsMultipliers []*SupsMultiplier `json:"supsMultiplier"`
}

type SupsMultiplier struct {
	Key       string    `json:"key"`
	Value     int       `json:"value"`
	ExpiredAt time.Time `json:"expiredAt"`
}

type WarMachineQueueStat struct {
	Hash           string  `json:"hash"`
	Position       *int    `json:"position,omitempty"`
	ContractReward *string `json:"contractReward,omitempty"`
}
