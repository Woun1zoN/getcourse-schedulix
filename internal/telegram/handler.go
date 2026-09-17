package telegram

import (
	"context"
	"fmt"
)

func (c *Client) handleMessage(message *Message) error {
	if message.Text != "/start" {
		return nil
	}

	user, err := c.userService.GetOrCreate(
		context.Background(),
		message.From.ID,
	)
	if err != nil {
		return fmt.Errorf("get or create user: %w", err)
	}

	fmt.Println("user:", user.TelegramID)

	return nil
}