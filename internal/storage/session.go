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

func (s *SessionStore) DeleteHashes(ctx context.Context, userID int64) error {
	return s.rdb.Del(ctx, HashesKey(userID)).Err()
}

func (s *SessionStore) MarkExpiredNotified(ctx context.Context, telegramID int64) (bool, error) {
	return s.rdb.SetNX(ctx, expiredKey(telegramID), "1", 0).Result()
}

func (s *SessionStore) ClearExpiredNotified(ctx context.Context, telegramID int64) error {
	return s.rdb.Del(ctx, expiredKey(telegramID)).Err()
}

func (s *SessionStore) SetCookiePromptMessage(ctx context.Context, telegramID int64, messageID int) error {
	return s.rdb.Set(ctx, cookiePromptKey(telegramID), messageID, awaitingCookieTTL).Err()
}

func (s *SessionStore) PopCookiePromptMessage(ctx context.Context, telegramID int64) (int, error) {
	val, err := s.rdb.GetDel(ctx, cookiePromptKey(telegramID)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (s *SessionStore) MarkStreamMissingNotified(ctx context.Context, telegramID, streamID int64) (bool, error) {
	return s.rdb.SetNX(ctx, staleStreamKey(telegramID, streamID), "1", 0).Result()
}

func (s *SessionStore) ClearStreamMissingNotified(ctx context.Context, telegramID, streamID int64) error {
	return s.rdb.Del(ctx, staleStreamKey(telegramID, streamID)).Err()
}

func cookiePromptKey(telegramID int64) string {
	return fmt.Sprintf("cookie_prompt:%d", telegramID)
}

func expiredKey(telegramID int64) string {
	return fmt.Sprintf("cookie_expired_notified:%d", telegramID)
}

func HashesKey(userID int64) string {
	return fmt.Sprintf("docs:hashes:%d", userID)
}

func key(telegramID int64) string {
	return fmt.Sprintf("session:%d", telegramID)
}

func staleStreamKey(telegramID, streamID int64) string {
	return fmt.Sprintf("stream_missing_notified:%d:%d", telegramID, streamID)
}
