package backend

import (
	"context"
	"testing"

	"github.com/ngutman/kaboo-server-go/api/types"

	"github.com/ngutman/kaboo-server-go/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TestingURI = "mongodb://localhost:27017"
	TestingDB  = "kaboo_test"
)

func TestCreatingNewGame(t *testing.T) {
	db, client := clearAndOpenDb(t)

	addUserToDB(t, client, "userid123", "user", "user@user.com")

	controller := NewGameController(db)
	context := contextWithUserID("userid123")
	result, err := controller.NewGame(context, "game1", 5, "password")
	if err != nil {
		t.Errorf("Error creating a new game, %v\n", err)
	}
	t.Logf("Created a new game result %v\n", result)
	result, err = controller.NewGame(context, "game1", 5, "password")
	if err == nil {
		t.Errorf("Should have failed creating a new game for user")
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
	_, err := client.Database(TestingDB).Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		t.Errorf("Couldn't add user %v\n", err)
		return nil
	}
	return &user
}
