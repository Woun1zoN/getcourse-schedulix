package getcourse

import (
	"fmt"
	"io"
	"net/http"
	"time"
	"errors"
	"context"
)

type Client struct {
	baseURL 	string
	cookies 	string
	client  	*http.Client
}

func NewClient(baseURL, cookies string) *Client {
	return &Client{
		baseURL: baseURL,
		cookies: cookies,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) get(url string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Cookie", c.cookies)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err

			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}

			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("getcourse: status %d", resp.StatusCode)
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		return data, err
	}

	return nil, fmt.Errorf("getcourse: timed out after retries: %w", lastErr)
}

func (c *Client) DownloadDocument(document Document) ([]byte, error) {
	return c.get(document.URL)
}