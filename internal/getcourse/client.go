package getcourse

import (
	"fmt"
	"io"
	"net/http"
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
		client:  &http.Client{},
	}
}

func (c *Client) get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Cookie", c.cookies)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getcourse: status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) DownloadDocument(document Document) ([]byte, error) {
	return c.get(document.URL)
}