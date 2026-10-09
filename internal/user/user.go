package user

import "time"

type User struct {
	ID                int64
	TelegramID        int64
	GetCourseCookie   *string
	GetCourseStreamID *int64
	TargetChatID      *int64
	TargetChatTitle   *string
	CreatedAt         time.Time
}

func (u *User) TargetChat() int64 {
	if u.TargetChatID != nil {
		return *u.TargetChatID
	}
	return u.TelegramID
}