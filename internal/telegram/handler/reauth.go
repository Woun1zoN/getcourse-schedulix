package handler

import "github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"

func (h *Handler) SendReauthPrompt(telegramID int64) error {
	kb := &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{{Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse}},
		},
	}

	return h.client.SendMessageWithKeyboardMode(
		telegramID,
		"*⚠️ Сессия GetCourse истекла*\\. Новые материалы приходить не будут, пока вы не переподключите аккаунт\\.",
		botapi.ParseModeMarkdownV2,
		kb,
	)
}
