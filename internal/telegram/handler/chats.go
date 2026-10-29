package handler

import (
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
)

type ChatPickerMode int

const (
	ChatPickerInitial ChatPickerMode = iota
	ChatPickerChange
)

const (
	chatRequestGroup   = 1
	chatRequestChannel = 2
)

func (h *Handler) SendChatPicker(chatID int64, mode ChatPickerMode) error {
	skipButtonText := "➡ Пропустить"

	if mode == ChatPickerChange {
		skipButtonText = "➖ Без чата"
	}

	kb := &botapi.ReplyKeyboardMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		Keyboard: [][]botapi.KeyboardButton{
			{
				{
					Text: "💬 Личные сообщения",
				},
			},
			{
				{
					Text: "👥 Выбрать группу",
					RequestChat: &botapi.KeyboardButtonRequestChat{
						RequestID:     chatRequestGroup,
						ChatIsChannel: false,
						BotIsMember:   true,
						RequestTitle:  true,
						UserAdministratorRights: &botapi.ChatAdministratorRights{
							CanManageChat: true,
						},
					},
				},
			},
			{
				{
					Text: "📢 Выбрать канал",
					RequestChat: &botapi.KeyboardButtonRequestChat{
						RequestID:     chatRequestChannel,
						ChatIsChannel: true,
						BotIsMember:   true,
						UserAdministratorRights: &botapi.ChatAdministratorRights{
							CanManageChat:   true,
							CanPostMessages: true,
						},
						BotAdministratorRights: &botapi.ChatAdministratorRights{
							CanPostMessages: true,
						},
					},
				},
			},
			{
				{
					Text: skipButtonText,
				},
			},
		},
	}

	return h.client.SendMessageWithKeyboardMode(
		chatID,
		"🔗 Выберите чат, куда Schedulix будет отправлять новые материалы.",
		"",
		kb,
	)
}
