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
	Velocity         int64         `json:"velocity" db:"velocity"`
	SharePercent     int64         `json:"sharePercent" db:"share_percent"`
	RecruitNumber    int64         `json:"recruitNumber" db:"recruit_number"`
	WinCount         int64         `json:"winCount" db:"win_count"`
	LossCount        int64         `json:"lossCount" db:"loss_count"`
	KillCount        int64         `json:"killCount" db:"kill_count"`
	DeathCount       int64         `json:"deathCount" db:"death_count"`
	MVP              string        `json:"mvp" db:"mvp"`
	LogoUrl          string        `json:"logoUrl,omitempty"`
	BackgroundUrl    string        `json:"backgroundUrl,omitempty"`
}
