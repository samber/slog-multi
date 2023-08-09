package main

import (
	"context"
	"reflect"

	"log/slog"
)

func errorFormattingMiddleware(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
	attrs := []slog.Attr{}

	record.Attrs(func(attr slog.Attr) bool {
		key := attr.Key
		value := attr.Value
		kind := attr.Value.Kind()

		if key == "error" && kind == slog.KindAny {
			if err, ok := value.Any().(error); ok {
				errType := reflect.TypeOf(err).String()
				msg := err.Error()

				attrs = append(
					attrs,
					slog.Group("error",
						slog.String("type", errType),
						slog.String("message", msg),
					),
				)
			} else {
				attrs = append(attrs, attr)
			}
		} else {
			attrs = append(attrs, attr)
		}

		return true
	})

	// new record with formatted error
	record = slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.AddAttrs(attrs...)

	return next(ctx, record)
}
