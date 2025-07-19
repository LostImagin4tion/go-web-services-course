package handlers

import (
	"fmt"
	"strings"
)

func (tb *TaskBot) handleTasks(chatID int64, userID int) {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	if len(tb.tasks) == 0 {
		tb.sendMessage(chatID, "Нет задач")
		return
	}

	var result []string
	for _, task := range tb.tasks {
		line := fmt.Sprintf("%d. %s by @%s", task.ID, task.Title, task.Author.UserName)
		if task.Assignee != nil {
			if task.AssigneeID == userID {
				line += "\nassignee: я"
				line += fmt.Sprintf("\n/unassign_%d /resolve_%d", task.ID, task.ID)
			} else {
				line += fmt.Sprintf("\nassignee: @%s", task.Assignee.UserName)
			}
		} else {
			line += fmt.Sprintf("\n/assign_%d", task.ID)
		}
		result = append(result, line)
	}

	tb.sendMessage(chatID, strings.Join(result, "\n\n"))
}

func (tb *TaskBot) handleMy(chatID int64, userID int) {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	var result []string
	for _, task := range tb.tasks {
		if task.Assignee != nil && task.AssigneeID == userID {
			line := fmt.Sprintf("%d. %s by @%s", task.ID, task.Title, task.Author.UserName)
			line += fmt.Sprintf("\n/unassign_%d /resolve_%d", task.ID, task.ID)
			result = append(result, line)
		}
	}

	if len(result) == 0 {
		tb.sendMessage(chatID, "Нет задач")
		return
	}

	tb.sendMessage(chatID, strings.Join(result, "\n\n"))
}

func (tb *TaskBot) handleOwner(chatID int64, userID int) {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()

	var result []string
	for _, task := range tb.tasks {
		if task.AuthorID == userID {
			line := fmt.Sprintf("%d. %s by @%s", task.ID, task.Title, task.Author.UserName)
			if task.Assignee == nil {
				line += fmt.Sprintf("\n/assign_%d", task.ID)
			}
			result = append(result, line)
		}
	}

	if len(result) == 0 {
		tb.sendMessage(chatID, "Нет задач")
		return
	}

	tb.sendMessage(chatID, strings.Join(result, "\n\n"))
}
