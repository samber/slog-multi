package slogmulti

import (
	"context"

	"golang.org/x/exp/slog"
)

// NewEnabledInlineMiddleware is shortcut to a middleware that implements only the `Enable` method.
func NewEnabledInlineMiddleware(enabledFunc func(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool) Middleware {
	return func(next slog.Handler) slog.Handler {
		return &EnabledInlineMiddleware{
			next:        next,
			enabledFunc: enabledFunc,
		}
	}
}

type EnabledInlineMiddleware struct {
	next slog.Handler
	// enableFunc func(context.Context, slog.Level) bool
	enabledFunc func(context.Context, slog.Level, func(context.Context, slog.Level) bool) bool
}

// Implements slog.Handler
func (h *EnabledInlineMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.enabledFunc(ctx, level, h.next.Enabled)
}

// Implements slog.Handler
func (h *EnabledInlineMiddleware) Handle(ctx context.Context, record slog.Record) error {
	return h.next.Handle(ctx, record)
}

// Implements slog.Handler
func (h *EnabledInlineMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewEnabledInlineMiddleware(h.enabledFunc)(h.next.WithAttrs(attrs))
}

// Implements slog.Handler
func (h *EnabledInlineMiddleware) WithGroup(name string) slog.Handler {
	return NewEnabledInlineMiddleware(h.enabledFunc)(h.next.WithGroup(name))
}
