package game_handlers

import "stepikGoWebServices/entities"

func (h *GameHandler) handleTake(player *entities.Player, item string) {
	if !hasItem(player.Wearing, "рюкзак") {
		player.SendMessage("некуда класть")
		return
	}

	locations := h.gameState["locations"].(map[string]map[string]interface{})
	location := locations[player.Location]
	items := location["items"].([]string)

	if !hasItem(items, item) {
		player.SendMessage("нет такого")
		return
	}

	newItems := removeItem(items, item)
	location["items"] = newItems

	player.Inventory = append(player.Inventory, item)
	player.SendMessage("предмет добавлен в инвентарь: " + item)
}
