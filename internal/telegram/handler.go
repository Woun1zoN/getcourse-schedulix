package telegram

import (
	"context"
	"fmt"
)

func (c *Client) handleMessage(message *Message) error {
    if message.Text != "/start" {
        return nil
    }

    c.logger.Info("telegram command",
        "user_id", message.From.ID,
        "command", message.Text,
    )

    _, err := c.userService.GetOrCreate(context.Background(), message.From.ID)
    if err != nil {
        c.logger.Error("get or create user failed",
            "user_id", message.From.ID,
            "error", err,
        )
        return fmt.Errorf("get or create user: %w", err)
    }

    return c.SendMessage(
        message.Chat.ID,
        "Добро пожаловать в Schedulix!\n\n"+"Для начала подключите GetCourse.",
    )
}