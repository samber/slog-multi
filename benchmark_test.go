package slogmulti

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// noopHandler is a minimal handler for benchmarks — no atomics, no allocations.
type noopHandler struct{}

func (noopHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (noopHandler) Handle(context.Context, slog.Record) error  { return nil }
func (h noopHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h noopHandler) WithGroup(string) slog.Handler            { return h }

type noopErrorHandler struct{ err error }

func (noopErrorHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (h noopErrorHandler) Handle(context.Context, slog.Record) error { return h.err }
func (h noopErrorHandler) WithAttrs([]slog.Attr) slog.Handler       { return h }
func (h noopErrorHandler) WithGroup(string) slog.Handler             { return h }

func benchRecord() slog.Record {
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "benchmark message", 0)
	r.AddAttrs(slog.String("key", "value"), slog.Int("count", 42))
	return r
}

func makeHandlers(n int) []slog.Handler {
	h := make([]slog.Handler, n)
	for i := range h {
		h[i] = noopHandler{}
	}
	return h
}

// ---------------------------------------------------------------------------
// Fanout benchmarks
// ---------------------------------------------------------------------------

func BenchmarkFanoutHandle(b *testing.B) {
	for _, n := range []int{1, 3, 5, 10} {
		b.Run(fmt.Sprintf("handlers=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			handler := Fanout(makeHandlers(n)...)
			ctx := context.Background()
			r := benchRecord()
			for i := 0; i < b.N; i++ {
				_ = handler.Handle(ctx, r)
			}
		})
	}
}

func BenchmarkFanoutHandleParallel(b *testing.B) {
	b.ReportAllocs()
	handler := Fanout(makeHandlers(3)...)
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.Handle(ctx, benchRecord())
		}
	})
}

func BenchmarkFanoutWithAttrs(b *testing.B) {
	for _, n := range []int{1, 3, 5} {
		b.Run(fmt.Sprintf("handlers=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			handler := Fanout(makeHandlers(n)...)
			attrs := []slog.Attr{slog.String("env", "prod"), slog.Int("version", 1)}
			for i := 0; i < b.N; i++ {
				_ = handler.WithAttrs(attrs)
			}
		})
	}
}

func BenchmarkFanoutWithGroup(b *testing.B) {
	b.ReportAllocs()
	handler := Fanout(makeHandlers(3)...)
	for i := 0; i < b.N; i++ {
		_ = handler.WithGroup("request")
	}
}

// ---------------------------------------------------------------------------
// Failover benchmarks
// ---------------------------------------------------------------------------

func BenchmarkFailoverHandle(b *testing.B) {
	b.Run("first_ok", func(b *testing.B) {
		b.ReportAllocs()
		handler := Failover()(noopHandler{}, noopHandler{}, noopHandler{})
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r)
		}
	})

	b.Run("first_fails", func(b *testing.B) {
		b.ReportAllocs()
		handler := Failover()(noopErrorHandler{err: errors.New("fail")}, noopHandler{}, noopHandler{})
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r)
		}
	})

	b.Run("all_fail", func(b *testing.B) {
		b.ReportAllocs()
		e := noopErrorHandler{err: errors.New("fail")}
		handler := Failover()(e, e, e)
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r)
		}
	})
}

func BenchmarkFailoverHandleParallel(b *testing.B) {
	b.ReportAllocs()
	handler := Failover()(noopErrorHandler{err: errors.New("fail")}, noopHandler{})
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.Handle(ctx, benchRecord())
		}
	})
}

// ---------------------------------------------------------------------------
// Pool benchmarks
// ---------------------------------------------------------------------------

func BenchmarkPoolHandle(b *testing.B) {
	for _, n := range []int{1, 3, 5, 10} {
		b.Run(fmt.Sprintf("handlers=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			handler := Pool()(makeHandlers(n)...)
			ctx := context.Background()
			r := benchRecord()
			for i := 0; i < b.N; i++ {
				_ = handler.Handle(ctx, r)
			}
		})
	}
}

func BenchmarkPoolHandleParallel(b *testing.B) {
	b.ReportAllocs()
	handler := Pool()(makeHandlers(5)...)
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.Handle(ctx, benchRecord())
		}
	})
}

// ---------------------------------------------------------------------------
// Router benchmarks
// ---------------------------------------------------------------------------

func BenchmarkRouterHandle(b *testing.B) {
	for _, n := range []int{1, 3, 5} {
		b.Run(fmt.Sprintf("predicates=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			r := Router()
			for i := 0; i < n; i++ {
				r = r.Add(noopHandler{}, LevelIs(slog.LevelError))
			}
			r = r.Add(noopHandler{}) // catch-all
			handler := r.Handler()
			ctx := context.Background()
			rec := benchRecord()
			for i := 0; i < b.N; i++ {
				_ = handler.Handle(ctx, rec)
			}
		})
	}
}

func BenchmarkRouterHandleParallel(b *testing.B) {
	b.ReportAllocs()
	handler := Router().
		Add(noopHandler{}, LevelIs(slog.LevelError)).
		Add(noopHandler{}, LevelIs(slog.LevelInfo)).
		Add(noopHandler{}).
		Handler()
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.Handle(ctx, benchRecord())
		}
	})
}

// ---------------------------------------------------------------------------
// FirstMatch benchmarks
// ---------------------------------------------------------------------------

func BenchmarkFirstMatchHandle(b *testing.B) {
	b.Run("match_first", func(b *testing.B) {
		b.ReportAllocs()
		handler := Router().
			Add(noopHandler{}, LevelIs(slog.LevelInfo)).
			Add(noopHandler{}, LevelIs(slog.LevelError)).
			Add(noopHandler{}).
			FirstMatch().
			Handler()
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r) // LevelInfo matches first
		}
	})

	b.Run("match_last", func(b *testing.B) {
		b.ReportAllocs()
		handler := Router().
			Add(noopHandler{}, LevelIs(slog.LevelError)).
			Add(noopHandler{}, LevelIs(slog.LevelWarn)).
			Add(noopHandler{}, LevelIs(slog.LevelInfo)).
			FirstMatch().
			Handler()
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r) // LevelInfo matches last
		}
	})

	b.Run("match_catchall", func(b *testing.B) {
		b.ReportAllocs()
		handler := Router().
			Add(noopHandler{}, LevelIs(slog.LevelError)).
			Add(noopHandler{}, LevelIs(slog.LevelWarn)).
			Add(noopHandler{}).
			FirstMatch().
			Handler()
		ctx := context.Background()
		r := slog.NewRecord(time.Now(), slog.LevelDebug, "debug", 0)
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r)
		}
	})
}

func BenchmarkFirstMatchHandleParallel(b *testing.B) {
	b.ReportAllocs()
	handler := Router().
		Add(noopHandler{}, LevelIs(slog.LevelError)).
		Add(noopHandler{}, LevelIs(slog.LevelInfo)).
		Add(noopHandler{}).
		FirstMatch().
		Handler()
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.Handle(ctx, benchRecord())
		}
	})
}

// ---------------------------------------------------------------------------
// Pipe benchmarks
// ---------------------------------------------------------------------------

func passthroughMiddleware() Middleware {
	return NewHandleInlineMiddleware(
		func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
			return next(ctx, record)
		},
	)
}

func BenchmarkPipeHandle(b *testing.B) {
	for _, n := range []int{1, 3, 5} {
		b.Run(fmt.Sprintf("middlewares=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			middlewares := make([]Middleware, n)
			for i := range middlewares {
				middlewares[i] = passthroughMiddleware()
			}
			handler := Pipe(middlewares...).Handler(noopHandler{})
			ctx := context.Background()
			r := benchRecord()
			for i := 0; i < b.N; i++ {
				_ = handler.Handle(ctx, r)
			}
		})
	}
}

func BenchmarkPipeHandleParallel(b *testing.B) {
	b.ReportAllocs()
	handler := Pipe(passthroughMiddleware(), passthroughMiddleware(), passthroughMiddleware()).Handler(noopHandler{})
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.Handle(ctx, benchRecord())
		}
	})
}

// ---------------------------------------------------------------------------
// Recovery benchmarks
// ---------------------------------------------------------------------------

func BenchmarkRecoveryHandle(b *testing.B) {
	noopRecovery := func(_ context.Context, _ slog.Record, _ error) {}

	b.Run("no_error", func(b *testing.B) {
		b.ReportAllocs()
		handler := RecoverHandlerError(noopRecovery)(noopHandler{})
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r)
		}
	})

	b.Run("with_error", func(b *testing.B) {
		b.ReportAllocs()
		handler := RecoverHandlerError(noopRecovery)(noopErrorHandler{err: errors.New("fail")})
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r)
		}
	})

	b.Run("with_panic", func(b *testing.B) {
		b.ReportAllocs()
		handler := RecoverHandlerError(noopRecovery)(&panickingHandler{panicValue: "boom"})
		ctx := context.Background()
		r := benchRecord()
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(ctx, r)
		}
	})
}

// ---------------------------------------------------------------------------
// Composition benchmarks
// ---------------------------------------------------------------------------

func BenchmarkComposedFanoutOfFailover(b *testing.B) {
	b.ReportAllocs()
	handler := Fanout(
		Failover()(noopErrorHandler{err: errors.New("fail")}, noopHandler{}),
		noopHandler{},
	)
	ctx := context.Background()
	r := benchRecord()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(ctx, r)
	}
}

func BenchmarkComposedPipeWithRouter(b *testing.B) {
	b.ReportAllocs()
	router := Router().
		Add(noopHandler{}, LevelIs(slog.LevelError)).
		Add(noopHandler{}, LevelIs(slog.LevelInfo)).
		Add(noopHandler{}).
		Handler()
	handler := Pipe(passthroughMiddleware()).Handler(router)
	ctx := context.Background()
	r := benchRecord()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(ctx, r)
	}
}

func BenchmarkComposedDeepNesting(b *testing.B) {
	b.ReportAllocs()
	// 5 levels of Fanout nesting
	var handler slog.Handler = noopHandler{}
	for i := 0; i < 5; i++ {
		handler = Fanout(handler, noopHandler{})
	}
	ctx := context.Background()
	r := benchRecord()
	for i := 0; i < b.N; i++ {
		_ = handler.Handle(ctx, r)
	}
}

func BenchmarkComposedDeepNestingParallel(b *testing.B) {
	b.ReportAllocs()
	var handler slog.Handler = noopHandler{}
	for i := 0; i < 5; i++ {
		handler = Fanout(handler, noopHandler{})
	}
	ctx := context.Background()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = handler.Handle(ctx, benchRecord())
		}
	})
}
