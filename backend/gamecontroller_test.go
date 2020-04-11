package backend

import (
	"context"
	"testing"

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

func Test_CreatingNewGame(t *testing.T) {
	db, client := clearAndOpenDb(t)
	user := addUserToDB(t, client, "userid123", "user", "user@user.com")

	controller := NewGameController(db)
	gameID, err := controller.NewGame(user, "game1", 5, "password")
	createdGameID, _ := primitive.ObjectIDFromHex(gameID)
	if err != nil {
		t.Errorf("Error creating a new game, %v\n", err)
	}
	t.Logf("Created a new game result %v\n", gameID)
	_, err = controller.NewGame(user, "game1", 5, "password")
	if err == nil {
		t.Errorf("Should have failed creating a new game for user")
	}
	var game models.KabooGame
	client.Database(TestingDB).Collection(models.GamesCollection).
		FindOne(context.Background(), bson.M{"_id": createdGameID}).Decode(&game)
	if game.Owner != user.ID {
		t.Errorf("Unexpected game created %v\n", game)
	}
}

func Test_LoadingActiveGames(t *testing.T) {
	db, client := clearAndOpenDb(t)
	user := addUserToDB(t, client, "userid123", "user", "user@user.com")
	// Create a game before loading the controller
	game, _ := db.GamesDAO.CreateGame(user, "game1", 4, "password")
	controller := NewGameController(db)
	if controller.userToActiveGames[user.ID] == nil || controller.activeGames[game.ID] == nil {
		t.Errorf("Should have loaded active game from db")
	}
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
