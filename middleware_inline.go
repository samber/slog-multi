package slogmulti

import (
	"context"

	"golang.org/x/exp/slog"
)

// Shortcut to a middleware that implements all methods.
func NewInlineMiddleware(
	enabledFunc func(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool,
	handleFunc func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error,
	withAttrsFunc func(attrs []slog.Attr, next func([]slog.Attr) slog.Handler) slog.Handler,
	withGroupFunc func(name string, next func(string) slog.Handler) slog.Handler,
) Middleware {
	return func(next slog.Handler) slog.Handler {
		return &InlineMiddleware{
			next:          next,
			enabledFunc:   enabledFunc,
			handleFunc:    handleFunc,
			withAttrsFunc: withAttrsFunc,
			withGroupFunc: withGroupFunc,
		}
	}
}

type InlineMiddleware struct {
	next          slog.Handler
	enabledFunc   func(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool
	handleFunc    func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error
	withAttrsFunc func(attrs []slog.Attr, next func([]slog.Attr) slog.Handler) slog.Handler
	withGroupFunc func(name string, next func(string) slog.Handler) slog.Handler
}

func (h *InlineMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return h.enabledFunc(ctx, level, h.next.Enabled)
}

func (h *InlineMiddleware) Handle(ctx context.Context, record slog.Record) error {
	return h.handleFunc(ctx, record, h.next.Handle)
}

func (h *InlineMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewInlineMiddleware(
		h.enabledFunc,
		h.handleFunc,
		h.withAttrsFunc,
		h.withGroupFunc,
	)(h.withAttrsFunc(attrs, h.next.WithAttrs))
}

func (h *InlineMiddleware) WithGroup(name string) slog.Handler {
	return NewInlineMiddleware(
		h.enabledFunc,
		h.handleFunc,
		h.withAttrsFunc,
		h.withGroupFunc,
	)(h.withGroupFunc(name, h.next.WithGroup))
}
