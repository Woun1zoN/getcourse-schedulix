package telegram

import (
	"context"
	"fmt"
	"log"
)

func (c *Client) sendDeleteConfirmation(chatID int64) error {
	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "⬅ Отмена", CallbackData: callbackCancel, Style: "primary"}},
			{{Text: "Да, удалить", CallbackData: callbackDeleteConfirm, Style: "danger"}},
		},
	}

	return c.SendMessageWithKeyboard(
		chatID,
		"⚠️ Все сохраненные настройки и привязка к GetCourse *будут удалены*. Это действие необратимо.\n\n*Вы уверены, что хотите удалить все данные?*",
		kb,
	)
}

func (c *Client) handleDeleteConfirm(ctx context.Context, cq *CallbackQuery) error {
	chatID := cq.Message.Chat.ID

	u, err := c.userService.GetByTelegramID(ctx, cq.From.ID)
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}
	if u == nil {
		c.cleanupMessage(chatID, cq.Message.MessageID)
		return c.SendMessage(chatID, "Нечего удалять, ваших данных нет 🥀")
	}

	if err := c.sessionStore.ClearExpiredNotified(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear expired flag: %w", err)
	}

	if err := c.sessionStore.ClearState(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	if err := c.sessionStore.DeleteHashes(ctx, u.ID); err != nil {
		return fmt.Errorf("delete redis hashes: %w", err)
	}

	if err := c.userService.Delete(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	c.cleanupMessage(chatID, cq.Message.MessageID)

	return c.SendMessage(chatID, "✅ Ваши данные удалены. Нажмите /start, чтобы начать заново.")
}

func (c *Client) cleanupMessage(chatID int64, messageID int) {
	if messageID == 0 {
		return
	}

	if err := c.DeleteMessage(chatID, messageID); err != nil {
		log.Printf("telegram: cleanup message %d in chat %d: %v", messageID, chatID, err)
	}
}
