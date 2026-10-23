module github.com/samber/slog-multi/examples/router

go 1.25

require (
	github.com/samber/slog-multi v1.0.0
	github.com/samber/slog-slack v1.0.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/samber/slog-common v0.22.0 // indirect
	github.com/slack-go/slack v0.23.1 // indirect
	golang.org/x/text v0.22.0 // indirect
)

replace github.com/samber/slog-multi => ../../
