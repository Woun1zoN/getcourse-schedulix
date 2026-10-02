package telegram

import (
	"context"
	"fmt"
    "strings"
    "log/slog"

    "github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
    "github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

const callbackConnectGetCourse = "connect_getcourse"
const callbackPickStream = "pick_stream"

func (c *Client) handleMessage(message *Message) error {
	ctx := context.Background()

	if message.Chat.Type != "private" {
		return nil
	}

	u, err := c.userService.GetOrCreate(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("get or create user: %w", err)
	}

	if message.Text == "/start" {
		if err := c.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		c.logger.Info(
			"telegram command",
			slog.String("command", "/start"),
			slog.Int64("telegram_id", message.From.ID),
		)

		switch {
		case u.GetCourseCookie == nil:
			kb := &InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
				},
			}
			return c.SendMessageWithKeyboardMode(message.Chat.ID, welcomeText, ParseModeMarkdownV2, kb)

		case u.GetCourseStreamID == nil:
			kb := &InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{{Text: "🔗 Выбрать тренинг", CallbackData: callbackPickStream}},
				},
			}

			return c.SendMessageWithKeyboard(message.Chat.ID,
				"С возвращением! GetCourse подключён, но тренинг ещё не выбран.\n\nБез тренинга расписания приходить не будут.", kb)

		default:
			kb := &InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{{Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse}},
					{{Text: "🔀 Сменить тренинг", CallbackData: callbackPickStream}},
				},
			}

			return c.SendMessageWithKeyboard(message.Chat.ID,
				"С возвращением! GetCourse уже подключён.", kb)
		}
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

func (c *Client) handleCallbackQuery(cq *CallbackQuery) error {
    _ = c.answerCallbackQuery(cq.ID)

	if strings.HasPrefix(cq.Data, callbackStreamPrefix) {
		return c.handleStreamSelected(cq)
	}

	if cq.Data == callbackPickStream {
		return c.handlePickStream(cq)
	}

    if cq.Data != callbackConnectGetCourse {
        return nil
    }

    if err := c.sessionStore.SetAwaitingCookie(context.Background(), cq.From.ID); err != nil {
        return err
    }

    return c.SendMessageMode(cq.Message.Chat.ID, connectGetCourseMessage, ParseModeMarkdownV2)
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

	return c.SendStreamPicker(message.Chat.ID, streams)
}

func (c *Client) handlePickStream(cq *CallbackQuery) error {
	u, err := c.userService.GetOrCreate(context.Background(), cq.From.ID)
	if err != nil {
		return err
	}

	if u.GetCourseCookie == nil {
		return c.SendMessage(cq.Message.Chat.ID, "Сначала подключите GetCourse.")
	}

	client := getcourse.NewClient(c.getCourseBaseURL, *u.GetCourseCookie)

	streams, err := client.GetStreams()
	if err != nil {
		return c.SendMessage(cq.Message.Chat.ID, "Не удалось получить список тренингов, попробуйте позже.")
	}

	return c.SendStreamPicker(cq.Message.Chat.ID, streams)
}