package db

var TableNames = struct {
	Blobs          string
	IssuedTokens   string
	Roles          string
	UserActivities string
	Users          string
	Factions       string
}{
	Blobs:          "blobs",
	IssuedTokens:   "issued_tokens",
	Roles:          "roles",
	UserActivities: "user_activities",
	Users:          "users",
	Factions:       "factions",
}
