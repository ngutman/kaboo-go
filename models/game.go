package models

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	"go.mongodb.org/mongo-driver/bson"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GameState game state type
type GameState int

// Game states
// GameStateInitializing - New game, first state
const (
	GameStateInitializing GameState = iota
	GameStateWaitingForPlayers
	GameStateOngoing
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

// CreateGame creates a game for the given user
func (d *Db) CreateGame(owner *User, name string, maxPlayers int, password string) (*KabooGame, error) {
	collection := d.database.Collection(GamesCollection)
	seed, err := generateGameSeed()
	log.Tracef("Generated seed - %v\n", seed)
	if err != nil {
		log.Errorf("Error generating level seed, %v\n", err)
		return nil, err
	}
	game := KabooGame{
		primitive.NilObjectID,
		owner.ID,
		GameStateInitializing,
		true,
		[]primitive.ObjectID{owner.ID},
		maxPlayers,
		name,
		password,
		seed,
	}
	res, err := collection.InsertOne(context.Background(), game)
	if err != nil {
		log.Fatalf("Couldn't insert level to db, %v\n", err)
		return nil, err
	}
	game.ID = res.InsertedID.(primitive.ObjectID)
	return &game, nil
}

// FetchActiveGames returns active games from the db
func (d *Db) FetchActiveGames() (results []*KabooGame, err error) {
	filter := bson.D{bson.E{Key: "active", Value: true}}
	cursor, err := d.database.Collection(GamesCollection).Find(context.Background(), filter)
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

func generateGameSeed() (seed string, err error) {
	b := make([]byte, GameSeedLength)
	_, err = rand.Read(b)
	if err != nil {
		return
	}
	seed = base64.URLEncoding.EncodeToString(b)
	return
}
