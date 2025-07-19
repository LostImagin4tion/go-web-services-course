package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (tb *TaskBot) handleResolve(chatID int64, userID int, user *tgbotapi.User, taskID int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	task, exists := tb.tasks[taskID]
	if !exists {
		return
	}

	tb.sendMessage(chatID, fmt.Sprintf(`Задача "%s" выполнена`, task.Title))

	// Уведомляем автора, если он не является исполнителем
	if task.AuthorID != userID {
		tb.sendMessage(int64(task.AuthorID), fmt.Sprintf(`Задача "%s" выполнена @%s`, task.Title, user.UserName))
	}

	delete(tb.tasks, taskID)
}
