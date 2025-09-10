# TodoService (Go + gRPC + GORM + Postgres)

Minimal, production-minded TodoService using Go, gRPC, protobuf, and Postgres with GORM.

## Requirements
- Go 1.21+
- protoc (protocol buffers compiler)
- Docker & Docker Compose (for local Postgres)
- `protoc-gen-go` and `protoc-gen-go-grpc` installed (`go install` as shown below)

## Quickstart (local)

1. copy `.env.example` to `.env` and adjust if needed.
2. Start Postgres:
   ```bash
   docker compose up -d
	 ```
3. Generate protos:
	```bash
	make proto
	```
4. Build & run:

   ```bash
   make build
   ./bin/todosvc
   ```

The server will auto-migrate the DB on startup.

## Ports

* gRPC: `50051` (configurable via `GRPC_PORT`)
* HTTP REST: `8080` (configurable via `HTTP_PORT`)

## Example REST requests

Create:

```bash
curl -X POST http://localhost:8080/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"buy milk","description":"2 liters"}'
```

List:

```bash
curl "http://localhost:8080/tasks?page=1&page_size=10"
```

Get:

```bash
curl http://localhost:8080/tasks/<id>
```

Mark complete:

```bash
curl -X PATCH http://localhost:8080/tasks/<id> \
  -H "Content-Type: application/json" \
  -d '{"completed": true}'
```

Delete:

```bash
curl -X DELETE http://localhost:8080/tasks/<id>
```

Health:

```bash
curl http://localhost:8080/healthz
```

## Example gRPC (grpcurl)

Install `grpcurl`. Then:

```bash
grpcurl -plaintext -d '{"title":"grpc task", "description":"via grpc"}' localhost:50051 todo.TodoService/CreateTask
grpcurl -plaintext localhost:50051 todo.TodoService/ListTasks
```

## Tests

Run unit tests (sqlite in-memory):

```bash
go test ./...
```

## Production notes

* Use TLS for gRPC (server certificates).
* Use connection pooling tuning & observability (metrics, traces).
* Replace AutoMigrate with versioned migrations (e.g., `gormigrate`) for production.
* Consider switching to `pgx` + `sqlc` for raw SQL performance-critical paths; the repo and service interfaces allow swapping implementations.

## Proto generation

Make sure `protoc` and `protoc-gen-go`/`protoc-gen-go-grpc` are installed:

```bash
# install generators
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# generate
make proto
```
## Makefile

* `make proto` - generates Go proto code
* `make build` - builds binary
* `make run` - runs binary
* `make migrate` - runs migrations (server auto-migrates on startup)
* `make test` - run unit tests

---

# 16) Notes: security, env handling, production suggestions
- **Secrets:** don't store DB credentials in plain `.env` in production — use secret managers (AWS Secrets Manager, HashiCorp Vault, etc.).
- **TLS:** enable TLS for gRPC; terminate TLS at edge or in-service.
- **Observability:** add logging (structured logger like `zerolog` or `zap`), metrics (Prometheus), and tracing (OpenTelemetry).
- **Migrations:** use `gormigrate` or an SQL-first tool for versioned migrations.
- **Connection Pooling:** tune `max_open_conns`, `max_idle_conns`, `conn_max_lifetime`.
- **Retries:** perform sensible retries on transient DB failures using backoff (not included to keep example minimal).
- **Swap to sqlc/pgx:** keep `Repository` interface small; implement a new repo using `sqlc + pgx` and swap easily.

---

# 17) Exact commands to run locally (step-by-step)

1. clone or create folder and files (structure above). Ensure `module` in `go.mod` matches repo path (example used `github.com/fuzail/todosvc`).

2. install tools:
```bash
# install protoc (system dependent) and the go codegen plugins:
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# get libs
go get ./...
````

3. start Postgres:

```bash
docker compose up -d
# wait until healthy: docker ps or docker logs
```

4. generate protos:

```bash
make proto
```

5. build:

```bash
make build
```

6. run:

```bash
./bin/todosvc
```

7. test:

```bash
# REST
curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" \
  -d '{"title":"buy milk","description":"2L"}'

# gRPC (grpcurl)
grpcurl -plaintext -d '{"title":"grpc task", "description":"via grpc"}' localhost:50051 todo.TodoService/CreateTask
```
## Protoc proto file compiler
```bash
protoc -I=proto -I="C:\protoc\include" --go_out=paths=source_relative:./proto --go-grpc_out=paths=source_relative:./proto proto/todo.proto
```