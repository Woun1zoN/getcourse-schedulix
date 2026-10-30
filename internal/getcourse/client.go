package getcourse

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	cookies string
	client  *http.Client
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

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Cookie", c.cookies)

		resp, err := c.client.Do(req)
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			lastErr = err

			if isTimeout(err) {
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

func (c *Client) DownloadDocument(ctx context.Context, document Document) ([]byte, error) {
	return c.get(ctx, document.URL)
}

func (c *Client) WithCookies(cookies string) *Client {
	return &Client{
		baseURL: c.baseURL,
		cookies: cookies,
		client:  c.client,
	}
}

func isTimeout(err error) bool {
	var ne net.Error
	return errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout())
}
