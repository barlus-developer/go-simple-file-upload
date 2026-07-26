# go-simple-file-upload

A small HTTP API built with Gin, Zap, Viper, and a DDD-style package layout. It exposes only file
upload and file retrieval — **there is no database**; uploaded files are stored directly on local
disk.

```http
GET /              health check
POST /files        upload a file
GET /files         list stored files
GET /files/:name   download a stored file
```

## Tech Stack

- Gin for HTTP routing and middleware.
- Zap for structured logging.
- Viper for configuration from defaults, files, and environment variables.
- Standard Go `net/http` server with graceful shutdown.
- Local filesystem for file storage — no database.

## Project Structure

```text
.
├── cmd/server                          # Application entrypoint
├── config.example.yaml                 # Example local configuration
├── internal/application/file           # File upload/download use cases + Storage port
├── internal/application/health         # Health use case
├── internal/bootstrap                  # Dependency wiring
├── internal/domain/file                # File domain model
├── internal/domain/health              # Health domain model
├── internal/infrastructure/config      # Configuration loading
├── internal/infrastructure/logger      # Logger setup
├── internal/infrastructure/storage     # Local disk storage adapter (implements Storage port)
└── internal/interfaces/http            # HTTP handlers, middleware, and routers
```

## Requirements

- Go 1.26.5 or compatible with the module version in `go.mod`.

## Run

```sh
go run ./cmd/server
```

By default, the server listens on `0.0.0.0:8080` and stores files under `./uploads`.

## Test

```sh
go test ./...
```

## Try the API

```sh
# Health check
curl http://localhost:8080/

# Upload a file
curl -F "file=@/path/to/local/file.txt" http://localhost:8080/files

# List stored files
curl http://localhost:8080/files

# Download a file
curl -OJ http://localhost:8080/files/file.txt
```

Upload response:

```json
{
  "name": "file.txt",
  "size": 1234,
  "content_type": "text/plain",
  "modified_at": "2026-07-26T00:00:00Z"
}
```

List response:

```json
{
  "files": [
    { "name": "file.txt", "size": 1234, "content_type": "", "modified_at": "2026-07-26T00:00:00Z" }
  ]
}
```

Notes on the file API:

- Upload accepts a single `multipart/form-data` field named `file`.
- File names are sanitized to their base name before being stored, so `../../etc/passwd` is saved
  as `passwd` inside the storage directory — there is no way to escape it.
- Uploading a file with a name that already exists overwrites the previous contents.
- Downloading a name that doesn't exist returns `404`.

## Configuration

The application has built-in defaults, so a config file is optional. To override locally, copy the
example file:

```sh
cp config.example.yaml config.yaml
```

Example config:

```yaml
app:
  environment: development

server:
  host: 0.0.0.0
  port: 8080

storage:
  dir: ./uploads
  max_upload_size_mb: 32
```

Environment variables use the `APP_` prefix and replace dots with underscores:

```sh
APP_APP_ENVIRONMENT=production APP_SERVER_PORT=3000 APP_STORAGE_DIR=/data/uploads go run ./cmd/server
```

## Logging

Each request is logged by the HTTP middleware with structured fields:

- `method`
- `path`
- `status`
- `body_size`
- `client_ip`

Production mode uses Zap production logging. Other environments use Zap development logging.

## Architecture

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the DDD package layout and request-flow diagrams.

## Contributors

- [barlus-developer](https://github.com/barlus-developer)
- Codex
