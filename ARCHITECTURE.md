# Architecture

This project uses a small DDD-style / hexagonal layout. There is no database: uploaded files are
persisted directly on the local filesystem behind a `Storage` port, so the storage technology can
be swapped later without touching handlers or use cases.

## Layers

- `cmd/server`: starts the process, creates the HTTP server, and handles graceful shutdown.
- `internal/bootstrap`: wires configuration, logging, services, handlers, middleware, and routes.
- `internal/domain`: contains business objects (`health.Status`, `file.File`).
- `internal/application`: contains use cases and application services that operate on domain
  objects. `application/file` also defines the `Storage` port that infrastructure adapters
  implement.
- `internal/interfaces/http`: contains Gin handlers, middleware, and router setup.
- `internal/infrastructure`: contains technical concerns such as configuration, logging, and the
  `storage.Local` disk adapter.

## Package Relationship

```mermaid
flowchart TD
    Main[cmd/server] --> Bootstrap[internal/bootstrap]
    Bootstrap --> Config[internal/infrastructure/config]
    Bootstrap --> Logger[internal/infrastructure/logger]
    Bootstrap --> LocalStorage[internal/infrastructure/storage]
    Bootstrap --> HealthService[internal/application/health]
    Bootstrap --> FileService[internal/application/file]
    Bootstrap --> Handler[internal/interfaces/http/handler]
    Bootstrap --> Router[internal/interfaces/http/router]
    Router --> Middleware[internal/interfaces/http/middleware]
    Router --> Handler
    Handler --> HealthService
    Handler --> FileService
    HealthService --> HealthDomain[internal/domain/health]
    FileService --> FileDomain[internal/domain/file]
    FileService -. Storage port .-> LocalStorage
```

## Health Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant Gin as Gin Router
    participant Log as Logger Middleware
    participant Handler as Health Handler
    participant Service as Health Service
    participant Domain as Domain Status

    Client->>Gin: GET /
    Gin->>Log: execute middleware chain
    Log->>Handler: c.Next()
    Handler->>Service: Status()
    Service->>Domain: create Status
    Domain-->>Service: status object
    Service-->>Handler: status object
    Handler-->>Client: 200 JSON response
    Log->>Log: record method, path, status, body_size, client_ip
```

## File Upload / Download Flow

```mermaid
sequenceDiagram
    participant Client
    participant Gin as Gin Router
    participant Handler as File Handler
    participant Service as File Service
    participant Storage as Local Disk Storage

    Client->>Gin: POST /files (multipart "file")
    Gin->>Handler: Upload(c)
    Handler->>Service: Upload(ctx, name, contentType, reader)
    Service->>Storage: Save(ctx, name, contentType, reader)
    Storage-->>Service: file.File metadata
    Service-->>Handler: file.File metadata
    Handler-->>Client: 201 JSON metadata

    Client->>Gin: GET /files/:name
    Gin->>Handler: Download(c)
    Handler->>Service: Download(ctx, name)
    Service->>Storage: Open(ctx, name)
    Storage-->>Service: file reader + metadata (or ErrNotFound)
    Service-->>Handler: file reader + metadata
    Handler-->>Client: 200 streamed bytes (or 404)
```

## Startup Algorithm

```mermaid
flowchart TD
    Start([Process starts]) --> LoadConfig[Load config with Viper]
    LoadConfig --> CreateLogger[Create Zap logger]
    CreateLogger --> BuildStorage[Create local disk storage, ensure dir exists]
    BuildStorage --> BuildServices[Create health and file application services]
    BuildServices --> BuildHandlers[Create HTTP handlers]
    BuildHandlers --> BuildRouter[Create Gin router and middleware]
    BuildRouter --> StartServer[Start net/http server]
    StartServer --> WaitSignal[Wait for SIGINT or SIGTERM]
    WaitSignal --> Shutdown[Gracefully shutdown server]
    Shutdown --> Stop([Process exits])
```

## Dependency Direction

Dependencies point inward toward the application and domain layers:

```mermaid
flowchart LR
    HTTP[interfaces/http] --> Application[application]
    Application --> Domain[domain]
    Infrastructure[infrastructure] -. implements Storage port .-> Application
    Bootstrap[bootstrap] --> HTTP
    Bootstrap --> Application
    Bootstrap --> Infrastructure
    Main[cmd/server] --> Bootstrap
```

The router knows about HTTP handlers and middleware. The handler knows about the application
service interface. The application service depends on the `Storage` port, not on a concrete
storage technology. `infrastructure/storage` implements that port by writing to local disk;
infrastructure packages do not contain request handling or business behavior.
