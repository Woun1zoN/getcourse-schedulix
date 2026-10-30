package getcourse

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"context"
)

var ErrCookieExpired = errors.New("getcourse: cookie invalid or expired")

func (c *Client) ValidateCookie(ctx context.Context) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/teach/control/",
		nil,
	)
	if err != nil {
		return fmt.Errorf("getcourse: validate: %w", err)
	}

	req.Header.Set("Cookie", c.cookies)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("getcourse: validate: %w", err)
	}
	defer resp.Body.Close()

	if strings.Contains(resp.Request.URL.Path, "/cms/system/login") {
		return ErrCookieExpired
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("getcourse: validate: server status %d", resp.StatusCode)
	}

	return nil
}
