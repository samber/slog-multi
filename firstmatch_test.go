package slogmulti

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

var remoteTimeReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func TestFirstMatch(t *testing.T) {
	t.Parallel()

	t.Run("routes to first matching handler", func(t *testing.T) {
		queryBuf := bytes.NewBufferString("")

		queryH := slog.NewTextHandler(queryBuf, &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: remoteTimeReplaceAttr,
		})

		commonBuf := bytes.NewBufferString("")
		commonH := slog.NewTextHandler(commonBuf, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: remoteTimeReplaceAttr,
		})

		handler := Router().
			Add(queryH, AttrKindIs("query", slog.KindString, "args", slog.KindAny)).
			Add(commonH).
			FirstMatch().
			Handler()

		logger := slog.New(handler).With("user_id", 123)
		// Test 1: Log matching first handler should only go to queryBuf
		logger.Info("get user by id", "query", "SELECT * FROM users id = ?", "args", []int{1})
		// Test 2: Debug level filtered by queryH, should not match any handler
		logger.Debug("get users", "query", "SELECT * FROM users", "args", []int{})
		// Test 3: Log not matching first handler should go to commonBuf
		logger.Warn("cache miss", "key", "user_1")

		if strings.TrimSpace(queryBuf.String()) != `level=INFO msg="get user by id" user_id=123 query="SELECT * FROM users id = ?" args=[1]` {
			t.Fatalf("query log buffer did not match")
		}

		if strings.TrimSpace(commonBuf.String()) != `level=WARN msg="cache miss" user_id=123 key=user_1` {
			t.Fatalf("common log buffer did not match")
		}
	})

	t.Run("stops at first match", func(t *testing.T) {
		buf1, buf2, buf3 := bytes.NewBufferString(""), bytes.NewBufferString(""), bytes.NewBufferString("")

		h1 := slog.NewTextHandler(buf1, &slog.HandlerOptions{
			ReplaceAttr: remoteTimeReplaceAttr,
		})
		h2 := slog.NewTextHandler(buf2, &slog.HandlerOptions{
			ReplaceAttr: remoteTimeReplaceAttr,
		})
		h3 := slog.NewTextHandler(buf3, &slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: remoteTimeReplaceAttr,
		})

		handler := Router().
			Add(h1, AttrValueIs("type", "error")).
			Add(h2, AttrValueIs("type", "error")). // Also matches, but should not receive
			Add(h3). // Fallback handler
			FirstMatch().
			Handler()

		logger := slog.New(handler).With("user_id", 123)
		logger.Info("test", "type", "error")
		logger.Debug("other_log", "type", "not_error")

		if buf1.String() != "level=INFO msg=test user_id=123 type=error\n" {
			t.Errorf("expected buf1 to contain log")
		}
		if buf2.Len() != 0 {
			t.Errorf("expected buf2 to NOT contain log (should stop at first match)")
		}
		if buf3.String() != "level=DEBUG msg=other_log user_id=123 type=not_error\n" {
			t.Errorf("expected buf3 to contain log")
		}
	})
}
