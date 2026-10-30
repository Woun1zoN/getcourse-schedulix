package handler

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

func (h *Handler) handleMessage(ctx context.Context, message *botapi.Message) error {
	if message.Chat.Type != "private" {
		return nil
	}

	u, err := h.userService.GetOrCreate(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("get or create user: %w", err)
	}

	if message.ChatShared != nil {
		return h.handleChatShared(ctx, message)
	}

	if message.Text == "/start" {
		if err := h.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		h.logger.Info(
			"telegram command",
			slog.String("command", "/start"),
			slog.Int64("telegram_id", message.From.ID),
		)

		return h.sendStartScreen(message.Chat.ID, u)
	}

	if message.Text == "/help" {
		if err := h.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		h.logger.Info(
			"telegram command",
			slog.String("command", "/help"),
			slog.Int64("telegram_id", message.From.ID),
		)

		return h.sendHelp(message.Chat.ID)
	}

	if message.Text == "/menu" {
		if err := h.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		h.logger.Info(
			"telegram command",
			slog.String("command", "/menu"),
			slog.Int64("telegram_id", message.From.ID),
		)

		return h.sendMenu(ctx, message.Chat.ID, u.TelegramID)
	}

	if message.Text == "/delete" {
		if err := h.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		return h.sendDeleteConfirmation(message.Chat.ID)
	}

	if message.Text == "💬 Личные сообщения" {
		if err := h.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		if err := h.userService.ResetTargetChat(ctx, message.From.ID); err != nil {
			return fmt.Errorf("reset target chat: %w", err)
		}

		text, kb, err := h.connectedMenu(ctx, message.From.ID)
		if err != nil {
			return fmt.Errorf("connected menu: %w", err)
		}

		return h.client.SendMessageWithKeyboardMode(message.Chat.ID, text, botapi.ParseModeMarkdownV2, kb)
	}

	if message.Text == "➡ Пропустить" || message.Text == "➖ Без чата" {
		if err := h.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		if err := h.userService.ClearTargetChat(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear target chat: %w", err)
		}

		text, kb, err := h.connectedMenu(ctx, message.From.ID)
		if err != nil {
			return fmt.Errorf("connected menu: %w", err)
		}

		return h.client.SendMessageWithKeyboardMode(message.Chat.ID, text, botapi.ParseModeMarkdownV2, kb)
	}

	awaiting, err := h.sessionStore.IsAwaitingCookie(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("check session state: %w", err)
	}

	if awaiting {
		return h.handleCookieInput(ctx, u, message)
	}

	return nil
}

func (h *Handler) handleCookieInput(ctx context.Context, u *user.User, message *botapi.Message) error {
	cookie := strings.TrimSpace(message.Text)

	h.cleanupMessage(message.Chat.ID, message.MessageID)

	client := getcourse.NewClient(h.getCourseBaseURL, cookie)

	if err := client.ValidateCookie(ctx); err != nil {
		return h.client.SendMessage(
			message.Chat.ID,
			"Не удалось подтвердить cookie. Убедитесь, что вы вошли в аккаунт и скопировали значение полностью.",
		)
	}

	if err := h.userService.ConnectGetCourse(ctx, u.TelegramID, cookie); err != nil {
		return fmt.Errorf("connect getcourse: %w", err)
	}

	if err := h.sessionStore.ClearExpiredNotified(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear expired flag: %w", err)
	}

	if err := h.sessionStore.ClearState(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	if promptID, err := h.sessionStore.PopCookiePromptMessage(ctx, u.TelegramID); err != nil {
		log.Printf("telegram: pop cookie prompt message: %v", err)
	} else if promptID != 0 {
		h.cleanupMessage(message.Chat.ID, promptID)
	}

	streams, err := client.GetStreams(ctx)
	if err != nil {
		return h.client.SendMessage(message.Chat.ID, "Не удалось получить список тренингов, попробуйте позже.")
	}

	if u.GetCourseStreamID != nil {
		for _, s := range streams {
			if s.ID == *u.GetCourseStreamID {
				text, kb, err := h.connectedMenu(ctx, message.From.ID)
				if err != nil {
					return fmt.Errorf("connected menu: %w", err)
				}

				return h.client.SendMessageWithKeyboardMode(message.Chat.ID, text, botapi.ParseModeMarkdownV2, kb)
			}
		}
	}

	if err := h.client.SendMessage(message.Chat.ID, "_✅ GetCourse успешно подключён!_"); err != nil {
		return err
	}

	return h.sendStreamPicker(message.Chat.ID, streams)
}

func (h *Handler) handlePickStream(ctx context.Context, cq *botapi.CallbackQuery) error {
	u, err := h.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	if u.GetCourseCookie == nil {
		return h.client.SendMessage(cq.From.ID, "Сначала подключите GetCourse.")
	}

	client := getcourse.NewClient(h.getCourseBaseURL, *u.GetCourseCookie)

	streams, err := client.GetStreams(ctx)
	if err != nil {
		return h.client.SendMessage(cq.From.ID, "Не удалось получить список тренингов, попробуйте позже.")
	}

	return h.sendStreamPicker(cq.From.ID, streams)
}

func (h *Handler) handleChatShared(ctx context.Context, message *botapi.Message) error {
	if message.ChatShared == nil {
		return nil
	}

	shared := message.ChatShared

	if shared.RequestID != chatRequestGroup && shared.RequestID != chatRequestChannel {
		return nil
	}

	if err := h.userService.SelectTargetChat(
		ctx,
		message.From.ID,
		shared.ChatID,
		shared.Title,
	); err != nil {
		return fmt.Errorf("select target chat: %w", err)
	}

	text, kb, err := h.connectedMenu(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("connected menu: %w", err)
	}

	return h.client.SendMessageWithKeyboardMode(message.Chat.ID, text, botapi.ParseModeMarkdownV2, kb)
}
