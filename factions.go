package passport

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

type Faction struct {
	ID            FactionID     `json:"id" db:"id"`
	Label         string        `json:"label" db:"label"`
	ImageUrl      string        `json:"imageUrl,omitempty" db:"image_url,omitempty"`
	Theme         *FactionTheme `json:"theme" db:"theme"`
	Description   string        `json:"description" db:"description"`
	Velocity      int64         `json:"velocity" db:"velocity"`
	SharePercent  int64         `json:"sharePercent" db:"share_percent"`
	RecruitNumber int64         `json:"recruitNumber" db:"recruit_number"`
	WinCount      int64         `json:"winCount" db:"win_count"`
	LossCount     int64         `json:"lossCount" db:"loss_count"`
	KillCount     int64         `json:"killCount" db:"kill_count"`
	DeathCount    int64         `json:"deathCount" db:"death_count"`
	MVP           string        `json:"mVP" db:"mvp"`
}
