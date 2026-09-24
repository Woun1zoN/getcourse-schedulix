package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

type Runner struct {
	baseClient     *getcourse.Client
	redisClient    *redis.Client
	telegramClient *telegram.Client
	userService    *user.Service
	interval       time.Duration
}

func NewRunner(baseClient *getcourse.Client, redisClient *redis.Client, telegramClient *telegram.Client, userService *user.Service, interval time.Duration) *Runner {
	return &Runner{
		baseClient:     baseClient,
		redisClient:    redisClient,
		telegramClient: telegramClient,
		userService:    userService,
		interval:       interval,
	}
}

func (r *Runner) tick(ctx context.Context) {
	users, err := r.userService.GetActive(ctx)
	if err != nil {
		log.Println("get active users:", err)
		return
	}

	for _, u := range users {
		client := r.baseClient.WithCookies(*u.GetCourseCookie)

		state := storage.NewRedisState(r.redisClient, fmt.Sprintf("docs:hashes:%d", u.ID))

		if err := InitApp(u.TelegramID, u.TelegramID, client, state, r.telegramClient); err != nil {
			log.Printf("user %d: %v", u.ID, err)
		}
	}
}

func (r *Runner) Run(ctx context.Context) {
	r.tick(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.tick(ctx)
		case <-ctx.Done():
			log.Println("shutdown signal received, stopping scheduler...")
			return
		}
	}
}