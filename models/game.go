package models

import (
	"context"
	"log"

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
)

// KabooGame represents a game, contains the game state
type KabooGame struct {
	ID         primitive.ObjectID   `bson:"_id,omitempty"`
	Owner      primitive.ObjectID   `bson:"owner"`
	State      GameState            `bson:"state"`
	Players    []primitive.ObjectID `bson:"players"`
	MaxPlayers int                  `bson:"max_players"`
	Name       string               `bson:"name"`
	Password   string               `bson:"password"`
}

// CreateGame creates a game for the given user
func (d *Db) CreateGame(owner *User, name string, maxPlayers int, password string) *KabooGame {
	collection := d.database.Collection(GamesCollection)
	game := KabooGame{
		primitive.NilObjectID,
		owner.ID,
		GameStateInitializing,
		[]primitive.ObjectID{owner.ID},
		maxPlayers,
		name,
		password,
	}
	res, err := collection.InsertOne(context.Background(), game)
	if err != nil {
		log.Fatalf("Couldn't create game %v\n", err)
		return nil
	}
	game.ID = res.InsertedID.(primitive.ObjectID)
	return &game
}
