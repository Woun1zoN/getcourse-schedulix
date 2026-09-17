package user

import "context"

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{
		repository: repository,
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