package main

import (
	"log"
	"os"
	"time"
	"context"
	"strconv"
	"os/signal"
	"syscall"
	"encoding/base64"

	"github.com/joho/godotenv"

	"github.com/Woun1zoN/getcourse-schedulix/internal/app"
	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram"
	"github.com/Woun1zoN/getcourse-schedulix/internal/database"
	"github.com/Woun1zoN/getcourse-schedulix/internal/database/migrations"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
	"github.com/Woun1zoN/getcourse-schedulix/internal/crypto"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	_ = godotenv.Load()

	// Configuration
	databaseURL := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	key, err := base64.StdEncoding.DecodeString(os.Getenv("COOKIE_ENC_KEY"))
	if err != nil {
    	log.Fatal(err)
	}

	cryptor, err := crypto.NewCryptor(key)
	if err != nil {
    	log.Fatal(err)
	}

	const getCourseBaseURL = "https://shtpt.getcourse.ru"

	// Database migrations
	if err := migrations.Run(
    	databaseURL,
    	"file:///app/internal/database/migrations/sql",
	); err != nil {
    	log.Fatal(err)
	}

	// Database initialization
	db, err := database.InitDB(ctx, databaseURL)
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

	sessionStore := storage.NewSessionStore(redisClient)

	// GetCourse client initialization
	getCourseClient := getcourse.NewClient(getCourseBaseURL, "")

	// User service initialization
	userRepository := user.NewRepository(db.DB)
	userService := user.NewService(userRepository, getCourseClient, cryptor)

	// Telegram client initialization
	telegramClient := telegram.NewClient(telegramToken, telegramChatID, getCourseBaseURL, userService, sessionStore)

	go func() {
    	if err := telegramClient.Run(ctx); err != nil {
        	log.Fatal(err)
    	}
	}()

	// Run the application
	runner := app.NewRunner(
    	getCourseClient,
    	redisClient,
    	telegramClient,
		userService,
    	5*time.Minute,
	)

	runner.Run(ctx)
	log.Println("app stopped gracefully")
}