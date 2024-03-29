package types

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
	LogoBlobID       BlobID        `json:"logo_blob_id" db:"logo_blob_id"`
	BackgroundBlobID BlobID        `json:"background_blob_id" db:"background_blob_id"`
	Theme            *FactionTheme `json:"theme" db:"theme"`
	Description      string        `json:"description" db:"description"`
}

type FactionStat struct {
	Velocity      int64      `json:"velocity"`
	RecruitNumber int64      `json:"recruit_number"`
	ID            FactionID  `json:"id" db:"id"`
	WinCount      int64      `json:"win_count"`
	LossCount     int64      `json:"loss_count"`
	KillCount     int64      `json:"kill_count"`
	DeathCount    int64      `json:"death_count"`
	MVP           *UserBrief `json:"mvp,omitempty"`
	SupsVoted     string     `json:"sups_voted"`
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
		Description: `Red Mountain is the leader in autonomous mining operations in the Supremacy era. It controls vast territories on Mars, as well as the entire continent formerly known as Australia on Earth. In addition to the production of War Machines, Red Mountain has an economy built on mining, space transportation and energy production. Its AI platforms are directed by REDNET and its leading human assistant is ChiefX. The main tiers of humans include Executives, Engineers and Mechanics. 

		By enlisting in Red Mountain, you are joining the greatest interplanetary mining syndicate ever assembled.  `,
	},
	{
		ID:    BostonCyberneticsFactionID,
		Label: "Boston Cybernetics",
		Theme: &FactionTheme{
			Primary:    "#428EC1",
			Secondary:  "#FFFFFF",
			Background: "#080C12",
		},
		Description: `Boston Cybernetics is the major commercial leader within the Supremacy Era. It has expansive territories comprising 275 Districts located along the east coast of the former United States. In addition to the production of War Machines, its economy is built on finance, memory production and the exploration of asteroid belts. Boston Cybernetics AI platforms are directed by BOSSDAN and the human assistant is Patron-A. The three main tiers of humans include Patrons, CyRiders and Stackers. AIs include Synths and Rexeon Guards.

		By enlisting in Boston Cybernetics, you are joining a financial and commercial superpower with plans for space colonization.`,
	},
	{
		ID:    ZaibatsuFactionID,
		Label: "Zaibatsu Heavy Industries",
		Theme: &FactionTheme{
			Primary:    "#FFFFFF",
			Secondary:  "#000000",
			Background: "#0D0D0D",
		},
		Description: `Zaibatsu is the industrial leader within the Supremacy era, with heavily populated territories dominated by neon city skyscrapers across the islands formerly known as Japan. In addition to the production of War Machines, Zaibatsu’s economy is built on production optimized by human and AI interaction, as well as the development of future cloud cities. Its AI platforms are directed by ZAIA and the leading human assistant is A1. The three main tiers of humans include APEXRs, KODRs and DENZRs. 

		By enlisting in Zaibatsu, you are joining a powerhouse in city construction and industrial production.`,
	},
}

type FactionSaleAvailable struct {
	ID            FactionID     `json:"id" db:"id"`
	Label         string        `json:"label" db:"label"`
	LogoBlobID    BlobID        `json:"logo_blob_id" db:"logo_blob_id"`
	Theme         *FactionTheme `json:"theme" db:"theme"`
	MegaAmount    int64         `json:"mega_amount" db:"mega_amount"`
	LootboxAmount int64         `json:"lootbox_amount" db:"lootbox_amount"`
}

type AssetQueueStat struct {
	Position       *int    `json:"position"`
	ContractReward *string `json:"contract_reward"`
}
