SHELL := /bin/sh

GO ?= go
BUF ?= buf

.PHONY: help check-tools fmt fmt-check vet test test-race proto-format proto-format-check proto-lint proto-generate proto-check check

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

proto-format: ## Format Protobuf contracts in place.
	@$(BUF) format -w

proto-format-check: ## Fail if Protobuf contracts are not formatted.
	@$(BUF) format --diff --exit-code

proto-lint: ## Lint Protobuf contracts.
	@$(BUF) lint

proto-generate: ## Generate Go and gRPC code from Protobuf contracts.
	@$(BUF) generate

proto-check: proto-format-check proto-lint ## Run contract formatting and lint gates.

check: fmt-check vet test test-race proto-check ## Run the repository quality gate.
