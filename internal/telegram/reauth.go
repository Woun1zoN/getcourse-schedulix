package telegram

func (c *Client) SendReauthPrompt(telegramID int64) error {
	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse}},
		},
	}

	return c.SendMessageWithKeyboardMode(
		telegramID,
		"*⚠️ Сессия GetCourse истекла*\\. Новые материалы приходить не будут, пока вы не переподключите аккаунт\\.",
		ParseModeMarkdownV2,
		kb,
	)
}
