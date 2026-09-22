package getcourse

import (
	"fmt"
	"net/http"
	"errors"
	"strings"
)

func (c *Client) ValidateCookie() error {
	req, err := http.NewRequest(
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
		return errors.New("getcourse: cookie невалидна или истекла")
	}

	return nil
}