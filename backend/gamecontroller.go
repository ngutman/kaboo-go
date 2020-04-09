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
	controller := GameController{
		activeGames: make(map[primitive.ObjectID]*models.KabooGame),
		db:          db,
	}
	controller.loadGames()
	return &controller
}

// NewGame create a new game returning the created game id on success
// A player can only create a game if he's not participating in any running games
func (g *GameController) NewGame(ctx context.Context, name string,
	maxPlayers int, password string) (*types.NewGameResult, error) {
	user, err := g.db.UserDAO.FetchUserByExternalID(ctx.Value(types.ContextUserKey).(string))
	if err != nil {
		return nil, err
	}
	// TODO: This is pretty bad, need to check if there's error (got a bit lazy)
	if playerActiveInGame, err := g.db.GamesDAO.IsPlayerInActiveGame(user.ID); err != nil || playerActiveInGame {
		log.Debugf("User %v (%v) already participating in a game\n", user.Username, user.ID.Hex())
		return nil, errors.New("User already in game")
	}
	game, err := g.db.GamesDAO.CreateGame(user, name, maxPlayers, password)
	if err != nil {
		return nil, errors.New("Error creating game")
	}
	g.activeGames[user.ID] = game

	var result types.NewGameResult
	result.GameID = game.ID.Hex()
	return &result, nil
}

func (g *GameController) loadGames() error {
	games, err := g.db.GamesDAO.FetchActiveGames()
	if err != nil {
		return err
	}
	for _, game := range games {
		g.activeGames[game.Owner] = game
	}
	log.Infof("Loaded %d active games\n", len(games))
	return nil
}
