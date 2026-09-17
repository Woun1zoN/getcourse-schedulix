package telegram

import (
	"encoding/json"
	"fmt"
)

type Update struct {
    UpdateID int `json:"update_id"`
    Message  *Message `json:"message"`
}

type Message struct {
    From *User `json:"from"`
	Chat Chat  `json:"chat"`
    Text string `json:"text"`
}

type Chat struct {
    ID int64 `json:"id"`
}

type User struct {
    ID int64 `json:"id"`
}

func (c *Client) getUpdates(offset int) ([]Update, error) {
    url := fmt.Sprintf(
        "https://api.telegram.org/bot%s/getUpdates?offset=%d",
        c.token,
        offset,
    )

    resp, err := c.client.Get(url)
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

func (c *Client) Run() error {
    offset := 0

    for {
        updates, err := c.getUpdates(offset)
        if err != nil {
            return err
        }

        for _, update := range updates {
            offset = update.UpdateID + 1

            if update.Message == nil {
                continue
            }

            if err := c.handleMessage(update.Message); err != nil {
                return err
            }
        }
    }
}