.PHONY: help setup test test-cover lint format

# Function to check if a command exists and install tools if not
# Usage: $(call check-command,command-name)
define check-command
	@if ! command -v $(1) >/dev/null 2>&1; then \
		echo "$(1) not found, installing tools..."; \
		$(MAKE) tools; \
	fi
endef

help:
	@echo "Makefile commands:"
	@echo "  setup              Installs all needed tools and dependecies"
	@echo "  test               Test the package with the defined tests"
	@echo "  test-cover         Test the package with the defined tests and generate a coverage report"
	@echo "  lint               Run golangci-lint and fix autofixable issues"
	@echo "  format             Run golangci-lint formaters only and fix autofixable issues"
	
setup:
	./scripts/install.sh
	go mod download

test:
	go test

test-cover:
	go test -covermode=count --coverprofile=coverage.out
	go tool cover -html coverage.out -o coverage.html

lint:
	$(call check-command,golangci-lint)
	golangci-lint run --fix
	
format:
	$(call check-command,golangci-lint)
	golangci-lint fmt
