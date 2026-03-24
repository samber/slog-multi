package slogmulti

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

type countingHandler struct {
	handleCount atomic.Int64
	minLevel    slog.Level
}

func newCountingHandler(minLevel slog.Level) *countingHandler {
	return &countingHandler{minLevel: minLevel}
}

func (h *countingHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.minLevel
}

func (h *countingHandler) Handle(_ context.Context, _ slog.Record) error {
	h.handleCount.Add(1)
	return nil
}

func (h *countingHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return &countingHandler{minLevel: h.minLevel}
}

func (h *countingHandler) WithGroup(_ string) slog.Handler {
	return &countingHandler{minLevel: h.minLevel}
}

type errorHandler struct {
	err         error
	handleCount atomic.Int64
}

func (h *errorHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *errorHandler) Handle(_ context.Context, _ slog.Record) error {
	h.handleCount.Add(1)
	return h.err
}

func (h *errorHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *errorHandler) WithGroup(_ string) slog.Handler      { return h }

type panickingHandler struct {
	panicValue any
}

func (h *panickingHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *panickingHandler) Handle(_ context.Context, _ slog.Record) error {
	panic(h.panicValue)
}

func (h *panickingHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *panickingHandler) WithGroup(_ string) slog.Handler      { return h }

// randomFailHandler fails on every Nth call (deterministic chaos).
type randomFailHandler struct {
	callCount atomic.Int64
	failEvery int64
	err       error
}

func (h *randomFailHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *randomFailHandler) Handle(_ context.Context, _ slog.Record) error {
	n := h.callCount.Add(1)
	if n%h.failEvery == 0 {
		return h.err
	}
	return nil
}

func (h *randomFailHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *randomFailHandler) WithGroup(_ string) slog.Handler      { return h }

// ---------------------------------------------------------------------------
// Edge case tests
// ---------------------------------------------------------------------------

func TestEdgeEmptyHandlers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)

	// Failover with 0 handlers
	assert.NoError(t, Failover()().Handle(ctx, r))
	assert.False(t, Failover()().Enabled(ctx, slog.LevelInfo))

	// Pool with 0 handlers
	assert.NoError(t, Pool()().Handle(ctx, r))
	assert.False(t, Pool()().Enabled(ctx, slog.LevelInfo))
}

func TestEdgeAllHandlersPanic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)

	p1 := &panickingHandler{panicValue: "panic1"}
	p2 := &panickingHandler{panicValue: errors.New("panic2")}

	// Fanout: should return joined error from both panics
	err := Fanout(p1, p2).Handle(ctx, r)
	assert.Error(t, err)

	// Failover: should try both and return last error
	err = Failover()(p1, p2).Handle(ctx, r)
	assert.Error(t, err)

	// Pool: should try all and return error
	err = Pool()(p1, p2).Handle(ctx, r)
	assert.Error(t, err)
}

func TestEdgeSingleHandlerFanout(t *testing.T) {
	t.Parallel()
	h := slog.NewJSONHandler(io.Discard, nil)
	result := Fanout(h)
	assert.Same(t, h, result)
}

func TestEdgeWithGroupEmpty(t *testing.T) {
	t.Parallel()

	h1 := newCountingHandler(slog.LevelDebug)
	h2 := newCountingHandler(slog.LevelDebug)

	fanout := Fanout(h1, h2)
	assert.Equal(t, fanout, fanout.WithGroup(""))

	failover := Failover()(h1, h2)
	assert.Equal(t, failover, failover.WithGroup(""))

	pool := Pool()(h1, h2)
	assert.Equal(t, pool, pool.WithGroup(""))

	recovery := RecoverHandlerError(func(_ context.Context, _ slog.Record, _ error) {})(h1)
	assert.Equal(t, recovery, recovery.WithGroup(""))
}

// ---------------------------------------------------------------------------
// Concurrent stress tests
// ---------------------------------------------------------------------------

const (
	stressGoroutines      = 100
	stressLogsPerRoutine  = 1000
)

func stressRun(t *testing.T, handler slog.Handler, level slog.Level) {
	t.Helper()
	var wg sync.WaitGroup
	wg.Add(stressGoroutines)
	for g := 0; g < stressGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < stressLogsPerRoutine; i++ {
				r := slog.NewRecord(time.Now(), level, fmt.Sprintf("msg-%d-%d", id, i), 0)
				r.AddAttrs(slog.Int("g", id), slog.Int("i", i))
				_ = handler.Handle(ctx, r)
			}
		}(g)
	}
	wg.Wait()
}

func TestStressFanoutConcurrent(t *testing.T) {
	t.Parallel()

	handlers := [3]*countingHandler{
		newCountingHandler(slog.LevelDebug),
		newCountingHandler(slog.LevelDebug),
		newCountingHandler(slog.LevelDebug),
	}
	fanout := Fanout(handlers[0], handlers[1], handlers[2])

	stressRun(t, fanout, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	for i, h := range handlers {
		assert.Equal(t, expected, h.handleCount.Load(), "handler %d", i)
	}
}

func TestStressFanoutWithAttrsConcurrent(t *testing.T) {
	t.Parallel()

	base := Fanout(
		slog.NewJSONHandler(io.Discard, nil),
		slog.NewJSONHandler(io.Discard, nil),
	)

	var wg sync.WaitGroup
	wg.Add(stressGoroutines * 2)
	for g := 0; g < stressGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			derived := base.WithAttrs([]slog.Attr{slog.Int("g", id)})
			r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
			_ = derived.Handle(context.Background(), r)
		}(g)
		go func(id int) {
			defer wg.Done()
			derived := base.WithGroup(fmt.Sprintf("group-%d", id))
			r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
			_ = derived.Handle(context.Background(), r)
		}(g)
	}
	wg.Wait()
}

func TestStressFailoverConcurrent(t *testing.T) {
	t.Parallel()

	primary := &errorHandler{err: errors.New("fail")}
	backup := newCountingHandler(slog.LevelDebug)
	handler := Failover()(primary, backup)

	stressRun(t, handler, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, primary.handleCount.Load(), "primary should have been called")
	assert.Equal(t, expected, backup.handleCount.Load(), "backup should have received all records")
}

func TestStressPoolConcurrent(t *testing.T) {
	t.Parallel()

	handlers := make([]*countingHandler, 5)
	slogHandlers := make([]slog.Handler, 5)
	for i := range handlers {
		handlers[i] = newCountingHandler(slog.LevelDebug)
		slogHandlers[i] = handlers[i]
	}
	pool := Pool()(slogHandlers...)

	stressRun(t, pool, slog.LevelInfo)

	var total int64
	for _, h := range handlers {
		total += h.handleCount.Load()
	}
	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, total, "total handle count across all pool handlers")
}

func TestStressRouterConcurrent(t *testing.T) {
	t.Parallel()

	errorSink := newCountingHandler(slog.LevelDebug)
	infoSink := newCountingHandler(slog.LevelDebug)
	catchAll := newCountingHandler(slog.LevelDebug)

	handler := Router().
		Add(errorSink, LevelIs(slog.LevelError)).
		Add(infoSink, LevelIs(slog.LevelInfo)).
		Add(catchAll).
		Handler()

	// Send half Info, half Error
	var wg sync.WaitGroup
	wg.Add(stressGoroutines)
	for g := 0; g < stressGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < stressLogsPerRoutine; i++ {
				level := slog.LevelInfo
				if i%2 == 0 {
					level = slog.LevelError
				}
				r := slog.NewRecord(time.Now(), level, "test", 0)
				_ = handler.Handle(ctx, r)
			}
		}(g)
	}
	wg.Wait()

	half := int64(stressGoroutines * stressLogsPerRoutine / 2)
	total := int64(stressGoroutines * stressLogsPerRoutine)

	assert.Equal(t, half, errorSink.handleCount.Load(), "error sink")
	assert.Equal(t, half, infoSink.handleCount.Load(), "info sink")
	assert.Equal(t, total, catchAll.handleCount.Load(), "catch-all sink")
}

func TestStressFirstMatchConcurrent(t *testing.T) {
	t.Parallel()

	errorSink := newCountingHandler(slog.LevelDebug)
	infoSink := newCountingHandler(slog.LevelDebug)
	catchAll := newCountingHandler(slog.LevelDebug)

	handler := Router().
		Add(errorSink, LevelIs(slog.LevelError)).
		Add(infoSink, LevelIs(slog.LevelInfo)).
		Add(catchAll).
		FirstMatch().
		Handler()

	// Send half Info, half Error
	var wg sync.WaitGroup
	wg.Add(stressGoroutines)
	for g := 0; g < stressGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < stressLogsPerRoutine; i++ {
				level := slog.LevelInfo
				if i%2 == 0 {
					level = slog.LevelError
				}
				r := slog.NewRecord(time.Now(), level, "test", 0)
				_ = handler.Handle(ctx, r)
			}
		}(g)
	}
	wg.Wait()

	half := int64(stressGoroutines * stressLogsPerRoutine / 2)

	assert.Equal(t, half, errorSink.handleCount.Load(), "error sink (first match)")
	assert.Equal(t, half, infoSink.handleCount.Load(), "info sink (first match)")
	assert.Equal(t, int64(0), catchAll.handleCount.Load(), "catch-all should get nothing in first-match")
}

func TestStressPipeConcurrent(t *testing.T) {
	t.Parallel()

	sink := newCountingHandler(slog.LevelDebug)

	middleware := NewHandleInlineMiddleware(
		func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
			record.AddAttrs(slog.String("middleware", "applied"))
			return next(ctx, record)
		},
	)

	handler := Pipe(middleware).Handler(sink)

	stressRun(t, handler, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, sink.handleCount.Load())
}

func TestStressRecoveryConcurrent(t *testing.T) {
	t.Parallel()

	var recoveryCount atomic.Int64
	recovery := RecoverHandlerError(func(_ context.Context, _ slog.Record, _ error) {
		recoveryCount.Add(1)
	})

	panicker := &panickingHandler{panicValue: "boom"}
	handler := recovery(panicker)

	stressRun(t, handler, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, recoveryCount.Load())
}

func TestStressComposedFanoutOfFailover(t *testing.T) {
	t.Parallel()

	primary := &errorHandler{err: errors.New("fail")}
	backup := newCountingHandler(slog.LevelDebug)
	direct := newCountingHandler(slog.LevelDebug)

	handler := Fanout(
		Failover()(primary, backup),
		direct,
	)

	stressRun(t, handler, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, backup.handleCount.Load(), "backup via failover")
	assert.Equal(t, expected, direct.handleCount.Load(), "direct via fanout")
}

func TestStressComposedPipeWithRouter(t *testing.T) {
	t.Parallel()

	errorSink := newCountingHandler(slog.LevelDebug)
	infoSink := newCountingHandler(slog.LevelDebug)

	var middlewareCalls atomic.Int64
	middleware := NewHandleInlineMiddleware(
		func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
			middlewareCalls.Add(1)
			return next(ctx, record)
		},
	)

	router := Router().
		Add(errorSink, LevelIs(slog.LevelError)).
		Add(infoSink, LevelIs(slog.LevelInfo)).
		Handler()

	handler := Pipe(middleware).Handler(router)

	// Alternate between Info and Error
	var wg sync.WaitGroup
	wg.Add(stressGoroutines)
	for g := 0; g < stressGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < stressLogsPerRoutine; i++ {
				level := slog.LevelInfo
				if i%2 == 0 {
					level = slog.LevelError
				}
				r := slog.NewRecord(time.Now(), level, "test", 0)
				_ = handler.Handle(ctx, r)
			}
		}(g)
	}
	wg.Wait()

	total := int64(stressGoroutines * stressLogsPerRoutine)
	half := total / 2

	assert.Equal(t, total, middlewareCalls.Load(), "middleware should see all records")
	assert.Equal(t, half, errorSink.handleCount.Load(), "error sink")
	assert.Equal(t, half, infoSink.handleCount.Load(), "info sink")
}

func TestStressInlineHandlerConcurrent(t *testing.T) {
	t.Parallel()

	var handleCount atomic.Int64
	handler := NewInlineHandler(
		func(_ context.Context, _ []string, _ []slog.Attr, level slog.Level) bool {
			return level >= slog.LevelInfo
		},
		func(_ context.Context, _ []string, _ []slog.Attr, _ slog.Record) error {
			handleCount.Add(1)
			return nil
		},
	)

	stressRun(t, handler, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, handleCount.Load())
}

func TestStressMultipleMiddlewaresConcurrent(t *testing.T) {
	t.Parallel()

	sink := newCountingHandler(slog.LevelDebug)

	var m1Calls, m2Calls, m3Calls atomic.Int64
	m1 := NewHandleInlineMiddleware(func(ctx context.Context, r slog.Record, next func(context.Context, slog.Record) error) error {
		m1Calls.Add(1)
		return next(ctx, r)
	})
	m2 := NewHandleInlineMiddleware(func(ctx context.Context, r slog.Record, next func(context.Context, slog.Record) error) error {
		m2Calls.Add(1)
		return next(ctx, r)
	})
	m3 := NewHandleInlineMiddleware(func(ctx context.Context, r slog.Record, next func(context.Context, slog.Record) error) error {
		m3Calls.Add(1)
		return next(ctx, r)
	})

	handler := Pipe(m1, m2, m3).Handler(sink)

	stressRun(t, handler, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, m1Calls.Load(), "middleware 1")
	assert.Equal(t, expected, m2Calls.Load(), "middleware 2")
	assert.Equal(t, expected, m3Calls.Load(), "middleware 3")
	assert.Equal(t, expected, sink.handleCount.Load(), "sink")
}

func TestStressMixedPanicsAndErrors(t *testing.T) {
	t.Parallel()

	panicker := &panickingHandler{panicValue: "boom"}
	errorer := &errorHandler{err: errors.New("fail")}
	good := newCountingHandler(slog.LevelDebug)

	// Fanout with a mix
	handler := Fanout(panicker, errorer, good)

	var wg sync.WaitGroup
	wg.Add(stressGoroutines)
	for g := 0; g < stressGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < stressLogsPerRoutine; i++ {
				r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
				err := handler.Handle(ctx, r)
				require.Error(t, err)
			}
		}(g)
	}
	wg.Wait()

	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, good.handleCount.Load())
}

// ---------------------------------------------------------------------------
// Adversarial tests
// ---------------------------------------------------------------------------

func TestAdversarialRandomFailures(t *testing.T) {
	t.Parallel()

	failing := &randomFailHandler{failEvery: 3, err: errors.New("random fail")}
	good := newCountingHandler(slog.LevelDebug)

	t.Run("fanout_with_random_failures", func(t *testing.T) {
		t.Parallel()
		handler := Fanout(failing, good)
		stressRun(t, handler, slog.LevelInfo)
		expected := int64(stressGoroutines * stressLogsPerRoutine)
		assert.Equal(t, expected, good.handleCount.Load(), "good handler must receive all records regardless of failures")
	})

	t.Run("failover_with_random_failures", func(t *testing.T) {
		t.Parallel()
		backup := newCountingHandler(slog.LevelDebug)
		rf := &randomFailHandler{failEvery: 2, err: errors.New("fail")}
		handler := Failover()(rf, backup)
		stressRun(t, handler, slog.LevelInfo)
		total := rf.callCount.Load() // primary was called for all
		expected := int64(stressGoroutines * stressLogsPerRoutine)
		assert.Equal(t, expected, total, "primary called for every record")
		// backup should have been called for ~half (every 2nd fails)
		assert.Greater(t, backup.handleCount.Load(), int64(0), "backup should receive some records")
	})
}

func TestAdversarialDeepNesting(t *testing.T) {
	t.Parallel()

	sink := newCountingHandler(slog.LevelDebug)
	// Fanout flattens nested FanoutHandlers, so each Fanout(prev, sink) adds one more ref.
	// After 5 iterations: 1 + 5 = 6 refs to sink.
	var handler slog.Handler = sink
	for i := 0; i < 5; i++ {
		handler = Fanout(handler, sink)
	}

	stressRun(t, handler, slog.LevelInfo)

	expected := int64(stressGoroutines * stressLogsPerRoutine * 6)
	assert.Equal(t, expected, sink.handleCount.Load())
}

func TestAdversarialConcurrentDeriveAndHandle(t *testing.T) {
	t.Parallel()

	sink := newCountingHandler(slog.LevelDebug)
	base := Fanout(sink, sink)

	var wg sync.WaitGroup
	// Half goroutines continuously derive, half continuously Handle
	wg.Add(stressGoroutines)
	for g := 0; g < stressGoroutines; g++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < stressLogsPerRoutine; i++ {
				switch i % 3 {
				case 0:
					derived := base.WithAttrs([]slog.Attr{slog.Int("id", id)})
					r := slog.NewRecord(time.Now(), slog.LevelInfo, "derived", 0)
					_ = derived.Handle(ctx, r)
				case 1:
					derived := base.WithGroup(fmt.Sprintf("g%d", id))
					r := slog.NewRecord(time.Now(), slog.LevelInfo, "grouped", 0)
					_ = derived.Handle(ctx, r)
				default:
					r := slog.NewRecord(time.Now(), slog.LevelInfo, "base", 0)
					_ = base.Handle(ctx, r)
				}
			}
		}(g)
	}
	wg.Wait()
	// No assertion on count — the test passes if no race/panic occurs
	assert.Greater(t, sink.handleCount.Load(), int64(0))
}

func TestAdversarialMixedPanicTypes(t *testing.T) {
	t.Parallel()

	panicValues := []any{
		"string panic",
		errors.New("error panic"),
		42,
		struct{ msg string }{"struct panic"},
	}

	for _, pv := range panicValues {
		handler := Fanout(&panickingHandler{panicValue: pv}, newCountingHandler(slog.LevelDebug))
		ctx := context.Background()

		var wg sync.WaitGroup
		var errCount atomic.Int64
		wg.Add(50)
		for g := 0; g < 50; g++ {
			go func() {
				defer wg.Done()
				for i := 0; i < 100; i++ {
					r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
					err := handler.Handle(ctx, r)
					if err != nil {
						errCount.Add(1)
					}
				}
			}()
		}
		wg.Wait()
		assert.Equal(t, int64(50*100), errCount.Load())
	}
}

func TestAdversarialPoolDistribution(t *testing.T) {
	t.Parallel()

	const numHandlers = 10
	handlers := make([]*countingHandler, numHandlers)
	slogHandlers := make([]slog.Handler, numHandlers)
	for i := range handlers {
		handlers[i] = newCountingHandler(slog.LevelDebug)
		slogHandlers[i] = handlers[i]
	}
	pool := Pool()(slogHandlers...)

	const goroutines = 200
	const logsPerRoutine = 500
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < logsPerRoutine; i++ {
				r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
				_ = pool.Handle(ctx, r)
			}
		}()
	}
	wg.Wait()

	total := int64(goroutines * logsPerRoutine)
	var sum int64
	for _, h := range handlers {
		count := h.handleCount.Load()
		sum += count
		// Each handler should get at least 1% of records (probabilistic but safe with 100k records)
		minExpected := total / 100
		assert.Greater(t, count, minExpected,
			"handler should get >1%% of records, got %d/%d", count, total)
	}
	assert.Equal(t, total, sum, "total records across all handlers")
}

func TestAdversarialRouterManyPredicates(t *testing.T) {
	t.Parallel()

	sink := newCountingHandler(slog.LevelDebug)
	r := Router()
	// 20 handlers with non-matching predicates
	for i := 0; i < 20; i++ {
		r = r.Add(newCountingHandler(slog.LevelDebug), MessageIs(fmt.Sprintf("match-%d", i)))
	}
	// catch-all
	r = r.Add(sink)
	handler := r.Handler()

	stressRun(t, handler, slog.LevelInfo)

	// All records have msg "msg-X-Y" which won't match "match-N", so catch-all gets everything
	expected := int64(stressGoroutines * stressLogsPerRoutine)
	assert.Equal(t, expected, sink.handleCount.Load())
}

func TestAdversarialRecoveryConcurrentPanicTypes(t *testing.T) {
	t.Parallel()

	var recoveryCount atomic.Int64
	recovery := RecoverHandlerError(func(_ context.Context, _ slog.Record, _ error) {
		recoveryCount.Add(1)
	})

	panicValues := []any{"string", errors.New("error"), 123, 3.14}

	var wg sync.WaitGroup
	wg.Add(len(panicValues) * 50)
	for _, pv := range panicValues {
		handler := recovery(&panickingHandler{panicValue: pv})
		for g := 0; g < 50; g++ {
			go func(h slog.Handler) {
				defer wg.Done()
				ctx := context.Background()
				for i := 0; i < 200; i++ {
					r := slog.NewRecord(time.Now(), slog.LevelInfo, "test", 0)
					_ = h.Handle(ctx, r)
				}
			}(handler)
		}
	}
	wg.Wait()

	expected := int64(len(panicValues) * 50 * 200)
	assert.Equal(t, expected, recoveryCount.Load())
}

func TestAdversarialHighContention(t *testing.T) {
	t.Parallel()

	// Single handler, maximum goroutines — stress the clone path
	sink := newCountingHandler(slog.LevelDebug)
	handler := Fanout(sink, sink) // 2 copies to force clone

	const goroutines = 1000
	const logsPerRoutine = 1000
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			ctx := context.Background()
			for i := 0; i < logsPerRoutine; i++ {
				r := slog.NewRecord(time.Now(), slog.LevelInfo, "contention", 0)
				r.AddAttrs(slog.Int("id", id), slog.Int("i", i))
				_ = handler.Handle(ctx, r)
			}
		}(g)
	}
	wg.Wait()

	// Each record goes to both copies of sink
	expected := int64(goroutines * logsPerRoutine * 2)
	assert.Equal(t, expected, sink.handleCount.Load())
}
