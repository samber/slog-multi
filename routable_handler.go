package slogmulti

import (
	"context"
	"log/slog"
	"slices"

	slogcommon "github.com/samber/slog-common"
)

// Ensure RoutableHandler implements the slog.Handler interface at compile time
var _ slog.Handler = (*RoutableHandler)(nil)

// RoutableHandler wraps a slog.Handler with conditional matching logic.
// It only forwards records to the underlying handler if all predicates return true.
// This enables sophisticated routing scenarios like level-based or attribute-based routing.
//
// @TODO: implement round robin strategy for load balancing across multiple handlers
type RoutableHandler struct {
	// predicates contains functions that determine if a record should be processed
	predicates []func(ctx context.Context, r slog.Record) bool
	// handler is the underlying slog.Handler that processes matching records
	handler slog.Handler
	// groups tracks the current group hierarchy for proper attribute handling
	groups []string
	// attrs contains accumulated attributes that should be added to records
	attrs []slog.Attr
	// skipMatchCheck indicates the caller MUST call IsMatch(ctx, record) and MUST NOT invoke the handler for a given record if IsMatch returns false.
	skipMatchCheck bool
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

// Handle processes a log record if all predicates return true.
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
	if h.skipMatchCheck || h.IsMatch(ctx, r) {
		return h.handler.Handle(ctx, r)
	}

	return nil
}

func (h *RoutableHandler) IsMatch(ctx context.Context, r slog.Record) bool {
	clone := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	clone.AddAttrs(
		slogcommon.AppendRecordAttrsToAttrs(h.attrs, h.groups, &r)...,
	)

	for _, predicate := range h.predicates {
		if !predicate(ctx, clone) {
			return false
		}
	}

	return true
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
		predicates: h.predicates,
		handler:    h.handler.WithAttrs(attrs),
		groups:     slices.Clone(h.groups),
		attrs:      slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
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
		predicates: h.predicates,
		handler:    h.handler.WithGroup(name),
		groups:     append(slices.Clone(h.groups), name),
		attrs:      h.attrs,
	}
}
