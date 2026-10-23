package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Woun1zoN/getcourse-schedulix/internal/logging"
	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

type Client struct {
	token            string
	chatID           string
	userService      *user.Service
	client           *http.Client
	logger           *slog.Logger
	sessionStore     *storage.SessionStore
	getCourseBaseURL string
}

func NewClient(token, chatID, getCourseBaseURL string, userService *user.Service, sessionStore *storage.SessionStore) *Client {
	return &Client{
		token:            token,
		chatID:           chatID,
		userService:      userService,
		client:           &http.Client{},
		logger:           logging.NewLogger(),
		sessionStore:     sessionStore,
		getCourseBaseURL: getCourseBaseURL,
	}
}

func (c *Client) answerCallbackQuery(id string) error {
	data, _ := json.Marshal(map[string]string{"callback_query_id": id})
	resp, err := c.client.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", c.token),
		"application/json", bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
