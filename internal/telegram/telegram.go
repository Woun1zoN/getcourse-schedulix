package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
)

type Client struct {
	token     string
	logChatID string
	chatID    string
	client    *http.Client
}

type response struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func NewClient(token, chatID, logChatID string) *Client {
	return &Client{
		token:     token,
		logChatID: logChatID,
		chatID:    chatID,
		client:    &http.Client{},
	}
}

func (c *Client) SendPhoto(path string, caption string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", c.chatID); err != nil {
		return err
	}

	if err := writer.WriteField("caption", caption); err != nil {
		return err
	}

	if err := writer.WriteField("parse_mode", "Markdown"); err != nil {
		return err
	}

	part, err := writer.CreateFormFile("photo", "schedule.jpg")
	if err != nil {
		return err
	}

	if _, err := file.WriteTo(part); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendPhoto",
		c.token,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		&body,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result response

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram: decode response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf(
			"telegram: %s",
			result.Description,
		)
	}

	return nil
}

func (c *Client) SendNewDocumentLog(documentName, documentURL, lessonID, lessonURL string, fileSize float64, checked, newDocuments int64) error {
    message := fmt.Sprintf(
    "INFO | Получен новый документ\n\n"+
        "📄 [%s](%s)\n\n"+
        "• ID урока: `%s`\n"+
        "• URL: %s\n"+
        "• Размер файла: %.1f КБ\n\n"+
        "Проверено уроков: %d · Найдено: %d",
    documentName, documentURL, lessonID, lessonURL, fileSize, checked, newDocuments,
    )

    payload := struct {
        ChatID string `json:"chat_id"`
        Text   string `json:"text"`
		ParseMode string `json:"parse_mode"`
    }{
        ChatID: c.logChatID,
        Text:   message,
		ParseMode: "Markdown",
    }

    data, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    url := fmt.Sprintf(
        "https://api.telegram.org/bot%s/sendMessage",
        c.token,
    )

    req, err := http.NewRequest(
        http.MethodPost,
        url,
        bytes.NewReader(data),
    )
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")

    resp, err := c.client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    var result response

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return fmt.Errorf("telegram: decode response: %w", err)
    }

    if !result.OK {
        return fmt.Errorf(
            "telegram: status %d: %s",
            resp.StatusCode,
            result.Description,
        )
    }

    return nil
}