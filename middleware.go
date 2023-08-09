package slogmulti

import (
	"log/slog"
)

// Middleware defines the handler used by slog.Handler as return value.
type Middleware func(slog.Handler) slog.Handler
