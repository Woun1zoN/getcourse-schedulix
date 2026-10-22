package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

func (c *Client) handleMessage(message *Message) error {
	ctx := context.Background()

	if message.Chat.Type != "private" {
		return nil
	}

	u, err := c.userService.GetOrCreate(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("get or create user: %w", err)
	}

	if message.ChatShared != nil {
		return c.handleChatShared(message)
	}

	if message.Text == "/start" {
		c.logger.Info(
			"telegram command",
			slog.String("command", "/start"),
			slog.Int64("telegram_id", message.From.ID),
		)

		return c.sendStartScreen(ctx, message.Chat.ID, u)
	}

	if message.Text == "/help" {
		c.logger.Info(
			"telegram command",
			slog.String("command", "/help"),
			slog.Int64("telegram_id", message.From.ID),
		)

		return c.sendHelp(message.Chat.ID)
	}

	if message.Text == "/menu" {
		c.logger.Info(
			"telegram command",
			slog.String("command", "/menu"),
			slog.Int64("telegram_id", message.From.ID),
		)

		return c.sendMenu(message.Chat.ID, u.TelegramID)
	}

	if message.Text == "💬 Личные сообщения" {
		if err := c.userService.ResetTargetChat(ctx, message.From.ID); err != nil {
			return fmt.Errorf("reset target chat: %w", err)
		}

		text, kb, err := c.connectedMenu(ctx, message.From.ID)
		if err != nil {
			return fmt.Errorf("connected menu: %w", err)
		}

		return c.SendMessageWithKeyboardMode(message.Chat.ID, text, ParseModeMarkdownV2, kb)
	}

	if message.Text == "➡ Пропустить" || message.Text == "➖ Без чата" {
		if err := c.userService.ClearTargetChat(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear target chat: %w", err)
		}

		text, kb, err := c.connectedMenu(ctx, message.From.ID)
		if err != nil {
			return fmt.Errorf("connected menu: %w", err)
		}

		return c.SendMessageWithKeyboardMode(message.Chat.ID, text, ParseModeMarkdownV2, kb)
	}

	awaiting, err := c.sessionStore.IsAwaitingCookie(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("check session state: %w", err)
	}

	if awaiting {
		return c.handleCookieInput(ctx, u, message)
	}

	return nil
}

func (c *Client) handleCookieInput(ctx context.Context, u *user.User, message *Message) error {
	cookie := strings.TrimSpace(message.Text)

	client := getcourse.NewClient(c.getCourseBaseURL, cookie)

	if err := client.ValidateCookie(); err != nil {
		return c.SendMessage(
			message.Chat.ID,
			"Не удалось подтвердить cookie. Убедитесь, что вы вошли в аккаунт и скопировали значение полностью.",
		)
	}

	if err := c.userService.ConnectGetCourse(ctx, u.TelegramID, cookie); err != nil {
		return fmt.Errorf("connect getcourse: %w", err)
	}

	if err := c.sessionStore.ClearState(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	streams, err := client.GetStreams()
	if err != nil {
		return c.SendMessage(message.Chat.ID, "Не удалось получить список тренингов, попробуйте позже.")
	}

	if u.GetCourseStreamID != nil {
		for _, s := range streams {
			if s.ID == *u.GetCourseStreamID {
				text, kb, err := c.connectedMenu(ctx, message.From.ID)
				if err != nil {
					return fmt.Errorf("connected menu: %w", err)
				}

				return c.SendMessageWithKeyboardMode(message.Chat.ID, text, ParseModeMarkdownV2, kb)
			}
		}
	}

	if err := c.SendMessage(message.Chat.ID, "_✅ GetCourse успешно подключён!_"); err != nil {
		return err
	}

	return c.SendStreamPicker(message.Chat.ID, streams)
}

func (c *Client) handlePickStream(cq *CallbackQuery) error {
	u, err := c.userService.GetOrCreate(context.Background(), cq.From.ID)
	if err != nil {
		return err
	}

	if u.GetCourseCookie == nil {
		return c.SendMessage(cq.From.ID, "Сначала подключите GetCourse.")
	}

	client := getcourse.NewClient(c.getCourseBaseURL, *u.GetCourseCookie)

	streams, err := client.GetStreams()
	if err != nil {
		return c.SendMessage(cq.From.ID, "Не удалось получить список тренингов, попробуйте позже.")
	}

	return c.SendStreamPicker(cq.From.ID, streams)
}

func (c *Client) handleChatShared(message *Message) error {
	ctx := context.Background()

	if message.ChatShared == nil {
		return nil
	}

	shared := message.ChatShared

	if shared.RequestID != chatRequestGroup && shared.RequestID != chatRequestChannel {
		return nil
	}

	if err := c.userService.SelectTargetChat(
		context.Background(),
		message.From.ID,
		shared.ChatID,
		shared.Title,
	); err != nil {
		return fmt.Errorf("select target chat: %w", err)
	}

	title := shared.Title
	if title == "" {
		title = fmt.Sprintf("%d", shared.ChatID)
	}

	text, kb, err := c.connectedMenu(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("connected menu: %w", err)
	}

	return c.SendMessageWithKeyboardMode(message.Chat.ID, text, ParseModeMarkdownV2, kb)
}