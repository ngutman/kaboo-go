package models

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User object in the system
type User struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	ExternalID string             `bson:"external_id"`
	Username   string             `bson:"username"`
}

// FetchUserByExternalID returns a user using his external id (e.g. Auth0)
// TODO: Add missing indices
func (d *Db) FetchUserByExternalID(externalID string) (user *User, err error) {
	var returnedUser User
	filter := bson.D{bson.E{Key: "external_id", Value: externalID}}
	err = d.database.Collection("users").FindOne(context.Background(), filter).Decode(&returnedUser)
	if err != nil {
		return nil, err
	}
	return &returnedUser, nil
}
