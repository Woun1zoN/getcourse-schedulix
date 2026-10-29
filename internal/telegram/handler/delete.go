package handler

import (
	"context"
	"fmt"
	"log"

	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
)

func (h *Handler) sendDeleteConfirmation(chatID int64) error {
	kb := &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{{Text: "⬅ Отмена", CallbackData: callbackCancel, Style: "primary"}},
			{{Text: "Да, удалить", CallbackData: callbackDeleteConfirm, Style: "danger"}},
		},
	}

	return h.client.SendMessageWithKeyboard(
		chatID,
		"⚠️ Все сохраненные настройки и привязка к GetCourse *будут удалены*. Это действие необратимо.\n\n*Вы уверены, что хотите удалить все данные?*",
		kb,
	)
}

func (h *Handler) handleDeleteConfirm(ctx context.Context, cq *botapi.CallbackQuery) error {
	chatID := cq.Message.Chat.ID

	u, err := h.userService.GetByTelegramID(ctx, cq.From.ID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if u == nil {
		h.cleanupMessage(chatID, cq.Message.MessageID)
		return h.client.SendMessage(chatID, "Нечего удалять, ваших данных нет 🥀")
	}

	if err := h.sessionStore.ClearExpiredNotified(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear expired flag: %w", err)
	}

	if err := h.sessionStore.ClearState(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	if err := h.sessionStore.DeleteHashes(ctx, u.ID); err != nil {
		return fmt.Errorf("delete redis hashes: %w", err)
	}

	if err := h.userService.Delete(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	h.cleanupMessage(chatID, cq.Message.MessageID)

	return h.client.SendMessage(chatID, "✅ Ваши данные удалены. Нажмите /start, чтобы начать заново.")
}

func (h *Handler) cleanupMessage(chatID int64, messageID int) {
	if messageID == 0 {
		return
	}

	if err := h.client.DeleteMessage(chatID, messageID); err != nil {
		log.Printf("telegram: cleanup message %d in chat %d: %v", messageID, chatID, err)
	}
}
