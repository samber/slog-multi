
build:
	go build -v ./...

test:
	go test -race -v ./...
watch-test:
	reflex -t 50ms -s -- sh -c 'gotest -race -v ./...'

bench:
	go test -benchmem -count 3 -bench ./...
watch-bench:
	reflex -t 50ms -s -- sh -c 'go test -benchmem -count 3 -bench ./...'

FUZZTIME ?= 10s

fuzz:
	go test -fuzz=FuzzFanoutHandle -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzFailoverHandle -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzPoolHandle -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzFirstMatchHandle -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzPipeHandle -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzRecoveryHandle -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzAttrValueIs -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzRouterWithAttrs -fuzztime=$(FUZZTIME) ./...
	go test -fuzz=FuzzRouterPredicates -fuzztime=$(FUZZTIME) ./...

coverage:
	go test -v -coverprofile=cover.out -covermode=atomic ./...
	go tool cover -html=cover.out -o cover.html

tools:
	go install github.com/cespare/reflex@latest
	go install github.com/rakyll/gotest@latest
	go install github.com/psampaz/go-mod-outdated@latest
	go install github.com/jondot/goweight@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go get -t -u golang.org/x/tools/cmd/cover
	go install github.com/sonatype-nexus-community/nancy@latest
	go mod tidy

lint:
	golangci-lint run --timeout 60s --max-same-issues 50 ./...
lint-fix:
	golangci-lint run --timeout 60s --max-same-issues 50 --fix ./...

audit:
	go list -json -m all | nancy sleuth

outdated:
	go list -u -m -json all | go-mod-outdated -update -direct

weight:
	goweight
