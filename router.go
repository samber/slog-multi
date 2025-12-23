package slogmulti

import (
	"context"
	"log/slog"

	"github.com/samber/lo"
)

type router struct {
	handlers   []slog.Handler
	firstMatch bool
}

// Router creates a new router instance for building conditional log routing.
// This function is the entry point for creating a routing configuration.
//
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.LevelIs(slog.LevelInfo)).
//	    Add(fileHandler, slogmulti.LevelIs(slog.LevelError)).
//	    Handler()
//
// Returns:
//
//	A new router instance ready for configuration
func Router() *router {
	return &router{
		handlers:   []slog.Handler{},
		firstMatch: false,
	}
}

// Add registers a new handler with optional predicates to the router.
// The handler will only process records if all provided predicates return true.
//
// Args:
//
//	handler: The slog.Handler to register
//	predicates: Optional functions that determine if a record should be routed to this handler
//
// Returns:
//
//	The router instance for method chaining
func (h *router) Add(handler slog.Handler, predicates ...func(ctx context.Context, r slog.Record) bool) *router {
	return &router{
		handlers: append(
			h.handlers,
			&RoutableHandler{
				predicates: predicates,
				handler:    handler,
				groups:     []string{},
				attrs:      []slog.Attr{},
			},
		),
	}
}

// Handler creates a slog.Handler from the configured router.
// This method finalizes the routing configuration and returns a handler
// that can be used with slog.New().
//
// Returns:
//
//	A slog.Handler that implements the routing logic
func (h *router) Handler() slog.Handler {
	if h.firstMatch {
		return FirstMatch(lo.Map(h.handlers, func(h slog.Handler, _ int) *RoutableHandler {
			return h.(*RoutableHandler)
		})...)
	} else {
		return Fanout(h.handlers...)
	}
}

func (h *router) FirstMatch() *router {
	return &router{
		handlers:   h.handlers,
		firstMatch: true,
	}
}
