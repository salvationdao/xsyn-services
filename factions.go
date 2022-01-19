package passport

type FactionTheme struct {
	Primary    string `json:"primary"`
	Secondary  string `json:"secondary"`
	Background string `json:"background"`
}

type Faction struct {
	ID       FactionID     `json:"id" db:"id"`
	Label    string        `json:"label" db:"label"`
	ImageUrl string        `json:"imageUrl"`
	Theme    *FactionTheme `json:"theme" db:"theme"`
}
