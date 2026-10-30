package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
)

const (
	callbackConnectGetCourse = "connect_getcourse"
	callbackPickStream       = "pick_stream"
	callbackPickChat         = "pick_chat"
	callbackCancel           = "cancel"
	callbackBackToMenu       = "back_to_menu"
	callbackDelete           = "delete"
	callbackDeleteConfirm    = "delete_confirm"
)

func (h *Handler) handleCallbackQuery(ctx context.Context, cq *botapi.CallbackQuery) error {
	_ = h.client.AnswerCallbackQuery(cq.ID)

	switch {
	case strings.HasPrefix(cq.Data, callbackStreamPrefix):
		return h.handleStreamSelected(ctx, cq)

	case cq.Data == callbackPickStream:
		return h.handlePickStream(ctx, cq)

	case cq.Data == callbackPickChat:
		return h.sendChatPicker(cq.Message.Chat.ID, ChatPickerChange)

	case cq.Data == callbackConnectGetCourse:
		if err := h.sessionStore.SetAwaitingCookie(ctx, cq.From.ID); err != nil {
			return err
		}
		msgID, err := h.client.SendMessageWithKeyboardModeID(
			cq.Message.Chat.ID, connectGetCourseMessage, botapi.ParseModeMarkdownV2, cancelKeyboard(),
		)
		if err != nil {
			return err
		}

		if err := h.sessionStore.SetCookiePromptMessage(ctx, cq.From.ID, msgID); err != nil {
			log.Printf("telegram: save cookie prompt message: %v", err)
		}

		return nil

	case cq.Data == callbackCancel:
		return h.handleCancel(ctx, cq)

	case cq.Data == callbackSkipStream:
		return h.handleSkipStream(ctx, cq)

	case cq.Data == callbackBackToMenu:
		return h.sendMenu(ctx, cq.Message.Chat.ID, cq.From.ID)

	case cq.Data == callbackDeleteConfirm:
		return h.handleDeleteConfirm(ctx, cq)

	case cq.Data == callbackDelete:
		return h.sendDeleteConfirmation(cq.Message.Chat.ID)
	}

	return nil
}

func (h *Handler) handleCancel(ctx context.Context, cq *botapi.CallbackQuery) error {
	if err := h.sessionStore.ClearState(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	u, err := h.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	chatID := cq.Message.Chat.ID

	if u.GetCourseCookie != nil {
		text, kb := h.renderConnectedMenu(u)

		if err := h.client.SendMessageWithKeyboardMode(chatID, text, botapi.ParseModeMarkdownV2, kb); err != nil {
			return err
		}

		h.cleanupMessage(chatID, cq.Message.MessageID)
		return nil
	}

	kb := &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
		},
	}

	if err := h.client.SendMessageWithKeyboardMode(chatID, welcomeText, botapi.ParseModeMarkdownV2, kb); err != nil {
		return err
	}

	h.cleanupMessage(chatID, cq.Message.MessageID)
	return nil
}

func (h *Handler) handleSkipStream(ctx context.Context, cq *botapi.CallbackQuery) error {
	if err := h.userService.ClearStream(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear stream: %w", err)
	}

	u, err := h.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	if u.TargetChatID == nil {
		return h.sendChatPicker(cq.From.ID, ChatPickerInitial)
	}

	text, kb := h.renderConnectedMenu(u)
	return h.client.SendMessageWithKeyboardMode(cq.From.ID, text, botapi.ParseModeMarkdownV2, kb)
}

func cancelKeyboard() *botapi.InlineKeyboardMarkup {
	return &botapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]botapi.InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: callbackCancel, Style: "danger"}},
		},
	}
}
