.PHONY: run build css watch test vet fmt-check lint vuln ci help
.DEFAULT_GOAL := help

TAILWIND_VERSION := v4.3.3
TAILWIND_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/macos/')
TAILWIND_ARCH := $(shell uname -m | sed 's/x86_64/x64/;s/aarch64/arm64/')

run: # Start the server locally.
	set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/server

build: # Build the server binary.
	mkdir -p bin
	go build -o bin/server ./cmd/server

tailwindcss: # Downloaded on demand: ~80 MB and platform-specific.
	curl -fsSLo $@ https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$(TAILWIND_OS)-$(TAILWIND_ARCH)
	chmod +x $@

css: tailwindcss # Build the production CSS.
	./tailwindcss -i static/css/input.css -o static/css/app.css --minify

watch: tailwindcss # Rebuild CSS when source files change.
	./tailwindcss -i static/css/input.css -o static/css/app.css --watch

test: # Run Go tests.
	go test -race -v ./...

vet: # Check Go code for suspicious constructs.
	go vet ./...

fmt-check: # Fail if Go files need formatting.
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then \
		printf 'Go files need formatting:\n%s\n' "$$files"; \
		exit 1; \
	fi

lint: # Run golangci-lint.
	go tool golangci-lint run

vuln: # Check for reachable Go dependency vulnerabilities.
	go tool govulncheck ./...

ci: fmt-check vet lint test vuln # Run all local CI checks.

help: # List available targets and their descriptions.
	@awk -F ':.*# ' '/^[[:alnum:]_-]+:.*# / { printf "%-14s %s\n", $$1, $$2 }' $(MAKEFILE_LIST) | sort
