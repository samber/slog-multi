package main

import (
	"fmt"
	"os"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slog"
)

func main() {
	// format go `error` type into an object {error: "*myCustomErrorType", message: "could not reach https://a.b/c"}
	errorFormattingMiddleware := slogmulti.NewHandleInlineMiddleware(errorFormattingMiddleware)

	// remove PII
	gdprMiddleware := NewGDPRMiddleware()

	sink := slog.HandlerOptions{}.NewJSONHandler(os.Stderr)

	logger := slog.New(
		slogmulti.
			Pipe(errorFormattingMiddleware).
			Pipe(gdprMiddleware).
			Handler(sink),
	)

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.String("email", "user-123"),
				slog.Time("created_at", time.Now()),
			),
		).
		With("environment", "dev").
		Error("A message",
			slog.String("foo", "bar"),
			slog.Any("error", fmt.Errorf("an error")))

	// output:
	// {
	//     "time":"2023-04-10T14:00:0.000000+00:00",
	//     "level":"ERROR",
	//     "msg":"A message",
	// 	   "user":{
	// 	       "id":"*******",
	// 	       "email":"*******",
	// 	       "created_at":"*******"
	//   	},
	//      "environment":"dev",
	//      "foo":"bar",
	// 	    "error":{
	// 	        "type":"*errors.errorString",
	// 	        "message":"an error"
	// 	    }
	// }
}
