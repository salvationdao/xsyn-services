package passport

import "github.com/gofrs/uuid"

var RedMountainFactionID = FactionID(uuid.Must(uuid.FromString("98bf7bb3-1a7c-4f21-8843-458d62884060")))
var BostonCyberneticsFactionID = FactionID(uuid.Must(uuid.FromString("7c6dde21-b067-46cf-9e56-155c88a520e2")))
var ZaibatsuFactionID = FactionID(uuid.Must(uuid.FromString("880db344-e405-428d-84e5-6ebebab1fe6d")))

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

type Faction struct {
	ID               FactionID     `json:"id" db:"id"`
	Label            string        `json:"label" db:"label"`
	LogoBlobID       BlobID        `json:"logoBlobID" db:"logo_blob_id"`
	BackgroundBlobID BlobID        `json:"backgroundBlobID" db:"background_blob_id"`
	Theme            *FactionTheme `json:"theme" db:"theme"`
	Description      string        `json:"description" db:"description"`
}

// from game server
type FactionStat struct {
	Velocity      int64     `json:"velocity"`
	RecruitNumber int64     `json:"recruitNumber"`
	ID            FactionID `json:"id" db:"id"`
	WinCount      int64     `json:"winCount"`
	LossCount     int64     `json:"lossCount"`
	KillCount     int64     `json:"killCount"`
	DeathCount    int64     `json:"deathCount"`
	MVP           *User     `json:"mvp,omitempty"`
	SupsVoted     string    `json:"supsVoted"`
}

var Factions = []*Faction{
	{
		ID:    RedMountainFactionID,
		Label: "Red Mountain Offworld Mining Corporation",
		Theme: &FactionTheme{
			Primary:    "#C24242",
			Secondary:  "#FFFFFF",
			Background: "#120E0E",
		},
		// NOTE: change content
		Description: "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
	},
	{
		ID:    BostonCyberneticsFactionID,
		Label: "Boston Cybernetics",
		Theme: &FactionTheme{
			Primary:    "#428EC1",
			Secondary:  "#FFFFFF",
			Background: "#080C12",
		},
		// NOTE: change content
		Description: "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
	},
	{
		ID:    ZaibatsuFactionID,
		Label: "Zaibatsu Heavy Industries",
		Theme: &FactionTheme{
			Primary:    "#FFFFFF",
			Secondary:  "#000000",
			Background: "#0D0D0D",
		},
		// NOTE: change content
		Description: "The battles spill over to the Terran economy, where SUPS are used as the keys to economic power. Terra operates a complex and interconnected economy, where everything is in limited supply, but there is also unlimited demand. If fighting isn’t your thing, Citizens can choose to be resource barons, arms manufacturers, defense contractors, tech labs and much more, with our expanding tree of resources and items to be crafted.",
	},
}
