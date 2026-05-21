.PHONY: help setup test test-cover lint format install-lefthook

help:
	@echo "Makefile commands:"
	@echo "  setup              Installs all needed tools and dependecies"
	@echo "  test               Test the package with the defined tests"
	@echo "  test-cover         Test the package with the defined tests and generate a coverage report"
	@echo "  lint               Run golangci-lint and fix autofixable issues"
	@echo "  format             Run golangci-lint formaters only and fix autofixable issues"
	@echo "  install-lefthook   Install lefthook git hooks"

setup:
	go mod download

test:
	go test

test-cover:
	go test -covermode=count --coverprofile=coverage.out
	go tool cover -html coverage.out -o coverage.html

lint:
	go tool golangci-lint run --fix

format:
	go tool golangci-lint fmt

install-lefthook:
	go tool leftook install
