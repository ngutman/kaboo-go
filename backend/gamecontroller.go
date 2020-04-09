package backend

import (
	"context"
	"errors"

	"github.com/ngutman/kaboo-server-go/api/types"
	"github.com/ngutman/kaboo-server-go/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameController manages the games, allows players to join, leave or create games
type GameController struct {
	activeGames map[primitive.ObjectID]*models.KabooGame
	db          *models.Db
}

// NewGameController returns a new game controller
func NewGameController(db *models.Db) *GameController {
	return &GameController{
		activeGames: make(map[primitive.ObjectID]*models.KabooGame),
		db:          db,
	}
}

// NewGame create a new game returning the created game id on success
// A player can only create a game if he's not participating in any running games
func (g *GameController) NewGame(ctx context.Context, name string,
	maxPlayers int, password string) (*types.NewGameResult, error) {
	user, err := g.db.FetchUserByExternalID(ctx.Value(types.ContextUserKey).(string))
	if err != nil {
		return nil, err
	}
	if g.activeGames[user.ID] != nil {
		log.Debug("User %v (%v) already participating in a game\n", user.Username, user.ID.Hex())
		return nil, errors.New("User already in game")
	}
	game, err := g.db.CreateGame(user, name, maxPlayers, password)
	if err != nil {
		return nil, errors.New("Error creating game")
	}
	g.activeGames[user.ID] = game

	var result types.NewGameResult
	result.GameID = game.ID.Hex()
	return &result, nil
}

func (g *GameController) loadGames() error {
	return nil
}
