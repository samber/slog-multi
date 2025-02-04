package slogmulti

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverHandlerError_ok(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	errored := false

	recover := RecoverHandlerError(
		func(ctx context.Context, record slog.Record, err error) {
			errored = true
			is.Equal(assert.AnError.Error(), err.Error())
		},
	)

	is.False(errored)
	recover(slog.NewJSONHandler(io.Discard, nil)).Handle(context.Background(), slog.Record{})
	is.False(errored)
}

func TestRecoverHandlerError_error(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	errored := false

	recover := RecoverHandlerError(
		func(ctx context.Context, record slog.Record, err error) {
			errored = true
			is.Equal(assert.AnError.Error(), err.Error())
		},
	)
	handler := NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
		return assert.AnError
	})

	is.False(errored)
	recover(handler(&slog.JSONHandler{})).Handle(context.Background(), slog.Record{})
	is.True(errored)
}

func TestRecoverHandlerError_panicError(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	errored := false

	recover := RecoverHandlerError(
		func(ctx context.Context, record slog.Record, err error) {
			errored = true
			is.Equal(assert.AnError.Error(), err.Error())
		},
	)
	handler := NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
		panic(assert.AnError)
	})

	is.False(errored)
	recover(handler(&slog.JSONHandler{})).Handle(context.Background(), slog.Record{})
	is.True(errored)
}

func TestRecoverHandlerError_panicAny(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	errored := false

	recover := RecoverHandlerError(
		func(ctx context.Context, record slog.Record, err error) {
			errored = true
			is.Equal(assert.AnError.Error(), err.Error())
		},
	)
	handler := NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
		panic(assert.AnError.Error())
	})

	is.False(errored)
	recover(handler(&slog.JSONHandler{})).Handle(context.Background(), slog.Record{})
	is.True(errored)
}
