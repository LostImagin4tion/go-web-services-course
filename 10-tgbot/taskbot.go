package main

import (
	"context"
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"io/ioutil"
	"log"
	"net/http"
	handlers "stepikGoWebServices/tgbot"
)

func startTaskBot(ctx context.Context, httpListenAddr string) error {
	bot, err := tgbotapi.NewBotAPI("fillme")
	if err != nil {
		return err
	}

	taskBot := handlers.NewTaskBot(bot)

	webhookURL := "http://127.0.0.1:8081"
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(webhookURL))
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return
		}

		var update tgbotapi.Update
		if err := json.Unmarshal(body, &update); err != nil {
			return
		}

		taskBot.HandleMessage(update)
	})

	server := &http.Server{
		Addr: httpListenAddr,
	}

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	return server.ListenAndServe()
}

func main() {
	err := startTaskBot(context.Background(), ":8081")
	if err != nil {
		log.Fatalln(err)
	}
}
