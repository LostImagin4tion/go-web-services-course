package game_handlers

import "stepikGoWebServices/entities"

func (h *GameHandler) handleSay(player *entities.Player, message string) {
	msg := player.Name + " говорит: " + message

	for _, p := range h.players {
		if p.Location == player.Location {
			p.SendMessage(msg)
		}
	}
}

func (h *GameHandler) handleSayToPlayer(player *entities.Player, targetName string, message string) {
	targetPlayer, exists := h.players[targetName]
	if !exists {
		player.SendMessage("тут нет такого игрока")
		return
	}

	if targetPlayer.Location != player.Location {
		player.SendMessage("тут нет такого игрока")
		return
	}

	if message == "" {
		msg := player.Name + " выразительно молчит, смотря на вас"
		targetPlayer.SendMessage(msg)
		return
	}

	msg := player.Name + " говорит вам: " + message
	targetPlayer.SendMessage(msg)
}
