package slogmulti

import (
	"io"
	"log/slog"
	"testing"
)

func TestFanoutIncrementalBuildFlattensHandlers(t *testing.T) {
	t.Parallel()

	h1 := slog.NewJSONHandler(io.Discard, nil)
	h2 := slog.NewJSONHandler(io.Discard, nil)
	h3 := slog.NewJSONHandler(io.Discard, nil)

	var handler slog.Handler = h1
	handler = Fanout(handler, h2)
	handler = Fanout(handler, h3)

	fanout, ok := handler.(*FanoutHandler)
	if !ok {
		t.Fatalf("expected FanoutHandler, got %T", handler)
	}
	if len(fanout.handlers) != 3 {
		t.Fatalf("expected 3 handlers, got %d", len(fanout.handlers))
	}
	if fanout.handlers[0] != h1 || fanout.handlers[1] != h2 || fanout.handlers[2] != h3 {
		t.Fatalf("handlers were not flattened correctly")
	}
}

func TestFanoutFlattensMixedNestedHandlers(t *testing.T) {
	t.Parallel()

	h1 := slog.NewJSONHandler(io.Discard, nil)
	h2 := slog.NewJSONHandler(io.Discard, nil)
	h3 := slog.NewJSONHandler(io.Discard, nil)
	h4 := slog.NewJSONHandler(io.Discard, nil)

	handler := Fanout(Fanout(h1, h2), h3, Fanout(h4))

	fanout, ok := handler.(*FanoutHandler)
	if !ok {
		t.Fatalf("expected FanoutHandler, got %T", handler)
	}
	if len(fanout.handlers) != 4 {
		t.Fatalf("expected 4 handlers, got %d", len(fanout.handlers))
	}

	if fanout.handlers[0] != h1 ||
		fanout.handlers[1] != h2 ||
		fanout.handlers[2] != h3 ||
		fanout.handlers[3] != h4 {
		t.Fatalf("handlers were not flattened correctly")
	}
}
