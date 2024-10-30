# Go DuckDuckGo Chat API

![DuckDuckGo Chat API ](https://duckduckgo.com/duckduckgo-help-pages/logo.v109.svg)

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/go-ddg-chat-api)](https://artifacthub.io/packages/search?repo=go-ddg-chat-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/nerdneilsfield/go-ddg-chat-api)](https://goreportcard.com/report/github.com/nerdneilsfield/go-ddg-chat-api)
[![Docker Image Size](https://img.shields.io/docker/image-size/nerdneils/go-ddg-chat-api)](https://hub.docker.com/r/nerdneils/go-ddg-chat-api)
[![Dockerhub Downloads Number](https://img.shields.io/docker/pulls/nerdneils/go-ddg-chat-api)](https://hub.docker.com/r/nerdneils/go-ddg-chat-api)
[![Github Downloads Number](https://img.shields.io/github/downloads/nerdneilsfield/go-ddg-chat-api/total)](https://github.com/nerdneilsfield/go-ddg-chat-api/releases)
[![GitHub Release](https://img.shields.io/github/release/nerdneilsfield/go-ddg-chat-api)](https://github.com/nerdneilsfield/go-ddg-chat-api/releases)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/nerdneilsfield/go-ddg-chat-api/goreleaser.yml?branch=master)](https://github.com/nerdneilsfield/go-ddg-chat-api/actions)

---

A Go implementation that provides an OpenAI-compatible API interface for DuckDuckGo's chat service.

## Features

- OpenAI-compatible API endpoints
- Support for streaming responses
- Multiple model mappings
- Token-based authentication
- Proxy support
- Health check endpoints
- CORS enabled

## Supported Models

```toml
"ddg/gpt-4o-mini" = "gpt-4o-mini"
"ddg/claude-3-haiku" = "claude-3-haiku-20240307"
"ddg/mixtral-8x7b" = "mistralai/Mixtral-8x7B-Instruct-v0.1"
"ddg/meta-Llama-3-1-70B-Instruct-Turbo" = "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo"
```

## Installation


- Install the binary:
```bash
go install github.com/nerdneilsfield/go-ddg-chat-api@latest
```

- Install from [Releases](https://github.com/nerdneilsfield/go-ddg-chat-api/releases)

- Install from docker:
```bash
# docker hub
docker pull nerdneils/go-ddg-chat-api
# ghcr
docker pull ghcr.io/nerdneilsfield/go-ddg-chat-api
```


## Configuration

Create a `config.toml` file:

```toml
port = 8085
host = "0.0.0.0"
user_agent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
tokens = ["duckduckgo-chat-api-token"]
ddg_chat_api_url = "https://duckduckgo.com/"

[model_mapping]
"ddg/gpt-4o-mini" = "gpt-4o-mini"
"ddg/claude-3-haiku" = "claude-3-haiku-20240307"
"ddg/mixtral-8x7b" = "mistralai/Mixtral-8x7B-Instruct-v0.1"
"ddg/meta-Llama-3-1-70B-Instruct-Turbo" = "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo"
```

## Usage

Run the server:

```bash
go-ddg-chat-api run config.toml
```

Check version:

```bash
go-ddg-chat-api version
```

Debug output:

```bash
go-ddg-chat-api run config.toml -v
```

Run with docker:

```bash
# normal output
docker run -d --name go-ddg-chat-api -p 8085:8085  -v $(pwd)/config.toml:/app/config.toml nerdneils/go-ddg-chat-api
# with proxy
docker run -d --name go-ddg-chat-api -p 8085:8085 -e HTTPS_PROXY=http://your-proxy-url:8080 -v $(pwd)/config.toml:/app/config.toml nerdneils/go-ddg-chat-api
# debug output
docker run -d --name go-ddg-chat-api -p 8085:8085  -v $(pwd)/config.toml:/app/config.toml nerdneils/go-ddg-chat-api /app/go-ddg-chat-api run /app/config.toml -v
```

## API Endpoints

- `GET /v1/models` - List available models
- `POST /v1/chat/completions` - Create chat completion
- `DELETE /v1/chat/completions/{id}` - Delete chat completion
- `GET /live` - Liveness probe
- `GET /ready` - Readiness probe

### Chat Completion Example

```bash
curl -X POST http://localhost:8085/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "model": "ddg/claude-3-haiku",
    "messages": [
      {"role": "user", "content": "Hello, how are you?"}
    ],
    "stream": true
  }'
```

## Environment Variables

- `HTTPS_PROXY` or `https_proxy` - Proxy server URL (optional)

## Development

Requirements:
- Go 1.21 or higher

Build from source:

```bash
git clone https://github.com/nerdneilsfield/go-ddg-chat-api
cd go-ddg-chat-api
go build
```

### The Chat Completion Process

```
                                     +-----------------+
                                     |                 |
                                     |  Client Request |
                                     |                 |
                                     +--------+--------+
                                              |
                                              v
+----------------+    +-----------------+    +----------------------+
|                |    |                 |    |                      |
| Auth Middleware+--->+ Chat Completion +--->+ Generate UUID        |
|                |    | Handler         |    | for Conversation     |
+----------------+    +-----------------+    +----------------------+
                                              |
                                              v
                     +-----------------+    +----------------------+
                     |                 |    |                      |
                     | Stream Response |<---+ Create Response      |
                     | Channel         |    | Channel             |
                     +--------+--------+    +----------------------+
                              |
                              v
+-----------------+    +---------------------+    +------------------+
|                 |    |                     |    |                  |
| Get VQD Token   +<---+ Chat with DDG API   +--->+ Process Messages |
|                 |    |                     |    |                  |
+-----------------+    +---------------------+    +------------------+
         |                      |                          |
         v                      v                          v
+-----------------+    +---------------------+    +------------------+
|                 |    |                     |    |                  |
| Random UserAgent|    | Stream DDG Response |    | Update History   |
|                 |    |                     |    |                  |
+-----------------+    +---------------------+    +------------------+
                              |
                              v
                     +-----------------+
                     |                 |
                     | Client Response |
                     |                 |
                     +-----------------+
```

## License

MIT License


## Acknowledgments

References: 
- [keyless-gpt-wrapper-api](https://github.com/callbacked/keyless-gpt-wrapper-api)

This project provides an OpenAI-compatible interface for DuckDuckGo's chat service. It is not affiliated with or endorsed by DuckDuckGo or OpenAI.


## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=nerdneilsfield/go-ddg-chat-api&type=Date)](https://star-history.com/#nerdneilsfield/go-ddg-chat-api&Date)
