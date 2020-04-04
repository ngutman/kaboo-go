package service

type GameController struct{}

type NewGameResult struct {
	GameID string
}

func (g *GameController) NewGame(name string, playersCount int, password string) (result NewGameResult, err error) {
	result.GameID = "1234"
	return result, nil
}
