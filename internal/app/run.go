package app

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

type Runner struct {
	baseClient     *getcourse.Client
	redisClient    *redis.Client
	telegramClient *botapi.Client
	userService    *user.Service
	sessionStore   *storage.SessionStore
	interval       time.Duration

	notify         func(telegramID int64) error

	notifyMaterial func(
        documentName, documentURL, lessonID, lessonURL string,
        fileSize float64,
        checked, newDocuments, logChatID int64,
    ) error
}

func NewRunner(
	baseClient *getcourse.Client,
	redisClient *redis.Client,
	telegramClient *botapi.Client,
	userService *user.Service,
	sessionStore *storage.SessionStore,
	interval time.Duration,
	notify func(telegramID int64) error,
	notifyMaterial func(
		documentName, documentURL, lessonID, lessonURL string,
		fileSize float64,
		checked, newDocuments, logChatID int64,
	) error,
) *Runner {
	return &Runner{
		baseClient:     baseClient,
		redisClient:    redisClient,
		telegramClient: telegramClient,
		userService:    userService,
		sessionStore:   sessionStore,
		interval:       interval,
		notify:         notify,
		notifyMaterial: notifyMaterial,
	}
}

func (r *Runner) tick(ctx context.Context) {
	users, err := r.userService.GetActive(ctx)
	if err != nil {
		log.Println("get active users:", err)
		return
	}

	for _, u := range users {
		if u.GetCourseStreamID == nil || u.TargetChatID == nil {
			continue
		}

		client := r.baseClient.WithCookies(*u.GetCourseCookie)
		state := storage.NewRedisState(r.redisClient, storage.HashesKey(u.ID))

		err := InitApp(u.TelegramID, *u.TargetChatID, *u.GetCourseStreamID, client, state, r.telegramClient, r.notifyMaterial)
		if err == nil {
			continue
		}

		if errors.Is(err, getcourse.ErrCookieExpired) {
			r.handleExpiredCookie(ctx, u)
			continue
		}

		log.Printf("user %d: %v", u.ID, err)
	}
}

func (r *Runner) handleExpiredCookie(ctx context.Context, u user.User) {
	first, err := r.sessionStore.MarkExpiredNotified(ctx, u.TelegramID)
	if err != nil {
		log.Printf("user %d: mark expired notified: %v", u.ID, err)
		return
	}
	if !first {
		return
	}

	if err := r.notify(u.TelegramID); err != nil {
		log.Printf("user %d: send reauth prompt: %v", u.ID, err)
		_ = r.sessionStore.ClearExpiredNotified(ctx, u.TelegramID)
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
