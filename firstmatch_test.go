package slogmulti

import (
	"bytes"
	"io"
	"log/slog"
	"testing"

	"github.com/samber/lo"
)

func TestFirstMatch(t *testing.T) {
	t.Parallel()

	var queryBuf bytes.Buffer
	queryH := slog.NewJSONHandler(&queryBuf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	var otherBuf bytes.Buffer
	otherH := slog.NewJSONHandler(&otherBuf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := Router().
		Add(queryH, AttrKeyTypeIs("query", slog.KindString, "args", slog.KindAny)).
		Add(otherH).
		FirstMatch().
		Handler()

	fanout, ok := handler.(*FirstMatchHandler)
	if !ok {
		t.Fatalf("expected FirstMatchHandler, got %T", handler)
	}
	if len(fanout.handlers) != 2 {
		t.Fatalf("expected 2 handlers, got %d", len(fanout.handlers))
	}

	logger := slog.New(handler)

	logger.Info("db log 1", "query", "SELECT * FROM users", "args", []int{1, 2, 3})
	// check queryBuf has recorded log
	if !bytes.Contains(lo.Must(io.ReadAll(&queryBuf)), []byte("db log 1")) {
		t.Fatalf("expected queryBuf to contain 'db log', but it doesn't")
	}

	logger.Debug("db log 2", "query", "SELECT * FROM users", "args", []int{1, 2, 3})
	if bytes.Contains(lo.Must(io.ReadAll(&queryBuf)), []byte("db log 2")) {
		t.Fatalf("expected queryBuf to not contain 'db log 2' at debug level, but it does")
	}

	if bytes.Contains(lo.Must(io.ReadAll(&otherBuf)), []byte("db log 2")) {
		t.Fatalf("expected queryBuf to not contain 'db log 2' at debug level, but it does")
	}

	logger.Info("other logs", "something", "value")
	if !bytes.Contains(lo.Must(io.ReadAll(&otherBuf)), []byte("ther logs")) {
		t.Fatalf("expected otherBuf to contain 'other logs', but it doesn't")
	}
}
