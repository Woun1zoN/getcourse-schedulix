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
const callbackPickChat = "pick_chat"

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
				"➡ _GetCourse подключён, но тренинг ещё не выбран_.\n\nБез тренинга материалы приходить не будут.", kb)

		default:
			text := "_✅ GetCourse подключён_\n\nНовые материалы будут приходить автоматически\\."

			if s := c.currentStream(u); s != nil {
				text = "_✅ GetCourse подключён_\n\n• Выбранный тренинг:\n>📚 *" +
					escapeMarkdownV2(s.Name) + "*" + "⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀"

				if info := s.Info(); info != "" {
					text += "\n>" + escapeMarkdownV2(info)
				}

				text += "\n\nНовые материалы будут приходить автоматически\\."
			}

			kb := &InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{{Text: "🔀 Сменить тренинг", CallbackData: callbackPickStream}},
					{{Text: "📨 Сменить чат", CallbackData: callbackPickChat}},
					{{Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse}},
				},
			}
			return c.SendMessageWithKeyboardMode(message.Chat.ID, text, ParseModeMarkdownV2, kb)
		}
	}

	if message.Text == "👤 Личные сообщения" {
		if err := c.userService.ResetTargetChat(ctx, message.From.ID); err != nil {
			return fmt.Errorf("reset target chat: %w", err)
		}
		return c.SendMessageAndRemoveKeyboard(message.Chat.ID, "✅ Материалы будут приходить в личные сообщения\\.")
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

	switch {
	case strings.HasPrefix(cq.Data, callbackStreamPrefix):
		return c.handleStreamSelected(cq)

	case cq.Data == callbackPickStream:
		return c.handlePickStream(cq)

	case cq.Data == callbackPickChat:
		return c.SendChatPicker(cq.Message.Chat.ID)

	case cq.Data == callbackConnectGetCourse:
		if err := c.sessionStore.SetAwaitingCookie(context.Background(), cq.From.ID); err != nil {
			return err
		}
		return c.SendMessageMode(cq.Message.Chat.ID, connectGetCourseMessage, ParseModeMarkdownV2)
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
				return c.SendMessage(message.Chat.ID, "_✅ GetCourse переподключён. Выбранный тренинг сохранён._")
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
		return c.SendMessage(cq.Message.Chat.ID, "Сначала подключите GetCourse.")
	}

	client := getcourse.NewClient(c.getCourseBaseURL, *u.GetCourseCookie)

	streams, err := client.GetStreams()
	if err != nil {
		return c.SendMessage(cq.Message.Chat.ID, "Не удалось получить список тренингов, попробуйте позже.")
	}

	return c.SendStreamPicker(cq.Message.Chat.ID, streams)
}

func (c *Client) handleChatShared(message *Message) error {
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

    if err := c.SendMessageAndRemoveKeyboard(message.Chat.ID, "✅ Чат выбран\\."); err != nil {
        return err
    }

    title := shared.Title
    if title == "" {
        title = fmt.Sprintf("%d", shared.ChatID)
    }

    return c.SendMessage(
        message.Chat.ID,
        fmt.Sprintf(
            "✅ Чат выбран.\n\nТеперь новые материалы будут отправляться в «%s».",
            title,
        ),
    )
}

func (c *Client) SendMessageAndRemoveKeyboard(chatID int64, text string) error {
	return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, &ReplyKeyboardRemove{
			RemoveKeyboard: true,
		},
	)
}