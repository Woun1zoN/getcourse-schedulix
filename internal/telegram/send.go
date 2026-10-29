package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
)

type response struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
}

func (c *Client) SendMessageWithKeyboard(chatID int64, text string, kb *InlineKeyboardMarkup) error {
	return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdown, kb)
}

func (c *Client) SendMessageMode(chatID int64, text, parseMode string) error {
	return c.SendMessageWithKeyboardMode(chatID, text, parseMode, nil)
}

func (c *Client) SendPhoto(chatID int64, path string, caption string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", fmt.Sprintf("%d", chatID)); err != nil {
		return err
	}

	if err := writer.WriteField("caption", caption); err != nil {
		return err
	}

	if err := writer.WriteField("parse_mode", ParseModeMarkdown); err != nil {
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

func (c *Client) SendNewDocumentLog(documentName, documentURL, lessonID, lessonURL string, fileSize float64, checked, newDocuments, logChatID int64) error {
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
		ChatID    string `json:"chat_id"`
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}{
		ChatID:    fmt.Sprintf("%d", logChatID),
		Text:      message,
		ParseMode: ParseModeMarkdown,
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

func (c *Client) SendMessage(chatID int64, text string) error {
	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage",
		c.token,
	)

	payload := struct {
		ChatID                int64  `json:"chat_id"`
		Text                  string `json:"text"`
		ParseMode             string `json:"parse_mode"`
		DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	}{
		ChatID:                chatID,
		Text:                  text,
		ParseMode:             ParseModeMarkdown,
		DisableWebPagePreview: true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(
		url,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram: sendMessage status %d", resp.StatusCode)
	}

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram: sendMessage: %s", result.Description)
	}

	return nil
}

func (c *Client) SendMessageWithKeyboardMode(chatID int64, text, parseMode string, replyMarkup any) error {
	_, err := c.SendMessageWithKeyboardModeID(chatID, text, parseMode, replyMarkup)
	return err
}

func (c *Client) SendMessageWithKeyboardModeID(chatID int64, text, parseMode string, replyMarkup any) (int, error) {
	payload := SendMessagePayload{
		ChatID:             chatID,
		Text:               text,
		ParseMode:          parseMode,
		ReplyMarkup:        replyMarkup,
		LinkPreviewOptions: &LinkPreviewOptions{IsDisabled: true},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.token),
		"application/json", bytes.NewReader(data),
	)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("telegram: decode response: %w", err)
	}
	if !result.OK {
		return 0, fmt.Errorf("telegram: sendMessage: %s", result.Description)
	}

	return result.Result.MessageID, nil
}

func (c *Client) SendMessageAndRemoveKeyboard(chatID int64, text string) error {
	return c.SendMessageWithKeyboardMode(chatID, text, ParseModeMarkdownV2, &ReplyKeyboardRemove{
		RemoveKeyboard: true,
	})
}

func (c *Client) DeleteMessage(chatID int64, messageID int) error {
	payload := struct {
		ChatID    int64 `json:"chat_id"`
		MessageID int   `json:"message_id"`
	}{
		ChatID:    chatID,
		MessageID: messageID,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := c.client.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/deleteMessage", c.token),
		"application/json", bytes.NewReader(data),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram: decode response: %w", err)
	}

	if !result.OK {
		return fmt.Errorf("telegram: deleteMessage: %s", result.Description)
	}

	return nil
}
