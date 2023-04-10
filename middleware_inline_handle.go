package slogmulti

import (
	"context"

	"golang.org/x/exp/slog"
)

// Shortcut to a middleware that implements only the `Handle` method.
func NewHandleInlineMiddleware(handleFunc func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error) Middleware {
	return func(next slog.Handler) slog.Handler {
		return &HandleInlineMiddleware{
			next:       next,
			handleFunc: handleFunc,
		}
	}
}

type HandleInlineMiddleware struct {
	next       slog.Handler
	handleFunc func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error
}

func (h *HandleInlineMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *HandleInlineMiddleware) Handle(ctx context.Context, record slog.Record) error {
	return h.handleFunc(ctx, record, h.next.Handle)
}

func (h *HandleInlineMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewHandleInlineMiddleware(h.handleFunc)(h.next.WithAttrs(attrs))
}

func (h *HandleInlineMiddleware) WithGroup(name string) slog.Handler {
	return NewHandleInlineMiddleware(h.handleFunc)(h.next.WithGroup(name))
}
