package slogmulti

import (
	"context"
	"log/slog"

	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
)

var _ slog.Handler = (*FanoutHandler)(nil)

type FanoutHandler struct {
	handlers []slog.Handler
}

// Fanout distributes records to multiple slog.Handler in parallel
func Fanout(handlers ...slog.Handler) slog.Handler {
	return &FanoutHandler{
		handlers: handlers,
	}
}

// Implements slog.Handler
func (h *FanoutHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// Implements slog.Handler
func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var result error
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := try(func() error {
				return h.handlers[i].Handle(ctx, r.Clone())
			})
			if err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result
}

// Implements slog.Handler
func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithAttrs(attrs)
	})
	return Fanout(handers...)
}

// Implements slog.Handler
func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithGroup(name)
	})
	return Fanout(handers...)
}
