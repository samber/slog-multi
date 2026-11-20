package slogmulti

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeBuilder_HandlerCanBeCalledMoreThanOnce(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	addTag := NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
		record.AddAttrs(slog.String("tag", "piped"))
		return next(ctx, record)
	})

	builder := Pipe(addTag)

	var first, second bytes.Buffer
	h1 := builder.Handler(slog.NewTextHandler(&first, &slog.HandlerOptions{ReplaceAttr: dropTime}))
	h2 := builder.Handler(slog.NewTextHandler(&second, &slog.HandlerOptions{ReplaceAttr: dropTime}))

	slog.New(h1).Info("hello")
	slog.New(h2).Info("hello")

	is.Equal("level=INFO msg=hello tag=piped\n", first.String())
	is.Equal("level=INFO msg=hello tag=piped\n", second.String())
}

func TestPipe_DoesNotShareCallerSlice(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	noop := func(next slog.Handler) slog.Handler { return next }
	mws := make([]Middleware, 1, 4)
	mws[0] = noop

	builder := Pipe(mws...).Pipe(noop)

	is.Len(builder.middlewares, 2)
	is.Nil(mws[:2][1])
}

func dropTime(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 && a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}
