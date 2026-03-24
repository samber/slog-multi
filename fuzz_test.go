package slogmulti

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func buildFuzzRecord(levelInt int, msg string, attrCount int) slog.Record {
	level := slog.Level(levelInt)
	r := slog.NewRecord(time.Now(), level, msg, 0)
	for i := 0; i < attrCount; i++ {
		r.AddAttrs(slog.Int(fmt.Sprintf("k%d", i), i))
	}
	return r
}

func FuzzFanoutHandle(f *testing.F) {
	f.Add(0, "hello", 3)
	f.Add(-4, "", 0)
	f.Add(4, "warn msg", 1)
	f.Add(8, strings.Repeat("x", 1000), 50)

	f.Fuzz(func(t *testing.T, levelInt int, msg string, attrCount int) {
		if attrCount < 0 {
			attrCount = 0
		}
		if attrCount > 50 {
			attrCount = 50
		}

		level := slog.Level(levelInt)
		h1 := newCountingHandler(slog.LevelDebug)
		h2 := newCountingHandler(slog.LevelWarn)
		h3 := newCountingHandler(slog.LevelDebug)

		fanout := Fanout(h1, h2, h3)
		r := buildFuzzRecord(levelInt, msg, attrCount)
		err := fanout.Handle(context.Background(), r)
		assert.NoError(t, err)

		// Fanout only calls Handle on enabled handlers
		var expected1, expected2, expected3 int64
		if level >= slog.LevelDebug {
			expected1 = 1
			expected3 = 1
		}
		if level >= slog.LevelWarn {
			expected2 = 1
		}

		assert.Equal(t, expected1, h1.handleCount.Load(), "h1")
		assert.Equal(t, expected2, h2.handleCount.Load(), "h2")
		assert.Equal(t, expected3, h3.handleCount.Load(), "h3")
	})
}

func FuzzFailoverHandle(f *testing.F) {
	f.Add(0, "test", true, false)
	f.Add(0, "test", false, false)
	f.Add(0, "test", true, true)
	f.Add(4, "", false, true)

	f.Fuzz(func(t *testing.T, levelInt int, msg string, firstFails bool, secondFails bool) {
		// errorHandler.Enabled always returns true, countingHandler checks minLevel.
		// Use errorHandler for both fail and success paths so Enabled is always true,
		// avoiding level-based skipping that complicates assertions.
		h1err := &errorHandler{err: errors.New("h1 fail")}
		h1ok := &errorHandler{err: nil}
		h2err := &errorHandler{err: errors.New("h2 fail")}
		h2ok := &errorHandler{err: nil}

		var h1, h2 *errorHandler
		if firstFails {
			h1 = h1err
		} else {
			h1 = h1ok
		}
		if secondFails {
			h2 = h2err
		} else {
			h2 = h2ok
		}

		handler := Failover()(h1, h2)
		r := buildFuzzRecord(levelInt, msg, 0)
		err := handler.Handle(context.Background(), r)

		if !firstFails {
			assert.NoError(t, err)
			assert.Equal(t, int64(1), h1.handleCount.Load())
			assert.Equal(t, int64(0), h2.handleCount.Load())
		} else if !secondFails {
			assert.NoError(t, err)
			assert.Equal(t, int64(1), h2.handleCount.Load())
		} else {
			assert.Error(t, err)
		}
	})
}

func FuzzPoolHandle(f *testing.F) {
	f.Add(0, "hello")
	f.Add(-4, "")
	f.Add(8, strings.Repeat("a", 500))

	f.Fuzz(func(t *testing.T, levelInt int, msg string) {
		// Use errorHandler (always enabled) to avoid level-based skipping
		handlers := make([]*errorHandler, 3)
		slogHandlers := make([]slog.Handler, 3)
		for i := range handlers {
			handlers[i] = &errorHandler{err: nil}
			slogHandlers[i] = handlers[i]
		}

		pool := Pool()(slogHandlers...)

		const iterations = 100
		for i := 0; i < iterations; i++ {
			r := buildFuzzRecord(levelInt, msg, 0)
			err := pool.Handle(context.Background(), r)
			assert.NoError(t, err)
		}

		var total int64
		for _, h := range handlers {
			total += h.handleCount.Load()
		}
		assert.Equal(t, int64(iterations), total)
	})
}

func FuzzFirstMatchHandle(f *testing.F) {
	f.Add(0, "hello world")
	f.Add(4, "error occurred")
	f.Add(-4, "debug")
	f.Add(8, "")

	f.Fuzz(func(t *testing.T, levelInt int, msg string) {
		// Use errorHandler (always enabled) as sinks to avoid level gating
		errorSink := &errorHandler{err: nil}
		infoSink := &errorHandler{err: nil}
		catchAll := &errorHandler{err: nil}

		handler := Router().
			Add(errorSink, LevelIs(slog.LevelError)).
			Add(infoSink, LevelIs(slog.LevelInfo)).
			Add(catchAll).
			FirstMatch().
			Handler()

		r := buildFuzzRecord(levelInt, msg, 0)
		err := handler.Handle(context.Background(), r)
		assert.NoError(t, err)

		level := slog.Level(levelInt)
		total := errorSink.handleCount.Load() + infoSink.handleCount.Load() + catchAll.handleCount.Load()

		if level == slog.LevelError {
			assert.Equal(t, int64(1), errorSink.handleCount.Load())
			assert.Equal(t, int64(1), total)
		} else if level == slog.LevelInfo {
			assert.Equal(t, int64(1), infoSink.handleCount.Load())
			assert.Equal(t, int64(1), total)
		} else {
			// Catch-all gets it (no predicate, always matches)
			assert.Equal(t, int64(1), catchAll.handleCount.Load())
			assert.Equal(t, int64(1), total)
		}
	})
}

func FuzzRouterPredicates(f *testing.F) {
	f.Add("hello world", 0)
	f.Add("", -4)
	f.Add("error in database", 8)
	f.Add(strings.Repeat("x", 1000), 4)

	f.Fuzz(func(t *testing.T, msg string, levelInt int) {
		level := slog.Level(levelInt)
		r := slog.NewRecord(time.Now(), level, msg, 0)
		r.AddAttrs(slog.String("key", "value"), slog.Int("num", 42))
		ctx := context.Background()

		// All of these must not panic regardless of input
		LevelIs(slog.LevelInfo, slog.LevelError)(ctx, r)
		LevelIsNot(slog.LevelInfo)(ctx, r)
		MessageIs(msg)(ctx, r)
		MessageIsNot(msg)(ctx, r)
		MessageContains("error")(ctx, r)
		MessageNotContains("error")(ctx, r)
		AttrValueIs("key", "value")(ctx, r)
		AttrKindIs("key", slog.KindString)(ctx, r)
	})
}
