package game_handlers

import "stepikGoWebServices/entities"

func (h *GameHandler) handleWear(player *entities.Player, item string) {
	locations := h.gameState["locations"].(map[string]map[string]interface{})
	location := locations[player.Location]
	items := location["items"].([]string)

	if !hasItem(items, item) {
		player.SendMessage("нет такого")
		return
	}

	newItems := removeItem(items, item)
	location["items"] = newItems

	player.Wearing = append(player.Wearing, item)
	player.SendMessage("вы одели: " + item)
}
