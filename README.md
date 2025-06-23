
# slog: Handler chaining, fanout, routing, failover, load balancing...

[![tag](https://img.shields.io/github/tag/samber/slog-multi.svg)](https://github.com/samber/slog-multi/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.21-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/slog-multi?status.svg)](https://pkg.go.dev/github.com/samber/slog-multi)
![Build Status](https://github.com/samber/slog-multi/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/slog-multi)](https://goreportcard.com/report/github.com/samber/slog-multi)
[![Coverage](https://img.shields.io/codecov/c/github/samber/slog-multi)](https://codecov.io/gh/samber/slog-multi)
[![Contributors](https://img.shields.io/github/contributors/samber/slog-multi)](https://github.com/samber/slog-multi/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/slog-multi)](./LICENSE)

Design workflows of [slog](https://pkg.go.dev/log/slog) handlers:
- **Fanout**: distribute `log.Record` to multiple `slog.Handler` in parallel
- **Pipe**: rewrite `log.Record` on the fly (eg: for privacy reasons)
- **Router**: forward `log.Record` to all matching `slog.Handler`
- **Failover**: forward `log.Record` to the first available `slog.Handler`
- **Pool**: increase log bandwidth by sending `log.Record` to a pool of `slog.Handler`
- **RecoverHandlerError**: catch panics and errors from handlers

Here is a simple workflow with both pipeline and fanout:

![workflow example with pipeline and fanout](./images/workflow.png)

Middlewares:
- [Inline handler](#inline-handler): a shortcut to implement `slog.Handler`
- [Inline middleware](#inline-middleware): a shortcut to implement `slogmulti.Middleware`

<div align="center">
  <hr>
  <sup><b>Sponsored by:</b></sup>
  <br>
  <a href="https://quickwit.io?utm_campaign=github_sponsorship&utm_medium=referral&utm_content=samber-slog-multi&utm_source=github">
    <div>
      <img src="https://github.com/samber/oops/assets/2951285/49aaaa2b-b8c6-4f21-909f-c12577bb6a2e" width="240" alt="Quickwit">
    </div>
    <div>
      Cloud-native search engine for observability - An OSS alternative to Splunk, Elasticsearch, Loki, and Tempo.
    </div>
  </a>
  <hr>
</div>

**See also:**

- [slog-multi](https://github.com/samber/slog-multi): `slog.Handler` chaining, fanout, routing, failover, load balancing...
- [slog-formatter](https://github.com/samber/slog-formatter): `slog` attribute formatting
- [slog-sampling](https://github.com/samber/slog-sampling): `slog` sampling policy
- [slog-mock](https://github.com/samber/slog-mock): `slog.Handler` for test purposes

**HTTP middlewares:**

- [slog-gin](https://github.com/samber/slog-gin): Gin middleware for `slog` logger
- [slog-echo](https://github.com/samber/slog-echo): Echo middleware for `slog` logger
- [slog-fiber](https://github.com/samber/slog-fiber): Fiber middleware for `slog` logger
- [slog-chi](https://github.com/samber/slog-chi): Chi middleware for `slog` logger
- [slog-http](https://github.com/samber/slog-http): `net/http` middleware for `slog` logger

**Loggers:**

- [slog-zap](https://github.com/samber/slog-zap): A `slog` handler for `Zap`
- [slog-zerolog](https://github.com/samber/slog-zerolog): A `slog` handler for `Zerolog`
- [slog-logrus](https://github.com/samber/slog-logrus): A `slog` handler for `Logrus`

**Log sinks:**

- [slog-datadog](https://github.com/samber/slog-datadog): A `slog` handler for `Datadog`
- [slog-betterstack](https://github.com/samber/slog-betterstack): A `slog` handler for `Betterstack`
- [slog-rollbar](https://github.com/samber/slog-rollbar): A `slog` handler for `Rollbar`
- [slog-loki](https://github.com/samber/slog-loki): A `slog` handler for `Loki`
- [slog-sentry](https://github.com/samber/slog-sentry): A `slog` handler for `Sentry`
- [slog-syslog](https://github.com/samber/slog-syslog): A `slog` handler for `Syslog`
- [slog-logstash](https://github.com/samber/slog-logstash): A `slog` handler for `Logstash`
- [slog-fluentd](https://github.com/samber/slog-fluentd): A `slog` handler for `Fluentd`
- [slog-graylog](https://github.com/samber/slog-graylog): A `slog` handler for `Graylog`
- [slog-quickwit](https://github.com/samber/slog-quickwit): A `slog` handler for `Quickwit`
- [slog-slack](https://github.com/samber/slog-slack): A `slog` handler for `Slack`
- [slog-telegram](https://github.com/samber/slog-telegram): A `slog` handler for `Telegram`
- [slog-mattermost](https://github.com/samber/slog-mattermost): A `slog` handler for `Mattermost`
- [slog-microsoft-teams](https://github.com/samber/slog-microsoft-teams): A `slog` handler for `Microsoft Teams`
- [slog-webhook](https://github.com/samber/slog-webhook): A `slog` handler for `Webhook`
- [slog-kafka](https://github.com/samber/slog-kafka): A `slog` handler for `Kafka`
- [slog-nats](https://github.com/samber/slog-nats): A `slog` handler for `NATS`
- [slog-parquet](https://github.com/samber/slog-parquet): A `slog` handler for `Parquet` + `Object Storage`
- [slog-channel](https://github.com/samber/slog-channel): A `slog` handler for Go channels

## 🚀 Install

```sh
go get github.com/samber/slog-multi
```

**Compatibility**: go >= 1.21

No breaking changes will be made to exported APIs before v2.0.0.

> [!WARNING]
> Use this library carefully, log processing can be very costly (!)
> 
> Excessive logging —with multiple processing steps and destinations— can introduce significant overhead, which is generally undesirable in performance-critical paths. Logging is always expensive, and sometimes, metrics or a sampling strategy are cheaper. The library itself does not generate extra load.

## 💡 Usage

GoDoc: [https://pkg.go.dev/github.com/samber/slog-multi](https://pkg.go.dev/github.com/samber/slog-multi)

### Broadcast: `slogmulti.Fanout()`

Distribute logs to multiple `slog.Handler` in parallel.

```go
import (
    slogmulti "github.com/samber/slog-multi"
    "log/slog"
)

func main() {
    logstash, _ := net.Dial("tcp", "logstash.acme:4242")    // use github.com/netbrain/goautosocket for auto-reconnect
    stderr := os.Stderr

    logger := slog.New(
        slogmulti.Fanout(
            slog.NewJSONHandler(logstash, &slog.HandlerOptions{}),  // pass to first handler: logstash over tcp
            slog.NewTextHandler(stderr, &slog.HandlerOptions{}),    // then to second handler: stderr
            // ...
        ),
    )

    logger.
        With(
            slog.Group("user",
                slog.String("id", "user-123"),
                slog.Time("created_at", time.Now()),
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

### Routing: `slogmulti.Router()`

Distribute logs to all matching `slog.Handler` in parallel.

```go
import (
    slogmulti "github.com/samber/slog-multi"
    slogslack "github.com/samber/slog-slack"
    "log/slog"
)

func main() {
    slackChannelUS := slogslack.Option{Level: slog.LevelError, WebhookURL: "xxx", Channel: "supervision-us"}.NewSlackHandler()
    slackChannelEU := slogslack.Option{Level: slog.LevelError, WebhookURL: "xxx", Channel: "supervision-eu"}.NewSlackHandler()
    slackChannelAPAC := slogslack.Option{Level: slog.LevelError, WebhookURL: "xxx", Channel: "supervision-apac"}.NewSlackHandler()

    logger := slog.New(
        slogmulti.Router().
            Add(slackChannelUS, recordMatchRegion("us")).
            Add(slackChannelEU, recordMatchRegion("eu")).
            Add(slackChannelAPAC, recordMatchRegion("apac")).
            Handler(),
    )

    logger.
        With("region", "us").
        With("pool", "us-east-1").
        Error("Server desynchronized")
}

func recordMatchRegion(region string) func(ctx context.Context, r slog.Record) bool {
    return func(ctx context.Context, r slog.Record) bool {
        ok := false

        r.Attrs(func(attr slog.Attr) bool {
            if attr.Key == "region" && attr.Value.Kind() == slog.KindString && attr.Value.String() == region {
                ok = true
                return false
            }

            return true
        })

        return ok
    }
}
```

### Failover: `slogmulti.Failover()`

List multiple targets for a `slog.Record` instead of retrying on the same unavailable log management system.

```go
import (
    "net"
    slogmulti "github.com/samber/slog-multi"
    "log/slog"
)


func main() {
    // ncat -l 1000 -k
    // ncat -l 1001 -k
    // ncat -l 1002 -k

    // list AZs
    // use github.com/netbrain/goautosocket for auto-reconnect
    logstash1, _ := net.Dial("tcp", "logstash.eu-west-3a.internal:1000")
    logstash2, _ := net.Dial("tcp", "logstash.eu-west-3b.internal:1000")
    logstash3, _ := net.Dial("tcp", "logstash.eu-west-3c.internal:1000")

    logger := slog.New(
        slogmulti.Failover()(
            slog.HandlerOptions{}.NewJSONHandler(logstash1, nil),    // send to this instance first
            slog.HandlerOptions{}.NewJSONHandler(logstash2, nil),    // then this instance in case of failure
            slog.HandlerOptions{}.NewJSONHandler(logstash3, nil),    // and finally this instance in case of double failure
        ),
    )

    logger.
        With(
            slog.Group("user",
                slog.String("id", "user-123"),
                slog.Time("created_at", time.Now()),
            ),
        ).
        With("environment", "dev").
        With("error", fmt.Errorf("an error")).
        Error("A message")
}
```

### Load balancing: `slogmulti.Pool()`

Increase log bandwidth by sending `log.Record` to a pool of `slog.Handler`.

```go
import (
    "net"
    slogmulti "github.com/samber/slog-multi"
    "log/slog"
)

func main() {
    // ncat -l 1000 -k
    // ncat -l 1001 -k
    // ncat -l 1002 -k

    // list AZs
    // use github.com/netbrain/goautosocket for auto-reconnect
    logstash1, _ := net.Dial("tcp", "logstash.eu-west-3a.internal:1000")
    logstash2, _ := net.Dial("tcp", "logstash.eu-west-3b.internal:1000")
    logstash3, _ := net.Dial("tcp", "logstash.eu-west-3c.internal:1000")

    logger := slog.New(
        slogmulti.Pool()(
            // a random handler will be picked
            slog.HandlerOptions{}.NewJSONHandler(logstash1, nil),
            slog.HandlerOptions{}.NewJSONHandler(logstash2, nil),
            slog.HandlerOptions{}.NewJSONHandler(logstash3, nil),
        ),
    )

    logger.
        With(
            slog.Group("user",
                slog.String("id", "user-123"),
                slog.Time("created_at", time.Now()),
            ),
        ).
        With("environment", "dev").
        With("error", fmt.Errorf("an error")).
        Error("A message")
}
```

### Recover errors: `slog.RecoverHandlerError()`

Returns a `slog.Handler` that recovers from panics or error of the chain of handlers.

```go
import (
	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
	"log/slog"
)

recovery := slogmulti.RecoverHandlerError(
    func(ctx context.Context, record slog.Record, err error) {
        // will be called only if subsequent handlers fail or return an error
        log.Println(err.Error())
    },
)
sink := NewSinkHandler(...)

logger := slog.New(
    slogmulti.
        Pipe(recovery).
        Handler(sink),
)

err := fmt.Errorf("an error")
logger.Error("a message",
    slog.Any("very_private_data", "abcd"),
    slog.Any("user", user),
    slog.Any("err", err))

// outputs:
// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="a message" error.message="an error" error.type="*errors.errorString" user="John doe" very_private_data="********"
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
    sink := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{})

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
                slog.Time("created_at", time.Now()),
            ),
        ).
        With("environment", "dev").
        Error("A message",
            slog.String("foo", "bar"),
            slog.Any("error", fmt.Errorf("an error")),
        )
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

#### Inline handler

An "inline handler" (aka. lambda), is a shortcut to implement `slog.Handler`, that hooks a single method and proxies others.

```go
mdw := slogmulti.NewHandleInlineHandler(
    // simulate "Handle()"
    func(ctx context.Context, groups []string, attrs []slog.Attr, record slog.Record) error {
        // [...]
        return nil
    },
)
```

```go
mdw := slogmulti.NewInlineHandler(
    // simulate "Enabled()"
    func(ctx context.Context, groups []string, attrs []slog.Attr, level slog.Level) bool {
        // [...]
        return true
    },
    // simulate "Handle()"
    func(ctx context.Context, groups []string, attrs []slog.Attr, record slog.Record) error {
        // [...]
        return nil
    },
)
```

#### Inline middleware

An "inline middleware" (aka. lambda), is a shortcut to implement middleware, that hooks a single method and proxies others.

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

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2023 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
