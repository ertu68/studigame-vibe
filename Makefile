.PHONY: run build test css watch

TAILWIND_VERSION := v4.3.3
TAILWIND_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/macos/')
TAILWIND_ARCH := $(shell uname -m | sed 's/x86_64/x64/;s/aarch64/arm64/')

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

# Downloaded on demand: ~80 MB and platform-specific, deshalb nicht im Repo.
tailwindcss:
	curl -fsSLo $@ https://github.com/tailwindlabs/tailwindcss/releases/download/$(TAILWIND_VERSION)/tailwindcss-$(TAILWIND_OS)-$(TAILWIND_ARCH)
	chmod +x $@

css: tailwindcss
	./tailwindcss -i static/css/input.css -o static/css/app.css --minify

watch: tailwindcss
	./tailwindcss -i static/css/input.css -o static/css/app.css --watch

test:
	go test ./...
