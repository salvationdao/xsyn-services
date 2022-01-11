package passport

type Faction struct {
	ID       FactionID `json:"id"`
	Label    string    `json:"label"`
	ImageUrl string    `json:"imageUrl"`
	Colour   string    `json:"colour"`
}
