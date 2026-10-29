package handler

import (
	"context"
	"fmt"

	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
)

func (h *Handler) connectedMenu(ctx context.Context, telegramID int64) (string, *botapi.InlineKeyboardMarkup, error) {
	u, err := h.userService.GetOrCreate(ctx, telegramID)
	if err != nil {
		return "", nil, fmt.Errorf("get user: %w", err)
	}

	text, kb := h.renderConnectedMenu(u)

	return text, kb, nil
}

func (h *Handler) renderConnectedMenu(u *user.User) (string, *botapi.InlineKeyboardMarkup) {
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

	if s := h.currentStream(u); s != nil {
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

	streamBtnText := "🔀 Сменить тренинг"
	if u.GetCourseStreamID == nil {
		streamBtnText = "📚 Выбрать тренинг"
	}

	chatBtnText := "💬 Сменить чат"
	if u.TargetChatID == nil {
		chatBtnText = "💬 Выбрать чат"
	}

	kb := &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{
				{Text: streamBtnText, CallbackData: callbackPickStream},
				{Text: chatBtnText, CallbackData: callbackPickChat},
			},
			{
				{Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse},
			},
			{
				{Text: "🗑 Удалить мои данные", CallbackData: callbackDelete},
			},
		},
	}

	return text, kb
}

func (h *Handler) sendStartScreen(chatID int64, u *user.User) error {
	if u.GetCourseCookie == nil {
		kb := &botapi.InlineKeyboardMarkup{InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
		}}
		return h.client.SendMessageWithKeyboardMode(chatID, welcomeText, botapi.ParseModeMarkdownV2, kb)
	}

	return h.client.SendMessageWithKeyboardMode(chatID, startCommandText(u), botapi.ParseModeMarkdownV2, h.startKeyboard(u))
}

func (h *Handler) sendHelp(chatID int64) error {
	kb := &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{{Text: "⬅ Вернуться в меню", CallbackData: callbackBackToMenu}},
		},
	}
	return h.client.SendMessageWithKeyboardMode(chatID, helpText, botapi.ParseModeMarkdownV2, kb)
}

func (h *Handler) sendMenu(ctx context.Context, chatID, telegramID int64) error {
	u, err := h.userService.GetOrCreate(ctx, telegramID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if err := h.sessionStore.ClearState(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	if u.GetCourseCookie == nil {
		kb := &botapi.InlineKeyboardMarkup{InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
		}}
		return h.client.SendMessageWithKeyboardMode(chatID, welcomeText, botapi.ParseModeMarkdownV2, kb)
	}

	text, kb := h.renderConnectedMenu(u)
	return h.client.SendMessageWithKeyboardMode(chatID, text, botapi.ParseModeMarkdownV2, kb)
}

func (h *Handler) startKeyboard(u *user.User) *botapi.InlineKeyboardMarkup {
	menuBtn := botapi.InlineKeyboardButton{Text: "⬅ Вернуться в меню", CallbackData: callbackBackToMenu}

	var setupBtn *botapi.InlineKeyboardButton

	switch {
	case u.GetCourseStreamID == nil:
		setupBtn = &botapi.InlineKeyboardButton{Text: "📚 Выбрать тренинг", CallbackData: callbackPickStream}
	case u.TargetChatID == nil:
		setupBtn = &botapi.InlineKeyboardButton{Text: "💬 Выбрать чат", CallbackData: callbackPickChat}
	}

	rows := [][]botapi.InlineKeyboardButton{{menuBtn}}
	if setupBtn != nil {
		rows = [][]botapi.InlineKeyboardButton{{menuBtn, *setupBtn}}
	}

	if u.GetCourseCookie != nil {
		rows = append(rows, []botapi.InlineKeyboardButton{
			{Text: "🗑 Удалить мои данные", CallbackData: callbackDelete},
		})
	}

	return &botapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}
