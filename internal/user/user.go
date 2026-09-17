package user

import "time"

type User struct {
	ID              int64
	TelegramID      int64
	GetCourseCookie *string
	CreatedAt       time.Time
}