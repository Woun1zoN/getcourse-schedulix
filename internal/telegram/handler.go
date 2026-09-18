package telegram

import (
	"context"
	"fmt"
)

func (c *Client) handleMessage(message *Message) error {
	if message.Text != "/start" {
		return nil
	}

	_, err := c.userService.GetOrCreate(
		context.Background(),
		message.From.ID,
	)
	if err != nil {
		return fmt.Errorf("get or create user: %w", err)
	}

	return c.SendMessage(
        message.Chat.ID,
        "Добро пожаловать в Schedulix!\n\n"+"Для начала подключите GetCourse.",
    )
}