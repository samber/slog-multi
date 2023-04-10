package slogmulti

import (
	"context"

	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

// @TODO: implement round robin strategy ?
type EitherHandler struct {
	handlers []slog.Handler
}

func Either(handlers ...slog.Handler) slog.Handler {
	return &EitherHandler{
		handlers: handlers,
	}
}

func (h *EitherHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

func (h *EitherHandler) Handle(ctx context.Context, r slog.Record) error {
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

func (h *EitherHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithAttrs(attrs)
	})
	return Either(handers...)
}

func (h *EitherHandler) WithGroup(name string) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithGroup(name)
	})
	return Either(handers...)
}
