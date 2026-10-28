package slogmulti_test

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	slogmulti "github.com/samber/slog-multi"
)

func TestWithAttrsIsolatesChildHandlers(t *testing.T) {
	factories := map[string]func(...slog.Handler) slog.Handler{
		"fanout":   func(hs ...slog.Handler) slog.Handler { return slogmulti.Fanout(hs...) },
		"failover": slogmulti.Failover(),
		"pool":     slogmulti.Pool(),
	}
	for name, factory := range factories {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			mutating := slogmulti.NewWithAttrsInlineMiddleware(
				func(attrs []slog.Attr, next func([]slog.Attr) slog.Handler) slog.Handler {
					// WithAttrs transfers ownership, so a child may modify its slice.
					attrs[0] = slog.String("service", "primary")
					return next(attrs)
				},
			)(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError}))
			secondary := slog.NewTextHandler(&output, &slog.HandlerOptions{
				ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
					if attr.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return attr
				},
			})

			logger := slog.New(factory(mutating, secondary)).With("service", "api")
			logger.Info("hello")

			if got, want := output.String(), "level=INFO msg=hello service=api\n"; got != want {
				t.Fatalf("secondary output = %q, want %q", got, want)
			}
		})
	}
}
