package main

import (
	"fmt"
	"net"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slog"
)

func main() {
	// ncat -l 1000 -k
	// ncat -l 1001 -k
	// ncat -l 1002 -k

	// logstash1, err := gas.Dial("tcp", "logstash.eu-west-3a.internal:1000")
	// logstash2, err := gas.Dial("tcp", "logstash.eu-west-3b.internal:1000")
	// logstash3, err := gas.Dial("tcp", "logstash.eu-west-3c.internal:1000")

	logstash1, _ := net.Dial("tcp", "localhost:1000")
	logstash2, _ := net.Dial("tcp", "localhost:1001")
	logstash3, _ := net.Dial("tcp", "localhost:1002")

	logger := slog.New(
		slogmulti.Either(
			slog.HandlerOptions{}.NewJSONHandler(logstash1),
			slog.HandlerOptions{}.NewJSONHandler(logstash2),
			slog.HandlerOptions{}.NewJSONHandler(logstash3),
		),
	)

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
