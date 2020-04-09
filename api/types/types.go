package types

import (
	"context"
)

const (
	ContextUserKey = "userKey"
)

type NewGameResult struct {
	GameID string
}

type GameBackend interface {
	NewGame(ctx context.Context, name string, playersCount int, password string) (result *NewGameResult, err error)
}
