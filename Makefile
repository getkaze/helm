VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)
BIN     := bin/helm

.PHONY: build build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-all clean test

build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN) .

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN)-linux-amd64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BIN)-linux-arm64 .

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN)-darwin-amd64 .

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BIN)-darwin-arm64 .

build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64

clean:
	rm -rf bin/

test:
	go test ./... -race -count=1
