package backend

import (
	"context"
	"errors"
	"sync"

	"github.com/ngutman/kaboo-server-go/api/types"
	"github.com/ngutman/kaboo-server-go/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrUserNotFound  = errors.New("User not found")
	ErrAlreadyInGame = errors.New("User already in game")
	ErrCreateGame    = errors.New("Failed creating game")
)

// GameController manages the games, allows players to join, leave or create games
type GameController struct {
	userToActiveGames map[primitive.ObjectID]*models.KabooGame
	activeGames       map[primitive.ObjectID]*models.KabooGame
	db                *models.Db
	gameMtx           *sync.Mutex
}

// NewGameController returns a new game controller
func NewGameController(db *models.Db) *GameController {
	controller := GameController{
		userToActiveGames: make(map[primitive.ObjectID]*models.KabooGame),
		activeGames:       make(map[primitive.ObjectID]*models.KabooGame),
		db:                db,
		gameMtx:           &sync.Mutex{},
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
		return nil, ErrUserNotFound
	}
	if g.db.GamesDAO.IsPlayerInActiveGame(user.ID) {
		log.Debugf("User %v (%v) already participating in a game\n", user.Username, user.ID.Hex())
		return nil, ErrAlreadyInGame
	}
	game, err := g.db.GamesDAO.CreateGame(user, name, maxPlayers, password)
	if err != nil {
		return nil, ErrCreateGame
	}

	g.registerGame(game)

	var result types.NewGameResult
	result.GameID = game.ID.Hex()
	return &result, nil
}

// JoinGameByGameID the user asks to join a specific game
func (g *GameController) JoinGameByGameID(ctx context.Context, gameID string) (*types.JoinGameResult, error) {
	user, err := g.db.UserDAO.FetchUserByExternalID(ctx.Value(types.ContextUserKey).(string))
	if err != nil {
		return nil, ErrUserNotFound
	}
	if g.db.GamesDAO.IsPlayerInActiveGame(user.ID) {
		log.Debugf("User %v (%v) already participating in a game\n", user.Username, user.ID.Hex())
		return nil, ErrAlreadyInGame
	}
	return nil, nil
}

func (g *GameController) loadGames() error {
	games, err := g.db.GamesDAO.FetchActiveGames()
	if err != nil {
		return err
	}
	for _, game := range games {
		g.registerGame(game)
	}
	log.Infof("Loaded %d active games\n", len(games))
	return nil
}

func (g *GameController) registerGame(game *models.KabooGame) {
	g.gameMtx.Lock()
	defer g.gameMtx.Unlock()

	g.userToActiveGames[game.Owner] = game
	g.activeGames[game.ID] = game
}
