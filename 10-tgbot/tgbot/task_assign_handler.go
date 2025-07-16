package handlers

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (tb *TaskBot) handleAssign(chatID int64, userID int, user *tgbotapi.User, taskID int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	task, exists := tb.tasks[taskID]
	if !exists {
		return
	}

	// Если задача уже назначена на кого-то, уведомляем старого исполнителя
	if task.Assignee != nil && task.AssigneeID != userID {
		tb.sendMessage(int64(task.AssigneeID), fmt.Sprintf(`Задача "%s" назначена на @%s`, task.Title, user.UserName))
	}

	task.Assignee = user
	task.AssigneeID = userID

	tb.sendMessage(chatID, fmt.Sprintf(`Задача "%s" назначена на вас`, task.Title))

	// Уведомляем автора, если он не является исполнителем
	if task.AuthorID != userID {
		tb.sendMessage(int64(task.AuthorID), fmt.Sprintf(`Задача "%s" назначена на @%s`, task.Title, user.UserName))
	}
}

func (tb *TaskBot) handleUnassign(chatID int64, userID int, user *tgbotapi.User, taskID int) {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	task, exists := tb.tasks[taskID]
	if !exists {
		return
	}

	if task.Assignee == nil || task.AssigneeID != userID {
		tb.sendMessage(chatID, "Задача не на вас")
		return
	}

	task.Assignee = nil
	task.AssigneeID = 0

	tb.sendMessage(chatID, "Принято")

	// Уведомляем автора, если он не является исполнителем
	if task.AuthorID != userID {
		tb.sendMessage(int64(task.AuthorID), fmt.Sprintf(`Задача "%s" осталась без исполнителя`, task.Title))
	}
}
