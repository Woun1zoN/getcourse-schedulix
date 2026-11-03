package user

import "time"

type User struct {
	ID                 int64
	TelegramID         int64
	GetCourseCookie    *string
	GetCourseStreamIDs []int64
	TargetChatID       *int64
	TargetChatTitle    *string
	CreatedAt          time.Time
}
