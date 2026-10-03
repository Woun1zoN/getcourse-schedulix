package telegram

import (
	"strings"
)

var mdV2Replacer = strings.NewReplacer(
	`\`, `\\`, "_", `\_`, "*", `\*`, "[", `\[`, "]", `\]`, "(", `\(`, ")", `\)`,
	"~", `\~`, "`", "\\`", ">", `\>`, "#", `\#`, "+", `\+`, "-", `\-`,
	"=", `\=`, "|", `\|`, "{", `\{`, "}", `\}`, ".", `\.`, "!", `\!`,
)

func escapeMarkdownV2(s string) string {
	return mdV2Replacer.Replace(s)
}

var welcomeText = strings.NewReplacer("{pad}", strings.Repeat("\u2800", 10)).Replace(`_Привет\! Добро пожаловать в GetCourse Schedulix\! 👋_

>Schedulix — бесплатный Telegram\-бот, который поможет следить за тренингами в GetCourse и получать новые материалы прямо в Telegram\.

Чтобы начать работу, нажмите кнопку ниже и следуйте инструкциям\. Вам необходимо:

>1\. Подключить аккаунт GetCourse
>2\. Выбрать тренинг, за которым нужно следить{pad}
>3\. Выбрать канал, куда отправлять материалы
>4\. Пройти остальные этапы настройки

❗ Ваши данные используются только для работы с GetCourse и не передаются другим пользователям\.`)

var connectGetCourseMessage = strings.NewReplacer("{pad}", strings.Repeat("\u2800", 25)).Replace(`_🔗 Подключение GetCourse_

Установите Cookie\-Editor для вашего браузера:
• [Google Chrome](https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)
• [Яндекс Браузер](https://chromewebstore.google.com/detail/cookie-editor/hlkenndednhfkekhgcdicdfddnkalmdm)
• [Mozilla Firefox](https://addons.mozilla.org/en-US/firefox/addon/cookie-editor/)

После установки:
>1\. Зайдите на *сайт вашей онлайн\-школы \(GetCourse\)* под своим аккаунтом\.
>2\. Кликните по иконке ` + "`Cookie-Editor`" + ` на этой же вкладке\.
>3\. Нажмите ` + "`Export`" + ` → выберите формат ` + "«`Header String`»" + `\.
>4\. Вставьте скопированную строку сюда одним сообщением\.

Не получается с расширением? Инструкция через DevTools:

>1\. Откройте *сайт вашей онлайн\-школы* и войдите в аккаунт\.
>2\. Нажмите F12 → вкладка ` + "`Network`" + `\.
>3\. Обновите страницу \(F5\) и кликните на любой запрос слева\.
>4\. Во вкладке Headers найдите ` + "«`cookie:`»" + ` среди ` + "`Request Headers`" + `\.
>5\. Скопируйте значение целиком, без слова ` + "«`cookie:`»" + `, и пришлите сюда\.

❗ Эта строка даёт полный доступ к вашему аккаунту GetCourse\. Никому её не передавайте, кроме этого бота\.`)