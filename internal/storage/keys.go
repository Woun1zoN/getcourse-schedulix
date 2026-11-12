package storage

import (
	"fmt"
	"time"
)

const (
    streamDraftTTL   = 30 * time.Minute
    awaitingCookieTTL = 10 * time.Minute
)

func baselineField(streamID int64) string {
	return fmt.Sprintf("baseline:%d", streamID)
}

func HashesKey(userID int64) string {
    return fmt.Sprintf("docs:hashes:%d", userID)
}

func cookiePromptKey(telegramID int64) string {
    return fmt.Sprintf("cookie_prompt:%d", telegramID)
}

func expiredKey(telegramID int64) string {
    return fmt.Sprintf("cookie_expired_notified:%d", telegramID)
}

func sessionKey(telegramID int64) string {
    return fmt.Sprintf("session:%d", telegramID)
}

func staleStreamKey(telegramID, streamID int64) string {
    return fmt.Sprintf("stream_missing_notified:%d:%d", telegramID, streamID)
}

func streamDraftKey(telegramID int64) string {
    return fmt.Sprintf("stream_draft:%d", telegramID)
}

func menuMessageKey(telegramID int64) string {
    return fmt.Sprintf("menu_message:%d", telegramID)
}