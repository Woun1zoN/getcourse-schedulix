package telegram

import (
	"context"
	"encoding/json"
	"fmt"
    "net/http"
)

type Update struct {
    UpdateID      int            `json:"update_id"`
    Message       *Message       `json:"message"`
    CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
    From *User `json:"from"`
	Chat Chat  `json:"chat"`
    Text string `json:"text"`
}

type CallbackQuery struct {
    ID      string  `json:"id"`
    From    User    `json:"from"`
    Message Message `json:"message"`
    Data    string  `json:"data"`
}

type Chat struct {
    ID int64 `json:"id"`
}

type User struct {
    ID int64 `json:"id"`
}

func (c *Client) getUpdates(ctx context.Context, offset int) ([]Update, error) {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d",
		c.token,
		offset,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("telegram: getUpdates failed")
	}

	return result.Result, nil
}

func (c *Client) Run(ctx context.Context) error {
	offset := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		updates, err := c.getUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}

		for _, update := range updates {
    		offset = update.UpdateID + 1

    		switch {
    		case update.CallbackQuery != nil:
        		if err := c.handleCallbackQuery(update.CallbackQuery); err != nil {
            		return err
        		}
    		case update.Message != nil:
        		if err := c.handleMessage(update.Message); err != nil {
        		    return err
        		}
			}
		}
	}
}