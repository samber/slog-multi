package slogmulti

import (
	"golang.org/x/exp/slog"
)

// Middleware defines the handler used by slog.Handler as return value.
type Middleware func(slog.Handler) slog.Handler
