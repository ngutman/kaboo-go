package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// UserCollection is the users mongo collection name
	UserCollection = "users"
)

// User object in the system
type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	ExternalID string             `bson:"external_id"`
	Username   string             `bson:"username"`
}

// UserDAO handles all user related db interactions
type UserDAO struct {
	collection *mongo.Collection
}

// FetchUserByExternalID returns a user using his external id (e.g. Auth0)
// TODO: Add missing indices
func (d *UserDAO) FetchUserByExternalID(externalID string) (user *User, err error) {
	var returnedUser User
	filter := bson.M{"external_id": externalID}
	err = d.collection.FindOne(context.Background(), filter).Decode(&returnedUser)
	if err != nil {
		return nil, err
	}
	return &returnedUser, nil
}
