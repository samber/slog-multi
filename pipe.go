package slogmulti

import (
	"golang.org/x/exp/slog"
)

type PipeBuilder struct {
	middlewares []Middleware
}

func Pipe(middlewares ...Middleware) *PipeBuilder {
	return &PipeBuilder{middlewares: middlewares}
}

func (h *PipeBuilder) Pipe(middleware Middleware) *PipeBuilder {
	h.middlewares = append(h.middlewares, middleware)
	return h
}

func (h *PipeBuilder) Handler(handler slog.Handler) slog.Handler {
	for len(h.middlewares) > 0 {
		middleware := h.middlewares[len(h.middlewares)-1]
		h.middlewares = h.middlewares[0 : len(h.middlewares)-1]
		handler = middleware(handler)
	}

	return handler
}
