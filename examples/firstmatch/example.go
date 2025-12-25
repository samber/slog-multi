package main

import (
	"context"

	"log/slog"

	slogmulti "github.com/samber/slog-multi"
	slogslack "github.com/samber/slog-slack"
)

func main() {
	queryLogLevel := slog.LevelDebug
	requestLogLevel := slog.LevelError
	influxdbLogLevel := slog.LevelInfo
	logLevel := slog.LevelError

	queryChannel := slogslack.Option{Level: queryLogLevel, WebhookURL: "xxx", Channel: "db queries"}.NewSlackHandler()
	requestChannel := slogslack.Option{Level: requestLogLevel, WebhookURL: "xxx", Channel: "service requests"}.NewSlackHandler()
	influxdbChannel := slogslack.Option{Level: influxdbLogLevel, WebhookURL: "xxx", Channel: "influxdb metrics"}.NewSlackHandler()
	fallbackChannel := slogslack.Option{Level: logLevel, WebhookURL: "xxx", Channel: "logs"}.NewSlackHandler()

	logger := slog.New(
		slogmulti.Router().
			Add(influxdbChannel, slogmulti.AttrValueIs("scope", "influx")).
			Add(queryChannel, slogmulti.AttrKindIs("query", slog.KindString, "args", slog.KindAny)).
			Add(requestChannel, slogmulti.AttrKindIs("method", slog.KindString, "body", slog.KindAny)).
			Add(fallbackChannel).
			FirstMatch().
			Handler(),
	)

	logger.Debug("Executing SQL query", "query", "SELECT * FROM users", "args", []int{1, 2, 3})
	logger.Error("Incoming request failed", "method", "POST", "body", "{'name':'test'}")
	logger.Error("An unexpected error occurred")

	influxLogger := logger.With("scope", "influx")

	// influx.NewClient(influxLogger) ...
	influxLogger.Info("InfluxDB client initialized")
}
