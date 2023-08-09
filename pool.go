package slogmulti

import (
	"context"
	"math/rand"
	"time"

	"log/slog"

	"github.com/samber/lo"
)

type PoolHandler struct {
	randSource rand.Source
	handlers   []slog.Handler
}

// Pool balances records between multiple slog.Handler in order to increase bandwidth.
// Uses a round robin strategy.
func Pool() func(...slog.Handler) slog.Handler {
	return func(handlers ...slog.Handler) slog.Handler {
		return &PoolHandler{
			randSource: rand.NewSource(time.Now().UnixNano()),
			handlers:   handlers,
		}
	}
}

// Implements slog.Handler
func (h *PoolHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

// Implements slog.Handler
func (h *PoolHandler) Handle(ctx context.Context, r slog.Record) error {
	// round robin
	rand := h.randSource.Int63() % int64(len(h.handlers))
	handlers := append(h.handlers[rand:], h.handlers[:rand]...)

	var err error

	for i := range handlers {
		if handlers[i].Enabled(ctx, r.Level) {
			err = try(func() error {
				return handlers[i].Handle(ctx, r.Clone())
			})
			if err == nil {
				return nil
			}
		}
	}

	return err
}

// Implements slog.Handler
func (h *PoolHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithAttrs(attrs)
	})
	return Pool()(handers...)
}

// Implements slog.Handler
func (h *PoolHandler) WithGroup(name string) slog.Handler {
	handers := lo.Map(h.handlers, func(h slog.Handler, _ int) slog.Handler {
		return h.WithGroup(name)
	})
	return Pool()(handers...)
}
