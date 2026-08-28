package handlers

// Character holds a player character for a specific edition.
type Character struct {
	ID             int64
	UserID         int64
	EditionID      int64
	EditionSlug    string
	Name           string
	DistrictID     int64
	DistrictName   string
	MoneyCents     int
	ContainerLevel int
	Mood           int
}

// Edition is a separate game world.
type Edition struct {
	ID     int64
	Slug   string
	Name   string
	Active bool
}

// District is a collection location within an edition.
type District struct {
	ID            int64
	EditionID     int64
	Name          string
	YieldFactor   float64
	UnlockAtCents int
	RiskFactor    float64
	DistanceRank  int
}
