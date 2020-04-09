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

// KabooGame represents a game, contains the game state
type KabooGame struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Owner primitive.ObjectID `bson:"owner"`
	State GameState          `bson:"state"`
}

// CreateEmptyGame creates a new empty game for user
func (d *Db) CreateEmptyGame(owner *User) *KabooGame {
	collection := d.database.Collection("games")
	game := KabooGame{
		primitive.NilObjectID,
		owner.ID,
		GameStateInitializing,
	}
	res, err := collection.InsertOne(context.Background(), game)
	if err != nil {
		log.Fatalf("Couldn't create game %v\n", err)
		return nil
	}
	game.ID = res.InsertedID.(primitive.ObjectID)
	return &game
}
