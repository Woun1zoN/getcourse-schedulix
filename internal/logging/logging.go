package logging

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func NewLogger() *slog.Logger {
	return slog.New(&pipeHandler{out: os.Stdout})
}

var acronyms = map[string]string{
	"id":  "ID",
	"url": "URL",
	"api": "API",
}

func labelFor(key string) string {
	words := strings.Split(key, "_")
	for i, w := range words {
		if up, ok := acronyms[w]; ok {
			words[i] = up
		} else if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

type pipeHandler struct {
	out   *os.File
	attrs []slog.Attr
}

func (h *pipeHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *pipeHandler) Handle(_ context.Context, r slog.Record) error {
	parts := []string{r.Level.String(), capitalize(r.Message)}

	r.Attrs(func(a slog.Attr) bool {
		parts = append(parts, fmt.Sprintf("%s: %v", labelFor(a.Key), a.Value))
		return true
	})
	for _, a := range h.attrs {
		parts = append(parts, fmt.Sprintf("%s: %v", labelFor(a.Key), a.Value))
	}

	_, err := fmt.Fprintf(h.out, "%s | %s\n",
		r.Time.Format("2006-01-02 15:04:05"),
		strings.Join(parts, " | "),
	)
	return err
}

func (h *pipeHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &pipeHandler{out: h.out, attrs: append(h.attrs, attrs...)}
}

func (h *pipeHandler) WithGroup(_ string) slog.Handler {
	return h
}