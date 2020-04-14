package websocket

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
