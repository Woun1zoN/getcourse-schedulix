// internal/app/run.go
package app

import (
	"context"
	"log"
	"time"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram"
)

type Runner struct {
	client         *getcourse.Client
	state          *storage.RedisState
	telegramClient *telegram.Client
	interval       time.Duration
}

func NewRunner(client *getcourse.Client, state *storage.RedisState, telegramClient *telegram.Client, interval time.Duration) *Runner {
	return &Runner{
		client:         client,
		state:          state,
		telegramClient: telegramClient,
		interval:       interval,
	}
}

func (r *Runner) Run(ctx context.Context) {
	if err := InitApp(r.client, r.state, r.telegramClient); err != nil {
		log.Println("run:", err)
	}

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := InitApp(r.client, r.state, r.telegramClient); err != nil {
				log.Println("run:", err)
			}
		case <-ctx.Done():
			log.Println("shutdown signal received, stopping scheduler...")
			return
		}
	}
}