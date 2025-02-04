package slogmulti

import (
	"context"
	"fmt"
	"log/slog"
)

type RecoveryFunc func(ctx context.Context, record slog.Record, err error)

var _ slog.Handler = (*HandlerErrorRecovery)(nil)

type HandlerErrorRecovery struct {
	recovery RecoveryFunc
	handler  slog.Handler
}

// RecoverHandlerError returns a slog.Handler that recovers from panics or error of the chain of handlers.
func RecoverHandlerError(recovery RecoveryFunc) func(slog.Handler) slog.Handler {
	return func(handler slog.Handler) slog.Handler {
		return &HandlerErrorRecovery{
			recovery: recovery,
			handler:  handler,
		}
	}
}

// Enabled implements slog.Handler.
func (h *HandlerErrorRecovery) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler.Enabled(ctx, l)
}

// Handle implements slog.Handler.
func (h *HandlerErrorRecovery) Handle(ctx context.Context, record slog.Record) error {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				h.recovery(ctx, record, e)
			} else {
				h.recovery(ctx, record, fmt.Errorf("%+v", r))
			}
		}
	}()

	err := h.handler.Handle(ctx, record)
	if err != nil {
		h.recovery(ctx, record, err)
	}

	// propagate error
	return err
}

// WithAttrs implements slog.Handler.
func (h *HandlerErrorRecovery) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &HandlerErrorRecovery{
		recovery: h.recovery,
		handler:  h.handler.WithAttrs(attrs),
	}
}

// WithGroup implements slog.Handler.
func (h *HandlerErrorRecovery) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &HandlerErrorRecovery{
		recovery: h.recovery,
		handler:  h.handler.WithGroup(name),
	}
}
