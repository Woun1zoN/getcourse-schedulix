package user

import "time"

type User struct {
	ID                int64
	TelegramID        int64
	GetCourseCookie   *string
	GetCourseStreamID *int64
	CreatedAt         time.Time
}