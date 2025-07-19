package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"stepikGoWebServices/entities"
)

func (tb *TaskBot) handleNew(chatID int64, userID int, user *tgbotapi.User, title string) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	task := &entities.Task{
		ID:       tb.nextID,
		Title:    title,
		Author:   user,
		AuthorID: userID,
	}

	tb.tasks[tb.nextID] = task
	tb.sendMessage(chatID, fmt.Sprintf(`Задача "%s" создана, id=%d`, title, tb.nextID))
	tb.nextID++
}
