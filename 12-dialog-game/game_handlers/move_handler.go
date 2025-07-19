package game_handlers

import (
	"stepikGoWebServices/entities"
	"strings"
)

func (h *GameHandler) handleMove(player *entities.Player, direction string) {
	locations := h.gameState["locations"].(map[string]map[string]interface{})
	location := locations[player.Location]
	exits := location["exits"].([]string)

	if !hasItem(exits, direction) {
		player.SendMessage("нет пути в " + direction)
		return
	}

	if direction == "улица" && !h.gameState["doorOpen"].(bool) {
		player.SendMessage("дверь закрыта")
		return
	}

	player.Location = direction

	newLocation := locations[direction]
	message := newLocation["description"].(string)
	newExits := newLocation["exits"].([]string)
	message += ". можно пройти - " + strings.Join(newExits, ", ")

	player.SendMessage(message)
}
