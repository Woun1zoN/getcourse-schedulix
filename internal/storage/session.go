package storage

import (
    "context"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

type SessionStore struct {
    rdb *redis.Client
}

func NewSessionStore(rdb *redis.Client) *SessionStore {
    return &SessionStore{rdb: rdb}
}

const awaitingCookieTTL = 10 * time.Minute

func (s *SessionStore) SetAwaitingCookie(ctx context.Context, telegramID int64) error {
    return s.rdb.Set(ctx, key(telegramID), "awaiting_cookie", awaitingCookieTTL).Err()
}

func (s *SessionStore) IsAwaitingCookie(ctx context.Context, telegramID int64) (bool, error) {
    val, err := s.rdb.Get(ctx, key(telegramID)).Result()
    if err == redis.Nil {
        return false, nil
    }
    if err != nil {
        return false, err
    }
    return val == "awaiting_cookie", nil
}

func (s *SessionStore) ClearState(ctx context.Context, telegramID int64) error {
    return s.rdb.Del(ctx, key(telegramID)).Err()
}

func key(telegramID int64) string {
    return fmt.Sprintf("session:%d", telegramID)
}