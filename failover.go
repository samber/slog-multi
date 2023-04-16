package slogmulti

import (
	"context"

	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

// @TODO: implement round robin strategy ?
type FailoverHandler struct {
	handlers []slog.Handler
}

// Failover forward record to the first available slog.Handler
func Failover() func(...slog.Handler) slog.Handler {
	return func(handlers ...slog.Handler) slog.Handler {
		return &FailoverHandler{
			handlers: handlers,
		}
	}
}

// Implements slog.Handler
func (h *FailoverHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// Implements slog.Handler
func (h *FailoverHandler) Handle(ctx context.Context, r slog.Record) error {
	var err error

	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err = try(func() error {
				return h.handlers[i].Handle(ctx, r.Clone())
			})
			if err == nil {
				return nil
			}
		}
	}

	return err
}

// Implements slog.Handler
func (h *FailoverHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithAttrs(attrs)
	})
	return Failover()(handers...)
}

// Implements slog.Handler
func (h *FailoverHandler) WithGroup(name string) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithGroup(name)
	})
	return Failover()(handers...)
}
