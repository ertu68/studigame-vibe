FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
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
