package passport

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
)

const (
	XsynTreasuryUsername  string = "Xsyn"
	SupremacyGameUsername string = "Supremacy"
	OnChainUsername       string = "OnChain"
)

var (
	XsynTreasuryUserID  = UserID(uuid.Must(uuid.FromString("ebf30ca0-875b-4e84-9a78-0b3fa36a1f87")))
	SupremacyGameUserID = UserID(uuid.Must(uuid.FromString("4fae8fdf-584f-46bb-9cb9-bb32ae20177e")))
	OnChainUserID       = UserID(uuid.Must(uuid.FromString("2fa1a63e-a4fa-4618-921f-4b4d28132069")))
)

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
	AvatarUrl                        string        `json:"avatarUrl"`
	Sups                             BigInt
	Online                           bool        `json:"online"`
	TwoFactorAuthenticationActivated bool        `json:"twoFactorAuthenticationActivated" db:"two_factor_authentication_activated"`
	TwoFactorAuthenticationSecret    string      `json:"twoFactorAuthenticationSecret" db:"two_factor_authentication_secret"`
	TwoFactorAuthenticationIsSet     bool        `json:"twoFactorAuthenticationIsSet" db:"two_factor_authentication_is_set"`
	HasRecoveryCode                  bool        `json:"hasRecoveryCode" db:"has_recovery_code"`
	Pass2FA                          bool        `json:"pass2FA"`
	Nonce                            null.String `json:"-" db:"nonce"`
	PublicAddress                    null.String `json:"publicAddress,omitempty" db:"public_address"`
	CreatedAt                        time.Time   `json:"createdAt" db:"created_at"`
	UpdatedAt                        time.Time   `json:"updatedAt" db:"updated_at"`
	DeletedAt                        *time.Time  `json:"deletedAt" db:"deleted_at"`
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
