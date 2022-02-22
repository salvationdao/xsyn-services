package db

var TableNames = struct {
	Blobs             string
	IssuedTokens      string
	Organisations     string
	Roles             string
	UserActivities    string
	UserOrganisations string
	Users             string
	Products          string
	Factions          string
}{
	Blobs:             "blobs",
	IssuedTokens:      "issued_tokens",
	Organisations:     "organisations",
	Roles:             "roles",
	UserActivities:    "user_activities",
	UserOrganisations: "user_organisations",
	Users:             "users",
	Products:          "products",
	Factions:          "factions",
}
