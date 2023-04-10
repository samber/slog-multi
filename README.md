
# slog: handler chaining and broadcasting

[![tag](https://img.shields.io/github/tag/samber/slog-multi.svg)](https://github.com/samber/slog-multi/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.20-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/slog-multi?status.svg)](https://pkg.go.dev/github.com/samber/slog-multi)
![Build Status](https://github.com/samber/slog-multi/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/slog-multi)](https://goreportcard.com/report/github.com/samber/slog-multi)
[![Coverage](https://img.shields.io/codecov/c/github/samber/slog-multi)](https://codecov.io/gh/samber/slog-multi)
[![Contributors](https://img.shields.io/github/contributors/samber/slog-multi)](https://github.com/samber/slog-multi/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/slog-multi)](./LICENSE)

Design workflows of [slog](https://pkg.go.dev/golang.org/x/exp/slog) handlers:
- **fanout**: distribute `log.Record` to multiple `slog.Handler` in parallel
- **pipelining**: rewrite `log.Record` on the fly (eg: for privacy reason)

![workflow example](./images/workflow.png)

## 🚀 Install

```sh
go get github.com/samber/slog-multi
```

**Compatibility**: go >= 1.20.1

This library is v0 and follows SemVer strictly. On `slog` final release (go 1.21), this library will go v1.

No breaking changes will be made to exported APIs before v1.0.0.

## 💡 Usage

GoDoc: [https://godoc.org/github.com/samber/slog-multi](https://godoc.org/github.com/samber/slog-multi)

### Fanout: `slogmulti.Multi()`

Distribute logs to multiple `slog.Handler` in parallel.

```go
import (
    slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slog"
)

func main() {
    logstash, _ := net.Dial("tcp", "logstash.acme:4242")
    stderr := os.Stderr

    logger := slog.New(slogmulti.NewMultiHandler(
        slog.HandlerOptions{}.NewJSONHandler(logstash),  // first handler: logstash over tcp
        slog.HandlerOptions{}.NewTextHandler(stderr),    // second handler: stderr
        // ...
    ))

    logger.
        With(
            slog.Group("user",
                slog.String("id", "user-123"),
                slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
            ),
        ).
        With("environment", "dev").
        With("error", fmt.Errorf("an error")).
        Error("A message")
}
```

Stderr output:

```
time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="A message" user.id=user-123 user.created_at=2023-04-10T14:00:0.000000+00:00 environment=dev error="an error"
```

Netcat output:

```json
{
	"time":"2023-04-10T14:00:0.000000+00:00",
	"level":"ERROR",
	"msg":"A message",
	"user":{
		"id":"user-123",
		"created_at":"2023-04-10T14:00:0.000000+00:00"
	},
	"environment":"dev",
	"error":"an error"
}
```

### Chaining: `slogmulti.Pipe()`

Rewrite `log.Record` on the fly (eg: for privacy reason).

```go
func main() {
	// first middleware: format go `error` type into an object {error: "*myCustomErrorType", message: "could not reach https://a.b/c"}
	errorFormattingMiddleware := slogmulti.NewHandleInlineMiddleware(errorFormattingMiddleware)

	// second middleware: remove PII
	gdprMiddleware := NewGDPRMiddleware()

    // final handler
	sink := slog.HandlerOptions{}.NewJSONHandler(os.Stderr)

	logger := slog.New(
		slogmulti.
			Pipe(errorFormattingMiddleware).
			Pipe(gdprMiddleware).
            // ...
			Handler(sink),
	)

	logger.
		With(
			slog.Group("user",
				slog.String("id", "user-123"),
				slog.String("email", "user-123"),
				slog.Time("created_at", time.Now().AddDate(0, 0, -1)),
			),
		).
		With("environment", "dev").
		Error("A message",
			slog.String("foo", "bar"),
			slog.Any("error", fmt.Errorf("an error")))
}
```

Stderr output:

```json
{
    "time":"2023-04-10T14:00:0.000000+00:00",
    "level":"ERROR",
    "msg":"A message",
    "user":{
        "id":"*******",
        "email":"*******",
        "created_at":"*******"
  	},
    "environment":"dev",
    "foo":"bar",
    "error":{
        "type":"*myCustomErrorType",
        "message":"an error"
    }
}
```

#### Custom middleware

Middleware must match the following prototype:

```go
type Middleware func(slog.Handler) slog.Handler
```

The example above uses:
- a custom middleware, [see here](./examples/pipe/gdpr.go)
- an inline middleware, [see here](./examples/pipe/errors.go)

Note: `WithAttrs` and `WithGroup` methods of custom middleware must return a new instance, instead of `this`.

#### Inline middleware

An "inline middleware" (aka. lambda), is a shortcut to middleware implementation, that hooks a single method and proxies others.

```go
// hook `logger.Enabled` method
mdw := slogmulti.NewEnabledInlineMiddleware(func(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool{
    // [...]
    return next(ctx, level)
})
```

```go
// hook `logger.Handle` method
mdw := slogmulti.NewHandleInlineMiddleware(func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
    // [...]
    return next(ctx, record)
})
```

```go
// hook `logger.WithAttrs` method
mdw := slogmulti.NewWithAttrsInlineMiddleware(func(attrs []slog.Attr, next func([]slog.Attr) slog.Handler) slog.Handler{
    // [...]
    return next(attrs)
})
```

```go
// hook `logger.WithGroup` method
mdw := slogmulti.NewWithGroupInlineMiddleware(func(name string, next func(string) slog.Handler) slog.Handler{
    // [...]
    return next(name)
})
```

A super inline middleware that hooks all methods.

> Warning: you would rather implement your own middleware.

```go
mdw := slogmulti.NewInlineMiddleware(
    func(ctx context.Context, level slog.Level, next func(context.Context, slog.Level) bool) bool{
        // [...]
        return next(ctx, level)
    },
    func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error{
        // [...]
        return next(ctx, record)
    },
    func(attrs []slog.Attr, next func([]slog.Attr) slog.Handler) slog.Handler{
        // [...]
        return next(attrs)
    },
    func(name string, next func(string) slog.Handler) slog.Handler{
        // [...]
        return next(name)
    },
)
```

## 🤝 Contributing

- Ping me on twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/slog-multi)
- Fix [open issues](https://github.com/samber/slog-multi/issues) or request new features

Don't hesitate ;)

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/slog-multi)

## 💫 Show your support

Give a ⭐️ if this project helped you!

![support](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2023 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.