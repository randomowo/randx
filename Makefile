GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOLANGCI_LINT_VERSION ?= v2.12.2

.PHONY: all lint test hooks install-lint

all: lint test

lint: install-lint
	$(GOLANGCI_LINT) run ./...

install-lint:
	@command -v $(GOLANGCI_LINT) >/dev/null || \
		$(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

test:
	$(GO) test -race ./...

hooks:
	git config core.hooksPath .githooks
