package telegram

import (
	"context"
	"fmt"
    "strings"
    "log/slog"

    "github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
    "github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

const callbackConnectGetCourse = "connect_getcourse"
const callbackPickStream = "pick_stream"

func (c *Client) handleMessage(message *Message) error {
	ctx := context.Background()

	if message.Chat.Type != "private" {
		return nil
	}

	u, err := c.userService.GetOrCreate(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("get or create user: %w", err)
	}

	if message.Text == "/start" {
		if err := c.sessionStore.ClearState(ctx, message.From.ID); err != nil {
			return fmt.Errorf("clear session state: %w", err)
		}

		c.logger.Info(
			"telegram command",
			slog.String("command", "/start"),
			slog.Int64("telegram_id", message.From.ID),
		)

		switch {
		case u.GetCourseCookie == nil:
			kb := &InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{{Text: "🔗 Подключить GetCourse", CallbackData: callbackConnectGetCourse}},
				},
			}
			return c.SendMessageWithKeyboard(message.Chat.ID,
				"Добро пожаловать в Schedulix!\n\nДля начала подключите GetCourse.", kb)

		case u.GetCourseStreamID == nil:
			kb := &InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{{Text: "📚 Выбрать поток", CallbackData: callbackPickStream}},
					{{Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse}},
				},
			}

			return c.SendMessageWithKeyboard(message.Chat.ID,
				"С возвращением! GetCourse подключён, но поток ещё не выбран.\n\nБез потока расписания приходить не будут.", kb)

		default:
			kb := &InlineKeyboardMarkup{
				InlineKeyboard: [][]InlineKeyboardButton{
					{{Text: "🔄 Переподключить GetCourse", CallbackData: callbackConnectGetCourse}},
					{{Text: "🔀 Сменить поток", CallbackData: callbackPickStream}},
				},
			}

			return c.SendMessageWithKeyboard(message.Chat.ID,
				"С возвращением! GetCourse уже подключён.", kb)
		}
	}

	awaiting, err := c.sessionStore.IsAwaitingCookie(ctx, message.From.ID)
	if err != nil {
		return fmt.Errorf("check session state: %w", err)
	}

	if awaiting {
		return c.handleCookieInput(ctx, u, message)
	}

	return nil
}

func (c *Client) handleCallbackQuery(cq *CallbackQuery) error {
    _ = c.answerCallbackQuery(cq.ID)

	if strings.HasPrefix(cq.Data, callbackStreamPrefix) {
		return c.handleStreamSelected(cq)
	}

	if cq.Data == callbackPickStream {
		return c.handlePickStream(cq)
	}

    if cq.Data != callbackConnectGetCourse {
        return nil
    }

    if err := c.sessionStore.SetAwaitingCookie(context.Background(), cq.From.ID); err != nil {
        return err
    }

    return c.SendMessage(cq.Message.Chat.ID,

    "🔗 Подключение GetCourse\n\n"+

        "Установите Cookie-Editor для вашего браузера:\n"+
        "• [Google Chrome](https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)\n"+
        "• [Яндекс Браузер](https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)\n"+
        "• [Mozilla Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)\n\n"+

        "После установки:\n"+
        "1. Войдите в shtpt.getcourse.ru под своим аккаунтом.\n"+
        "2. Кликните по иконке Cookie-Editor на этой же вкладке.\n"+
        "3. Нажмите Export → выберите формат «Header String».\n"+
        "4. Вставьте скопированную строку сюда одним сообщением.\n\n"+

        "Не получается с расширением? Инструкция через DevTools:\n"+
        "1. Откройте shtpt.getcourse.ru и войдите в аккаунт.\n"+
        "2. Нажмите F12 → вкладка Network.\n"+
        "3. Обновите страницу (F5) и кликните на любой запрос слева.\n"+
        "4. Во вкладке Headers найдите «cookie:» среди Request Headers.\n"+
        "5. Скопируйте значение целиком, без слова «cookie:», и пришлите сюда.\n\n"+

        "⚠️ Эта строка даёт полный доступ к вашему аккаунту GetCourse. "+
        "Никому её не передавайте, кроме этого бота.",

)
}

func (c *Client) handleCookieInput(ctx context.Context, u *user.User, message *Message) error {
	cookie := strings.TrimSpace(message.Text)

	if cookie == "" {
		return c.SendMessage(message.Chat.ID, "Cookie пустая, попробуйте ещё раз.")
	}

	client := getcourse.NewClient(c.getCourseBaseURL, cookie)

	if err := client.ValidateCookie(); err != nil {
		return c.SendMessage(
			message.Chat.ID,
			"Не удалось подтвердить cookie. Убедитесь, что вы вошли в аккаунт и скопировали значение полностью.",
		)
	}

	if err := c.userService.ConnectGetCourse(ctx, u.TelegramID, cookie); err != nil {
		return fmt.Errorf("connect getcourse: %w", err)
	}

	if err := c.sessionStore.ClearState(ctx, u.TelegramID); err != nil {
		return fmt.Errorf("clear session state: %w", err)
	}

	if err := c.SendMessage(message.Chat.ID, "✅ GetCourse успешно подключён!"); err != nil {
		return err
	}

	streams, err := client.GetStreams()
	if err != nil {
		return c.SendMessage(message.Chat.ID, "Не удалось получить список потоков, попробуйте позже.")
	}

	return c.SendStreamPicker(message.Chat.ID, streams)
}

func (c *Client) handlePickStream(cq *CallbackQuery) error {
	u, err := c.userService.GetOrCreate(context.Background(), cq.From.ID)
	if err != nil {
		return err
	}

	if u.GetCourseCookie == nil {
		return c.SendMessage(cq.Message.Chat.ID, "Сначала подключите GetCourse.")
	}

	client := getcourse.NewClient(c.getCourseBaseURL, *u.GetCourseCookie)

	streams, err := client.GetStreams()
	if err != nil {
		return c.SendMessage(cq.Message.Chat.ID, "Не удалось получить список потоков, попробуйте позже.")
	}

	return c.SendStreamPicker(cq.Message.Chat.ID, streams)
}