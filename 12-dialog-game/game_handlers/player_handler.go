package game_handlers

import "stepikGoWebServices/entities"

func (h *GameHandler) HandlePlayer(player *entities.Player) {
	for command := range player.Input {
		h.HandleCommand(player, command)
	}
}
