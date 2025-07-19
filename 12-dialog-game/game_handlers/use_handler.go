package game_handlers

import "stepikGoWebServices/entities"

func (h *GameHandler) handleUse(player *entities.Player, item string, target string) {
	if !hasItem(player.Inventory, item) {
		player.SendMessage("нет предмета в инвентаре - " + item)
		return
	}

	if item == "ключи" && target == "дверь" {
		h.gameState["doorOpen"] = true
		player.SendMessage("дверь открыта")
		return
	}

	player.SendMessage("не к чему применить")
}
