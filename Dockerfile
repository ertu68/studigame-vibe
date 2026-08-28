FROM golang:1.27-alpine AS build
ARG TARGETARCH
WORKDIR /src
RUN apk add --no-cache curl
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# The tailwindcss binary is platform-specific and not in the repo, so fetch the musl build here.
RUN ARCH="${TARGETARCH:-$(uname -m)}"; \
    case "$ARCH" in amd64|x86_64) ARCH=x64 ;; arm64|aarch64) ARCH=arm64 ;; esac; \
    curl -fsSLo /usr/local/bin/tailwindcss \
      "https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.3/tailwindcss-linux-${ARCH}-musl" && \
    chmod +x /usr/local/bin/tailwindcss && \
    tailwindcss -i static/css/input.css -o static/css/app.css --minify
# CGO_ENABLED=0 works because modernc.org/sqlite is pure Go.
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/server ./server
COPY --from=build /src/static ./static
USER app
EXPOSE 8080
ENTRYPOINT ["./server"]
