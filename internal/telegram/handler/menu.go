package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

func (h *Handler) connectedMenu(ctx context.Context, telegramID int64) (string, *botapi.InlineKeyboardMarkup, error) {
	u, err := h.userService.GetOrCreate(ctx, telegramID)
	if err != nil {
		return "", nil, fmt.Errorf("get user: %w", err)
	}

	text, kb := h.renderConnectedMenu(ctx, u)

	return text, kb, nil
}

func (h *Handler) renderConnectedMenu(ctx context.Context, u *user.User) (string, *botapi.InlineKeyboardMarkup) {
	var text string

	noStreams := len(u.GetCourseStreamIDs) == 0

	switch {
	case noStreams && u.TargetChatID == nil:
		text += "\n\n❗ Пока тренинг и чат не выбраны, материалы приходить не будут\\."
	case noStreams:
		text += "\n\n❗ Пока тренинг не выбран, материалы приходить не будут\\."
	case u.TargetChatID == nil:
		text += "\n\n❗ Пока чат не выбран, материалы приходить не будут\\."
	default:
		text += "\n\n✅ GetCourse подключён, новые материалы будут приходить автоматически\\."
	}

	if noStreams {
		text += "\n\n• Выбранный тренинг:\n➖ Не выбран"
	} else {
		streams, err := h.currentStreams(ctx, u)

		switch {
		case err != nil:
			log.Printf("telegram: current streams for user %d: %v", u.TelegramID, err)
			text += "\n\n• Выбранные тренинги:\n⚠️ Не удалось загрузить названия"

		case len(streams) == 0:
			text += "\n\n• Выбранные тренинги:\n⚠️ Выбранные тренинги больше не найдены\\. Выберите другие\\."

		default:
			if len(streams) == 1 {
				text += "\n\n• Выбранный тренинг:"
			} else {
				text += "\n\n• Выбранные тренинги:"
			}
			for i, s := range streams {
				prefix := ""
				if len(streams) > 1 {
					prefix = fmt.Sprintf("%d\\. ", i+1)
				}

				text += fmt.Sprintf("\n%s📚 *%s*", prefix, escapeMarkdownV2(s.Name))
				if info := s.Info(); info != "" {
					text += "\n" + escapeMarkdownV2(info)
				}
			}
		}
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
	if noStreams {
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

	text, kb := h.renderConnectedMenu(ctx, u)
	return h.client.SendMessageWithKeyboardMode(chatID, text, botapi.ParseModeMarkdownV2, kb)
}

func (h *Handler) startKeyboard(u *user.User) *botapi.InlineKeyboardMarkup {
	menuBtn := botapi.InlineKeyboardButton{Text: "⬅ Вернуться в меню", CallbackData: callbackBackToMenu}

	var setupBtn *botapi.InlineKeyboardButton

	switch {
	case len(u.GetCourseStreamIDs) == 0:
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
