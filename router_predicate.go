package slogmulti

import (
	"context"
	"log/slog"
	"strings"
)

// LevelIs returns a function that checks if the record level is in the given levels.
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.LevelIs(slog.LevelInfo)).
//	    Add(fileHandler, slogmulti.LevelIs(slog.LevelError)).
//	    Handler()
//
// Args:
//
//	levels: The levels to match
//
// Returns:
//
//	A function that checks if the record level is in the given levels
func LevelIs(levels ...slog.Level) func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		for _, level := range levels {
			if r.Level == level {
				return true
			}
		}
		return false
	}
}

// LevelIsNot returns a function that checks if the record level is not in the given levels.
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.LevelIsNot(slog.LevelInfo)).
//	    Add(fileHandler, slogmulti.LevelIsNot(slog.LevelError)).
//	    Handler()
//
// Args:
//
//	levels: The levels to check
//
// Returns:
//
//	A function that checks if the record level is not in the given levels
func LevelIsNot(levels ...slog.Level) func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		for _, level := range levels {
			if r.Level == level {
				return false
			}
		}
		return true
	}
}

// MessageIs returns a function that checks if the record message is equal to the given message.
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.MessageIs("database error")).
//	    Add(fileHandler, slogmulti.MessageIs("database error")).
//	    Handler()
//
// Args:
//
//	msg: The message to check
//
// Returns:
//
//	A function that checks if the record message is equal to the given message
func MessageIs(msg string) func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		return r.Message == msg
	}
}

// MessageIsNot returns a function that checks if the record message is not equal to the given message.
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.MessageIsNot("database error")).
//	    Add(fileHandler, slogmulti.MessageIsNot("database error")).
//	    Handler()
//
// Args:
//
//	msg: The message to check
//
// Returns:
//
//	A function that checks if the record message is not equal to the given message
func MessageIsNot(msg string) func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		return r.Message != msg
	}
}

// MessageContains returns a function that checks if the record message contains the given part.
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.MessageContains("database error")).
//	    Add(fileHandler, slogmulti.MessageContains("database error")).
//	    Handler()
//
// Args:
//
//	part: The part to check
//
// Returns:
//
//	A function that checks if the record message contains the given part
func MessageContains(part string) func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		return strings.Contains(r.Message, part)
	}
}

// MessageNotContains returns a function that checks if the record message does not contain the given part.
// Example usage:
//
//	r := slogmulti.Router().
//	    Add(consoleHandler, slogmulti.MessageNotContains("database error")).
//	    Add(fileHandler, slogmulti.MessageNotContains("database error")).
//	    Handler()
//
// Args:
//
//	part: The part to check
//
// Returns:
//
//	A function that checks if the record message does not contain the given part
func MessageNotContains(part string) func(ctx context.Context, r slog.Record) bool {
	return func(ctx context.Context, r slog.Record) bool {
		return !strings.Contains(r.Message, part)
	}
}
