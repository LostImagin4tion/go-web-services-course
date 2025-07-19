package game_handlers

import (
	"stepikGoWebServices/entities"
	"strings"
)

func (h *GameHandler) handleLook(player *entities.Player) {
	locations := h.gameState["locations"].(map[string]map[string]interface{})
	location := locations[player.Location]

	var description string
	items := location["items"].([]string)

	if player.Location == "кухня" {
		if hasItem(player.Inventory, "ключи") && hasItem(player.Inventory, "конспекты") {
			description = "ты находишься на кухне, на столе чай, надо идти в универ"
		} else {
			description = "ты находишься на кухне, на столе чай, надо собрать рюкзак и идти в универ"
		}
	} else if player.Location == "комната" {
		if len(items) == 0 {
			description = "пустая комната"
		} else {
			description = formatRoomItems(items)
		}
	} else {
		description = location["description"].(string)
	}

	exits := location["exits"].([]string)
	message := description + ". можно пройти - " + strings.Join(exits, ", ")

	otherPlayers := h.getOtherPlayersInLocation(player)
	if len(otherPlayers) > 0 {
		message += ". Кроме вас тут ещё " + strings.Join(otherPlayers, ", ")
	}

	player.SendMessage(message)
}

func (h *GameHandler) getOtherPlayersInLocation(player *entities.Player) []string {
	var others []string
	for _, p := range h.players {
		if p.Name != player.Name && p.Location == player.Location {
			others = append(others, p.Name)
		}
	}
	return others
}
