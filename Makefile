SHELL := /bin/sh

GO ?= go

.PHONY: help check-tools fmt fmt-check vet test test-race check

help: ## Show available commands.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z_-]+:.*## / {printf "%-18s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check-tools: ## Verify the pinned local toolchain.
	@./scripts/check-toolchain.sh

fmt: ## Format handwritten Go source.
	@files="$$(find . -type f -name '*.go' -not -path './gen/*')"; \
	if [ -n "$$files" ]; then gofmt -w $$files; fi

fmt-check: ## Fail if handwritten Go source is not formatted.
	@files="$$(find . -type f -name '*.go' -not -path './gen/*')"; \
	if [ -n "$$files" ]; then \
		unformatted="$$(gofmt -l $$files)"; \
		if [ -n "$$unformatted" ]; then \
			echo "Unformatted Go files:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi; \
	fi

vet: ## Run Go's static analyzer.
	@$(GO) vet ./...

test: ## Run the fast test suite.
	@$(GO) test ./...

test-race: ## Run tests with the race detector.
	@$(GO) test -race ./...

check: fmt-check vet test test-race ## Run the Milestone 0 quality gate.

