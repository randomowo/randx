GO ?= go
GOLANGCI_LINT ?= golangci-lint

.PHONY: all lint test hooks

all: lint test

lint:
	$(GOLANGCI_LINT) run ./...

test:
	$(GO) test -race ./...

hooks:
	git config core.hooksPath .githooks
