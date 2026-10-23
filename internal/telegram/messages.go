package telegram

import (
	"strings"

	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

const (
	ParseModeMarkdown   = "Markdown"
	ParseModeMarkdownV2 = "MarkdownV2"
)

var mdV2Replacer = strings.NewReplacer(
	`\`, `\\`, "_", `\_`, "*", `\*`, "[", `\[`, "]", `\]`, "(", `\(`, ")", `\)`,
	"~", `\~`, "`", "\\`", ">", `\>`, "#", `\#`, "+", `\+`, "-", `\-`,
	"=", `\=`, "|", `\|`, "{", `\{`, "}", `\}`, ".", `\.`, "!", `\!`,
)

func escapeMarkdownV2(s string) string {
	return mdV2Replacer.Replace(s)
}

var welcomeText = strings.NewReplacer("{pad}", strings.Repeat("\u2800", 13)).Replace(`_👋 Добро пожаловать в GetCourse Schedulix\!_


>*Schedulix* — бесплатный Telegram\-бот, который помогает отслеживать тренинги в GetCourse и получать новые материалы прямо в Telegram\.

🛠 Для начала работы нужно:

>1\. Подключить *аккаунт GetCourse*\.
>2\. Выбрать *тренинг* для отслеживания\.
>3\. Выбрать *канал* для *отправки материалов*\.{pad}

❗ Ваши данные используются только для работы с GetCourse и не передаются другим пользователям\.

ℹ️ /help — Список доступных команд`)

var connectGetCourseMessage = strings.NewReplacer("{pad}", strings.Repeat("\u2800", 25)).Replace(`_🔗 Подключение GetCourse_

Установите Cookie\-Editor для вашего браузера:
• [Google Chrome](https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)
• [Яндекс Браузер](https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)
• [Mozilla Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)

После установки:
>1\. Зайдите на *сайт вашей онлайн\-школы \(GetCourse\)* под своим аккаунтом\.
>2\. Кликните по иконке ` + "`Cookie-Editor`" + ` на этой же вкладке\.
>3\. Нажмите ` + "`Export`" + ` → выберите формат ` + "«`Header String`»" + `\.
>4\. Вставьте скопированную строку сюда *одним сообщением*\.

Не получается с расширением? Попробуйте вручную через DevTools:

>1\. Откройте *сайт вашей онлайн\-школы* и войдите в аккаунт\.
>2\. Нажмите F12 → вкладка ` + "`Network`" + `\.
>3\. Обновите страницу \(F5\) и кликните на любой запрос слева\.
>4\. Во вкладке Headers найдите ` + "«`cookie:`»" + ` среди ` + "`Request Headers`" + `\.
>5\. Скопируйте значение целиком, и *вставьте сюда одним сообщением*\.

❗ Эта строка даёт полный доступ к вашему аккаунту GetCourse\. Никому её не передавайте, кроме этого бота\.`)

var helpText = `*📖 Доступные команды*

*Меню*
/start — Открыть приветственное меню
/help — Cписок доступных команд
/menu — Открыть главное меню

*Опции*
/delete — Удалить все ваши данные`

func startCommandText(u *user.User) string {
	var status string

	switch {
	case u.GetCourseCookie == nil:
		status = "❗ Статус: GetCourse не подключён\\."
	case u.GetCourseStreamID == nil && u.TargetChatID == nil:
		status = "❗ Статус: тренинг и чат не выбраны, материалы приходить не будут\\."
	case u.GetCourseStreamID == nil:
		status = "❗ Статус: тренинг не выбран, материалы приходить не будут\\."
	case u.TargetChatID == nil:
		status = "❗ Статус: чат не выбран, материалы приходить не будут\\."
	default:
		status = "✅ Статус: GetCourse подключён, новые материалы будут приходить автоматически\\."
	}

	return `_👋 Добро пожаловать в GetCourse Schedulix\!_


>*Schedulix* — бесплатный Telegram\-бот, который помогает отслеживать тренинги в GetCourse и получать новые материалы прямо в Telegram\.

` + status + `

ℹ️ /help — Список доступных команд`
}