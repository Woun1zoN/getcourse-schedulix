// internal/app/run_test.go
package app

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/Woun1zoN/getcourse-schedulix/internal/storage"
	"github.com/Woun1zoN/getcourse-schedulix/internal/user"
)

func TestHandleExpiredCookie(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := storage.NewSessionStore(rdb)
	ctx := context.Background()

	calls := 0
	var sendErr error

	r := &Runner{
		sessionStore: store,
		notify: func(int64) error {
			calls++
			return sendErr
		},
	}
	u := user.User{ID: 67, TelegramID: 67}

	r.handleExpiredCookie(ctx, u)
	if calls != 1 {
		t.Fatalf("want 1 notification, got %d", calls)
	}

	r.handleExpiredCookie(ctx, u)
	r.handleExpiredCookie(ctx, u)
	if calls != 1 {
		t.Fatalf("want still 1 notification, got %d", calls)
	}

	if err := store.ClearExpiredNotified(ctx, u.TelegramID); err != nil {
		t.Fatal(err)
	}
	r.handleExpiredCookie(ctx, u)
	if calls != 2 {
		t.Fatalf("want 2 notifications after reset, got %d", calls)
	}

	_ = store.ClearExpiredNotified(ctx, u.TelegramID)
	sendErr = errors.New("telegram down")
	r.handleExpiredCookie(ctx, u)
	sendErr = nil
	r.handleExpiredCookie(ctx, u)
	if calls != 4 {
		t.Fatalf("want 4 total calls (fail + retry), got %d", calls)
	}
}
