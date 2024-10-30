FROM alpine:latest

COPY go-ddg-chat-api /app/go-ddg-chat-api
COPY config.toml /app/config.toml

ENTRYPOINT ["/app/go-ddg-chat-api", "run", "/app/config.toml"]