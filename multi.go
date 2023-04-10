package slogmulti

import (
	"context"

	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &MultiHandler{
		handlers: handlers,
	}
}

func (h *MultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := h.handlers[i].Handle(ctx, r.Clone())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithAttrs(attrs)
	})
	return NewMultiHandler(handers...)
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithGroup(name)
	})
	return NewMultiHandler(handers...)
}
