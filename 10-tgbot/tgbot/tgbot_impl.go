package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"stepikGoWebServices/entities"
	"sync"
)

type TaskBot struct {
	bot    *tgbotapi.BotAPI
	tasks  map[int]*entities.Task
	mutex  sync.RWMutex
	nextID int
}

func NewTaskBot(bot *tgbotapi.BotAPI) *TaskBot {
	return &TaskBot{
		bot:    bot,
		tasks:  make(map[int]*entities.Task),
		mutex:  sync.RWMutex{},
		nextID: 1,
	}
}
