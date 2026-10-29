package handler

import (
	"context"
	"log"
	"time"
	"log/slog"

	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
)

type Handler struct {
    client *botapi.Client
	userService *user.Service
	sessionStore *storage.SessionStore
	getCourseBaseURL string
	logger *slog.Logger
}

func New(client *botapi.Client, userService *user.Service, sessionStore *storage.SessionStore, getCourseBaseURL string, logger *slog.Logger) *Handler {
	return &Handler{
		client:      client,
		userService: userService,
		sessionStore: sessionStore,
		getCourseBaseURL: getCourseBaseURL,
		logger:       logger,
	}
}

func (h *Handler) Run(ctx context.Context) error {
	offset := 0

	for {
		if ctx.Err() != nil {
			return nil
		}

		updates, err := h.client.GetUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		for _, update := range updates {
			offset = update.UpdateID + 1
			h.handleUpdate(ctx, update)
		}
	}
}

func (h *Handler) handleUpdate(ctx context.Context, update botapi.Update) {
	uctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var err error

	switch {
	case update.CallbackQuery != nil:
		err = h.handleCallbackQuery(uctx, update.CallbackQuery)
	case update.Message != nil:
		err = h.handleMessage(uctx, update.Message)
	}

	if err != nil {
		log.Printf("telegram: update %d: %v", update.UpdateID, err)
	}
}