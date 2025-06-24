package slogmulti

import (
	"context"
	"slices"

	"log/slog"

	slogcommon "github.com/samber/slog-common"
)

type router struct {
	handlers []slog.Handler
}

// Router creates a new router instance for building conditional log routing.
// This function is the entry point for creating a routing configuration.
//
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.Level(slog.LevelInfo)).
//	    Add(fileHandler, slogmulti.Level(slog.LevelError)).
//	    Handler()
//
// Returns:
//
//	A new router instance ready for configuration
func Router() *router {
	return &router{
		handlers: []slog.Handler{},
	}
}

// Add registers a new handler with optional matchers to the router.
// The handler will only process records if all provided matchers return true.
//
// Args:
//
//	handler: The slog.Handler to register
//	matchers: Optional functions that determine if a record should be routed to this handler
//
// Returns:
//
//	The router instance for method chaining
func (h *router) Add(handler slog.Handler, matchers ...func(ctx context.Context, r slog.Record) bool) *router {
	return &router{
		handlers: append(
			h.handlers,
			&RoutableHandler{
				matchers: matchers,
				handler:  handler,
				groups:   []string{},
				attrs:    []slog.Attr{},
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
	return Fanout(h.handlers...)
}

// Ensure RoutableHandler implements the slog.Handler interface at compile time
var _ slog.Handler = (*RoutableHandler)(nil)

// RoutableHandler wraps a slog.Handler with conditional matching logic.
// It only forwards records to the underlying handler if all matchers return true.
// This enables sophisticated routing scenarios like level-based or attribute-based routing.
//
// @TODO: implement round robin strategy for load balancing across multiple handlers
type RoutableHandler struct {
	// matchers contains functions that determine if a record should be processed
	matchers []func(ctx context.Context, r slog.Record) bool
	// handler is the underlying slog.Handler that processes matching records
	handler slog.Handler
	// groups tracks the current group hierarchy for proper attribute handling
	groups []string
	// attrs contains accumulated attributes that should be added to records
	attrs []slog.Attr
}

// Enabled checks if the underlying handler is enabled for the given log level.
// This method implements the slog.Handler interface requirement.
//
// Args:
//
//	ctx: The context for the logging operation
//	l: The log level to check
//
// Returns:
//
//	true if the underlying handler is enabled for the level, false otherwise
func (h *RoutableHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler.Enabled(ctx, l)
}

// Handle processes a log record if all matchers return true.
// This method implements the slog.Handler interface requirement.
//
// Args:
//
//	ctx: The context for the logging operation
//	r: The log record to process
//
// Returns:
//
//	An error if the underlying handler failed to process the record, nil otherwise
func (h *RoutableHandler) Handle(ctx context.Context, r slog.Record) error {
	clone := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	clone.AddAttrs(
		slogcommon.AppendRecordAttrsToAttrs(h.attrs, h.groups, &r)...,
	)

	for _, matcher := range h.matchers {
		if !matcher(ctx, clone) {
			return nil
		}
	}

	return h.handler.Handle(ctx, r)
}

// WithAttrs creates a new RoutableHandler with additional attributes.
// This method implements the slog.Handler interface requirement.
//
// The method properly handles attribute accumulation within the current group context,
// ensuring that attributes are correctly applied to records when they are processed.
//
// Args:
//
//	attrs: The attributes to add to the handler
//
// Returns:
//
//	A new RoutableHandler with the additional attributes
func (h *RoutableHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RoutableHandler{
		matchers: h.matchers,
		handler:  h.handler.WithAttrs(attrs),
		groups:   slices.Clone(h.groups),
		attrs:    slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
	}
}

// WithGroup creates a new RoutableHandler with a group name.
// This method implements the slog.Handler interface requirement.
//
// The method follows the same pattern as the standard slog implementation:
// - If the group name is empty, returns the original handler unchanged
// - Otherwise, creates a new handler with the group name added to the group hierarchy
//
// Args:
//
//	name: The group name to apply to the handler
//
// Returns:
//
//	A new RoutableHandler with the group name, or the original handler if the name is empty
func (h *RoutableHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &RoutableHandler{
		matchers: h.matchers,
		handler:  h.handler.WithGroup(name),
		groups:   append(slices.Clone(h.groups), name),
		attrs:    h.attrs,
	}
}
