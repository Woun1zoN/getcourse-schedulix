package handler

import (
    "fmt"

    "github.com/Woun1zoN/getcourse-schedulix/internal/telegram/botapi"
)

func (h *Handler) SendNewMaterialNotice(
    documentName, documentURL, lessonID, lessonURL string,
    fileSize float64,
    checked, newDocuments, logChatID int64,
) error {
    message := fmt.Sprintf(
        "📚 Получен новый материал\n\n"+
            "📄 [%s](%s)\n\n"+
            "• ID урока: `%s`\n"+
            "• URL: %s\n"+
            "• Размер файла: %.1f КБ\n\n"+
            "Проверено уроков: %d · Найдено: %d",
        documentName, documentURL, lessonID, lessonURL,
        fileSize, checked, newDocuments,
    )

    kb := &botapi.InlineKeyboardMarkup{
        InlineKeyboard: [][]botapi.InlineKeyboardButton{
            {{Text: "⬅ Меню", CallbackData: callbackBackToMenu}},
        },
    }

    return h.client.SendMessageWithKeyboardMode(
        logChatID,
        message,
        botapi.ParseModeMarkdown,
        kb,
    )
}