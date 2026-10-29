package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

const callbackStreamPrefix = "stream:"
const callbackSkipStream = "skip_stream"

func (c *Client) SendStreamPicker(chatID int64, streams []getcourse.Stream) error {
	if len(streams) == 0 {
		return c.SendMessage(chatID, "На вашем аккаунте нет ниодного тренинга.")
	}

	const perRow = 2

	rows := make([][]InlineKeyboardButton, 0, len(streams))

	for i, s := range streams {
		btn := InlineKeyboardButton{
			Text:         s.Name,
			CallbackData: callbackStreamPrefix + strconv.FormatInt(s.ID, 10),
		}

		if i%perRow == 0 {
			rows = append(rows, []InlineKeyboardButton{btn})
		} else {
			last := len(rows) - 1
			rows[last] = append(rows[last], btn)
		}
	}

	rows = append(rows, []InlineKeyboardButton{
		{Text: "Пропустить", CallbackData: callbackSkipStream, Style: "primary"},
	})

	rows = append(rows, []InlineKeyboardButton{
		{Text: "Отмена", CallbackData: callbackCancel, Style: "danger"},
	})

	return c.SendMessageWithKeyboard(
		chatID,
		"🔗 Выберите тренинг, за которым следить:",
		&InlineKeyboardMarkup{InlineKeyboard: rows},
	)
}

func (c *Client) handleStreamSelected(ctx context.Context, cq *CallbackQuery) error {
	streamID, err := strconv.ParseInt(strings.TrimPrefix(cq.Data, callbackStreamPrefix), 10, 64)
	if err != nil {
		return nil
	}

	if err := c.userService.SelectStream(ctx, cq.From.ID, streamID); err != nil {
		return fmt.Errorf("select stream: %w", err)
	}

	u, err := c.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	if u.TargetChatID == nil {
		if err := c.SendChatPicker(cq.From.ID, ChatPickerInitial); err != nil {
			return err
		}

		c.cleanupMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		return nil
	}

	text, kb := c.renderConnectedMenu(u)

	if err := c.SendMessageWithKeyboardMode(cq.From.ID, text, ParseModeMarkdownV2, kb); err != nil {
		return err
	}

	c.cleanupMessage(cq.Message.Chat.ID, cq.Message.MessageID)
	return nil
}

func (c *Client) currentStream(u *user.User) *getcourse.Stream {
	if u.GetCourseCookie == nil || u.GetCourseStreamID == nil {
		return nil
	}

	client := getcourse.NewClient(c.getCourseBaseURL, *u.GetCourseCookie)
	streams, err := client.GetStreams()
	if err != nil {
		return nil
	}

	for _, s := range streams {
		if s.ID == *u.GetCourseStreamID {
			return &s
		}
	}
	return nil
}
