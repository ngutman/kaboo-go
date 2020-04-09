package backend

import (
	"context"
	"testing"

	"github.com/ngutman/kaboo-server-go/api/types"

	"github.com/ngutman/kaboo-server-go/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TestingURI = "mongodb://localhost:27017"
	TestingDB  = "kaboo_test"
)

func TestCreatingNewGame(t *testing.T) {
	db, client := clearAndOpenDb(t)

	user := addUserToDB(t, client, "userid123", "user", "user@user.com")

	controller := NewGameController(db)
	ctx := contextWithUserID("userid123")
	result, err := controller.NewGame(ctx, "game1", 5, "password")
	createdGameID, _ := primitive.ObjectIDFromHex(result.GameID)
	if err != nil {
		t.Errorf("Error creating a new game, %v\n", err)
	}
	t.Logf("Created a new game result %v\n", result)
	result, err = controller.NewGame(ctx, "game1", 5, "password")
	if err == nil {
		t.Errorf("Should have failed creating a new game for user")
	}
	var game models.KabooGame
	client.Database(TestingDB).Collection(models.GamesCollection).
		FindOne(context.Background(), bson.D{bson.E{Key: "_id", Value: createdGameID}}).Decode(&game)
	if game.Owner != user.ID {
		t.Errorf("Unexpected game created %v\n", game)
	}
}

func contextWithUserID(externalUserID string) context.Context {
	return context.WithValue(context.Background(), types.ContextUserKey, externalUserID)
}

func clearAndOpenDb(t *testing.T) (*models.Db, *mongo.Client) {
	clientOptions := options.Client().ApplyURI(TestingURI)
	client, _ := mongo.Connect(context.Background(), clientOptions)
	if err := client.Database(TestingDB).Drop(context.Background()); err != nil {
		t.Error(err)
		return nil, nil
	}

	var db models.Db
	db.Open(TestingURI, TestingDB)
	return &db, client
}

func addUserToDB(t *testing.T, client *mongo.Client, externalUserID string, username string, email string) *models.User {
	user := models.User{
		ExternalID: externalUserID,
		Username:   username,
	}
	userID, err := client.Database(TestingDB).Collection("users").InsertOne(context.Background(), user)
	user.ID = userID.InsertedID.(primitive.ObjectID)
	if err != nil {
		t.Errorf("Couldn't add user %v\n", err)
		return nil
	}
	return &user
}
