package handlers

// Character holds a player character for a specific edition.
type Character struct {
	ID             int64
	UserID         int64
	EditionID      int64
	EditionSlug    string
	Name           string
	SubjectID      int64
	SubjectName    string
	MoneyCents     int
	ContainerLevel int
	Mood           int
}

// Edition is a separate game world (MVP: one fictional uni-city).
type Edition struct {
	ID     int64
	Slug   string
	Name   string
	Active bool
}

// Subject is a Studienfach, the class system of the game.
type Subject struct {
	ID        int64
	EditionID int64
	Name      string
}
