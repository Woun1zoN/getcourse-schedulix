package telegram

type Update struct {
    UpdateID      int            `json:"update_id"`
    Message       *Message       `json:"message"`
    CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
    From       *User        `json:"from"`
    Chat       Chat         `json:"chat"`
    Text       string       `json:"text"`
    ChatShared *ChatShared  `json:"chat_shared,omitempty"`
}

type CallbackQuery struct {
    ID      string  `json:"id"`
    From    User    `json:"from"`
    Message Message `json:"message"`
    Data    string  `json:"data"`
}

type Chat struct {
    ID   int64  `json:"id"`
	Type string `json:"type"`
}

type ChatShared struct {
    RequestID int64  `json:"request_id"`
    ChatID    int64  `json:"chat_id"`
    Title     string `json:"title,omitempty"`
    Username  string `json:"username,omitempty"`
}

type ReplyKeyboardMarkup struct {
    Keyboard        [][]KeyboardButton `json:"keyboard"`
    ResizeKeyboard  bool               `json:"resize_keyboard,omitempty"`
    OneTimeKeyboard bool               `json:"one_time_keyboard,omitempty"`
}

type ReplyKeyboardRemove struct {
    RemoveKeyboard bool `json:"remove_keyboard"`
}

type KeyboardButton struct {
    Text        string                     `json:"text"`
    RequestChat *KeyboardButtonRequestChat `json:"request_chat,omitempty"`
}

type KeyboardButtonRequestChat struct {
    RequestID               int                  	 `json:"request_id"`
    ChatIsChannel           bool                 	 `json:"chat_is_channel"`
    UserAdministratorRights *ChatAdministratorRights `json:"user_administrator_rights,omitempty"`
    BotAdministratorRights  *ChatAdministratorRights `json:"bot_administrator_rights,omitempty"`
    BotIsMember             bool                 	 `json:"bot_is_member,omitempty"`
}

type ChatAdministratorRights struct {
    CanManageChat      bool `json:"can_manage_chat,omitempty"`
    CanChangeInfo      bool `json:"can_change_info,omitempty"`
    CanPostMessages    bool `json:"can_post_messages,omitempty"`
    CanEditMessages    bool `json:"can_edit_messages,omitempty"`
    CanDeleteMessages  bool `json:"can_delete_messages,omitempty"`
    CanInviteUsers     bool `json:"can_invite_users,omitempty"`
    CanRestrictMembers bool `json:"can_restrict_members,omitempty"`
    CanPinMessages     bool `json:"can_pin_messages,omitempty"`
    CanManageTopics    bool `json:"can_manage_topics,omitempty"`
}

type SendMessagePayload struct {
	ChatID             int64                  `json:"chat_id"`
	Text               string                 `json:"text"`
	ParseMode          string                 `json:"parse_mode"`
	ReplyMarkup        any                    `json:"reply_markup,omitempty"`
	LinkPreviewOptions *linkPreviewOptions    `json:"link_preview_options,omitempty"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}