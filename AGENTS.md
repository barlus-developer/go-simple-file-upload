# AGENTS.md

Guidance for coding agents working in this repository.

## Project Overview

`go-simple-file-upload` is a small Go HTTP API using Gin, Zap, Viper, and a DDD-style package layout. There is no database — uploaded files are persisted directly on local disk. The API surface is intentionally minimal:

```http
GET /            health check
POST /files      upload a file (multipart form field "file")
GET /files       list stored files
GET /files/:name download a stored file
```

## Architecture

Keep changes aligned with the existing package boundaries:

- `cmd/server`: process entrypoint and graceful shutdown.
- `internal/bootstrap`: dependency wiring.
- `internal/domain`: domain models (`health`, `file`).
- `internal/application`: application services and use cases, including the `file.Storage` port.
- `internal/interfaces/http`: HTTP handlers, middleware, and router setup.
- `internal/infrastructure`: configuration, logging, and technical adapters — including the `storage.Local` disk adapter that implements the `file.Storage` port. There is no database adapter.

See `ARCHITECTURE.md` for Mermaid diagrams and request flow details.

## Development Commands

Run the service:

```sh
go run ./cmd/server
```

Run all tests:

```sh
go test ./...
```

Build the server:

```sh
go build ./cmd/server
```

Format changed Go files:

```sh
gofmt -w <files>
```

## Testing Expectations

When changing behavior, add or update tests near the affected package:

- Config and error behavior: `internal/infrastructure/config`.
- Application service behavior: `internal/application`.
- Local disk storage behavior: `internal/infrastructure/storage`.
- HTTP API behavior: `internal/interfaces/http/router`.
- HTTP middleware behavior: `internal/interfaces/http/middleware`.

Before committing, run:

```sh
go test ./...
go build ./cmd/server
```

## Configuration

Runtime defaults are built in. Local overrides can use `config.yaml` or environment variables with the `APP_` prefix, such as:

```sh
APP_APP_ENVIRONMENT=production APP_SERVER_PORT=3000 go run ./cmd/server
```

Do not commit local secrets or environment-specific configuration.

## Git Notes

- Keep commits focused and describe the behavior or documentation change.
- When an agent contributes to a commit, add that agent as a Git co-author using a `Co-authored-by:` trailer in the commit message.
- For Codex-authored changes, use `Co-authored-by: Codex <codex@openai.com>`.
- Do not remove generated project documentation unless the user asks.
- Do not overwrite local config files or unrelated user changes.
