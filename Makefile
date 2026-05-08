BIN          := movidesk-cli
PKG          := github.com/jonasandre/movidesk-cli
VERSION_PKG  := $(PKG)/internal/version
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE         ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS      := -s -w \
                -X $(VERSION_PKG).Version=$(VERSION) \
                -X $(VERSION_PKG).Commit=$(COMMIT) \
                -X $(VERSION_PKG).Date=$(DATE)

.PHONY: all build test lint fmt vet tidy run clean install release-snapshot

all: lint test build

build:
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BIN) ./cmd/$(BIN)

install:
	go install -trimpath -ldflags '$(LDFLAGS)' ./cmd/$(BIN)

test:
	go test -race -count=1 ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

lint:
	@command -v golangci-lint >/dev/null || { echo "golangci-lint not installed (https://golangci-lint.run)"; exit 1; }
	golangci-lint run ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

run: build
	./bin/$(BIN) $(ARGS)

clean:
	rm -rf bin dist coverage.out

release-snapshot:
	@command -v goreleaser >/dev/null || { echo "goreleaser not installed (https://goreleaser.com)"; exit 1; }
	goreleaser release --snapshot --clean
