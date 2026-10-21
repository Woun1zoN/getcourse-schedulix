package telegram

type ChatPickerMode int

const (
    ChatPickerInitial ChatPickerMode = iota
    ChatPickerChange
)

const (
	chatRequestGroup = 1
	chatRequestChannel = 2
)

func (c *Client) SendChatPicker(chatID int64, mode ChatPickerMode) error {
    skipButtonText := "➡ Пропустить"

    if mode == ChatPickerChange {
        skipButtonText = "➖ Без чата"
    }

    kb := &ReplyKeyboardMarkup{
	    ResizeKeyboard:  true,
	    OneTimeKeyboard: true,
	    Keyboard: [][]KeyboardButton{
	    	{
	    		{
	    			Text: "💬 Личные сообщения",
	    		},
	    	},
	    	{
	    		{
	    			Text: "👥 Выбрать группу",
	    			RequestChat: &KeyboardButtonRequestChat{
	    				RequestID:     chatRequestGroup,
	    				ChatIsChannel: false,
	    				BotIsMember:   true,
                        RequestTitle:  true,
	    				UserAdministratorRights: &ChatAdministratorRights{
	    					CanManageChat: true,
		    			},
		    		},
		    	},
		    },
		    {
		    	{
		    		Text: "📢 Выбрать канал",
		    		RequestChat: &KeyboardButtonRequestChat{
		    			RequestID:     chatRequestChannel,
		    			ChatIsChannel: true,
		    			BotIsMember:   true,
		    			UserAdministratorRights: &ChatAdministratorRights{
		    				CanManageChat: true,
                            CanPostMessages: true,
		    			},
		    			BotAdministratorRights: &ChatAdministratorRights{
		    				CanPostMessages: true,
		    			},
		    		},
		    	},
		    },
            {
                {
                    Text: skipButtonText,
                },
            },
	    },
    }

    return c.SendMessageWithKeyboardMode(
        chatID,
        "🔗 Выберите чат, куда Schedulix будет отправлять новые материалы.",
        "",
        kb,
    )
}