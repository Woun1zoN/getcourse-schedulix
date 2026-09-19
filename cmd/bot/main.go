package main

import (
	"log"
	"os"
	"time"
	"context"
	"strconv"

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

	// Configuration
	databaseURL := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")
	telegramLogChatID := os.Getenv("TELEGRAM_LOG_CHAT_ID")
	getcourseCookies := os.Getenv("GETCOURSE_COOKIES")

	// Database migrations
	if err := migrations.Run(
    	databaseURL,
    	"file:///app/internal/database/migrations/sql",
	); err != nil {
    	log.Fatal(err)
	}

	// Database initialization
	db, err := database.InitDB(context.Background(), databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()

	// Redis initialization
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
    	log.Fatalf("parse REDIS_DB: %v", err)
	}

	redisClient, err := storage.NewRedisClient(
    	redisAddr,
		redisPassword,
    	redisDB,
	)
	if err != nil {
    	log.Fatal(err)
	}
	defer redisClient.Close()

	// User service initialization
	userRepository := user.NewRepository(db.DB)
	userService := user.NewService(userRepository)

	// GetCourse client initialization
	getCourseClient := getcourse.NewClient("https://shtpt.getcourse.ru", getcourseCookies)

	state := storage.NewRedisState(redisClient, "docs:hashes")

	// Telegram client initialization
	telegramClient := telegram.NewClient(telegramToken, telegramChatID, telegramLogChatID, userService)

	go func() {
    	if err := telegramClient.Run(); err != nil {
        	log.Fatal(err)
    	}
	}()

	// Run the application
	if err := app.InitApp(getCourseClient, state, telegramClient); err != nil {
		log.Println("run:", err)
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := app.InitApp(getCourseClient, state, telegramClient); err != nil {
			log.Println("run:", err)
		}
	}
}