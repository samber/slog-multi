package slogmulti

import (
	"context"
	"slices"

	"log/slog"

	slogcommon "github.com/samber/slog-common"
)

type router struct {
	handlers []slog.Handler
}

// Router forwards records to all matching slog.Handler.
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
				groups:   []string{},
				attrs:    []slog.Attr{},
			},
		),
	}
}

func (h *router) Handler() slog.Handler {
	return Fanout(h.handlers...)
}

var _ slog.Handler = (*RoutableHandler)(nil)

// @TODO: implement round robin strategy ?
type RoutableHandler struct {
	matchers []func(ctx context.Context, r slog.Record) bool
	handler  slog.Handler
	groups   []string
	attrs    []slog.Attr
}

// Implements slog.Handler
func (h *RoutableHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler.Enabled(ctx, l)
}

// Implements slog.Handler
func (h *RoutableHandler) Handle(ctx context.Context, r slog.Record) error {
	clone := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	clone.AddAttrs(
		slogcommon.AppendRecordAttrsToAttrs(h.attrs, h.groups, &r)...,
	)

	for _, matcher := range h.matchers {
		if !matcher(ctx, clone) {
			return nil
		}
	}

	return h.handler.Handle(ctx, r)
}

// Implements slog.Handler
func (h *RoutableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RoutableHandler{
		matchers: h.matchers,
		handler:  h.handler.WithAttrs(attrs),
		groups:   slices.Clone(h.groups),
		attrs:    slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
	}
}

// Implements slog.Handler
func (h *RoutableHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &RoutableHandler{
		matchers: h.matchers,
		handler:  h.handler.WithGroup(name),
		groups:   append(slices.Clone(h.groups), name),
		attrs:    h.attrs,
	}
}
