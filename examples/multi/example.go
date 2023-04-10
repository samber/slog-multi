package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"golang.org/x/exp/slog"
)

func connectLogstash() *net.TCPConn {
	// ncat -l 4242 -k
	addr, err := net.ResolveTCPAddr("tcp", "localhost:4242")
	if err != nil {
		log.Fatal("TCP connection failed:", err.Error())
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal("TCP connection failed:", err.Error())
	}

	return conn
}

func main() {
	logstash := connectLogstash()
	stderr := os.Stderr

	logger := slog.New(
		slogmulti.NewMultiHandler(
			slog.HandlerOptions{}.NewJSONHandler(logstash),
			slog.HandlerOptions{}.NewTextHandler(stderr),
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

	// stderr output:
	// time=2023-04-10T14:00:0.000000+00:00 level=ERROR msg="A message" user.id=user-123 user.created_at=2023-04-10T14:00:0.000000+00:00 environment=dev error="an error"

	// netcat output:
	// {
	// 	"time":"2023-04-10T14:00:0.000000+00:00",
	// 	"level":"ERROR",
	// 	"msg":"A message",
	// 	"user":{
	// 		"id":"user-123",
	// 		"created_at":"2023-04-10T14:00:0.000000+00:00"
	// 	},
	// 	"environment":"dev",
	// 	"error":"an error"
	// }
}
