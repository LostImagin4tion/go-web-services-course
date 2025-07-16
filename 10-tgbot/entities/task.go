package entities

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

type Task struct {
	ID         int
	Title      string
	Author     *tgbotapi.User
	Assignee   *tgbotapi.User
	AuthorID   int
	AssigneeID int
}
