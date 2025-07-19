package game_handlers

import (
	"stepikGoWebServices/entities"
	"sync"
)

type GameHandler struct {
	gameMutex *sync.RWMutex
	players   map[string]*entities.Player
	gameState map[string]interface{}
}

func NewGameHandler() *GameHandler {
	return &GameHandler{
		gameMutex: &sync.RWMutex{},
		players:   make(map[string]*entities.Player),
		gameState: map[string]interface{}{
			"doorOpen": false,
			"locations": map[string]map[string]interface{}{
				"кухня": {
					"description": "ты находишься на кухне, на столе чай, надо собрать рюкзак и идти в универ",
					"exits":       []string{"коридор"},
					"items":       []string{},
				},
				"коридор": {
					"description": "ничего интересного",
					"exits":       []string{"кухня", "комната", "улица"},
					"items":       []string{},
				},
				"комната": {
					"description": "ты в своей комнате",
					"exits":       []string{"коридор"},
					"items":       []string{"ключи", "конспекты", "рюкзак"},
				},
				"улица": {
					"description": "на улице весна",
					"exits":       []string{"домой"},
					"items":       []string{},
				},
			},
		},
	}
}

func (h *GameHandler) AddPlayer(player *entities.Player) {
	h.gameMutex.Lock()
	h.players[player.Name] = player
	h.gameMutex.Unlock()

	go h.HandlePlayer(player)
}
