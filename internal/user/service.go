package user

import (
	"context"
	"fmt"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
)

type Service struct {
	repository 	    *Repository
	getCourseClient *getcourse.Client
}

func NewService(repository *Repository, getCourseClient *getcourse.Client) *Service {
	return &Service{
		repository: repository,
		getCourseClient: getCourseClient,
	}
}

func (s *Service) GetOrCreate(ctx context.Context, telegramID int64) (*User, error) {
	user, err := s.repository.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	if user != nil {
		return user, nil
	}

	return s.repository.Create(ctx, telegramID)
}

func (s *Service) ConnectGetCourse(ctx context.Context, telegramID int64, cookie string,) error {
	client := s.getCourseClient.WithCookies(cookie)

	if err := client.ValidateCookie(); err != nil {
		return fmt.Errorf("validate getcourse cookie: %w", err)
	}

	if err := s.repository.UpdateCookie(ctx, telegramID, cookie); err != nil {
		return fmt.Errorf("update getcourse cookie: %w", err)
	}

	return nil
}