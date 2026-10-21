package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

const callbackConnectGetCourse = "connect_getcourse"
const callbackPickStream = "pick_stream"
const callbackPickChat = "pick_chat"
const callbackCancel = "cancel"
const callbackBackToMenu = "back_to_menu"

func (c *Client) connectedMenu(ctx context.Context, telegramID int64) (string, *InlineKeyboardMarkup, error) {
	u, err := c.userService.GetOrCreate(ctx, telegramID)
	if err != nil {
		return "", nil, fmt.Errorf("get user: %w", err)
	}

	text, kb := c.renderConnectedMenu(u)

	return text, kb, nil
}

func (c *Client) renderConnectedMenu(u *user.User) (string, *InlineKeyboardMarkup) {
	var text string

	switch {
	case u.GetCourseStreamID == nil && u.TargetChatID == nil:
		text += "\n\n❗ Пока тренинг и чат не выбраны, материалы приходить не будут\\."
	case u.GetCourseStreamID == nil:
		text += "\n\n❗ Пока тренинг не выбран, материалы приходить не будут\\."
	case u.TargetChatID == nil:
		text += "\n\n❗ Пока чат не выбран, материалы приходить не будут\\."
	default:
		text += "\n\n✅ GetCourse подключён, новые материалы будут приходить автоматически\\."
	}

	if s := c.currentStream(u); s != nil {
		text += "\n\n• Выбранный тренинг:\n📚 *" + escapeMarkdownV2(s.Name) + "*"
		if info := s.Info(); info != "" {
			text += "\n" + escapeMarkdownV2(info)
		}
	} else if u.GetCourseStreamID == nil {
		text += "\n\n• Выбранный тренинг:\n➖ Не выбран"
	}

	chatName := "*💬 Личные сообщения*"
	switch {
	case u.TargetChatID == nil:
		chatName = "➖ Не выбран"
	case *u.TargetChatID != u.TelegramID:
		if u.TargetChatTitle != nil && *u.TargetChatTitle != "" {
			chatName = "💬 *" + escapeMarkdownV2(*u.TargetChatTitle) + "*"
		} else {
			chatName = fmt.Sprintf("💬 Чат %d", *u.TargetChatID)
		}
	}


	text += "\n\n• Выбранный чат:\n" + chatName

	text += "\n\nℹ️ /help — Список доступных команд"

	kb := &InlineKeyboardMarkup{
    	InlineKeyboard: [][]InlineKeyboardButton{
        	{
        	    {Text: "🔀 Сменить тренинг", CallbackData: callbackPickStream},
        	    {Text: "💬 Сменить чат", CallbackData: callbackPickChat},
        	},
        	{
        	    {Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse},
        	},
    	},
	}

	return text, kb
}

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

func (c *Client) handleCallbackQuery(cq *CallbackQuery) error {
	_ = c.answerCallbackQuery(cq.ID)

	switch {
	case strings.HasPrefix(cq.Data, callbackStreamPrefix):
		return c.handleStreamSelected(cq)

	case cq.Data == callbackPickStream:
		return c.handlePickStream(cq)

	case cq.Data == callbackPickChat:
		return c.SendChatPicker(cq.Message.Chat.ID, ChatPickerChange)

	case cq.Data == callbackConnectGetCourse:
		if err := c.sessionStore.SetAwaitingCookie(context.Background(), cq.From.ID); err != nil {
			return err
		}
		return c.SendMessageWithKeyboardMode(cq.Message.Chat.ID, connectGetCourseMessage, ParseModeMarkdownV2, cancelKeyboard())

	case cq.Data == callbackCancel:
		return c.handleCancel(cq)

	case cq.Data == callbackSkipStream:
		return c.handleSkipStream(cq)

	case cq.Data == callbackBackToMenu:
		ctx := context.Background()

		u, err := c.userService.GetOrCreate(ctx, cq.From.ID)
		if err != nil {
			return err
		}

		return c.sendStartScreen(ctx, cq.Message.Chat.ID, u)
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

func (c *Client) handleCancel(cq *CallbackQuery) error {
	ctx := context.Background()

	if err := c.sessionStore.ClearState(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	u, err := c.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	chatID := cq.Message.Chat.ID

	if u.GetCourseCookie != nil {
		text, kb := c.renderConnectedMenu(u)
		return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, kb)
	}

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
		},
	}

	return c.SendMessageWithKeyboardMode(chatID, welcomeText, ParseModeMarkdownV2, kb)
}

func (c *Client) handleSkipStream(cq *CallbackQuery) error {
	ctx := context.Background()

	if err := c.userService.ClearStream(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear stream: %w", err)
	}

	u, err := c.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	if u.TargetChatID == nil {
		return c.SendChatPicker(cq.From.ID, ChatPickerInitial)
	}

	text, kb := c.renderConnectedMenu(u)
	return c.SendMessageWithKeyboardMode(cq.From.ID, text, ParseModeMarkdownV2, kb)
}

func cancelKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: callbackCancel, Style: "danger"}},
		},
	}
}

func (c *Client) SendMessageAndRemoveKeyboard(chatID int64, text string) error {
	return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, &ReplyKeyboardRemove{
		RemoveKeyboard: true,
	})
}

func (c *Client) sendStartScreen(ctx context.Context, chatID int64, u *user.User) error {
	if err := c.sessionStore.ClearState(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	switch {
	case u.GetCourseCookie == nil:
		kb := &InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{
				{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
			},
		}
		return c.SendMessageWithKeyboardMode(chatID, welcomeText, ParseModeMarkdownV2, kb)

	case u.GetCourseStreamID == nil:
		kb := &InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{
				{{Text: "🔗 Выбрать тренинг", CallbackData: callbackPickStream}},
			},
		}
		return c.SendMessageWithKeyboard(chatID,
			"➡ _GetCourse подключён, но тренинг ещё не выбран_.\n\nБез тренинга материалы приходить не будут.", kb)

	case u.TargetChatID == nil:
		return c.SendChatPicker(chatID, ChatPickerInitial)

	default:
		text, kb, err := c.connectedMenu(ctx, u.TelegramID)
		if err != nil {
			return fmt.Errorf("connected menu: %w", err)
		}
		return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, kb)
	}
}

func (c *Client) sendHelp(chatID int64) error {
	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "⬅ Вернуться в меню", CallbackData: callbackBackToMenu}},
		},
	}
	return c.SendMessageWithKeyboardMode(chatID, helpText, ParseModeMarkdownV2, kb)
}