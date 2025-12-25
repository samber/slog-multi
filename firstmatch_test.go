package slogmulti

import (
	"bytes"
	"log/slog"
	"testing"
)

func TestFirstMatch(t *testing.T) {
	t.Parallel()

	t.Run("routes to first matching handler", func(t *testing.T) {
		var queryBuf bytes.Buffer
		queryH := slog.NewJSONHandler(&queryBuf, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})

		var otherBuf bytes.Buffer
		otherH := slog.NewJSONHandler(&otherBuf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		handler := Router().
			Add(queryH, AttrKindIs("query", slog.KindString, "args", slog.KindAny)).
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

		// Test 1: Log matching first handler should only go to queryBuf
		logger.Info("db log 1", "query", "SELECT * FROM users", "args", []int{1, 2, 3})

		if !bytes.Contains(queryBuf.Bytes(), []byte("db log 1")) {
			t.Errorf("expected queryBuf to contain 'db log 1', but it doesn't")
		}
		if bytes.Contains(otherBuf.Bytes(), []byte("db log 1")) {
			t.Errorf("expected otherBuf to NOT contain 'db log 1' (should stop at first match), but it does")
		}

		queryBuf.Reset()
		otherBuf.Reset()

		// Test 2: Debug level filtered by queryH, should not match any handler
		logger.Debug("db log 2", "query", "SELECT * FROM users", "args", []int{1, 2, 3})

		if bytes.Contains(queryBuf.Bytes(), []byte("db log 2")) {
			t.Errorf("expected queryBuf to not contain 'db log 2' at debug level, but it does")
		}
		if bytes.Contains(otherBuf.Bytes(), []byte("db log 2")) {
			t.Errorf("expected otherBuf to not contain 'db log 2' (filtered by first handler), but it does")
		}

		queryBuf.Reset()
		otherBuf.Reset()

		// Test 3: Log not matching first handler should go to second (otherH)
		logger.Info("other logs", "something", "value")

		if bytes.Contains(queryBuf.Bytes(), []byte("other logs")) {
			t.Errorf("expected queryBuf to NOT contain 'other logs', but it does")
		}
		if !bytes.Contains(otherBuf.Bytes(), []byte("other logs")) {
			t.Errorf("expected otherBuf to contain 'other logs', but it doesn't")
		}
	})

	t.Run("stops at first match", func(t *testing.T) {
		var buf1, buf2, buf3 bytes.Buffer
		h1 := slog.NewJSONHandler(&buf1, nil)
		h2 := slog.NewJSONHandler(&buf2, nil)
		h3 := slog.NewJSONHandler(&buf3, nil)

		handler := Router().
			Add(h1, AttrValueIs("type", "error")).
			Add(h2, AttrValueIs("type", "error")). // Also matches, but should not receive
			Add(h3).                               // Catch-all
			FirstMatch().
			Handler()

		logger := slog.New(handler)
		logger.Info("test", "type", "error")

		if !bytes.Contains(buf1.Bytes(), []byte("test")) {
			t.Errorf("expected buf1 to contain log")
		}
		if bytes.Contains(buf2.Bytes(), []byte("test")) {
			t.Errorf("expected buf2 to NOT contain log (should stop at first match)")
		}
		if bytes.Contains(buf3.Bytes(), []byte("test")) {
			t.Errorf("expected buf3 to NOT contain log (should stop at first match)")
		}
	})
}
