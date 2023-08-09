package slogmulti

import (
	"log/slog"
)

// Pipe defines a chain of Middleware.
type PipeBuilder struct {
	middlewares []Middleware
}

// Pipe builds a chain of Middleware.
// Eg: rewrite log.Record on the fly for privacy reason.
func Pipe(middlewares ...Middleware) *PipeBuilder {
	return &PipeBuilder{middlewares: middlewares}
}

// Implements slog.Handler
func (h *PipeBuilder) Pipe(middleware Middleware) *PipeBuilder {
	h.middlewares = append(h.middlewares, middleware)
	return h
}

// Implements slog.Handler
func (h *PipeBuilder) Handler(handler slog.Handler) slog.Handler {
	for len(h.middlewares) > 0 {
		middleware := h.middlewares[len(h.middlewares)-1]
		h.middlewares = h.middlewares[0 : len(h.middlewares)-1]
		handler = middleware(handler)
	}

	return handler
}
