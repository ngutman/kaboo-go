package types

import (
	"github.com/ngutman/kaboo-server-go/models"
)

type GameBackend interface {
	NewGame(user *models.User, name string, playersCount int, password string) (string, error)
	JoinGameByGameID(user *models.User, strGameID string, password string) (bool, error)
}
