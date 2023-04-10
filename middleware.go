package slogmulti

import (
	"golang.org/x/exp/slog"
)

type Middleware func(slog.Handler) slog.Handler
