package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
)

func (tb *TaskBot) HandleMessage(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if strings.HasPrefix(text, "/tasks") {
		tb.handleTasks(chatID, userID)
	} else if strings.HasPrefix(text, "/new ") {
		tb.handleNew(chatID, userID, update.Message.From, text[5:])
	} else if strings.HasPrefix(text, "/assign_") {
		taskIDStr := text[8:]
		taskID, _ := strconv.Atoi(taskIDStr)
		tb.handleAssign(chatID, userID, update.Message.From, taskID)
	} else if strings.HasPrefix(text, "/unassign_") {
		taskIDStr := text[10:]
		taskID, _ := strconv.Atoi(taskIDStr)
		tb.handleUnassign(chatID, userID, update.Message.From, taskID)
	} else if strings.HasPrefix(text, "/resolve_") {
		taskIDStr := text[9:]
		taskID, _ := strconv.Atoi(taskIDStr)
		tb.handleResolve(chatID, userID, update.Message.From, taskID)
	} else if strings.HasPrefix(text, "/my") {
		tb.handleMy(chatID, userID)
	} else if strings.HasPrefix(text, "/owner") {
		tb.handleOwner(chatID, userID)
	}
}

func (tb *TaskBot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	tb.bot.Send(msg)
}
