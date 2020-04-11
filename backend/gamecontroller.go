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
	// ErrUserNotFound user not found error
	ErrUserNotFound = errors.New("User not found")

	// ErrAlreadyInGame user already in game
	ErrAlreadyInGame = errors.New("User already in game")

	// ErrCreateGame failure creating game
	ErrCreateGame = errors.New("Failed creating game")
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
	maxPlayers int, password string) (string, error) {
	user, err := g.db.UserDAO.FetchUserByExternalID(ctx.Value(types.ContextUserKey).(string))
	if err != nil {
		return "", ErrUserNotFound
	}
	if g.db.GamesDAO.IsPlayerInActiveGame(user.ID) {
		log.Debugf("User %v (%v) already participating in a game\n", user.Username, user.ID.Hex())
		return "", ErrAlreadyInGame
	}
	game, err := g.db.GamesDAO.CreateGame(user, name, maxPlayers, password)
	if err != nil {
		return "", ErrCreateGame
	}

	g.registerActiveGame(game)

	return game.ID.Hex(), nil
}

// JoinGameByGameID the user asks to join a specific game
func (g *GameController) JoinGameByGameID(ctx context.Context, strGameID string, password string) (bool, error) {
	user, err := g.db.UserDAO.FetchUserByExternalID(ctx.Value(types.ContextUserKey).(string))
	if err != nil {
		return false, ErrUserNotFound
	}
	if g.db.GamesDAO.IsPlayerInActiveGame(user.ID) {
		log.Debugf("User %v (%v) already participating in a game\n", user.Username, user.ID.Hex())
		return false, ErrAlreadyInGame
	}
	gameID, _ := primitive.ObjectIDFromHex(strGameID)
	game := g.activeGames[gameID]
	if game.State != models.GameStateWaitingForPlayers {
		return true, nil
	}
	// 1. Check game status
	// 2. Check enough players
	// 3. Check password
	return true, nil
}

func (g *GameController) loadGames() error {
	games, err := g.db.GamesDAO.FetchActiveGames()
	if err != nil {
		return err
	}
	for _, game := range games {
		g.registerActiveGame(game)
	}
	log.Infof("Loaded %d active games\n", len(games))
	return nil
}

func (g *GameController) registerActiveGame(game *models.KabooGame) {
	g.gameMtx.Lock()
	defer g.gameMtx.Unlock()

	g.userToActiveGames[game.Owner] = game
	g.activeGames[game.ID] = game
}
