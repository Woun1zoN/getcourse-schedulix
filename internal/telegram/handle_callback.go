package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
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

func (c *Client) handleCallbackQuery(ctx context.Context, cq *CallbackQuery) error {
	_ = c.answerCallbackQuery(cq.ID)

	switch {
	case strings.HasPrefix(cq.Data, callbackStreamPrefix):
		return c.handleStreamSelected(ctx, cq)

	case cq.Data == callbackPickStream:
		return c.handlePickStream(ctx, cq)

	case cq.Data == callbackPickChat:
		return c.SendChatPicker(cq.Message.Chat.ID, ChatPickerChange)

	case cq.Data == callbackConnectGetCourse:
		if err := c.sessionStore.SetAwaitingCookie(ctx, cq.From.ID); err != nil {
			return err
		}
		msgID, err := c.SendMessageWithKeyboardModeID(
			cq.Message.Chat.ID, connectGetCourseMessage, ParseModeMarkdownV2, cancelKeyboard(),
		)
		if err != nil {
			return err
		}

		if err := c.sessionStore.SetCookiePromptMessage(ctx, cq.From.ID, msgID); err != nil {
			log.Printf("telegram: save cookie prompt message: %v", err)
		}

		return nil

	case cq.Data == callbackCancel:
		return c.handleCancel(ctx, cq)

	case cq.Data == callbackSkipStream:
		return c.handleSkipStream(ctx, cq)

	case cq.Data == callbackBackToMenu:
		return c.sendMenu(ctx, cq.Message.Chat.ID, cq.From.ID)

	case cq.Data == callbackDeleteConfirm:
		return c.handleDeleteConfirm(ctx, cq)

	case cq.Data == callbackDelete:
		return c.sendDeleteConfirmation(cq.Message.Chat.ID)
	}

	return nil
}

func (c *Client) handleCancel(ctx context.Context, cq *CallbackQuery) error {
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

		if err := c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, kb); err != nil {
			return err
		}

		c.cleanupMessage(chatID, cq.Message.MessageID)
		return nil
	}

	kb := &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
		},
	}

	if err := c.SendMessageWithKeyboardMode(chatID, welcomeText, ParseModeMarkdownV2, kb); err != nil {
		return err
	}

	c.cleanupMessage(chatID, cq.Message.MessageID)
	return nil
}

func (c *Client) handleSkipStream(ctx context.Context, cq *CallbackQuery) error {
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
