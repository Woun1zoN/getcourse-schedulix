package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

const callbackStreamPrefix = "stream:"
const callbackSkipStream = "skip_stream"

func (h *Handler) SendStreamPicker(chatID int64, streams []getcourse.Stream) error {
	if len(streams) == 0 {
		return h.client.SendMessage(chatID, "На вашем аккаунте нет ниодного тренинга.")
	}

	const perRow = 2

	rows := make([][]botapi.InlineKeyboardButton, 0, len(streams))

	for i, s := range streams {
		btn := botapi.InlineKeyboardButton{
			Text:         s.Name,
			CallbackData: callbackStreamPrefix + strconv.FormatInt(s.ID, 10),
		}

		if i%perRow == 0 {
			rows = append(rows, []botapi.InlineKeyboardButton{btn})
		} else {
			last := len(rows) - 1
			rows[last] = append(rows[last], btn)
		}
	}

	rows = append(rows, []botapi.InlineKeyboardButton{
		{Text: "Пропустить", CallbackData: callbackSkipStream, Style: "primary"},
	})

	rows = append(rows, []botapi.InlineKeyboardButton{
		{Text: "Отмена", CallbackData: callbackCancel, Style: "danger"},
	})

	return h.client.SendMessageWithKeyboard(
		chatID,
		"🔗 Выберите тренинг, за которым следить:",
		&botapi.InlineKeyboardMarkup{InlineKeyboard: rows},
	)
}

func (h *Handler) handleStreamSelected(ctx context.Context, cq *botapi.CallbackQuery) error {
	streamID, err := strconv.ParseInt(strings.TrimPrefix(cq.Data, callbackStreamPrefix), 10, 64)
	if err != nil {
		return nil
	}

	if err := h.userService.SelectStream(ctx, cq.From.ID, streamID); err != nil {
		return fmt.Errorf("select stream: %w", err)
	}

	u, err := h.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	if u.TargetChatID == nil {
		if err := h.SendChatPicker(cq.From.ID, ChatPickerInitial); err != nil {
			return err
		}

		h.cleanupMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		return nil
	}

	text, kb := h.renderConnectedMenu(u)

	if err := h.client.SendMessageWithKeyboardMode(cq.From.ID, text, botapi.ParseModeMarkdownV2, kb); err != nil {
		return err
	}

	h.cleanupMessage(cq.Message.Chat.ID, cq.Message.MessageID)
	return nil
}

func (h *Handler) currentStream(u *user.User) *getcourse.Stream {
	if u.GetCourseCookie == nil || u.GetCourseStreamID == nil {
		return nil
	}

	client := getcourse.NewClient(h.getCourseBaseURL, *u.GetCourseCookie)
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
