package botapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	ID int64 `json:"id"`
}

func (c *Client) GetUpdates(ctx context.Context, offset int) ([]Update, error) {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30",
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
