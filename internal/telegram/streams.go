package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
)

const callbackStreamPrefix = "stream:"

func (c *Client) SendStreamPicker(chatID int64, streams []getcourse.Stream) error {
	if len(streams) == 0 {
		return c.SendMessage(chatID, "Не нашёл ни одного потока на вашем аккаунте.")
	}

	rows := make([][]InlineKeyboardButton, 0, len(streams))
	for _, s := range streams {
		rows = append(rows, []InlineKeyboardButton{{
			Text:         s.Name,
			CallbackData: callbackStreamPrefix + strconv.FormatInt(s.ID, 10),
		}})
	}

	return c.SendMessageWithKeyboard(
		chatID,
		"Выберите поток, за которым следить:",
		&InlineKeyboardMarkup{InlineKeyboard: rows},
	)
}

func (c *Client) handleStreamSelected(cq *CallbackQuery) error {
	streamID, err := strconv.ParseInt(strings.TrimPrefix(cq.Data, callbackStreamPrefix), 10, 64)
	if err != nil {
		return nil
	}

	if err := c.userService.SelectStream(context.Background(), cq.From.ID, streamID); err != nil {
		return fmt.Errorf("select stream: %w", err)
	}

	return c.SendMessage(cq.Message.Chat.ID, "✅ Поток выбран. Новые расписания будут приходить автоматически.")
}