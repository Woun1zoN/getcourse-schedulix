package telegram

import (
	"context"
	"fmt"

	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

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

func (c *Client) sendStartScreen(chatID int64, u *user.User) error {
	return c.SendMessageWithKeyboardMode(chatID, startCommandText(u), ParseModeMarkdownV2, c.startKeyboard(u))
}

func (c *Client) sendHelp(chatID int64) error {
	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "⬅ Вернуться в меню", CallbackData: callbackBackToMenu}},
		},
	}
	return c.SendMessageWithKeyboardMode(chatID, helpText, ParseModeMarkdownV2, kb)
}

func (c *Client) sendMenu(chatID, telegramID int64) error {
	ctx := context.Background()

	u, err := c.userService.GetOrCreate(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if err := c.sessionStore.ClearState(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	if u.GetCourseCookie == nil {
		kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
		}}
		return c.SendMessageWithKeyboardMode(chatID, welcomeText, ParseModeMarkdownV2, kb)
	}

	text, kb := c.renderConnectedMenu(u)
	return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, kb)
}

func (c *Client) startKeyboard(u *user.User) *InlineKeyboardMarkup {
	menuBtn := InlineKeyboardButton{Text: "⬅ Вернуться в меню", CallbackData: callbackBackToMenu}

	var setupBtn *InlineKeyboardButton

	switch {
	case u.GetCourseCookie == nil:
		setupBtn = &InlineKeyboardButton{Text: "🔗 Подключить", CallbackData: callbackConnectGetCourse}
	case u.GetCourseStreamID == nil:
		setupBtn = &InlineKeyboardButton{Text: "📚 Выбрать тренинг", CallbackData: callbackPickStream}
	case u.TargetChatID == nil:
		setupBtn = &InlineKeyboardButton{Text: "💬 Выбрать чат", CallbackData: callbackPickChat}
	}

	if setupBtn == nil {
		return &InlineKeyboardMarkup{
			InlineKeyboard: [][]InlineKeyboardButton{{menuBtn}},
		}
	}

	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{{menuBtn, *setupBtn}},
	}
}
