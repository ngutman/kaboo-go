package models

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameState game state type
type GameState int

// Game states
const (
	GameStateWaitingForPlayers GameState = iota
	GameStateOngoing
	GameStateEnded
)

const (
	// GamesCollection name of the games collection
	GamesCollection = "games"
	// GameSeedLength length of level seed
	GameSeedLength = 32
)

// KabooGame represents a game, contains the game state
type KabooGame struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty"`
	Owner      primitive.ObjectID   `bson:"owner"`
	State      GameState            `bson:"state"`
	Active     bool                 `bson:"active"`
	Players    []primitive.ObjectID `bson:"players"`
	MaxPlayers int                  `bson:"max_players"`
	Name       string               `bson:"name"`
	Password   string               `bson:"password"`
	Seed       string               `bson:"seed"`
}

// GamesDAO is handling all game related actions against the db
type GamesDAO struct {
	collection *mongo.Collection
	gmtx       sync.Mutex
}

// CreateGame creates a game for the given user
func (g *GamesDAO) CreateGame(owner *User, name string, maxPlayers int, password string) (*KabooGame, error) {
	seed, err := generateGameSeed()
	log.Tracef("Generated seed - %v\n", seed)
	if err != nil {
		log.Errorf("Error generating level seed, %v\n", err)
		return nil, err
	}
	game := KabooGame{
		primitive.NilObjectID,
		owner.ID,
		GameStateWaitingForPlayers,
		true,
		[]primitive.ObjectID{owner.ID},
		maxPlayers,
		name,
		password,
		seed,
	}
	res, err := g.collection.InsertOne(context.Background(), game)
	if err != nil {
		log.Fatalf("Couldn't insert level to db, %v\n", err)
		return nil, err
	}
	game.ID = res.InsertedID.(primitive.ObjectID)
	return &game, nil
}

// FetchActiveGames returns active games from the db
func (g *GamesDAO) FetchActiveGames() (results []*KabooGame, err error) {
	filter := bson.M{"active": true}
	cursor, err := g.collection.Find(context.Background(), filter)
	if err != nil {
		log.Errorf("Error fetching active games, %v\n", err)
		return results, err
	}
	err = cursor.All(context.Background(), &results)
	if err != nil {
		log.Errorf("Error fetching active games, %v\n", err)
		return results, err
	}
	return results, nil
}

// IsPlayerInActiveGame returns if given player is participating in any active game
func (g *GamesDAO) IsPlayerInActiveGame(user primitive.ObjectID) bool {
	filter := bson.M{"players": user, "active": true}
	count, err := g.collection.CountDocuments(context.Background(), filter)
	if err != nil {
		log.Panicf("Error fetching player games, %v\n", err)
	}
	return count > 0
}

// TryToAddPlayerToGame attempts to add the player to the given game, will fail if there are too many players
func (g *GamesDAO) TryToAddPlayerToGame(game *KabooGame, user *User) (bool, error) {
	g.gmtx.Lock()
	defer g.gmtx.Unlock()

	if len(game.Players) >= game.MaxPlayers {
		return false, fmt.Errorf("Too many players in game (maximum %d, current %d)", game.MaxPlayers, len(game.Players))
	}

	game.Players = append(game.Players, user.ID)
	if _, err := g.collection.UpdateOne(context.Background(), bson.M{"_id": game.ID}, bson.M{"$set": bson.M{"players": game.Players}}); err != nil {
		return false, err
	}

	return true, nil
}

func generateGameSeed() (seed string, err error) {
	b := make([]byte, GameSeedLength)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	seed = base64.URLEncoding.EncodeToString(b)
	return
}
