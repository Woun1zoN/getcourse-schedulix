package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

type StreamPickerMode int

const (
	StreamPickerInitial StreamPickerMode = iota
	StreamPickerChange
)

const (
	callbackStreamPrefix = "stream:"
	callbackStreamsDone  = "streams_done"
	callbackSkipStream   = "skip_stream"
)

const streamPickerText = `🔗 Выберите тренинги, за которыми следить.

Повторно нажмите на выбранный тренинг, чтобы снять выбор.`

type streamDraft struct {
	Streams  []getcourse.Stream `json:"streams"`
	Selected []int64            `json:"selected"`
	Touched  bool               `json:"touched"`
	Mode     StreamPickerMode   `json:"mode"`
}

func (d *streamDraft) isSelected(id int64) bool {
	return slices.Contains(d.Selected, id)
}

func (d *streamDraft) toggle(id int64) bool {
	if !slices.ContainsFunc(d.Streams, func(s getcourse.Stream) bool { return s.ID == id }) {
		return false
	}

	if i := slices.Index(d.Selected, id); i != -1 {
		d.Selected = slices.Delete(d.Selected, i, i+1)
	} else {
		d.Selected = append(d.Selected, id)
	}

	d.Touched = true
	return true
}

func (d *streamDraft) keyboard() *botapi.InlineKeyboardMarkup {
	const perRow = 2

	rows := make([][]botapi.InlineKeyboardButton, 0, len(d.Streams)/perRow+2)

	for i, s := range d.Streams {
		btn := botapi.InlineKeyboardButton{
			Text:         s.Name,
			CallbackData: callbackStreamPrefix + strconv.FormatInt(s.ID, 10),
		}

		if d.isSelected(s.ID) {
			btn.Text = s.Name + " · Выбрано"
			btn.Style = "primary"
		}

		if i%perRow == 0 {
			rows = append(rows, []botapi.InlineKeyboardButton{btn})
		} else {
			last := len(rows) - 1
			rows[last] = append(rows[last], btn)
		}
	}

	if d.Touched || len(d.Selected) > 0 {
		rows = append(rows,
			[]botapi.InlineKeyboardButton{
				{Text: "Готово", CallbackData: callbackStreamsDone, Style: "success"},
			},
			[]botapi.InlineKeyboardButton{
				{Text: "Отмена", CallbackData: callbackCancel, Style: "danger"},
			},
		)
	} else {
		if d.Mode == StreamPickerInitial {
			rows = append(rows, []botapi.InlineKeyboardButton{
				{Text: "Пропустить", CallbackData: callbackSkipStream, Style: "primary"},
			})
		}
		rows = append(rows, []botapi.InlineKeyboardButton{
			{Text: "Отмена", CallbackData: callbackCancel, Style: "danger"},
		})
	}

	return &botapi.InlineKeyboardMarkup{InlineKeyboard: rows}
}

func (h *Handler) saveDraft(ctx context.Context, telegramID int64, d *streamDraft) error {
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}
	return h.sessionStore.SetStreamDraft(ctx, telegramID, data)
}

func (h *Handler) loadDraft(ctx context.Context, telegramID int64) (*streamDraft, error) {
	data, ok, err := h.sessionStore.GetStreamDraft(ctx, telegramID)
	if err != nil || !ok {
		return nil, err
	}

	d := &streamDraft{}
	if err := json.Unmarshal(data, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (h *Handler) sendStreamPicker(ctx context.Context, chatID int64, streams []getcourse.Stream, selected []int64, mode StreamPickerMode) error {
	if len(streams) == 0 {
		return h.client.SendMessage(chatID, "На вашем аккаунте нет ниодного тренинга.")
	}

	d := &streamDraft{Streams: streams, Mode: mode}

	for _, id := range selected {
		if slices.ContainsFunc(streams, func(s getcourse.Stream) bool { return s.ID == id }) {
			d.Selected = append(d.Selected, id)
		}
	}
	d.Touched = len(d.Selected) > 0

	if err := h.saveDraft(ctx, chatID, d); err != nil {
		return fmt.Errorf("save stream draft: %w", err)
	}

	return h.client.SendMessageWithKeyboard(chatID, streamPickerText, d.keyboard())
}

func (h *Handler) handleStreamToggle(ctx context.Context, cq *botapi.CallbackQuery) error {
	streamID, err := strconv.ParseInt(strings.TrimPrefix(cq.Data, callbackStreamPrefix), 10, 64)
	if err != nil {
		return nil
	}

	d, err := h.loadDraft(ctx, cq.From.ID)
	if err != nil {
		return fmt.Errorf("load stream draft: %w", err)
	}
	if d == nil {
		return h.client.SendMessage(cq.From.ID, "Список устарел, откройте выбор тренингов заново.")
	}

	if !d.toggle(streamID) {
		return nil
	}

	if err := h.saveDraft(ctx, cq.From.ID, d); err != nil {
		return fmt.Errorf("save stream draft: %w", err)
	}

	return h.client.EditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, d.keyboard())
}

func (h *Handler) handleStreamsDone(ctx context.Context, cq *botapi.CallbackQuery) error {
	d, err := h.loadDraft(ctx, cq.From.ID)
	if err != nil {
		return fmt.Errorf("load stream draft: %w", err)
	}
	if d == nil {
		return h.client.SendMessage(cq.From.ID, "Список устарел, откройте выбор тренингов заново.")
	}

	if err := h.userService.SelectStreams(ctx, cq.From.ID, d.Selected); err != nil {
		return fmt.Errorf("select streams: %w", err)
	}

	if err := h.sessionStore.DeleteStreamDraft(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("delete stream draft: %w", err)
	}

	u, err := h.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	if u.TargetChatID == nil {
		if err := h.sendChatPicker(cq.From.ID, ChatPickerInitial); err != nil {
			return err
		}
		h.cleanupMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		return nil
	}

	text, kb := h.renderConnectedMenu(ctx, u)
	if err := h.sendMenuMessage(ctx, cq.From.ID, cq.From.ID, text, kb); err != nil {
		return err
	}

	h.cleanupMessage(cq.Message.Chat.ID, cq.Message.MessageID)
	return nil
}

func (h *Handler) currentStreams(ctx context.Context, u *user.User) ([]getcourse.Stream, error) {
	if u.GetCourseCookie == nil || len(u.GetCourseStreamIDs) == 0 {
		return nil, nil
	}

	client := getcourse.NewClient(h.getCourseBaseURL, *u.GetCourseCookie)
	all, err := client.GetStreams(ctx)
	if err != nil {
		return nil, fmt.Errorf("get streams: %w", err)
	}

	var result []getcourse.Stream
	for _, s := range all {
		if slices.Contains(u.GetCourseStreamIDs, s.ID) {
			result = append(result, s)
		}
	}
	return result, nil
}
