.PHONY: build test clean

VERSION := $(shell cat VERSION)
LDFLAGS := -X main.version=$(VERSION)
BIN := bin/tmux-jump

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o $(BIN) .
	@echo "built $(BIN) (v$(VERSION))"

test:
	go test ./...

clean:
	rm -rf bin tmux-jump
