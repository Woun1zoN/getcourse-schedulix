package user

import (
	"context"
	"fmt"

	"github.com/Woun1zoN/getcourse-schedulix/internal/getcourse"
	"github.com/Woun1zoN/getcourse-schedulix/internal/crypto"
)

type Service struct {
	repository 	    *Repository
	getCourseClient *getcourse.Client
	cryptor         *crypto.Cryptor
}

func NewService(repository *Repository, getCourseClient *getcourse.Client, cryptor *crypto.Cryptor) *Service {
	return &Service{
		repository: repository,
		getCourseClient: getCourseClient,
		cryptor: cryptor,
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

	encrypted, err := s.cryptor.Encrypt(cookie)
    if err != nil {
        return fmt.Errorf("encrypt getcourse cookie: %w", err)
    }

	if err := s.repository.UpdateCookie(ctx, telegramID, encrypted); err != nil {
		return fmt.Errorf("update getcourse cookie: %w", err)
	}

	return nil
}