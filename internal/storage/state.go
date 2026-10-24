package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisState struct {
	rdb *redis.Client
	key string
}

func NewRedisState(rdb *redis.Client, key string) *RedisState {
	return &RedisState{rdb: rdb, key: key}
}

func Hash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func (s *RedisState) HasChanged(ctx context.Context, url string, data []byte) (bool, error) {
	hash := Hash(data)

	oldHash, err := s.rdb.HGet(ctx, s.key, url).Result()
	if err == redis.Nil {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return oldHash != hash, nil
}

func (s *RedisState) MarkProcessed(ctx context.Context, url string, data []byte) error {
	return s.rdb.HSet(ctx, s.key, url, Hash(data)).Err()
}

func baselineField(streamID int64) string {
	return fmt.Sprintf("baseline:%d", streamID)
}

func (s *RedisState) IsBaselined(ctx context.Context, streamID int64) (bool, error) {
	return s.rdb.HExists(ctx, s.key, baselineField(streamID)).Result()
}

func (s *RedisState) MarkBaselined(ctx context.Context, streamID int64) error {
	return s.rdb.HSet(ctx, s.key, baselineField(streamID), "1").Err()
}
