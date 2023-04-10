package slogmulti

import (
	"context"

	"golang.org/x/exp/slog"
)

// Shortcut to a middleware that implements only the `WithAttrs` method.
func NewWithGroupInlineMiddleware(withGroupFunc func(name string, next func(string) slog.Handler) slog.Handler) Middleware {
	return func(next slog.Handler) slog.Handler {
		return &WithGroupInlineMiddleware{
			next:          next,
			withGroupFunc: withGroupFunc,
		}
	}
}

type WithGroupInlineMiddleware struct {
	next          slog.Handler
	withGroupFunc func(name string, next func(string) slog.Handler) slog.Handler
}

func (h *WithGroupInlineMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *WithGroupInlineMiddleware) Handle(ctx context.Context, record slog.Record) error {
	return h.next.Handle(ctx, record)
}

func (h *WithGroupInlineMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewWithGroupInlineMiddleware(h.withGroupFunc)(h.next.WithAttrs(attrs))
}

func (h *WithGroupInlineMiddleware) WithGroup(name string) slog.Handler {
	return NewWithGroupInlineMiddleware(h.withGroupFunc)(h.withGroupFunc(name, h.next.WithGroup))
}
