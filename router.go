package slogmulti

import (
	"context"

	"golang.org/x/exp/slog"
)

type router struct {
	handlers []slog.Handler
}

// Router forward record to all matching slog.Handler.
func Router() *router {
	return &router{
		handlers: []slog.Handler{},
	}
}

// Add a new handler to the router. The handler will be called if all matchers return true.
func (h *router) Add(handler slog.Handler, matchers ...func(ctx context.Context, r slog.Record) bool) *router {
	return &router{
		handlers: append(
			h.handlers,
			&RoutableHandler{
				matchers: matchers,
				handler:  handler,
			},
		),
	}
}

func (h *router) Handler() slog.Handler {
	return Fanout(h.handlers...)
}

// @TODO: implement round robin strategy ?
type RoutableHandler struct {
	matchers []func(ctx context.Context, r slog.Record) bool
	handler  slog.Handler
}

// Implements slog.Handler
func (h *RoutableHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return true
}

// Implements slog.Handler
func (h *RoutableHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, matcher := range h.matchers {
		if !matcher(ctx, r) {
			return nil
		}
	}

	return h.Handle(ctx, r)
}

// Implements slog.Handler
func (h *RoutableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RoutableHandler{
		matchers: h.matchers,
		handler:  h.handler.WithAttrs(attrs),
	}
}

// Implements slog.Handler
func (h *RoutableHandler) WithGroup(name string) slog.Handler {
	return &RoutableHandler{
		matchers: h.matchers,
		handler:  h.handler.WithGroup(name),
	}
}
