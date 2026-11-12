package storage

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type SessionStore struct {
	rdb *redis.Client
}

func NewSessionStore(rdb *redis.Client) *SessionStore {
	return &SessionStore{rdb: rdb}
}

func (s *SessionStore) SetAwaitingCookie(ctx context.Context, telegramID int64) error {
	return s.rdb.Set(ctx, sessionKey(telegramID), "awaiting_cookie", awaitingCookieTTL).Err()
}

func (s *SessionStore) IsAwaitingCookie(ctx context.Context, telegramID int64) (bool, error) {
	val, err := s.rdb.Get(ctx, sessionKey(telegramID)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == "awaiting_cookie", nil
}

func (s *SessionStore) ClearState(ctx context.Context, telegramID int64) error {
	return s.rdb.Del(ctx, sessionKey(telegramID)).Err()
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

func (s *SessionStore) SetStreamDraft(ctx context.Context, telegramID int64, data []byte) error {
	return s.rdb.Set(ctx, streamDraftKey(telegramID), data, streamDraftTTL).Err()
}

func (s *SessionStore) GetStreamDraft(ctx context.Context, telegramID int64) ([]byte, bool, error) {
	data, err := s.rdb.Get(ctx, streamDraftKey(telegramID)).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (s *SessionStore) DeleteStreamDraft(ctx context.Context, telegramID int64) error {
	return s.rdb.Del(ctx, streamDraftKey(telegramID)).Err()
}

func (s *SessionStore) GetMenuMessage(ctx context.Context, telegramID int64) (int, error) {
	val, err := s.rdb.Get(ctx, menuMessageKey(telegramID)).Int()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (s *SessionStore) SetMenuMessage(ctx context.Context, telegramID int64, messageID int) error {
	return s.rdb.Set(ctx, menuMessageKey(telegramID), messageID, 0).Err()
}

func (s *SessionStore) ClearMenuMessage(ctx context.Context, telegramID int64) error {
	return s.rdb.Del(ctx, menuMessageKey(telegramID)).Err()
}