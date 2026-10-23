package telegram

import (
	"context"
	"fmt"
	"strings"
)

const (
	callbackConnectGetCourse = "connect_getcourse"
	callbackPickStream       = "pick_stream"
	callbackPickChat         = "pick_chat"
	callbackCancel           = "cancel"
	callbackBackToMenu       = "back_to_menu"
)

func (c *Client) handleCallbackQuery(cq *CallbackQuery) error {
	_ = c.answerCallbackQuery(cq.ID)

	switch {
	case strings.HasPrefix(cq.Data, callbackStreamPrefix):
		return c.handleStreamSelected(cq)

	case cq.Data == callbackPickStream:
		return c.handlePickStream(cq)

	case cq.Data == callbackPickChat:
		return c.SendChatPicker(cq.Message.Chat.ID, ChatPickerChange)

	case cq.Data == callbackConnectGetCourse:
		if err := c.sessionStore.SetAwaitingCookie(context.Background(), cq.From.ID); err != nil {
			return err
		}
		return c.SendMessageWithKeyboardMode(cq.Message.Chat.ID, connectGetCourseMessage, ParseModeMarkdownV2, cancelKeyboard())

	case cq.Data == callbackCancel:
		return c.handleCancel(cq)

	case cq.Data == callbackSkipStream:
		return c.handleSkipStream(cq)

	case cq.Data == callbackBackToMenu:
		return c.sendMenu(cq.Message.Chat.ID, cq.From.ID)

	case cq.Data == callbackDeleteConfirm:
		return c.handleDeleteConfirm(cq)
	}

	return nil
}

func (c *Client) handleCancel(cq *CallbackQuery) error {
	ctx := context.Background()

	if err := c.sessionStore.ClearState(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	u, err := c.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	chatID := cq.Message.Chat.ID

	if u.GetCourseCookie != nil {
		text, kb := c.renderConnectedMenu(u)
		return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, kb)
	}

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
		},
	}

	return c.SendMessageWithKeyboardMode(chatID, welcomeText, ParseModeMarkdownV2, kb)
}

func (c *Client) handleSkipStream(cq *CallbackQuery) error {
	ctx := context.Background()

	if err := c.userService.ClearStream(ctx, cq.From.ID); err != nil {
		return fmt.Errorf("clear stream: %w", err)
	}

	u, err := c.userService.GetOrCreate(ctx, cq.From.ID)
	if err != nil {
		return err
	}

	if u.TargetChatID == nil {
		return c.SendChatPicker(cq.From.ID, ChatPickerInitial)
	}

	text, kb := c.renderConnectedMenu(u)
	return c.SendMessageWithKeyboardMode(cq.From.ID, text, ParseModeMarkdownV2, kb)
}

func cancelKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "Отмена", CallbackData: callbackCancel, Style: "danger"}},
		},
	}
}
