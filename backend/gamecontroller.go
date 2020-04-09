package backend

import (
	"context"
	"errors"
	"github.com/ngutman/kaboo-server-go/api/types"
	"github.com/ngutman/kaboo-server-go/models"
)

// KabooGame represents a game, contains the game state
type KabooGame struct {
}

// GameController manages the games, allows players to join, leave or create games
type GameController struct {
	games map[models.User]*KabooGame
	db    *models.Db
}

// NewGameController returns a new game controller
func NewGameController(db *models.Db) *GameController {
	return &GameController{
		games: make(map[models.User]*KabooGame),
		db:    db,
	}
}

// NewGame create a new game returning the created game id on success
// A player can only create a game if he's not participating in any running games
func (g *GameController) NewGame(ctx context.Context, name string,
	playersCount int, password string) (*types.NewGameResult, error) {
	var result types.NewGameResult
	user, err := g.db.FetchUserByExternalID(ctx.Value(types.ContextUserKey).(string))
	if err != nil {
		return nil, err
	}
	if g.games[*user] != nil {
		return nil, errors.New("User already in game")
	}
	return &result, nil
}
