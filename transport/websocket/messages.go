package websocket

import "github.com/ngutman/kaboo-server-go/models"

const (
	WSMessageTypeUserJoinsGame = iota
)

// User websocket user struct
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// WSMessageUserJoinedGame user joined a game message
type WSMessageUserJoinedGame struct {
	MessageType int    `json:"type"`
	GameID      string `json:"gameid"`
	User        User   `json:"user"`
}

// NewWSMessageUserJoinedGame create a return a new user joined game message
func NewWSMessageUserJoinedGame(game *models.KabooGame, user *models.User) WSMessageUserJoinedGame {
	return WSMessageUserJoinedGame{
		MessageType: WSMessageTypeUserJoinsGame,
		GameID:      game.ID.Hex(),
		User: User{
			ID:   user.ID.Hex(),
			Name: user.Username,
		},
	}
}
