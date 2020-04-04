package service

// User is the user representation in the system
type User struct {
	ID         string
	ExternalID string
	Name       string
	Email      string
}

type Player struct {
	User User
	// TODO: Put cards.. etc'
}

type Game struct {
	Owner           User
	Name            string
	NumberOfPlayers int
	Password        string
	Players         []Player // Includes owner
}
