package main

import (
	"log"
	"os"
	"time"
	"context"

	"github.com/joho/godotenv"

	"github.com/Woun1zoN/schedule-bot/internal/app"
	"github.com/Woun1zoN/schedule-bot/internal/getcourse"
	"github.com/Woun1zoN/schedule-bot/internal/storage"
	"github.com/Woun1zoN/schedule-bot/internal/telegram"
	"github.com/Woun1zoN/schedule-bot/internal/database"
	"github.com/Woun1zoN/schedule-bot/internal/database/migrations"
	"github.com/Woun1zoN/schedule-bot/internal/user"
)

func main() {
	_ = godotenv.Load()

	cookies := os.Getenv("GETCOURSE_COOKIES")

	if err := migrations.Run(
    	os.Getenv("DATABASE_URL"),
    	"file:///app/internal/database/migrations/sql",
	); err != nil {
    	log.Fatal(err)
	}

	db, err := database.InitDB(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	userRepository := user.NewRepository(db.DB)
	userService := user.NewService(userRepository)

	client := getcourse.NewClient("https://shtpt.getcourse.ru", cookies)

	state, err := storage.Load("state.json")
	if err != nil {
		log.Fatal(err)
	}

	telegramClient := telegram.NewClient(os.Getenv("TELEGRAM_BOT_TOKEN"), os.Getenv("TELEGRAM_CHAT_ID"), os.Getenv("TELEGRAM_LOG_CHAT_ID"), userService)

	if err := telegramClient.Run(); err != nil {
    	log.Fatal(err)
	}

	if err := app.InitApp(client, state, telegramClient); err != nil {
		log.Println("run:", err)
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := app.InitApp(client, state, telegramClient); err != nil {
			log.Println("run:", err)
		}
	}
}