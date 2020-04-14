package backend

import (
	"errors"
	"sync"

	"github.com/ngutman/kaboo-server-go/models"
	"github.com/ngutman/kaboo-server-go/transport/websocket"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	// ErrAlreadyInGame user already in game
	ErrAlreadyInGame = errors.New("User already in game")

	// ErrCreateGame failure creating game
	ErrCreateGame = errors.New("Failed creating game")

	// ErrJoinGameAlreadyStarted error since game already started
	ErrJoinGameAlreadyStarted = errors.New("Game already started")

	// ErrWrongGamePassword wrong password
	ErrWrongGamePassword = errors.New("Wrong password")

	// ErrGameDoesntExist game doesn't exist
	ErrGameDoesntExist = errors.New("Game doesn't exist")
)

// MessageSender websocket message sender interface
type MessageSender interface {
	BroadcastMessageToUsers(users []primitive.ObjectID, message interface{})
}

// GameController manages the games, allows players to join, leave or create games
type GameController struct {
	userToActiveGames map[primitive.ObjectID]*models.KabooGame
	activeGames       map[primitive.ObjectID]*models.KabooGame
	db                *models.Db
	sender            MessageSender
	gameMtx           *sync.Mutex
}

// NewGameController returns a new game controller
func NewGameController(db *models.Db, sender MessageSender) *GameController {
	controller := GameController{
		userToActiveGames: make(map[primitive.ObjectID]*models.KabooGame),
		activeGames:       make(map[primitive.ObjectID]*models.KabooGame),
		db:                db,
		sender:            sender,
		gameMtx:           &sync.Mutex{},
	}
	controller.loadGames()
	return &controller
}

// NewGame create a new game returning the created game id on success
// A player can only create a game if he's not participating in any running games
func (g *GameController) NewGame(user *models.User, name string,
	maxPlayers int, password string) (string, error) {
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
func (g *GameController) JoinGameByGameID(user *models.User, strGameID string, password string) (bool, error) {
	if g.db.GamesDAO.IsPlayerInActiveGame(user.ID) {
		log.Debugf("User %v (%v) already participating in a game\n", user.Username, user.ID.Hex())
		return false, ErrAlreadyInGame
	}
	gameID, _ := primitive.ObjectIDFromHex(strGameID)
	game := g.activeGames[gameID]
	if game == nil {
		return false, ErrGameDoesntExist
	}
	if game.State != models.GameStateWaitingForPlayers {
		return false, ErrJoinGameAlreadyStarted
	}
	if password != game.Password {
		return false, ErrWrongGamePassword
	}
	success, err := g.db.GamesDAO.TryToAddPlayerToGame(game, user)
	if err != nil {
		return false, err
	}
	g.sender.BroadcastMessageToUsers(game.Players, websocket.WSMessageUserJoinedGame{
		MessageType: 0, GameID: game.ID.Hex(), User: websocket.User{ID: user.ID.Hex(), Name: user.Username},
	})
	return success, nil
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
