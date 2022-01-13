package passport

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/volatiletech/null/v8"
)

// User is a single user on the platform
type User struct {
	ID                               UserID        `json:"id" db:"id"`
	FirstName                        string        `json:"firstName" db:"first_name"`
	LastName                         string        `json:"lastName" db:"last_name"`
	Email                            null.String   `json:"email" db:"email"`
	FacebookID                       null.String   `json:"-" db:"facebook_id"`
	GoogleID                         null.String   `json:"-" db:"google_id"`
	TwitchID                         null.String   `json:"-" db:"twitch_id"`
	FactionID                        *FactionID    `json:"factionID" db:"faction_id"`
	Faction                          *Faction      `json:"faction"`
	Username                         string        `json:"username" db:"username"`
	Verified                         bool          `json:"verified" db:"verified"`
	OldPasswordRequired              bool          `json:"oldPasswordRequired" db:"old_password_required"`
	RoleID                           RoleID        `json:"roleID" db:"role_id"`
	Role                             Role          `json:"role" db:"role"`
	Organisation                     *Organisation `json:"organisation" db:"organisation"`
	AvatarID                         *BlobID       `json:"avatarID" db:"avatar_id"`
	Sups                             int64         `json:"sups" db:"sups"`
	Online                           bool          `json:"online"`
	TwoFactorAuthenticationActivated bool          `json:"twoFactorAuthenticationActivated" db:"two_factor_authentication_activated"`
	TwoFactorAuthenticationSecret    string        `json:"twoFactorAuthenticationSecret" db:"two_factor_authentication_secret"`
	TwoFactorAuthenticationIsSet     bool          `json:"twoFactorAuthenticationIsSet" db:"two_factor_authentication_is_set"`
	HasRecoveryCode                  bool          `json:"hasRecoveryCode" db:"has_recovery_code"`
	Pass2FA                          bool          `json:"pass2FA"`
	Nonce                            null.String   `json:"-" db:"nonce"`
	PublicAddress                    null.String   `json:"publicAddress,omitempty" db:"public_address"`
	CreatedAt                        time.Time     `json:"createdAt" db:"created_at"`
	UpdatedAt                        time.Time     `json:"updatedAt" db:"updated_at"`
	DeletedAt                        *time.Time    `json:"deletedAt" db:"deleted_at"`
}

// IsAdmin returns true if user is a admin role
func (user *User) IsAdmin() bool {
	return user.RoleID == UserRoleAdminID || user.RoleID == UserRoleSuperAdminID
}

// IsSuperAdmin returns true if user is a super admin role
func (user *User) IsSuperAdmin() bool {
	return user.RoleID == UserRoleSuperAdminID
}

// IsMember returns true if user is a member
func (user *User) IsMember() bool {
	return user.RoleID == UserRoleMemberID
}

// IssueToken contains token information used for login and verifying accounts
type IssueToken struct {
	ID     IssueTokenID `json:"id" db:"id"`
	UserID UserID       `json:"userId" db:"user_id"`
}

func (i IssueToken) Whitelisted() bool {
	if !i.ID.IsNil() {
		return true
	}
	return false
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
