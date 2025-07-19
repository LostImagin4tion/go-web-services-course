package game_handlers

import (
	"stepikGoWebServices/entities"
	"strings"
)

func (h *GameHandler) HandleCommand(player *entities.Player, command string) {
	h.gameMutex.Lock()
	defer h.gameMutex.Unlock()

	parts := strings.Fields(command)
	if len(parts) == 0 {
		player.SendMessage("неизвестная команда")
		return
	}

	cmd := parts[0]

	switch cmd {
	case "осмотреться":
		h.handleLook(player)
	case "идти":
		if len(parts) < 2 {
			player.SendMessage("неизвестная команда")
			return
		}
		h.handleMove(player, parts[1])
	case "взять":
		if len(parts) < 2 {
			player.SendMessage("неизвестная команда")
			return
		}
		h.handleTake(player, parts[1])
	case "одеть":
		if len(parts) < 2 {
			player.SendMessage("неизвестная команда")
			return
		}
		h.handleWear(player, parts[1])
	case "применить":
		if len(parts) < 3 {
			player.SendMessage("неизвестная команда")
			return
		}
		h.handleUse(player, parts[1], parts[2])
	case "сказать":
		if len(parts) < 2 {
			player.SendMessage("неизвестная команда")
			return
		}
		message := strings.Join(parts[1:], " ")
		h.handleSay(player, message)
	case "сказать_игроку":
		if len(parts) < 2 {
			player.SendMessage("неизвестная команда")
			return
		}
		targetPlayer := parts[1]
		var message string
		if len(parts) > 2 {
			message = strings.Join(parts[2:], " ")
		}
		h.handleSayToPlayer(player, targetPlayer, message)
	default:
		player.SendMessage("неизвестная команда")
	}
}

// formatRoomItems форматирует предметы в комнате
func formatRoomItems(items []string) string {
	if len(items) == 0 {
		return "пустая комната"
	}

	tableItems := []string{}
	chairItems := []string{}

	for _, item := range items {
		if item == "рюкзак" {
			chairItems = append(chairItems, item)
		} else {
			tableItems = append(tableItems, item)
		}
	}

	var parts []string
	if len(tableItems) > 0 {
		parts = append(parts, "на столе: "+strings.Join(tableItems, ", "))
	}
	if len(chairItems) > 0 {
		parts = append(parts, "на стуле - "+strings.Join(chairItems, ", "))
	}

	return strings.Join(parts, ", ")
}

// hasItem проверяет наличие предмета в списке
func hasItem(items []string, item string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}

// removeItem удаляет предмет из списка
func removeItem(items []string, item string) []string {
	var result []string
	for _, i := range items {
		if i != item {
			result = append(result, i)
		}
	}
	return result
}
