# Market REST API

A production-style RESTful CRUD API built in Go — no frameworks, no ORM. The project explores interface-driven design, PostgreSQL integration via `database/sql`, a Cobra CLI, gRPC, and full Docker containerisation.

This is a deliberate follow-up to an earlier stdlib-only task manager, carrying forward everything learned and adding real infrastructure on top.

---

## Tech Stack

| Layer | Choice | Why |
|---|---|---|
| Language | Go | Compiled, statically typed, excellent concurrency model |
| Router | [chi](https://github.com/go-chi/chi) | Lightweight, stdlib-compatible, no magic |
| Database | PostgreSQL 16 | Production-grade relational DB |
| DB Driver | `database/sql` + `lib/pq` | No ORM — raw SQL, full control |
| CLI | [Cobra](https://github.com/spf13/cobra) | Industry-standard Go CLI framework |
| RPC | gRPC + Protocol Buffers | High-performance binary transport |
| Containers | Docker + Docker Compose | One-command reproducible environment |

---

## Features

- Full CRUD over HTTP — GET, POST, PUT, DELETE
- Interface-driven storage layer — `MarketStore` interface with a PostgreSQL implementation, swappable without touching the server
- `database/sql` directly — parameterised queries, no ORM, no magic
- Cobra CLI with all CRUD commands runnable from the terminal
- gRPC server running alongside the REST API on a separate port
- Multi-stage Dockerfile — final image is ~15MB, build tools excluded
- Docker Compose with health checks — API container waits for Postgres to be genuinely ready before starting
- SQL migrations in `/db/migrations` — schema versioned and reproducible
- Environment-based config — credentials never hardcoded, injected via `.env`

---

## Project Structure

```
├── cmd/                        # Entry points
│   └── webserver/main.go       # Starts REST + gRPC servers
├── rest/                       # REST layer
│   ├── server.go               # HTTP handlers (chi router)
│   ├── market_store.go         # MarketStore interface
│   └── postgres_store.go       # PostgreSQL implementation
├── marketgrpc/                 # gRPC layer
│   └── server.go               # gRPC service implementation
├── protobuf/                   # Protocol Buffer definitions
│   └── market.proto            # Service + message definitions
├── db/
│   └── migrations/
│       └── 001_create_market.sql  # Table schema
├── Dockerfile                  # Multi-stage build
├── docker-compose.yml          # Postgres + API wired together
└── Makefile                    # protoc code generation
```

---

## Getting Started

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- That's it — Go does not need to be installed locally

### Run with Docker

```bash
# Clone the repo
git clone https://github.com/Wreker-creator/CRUD2.git
cd CRUD2

# Create a .env file in the project root
cat > .env << EOF
POSTGRES_USER=_name_
POSTGRES_PASSWORD=_password_
POSTGRES_DB=marketdb
DATABASE_URL=postgres://_name_:_password_@db:_port_/marketdb?sslmode=disable
EOF

# Start everything — Postgres + API
docker-compose up --build
```

The API is available at `http://localhost:8080`.
The gRPC server listens on port `50051`.

---

## API Endpoints

| Method | Endpoint | Description | Status Codes |
|---|---|---|---|
| GET | `/market` | Get all items | 200 |
| GET | `/market/{name}` | Get item by name | 200, 404 |
| POST | `/market` | Create a new item | 201, 400 |
| PUT | `/market/{name}` | Update an item | 200, 400, 404 |
| DELETE | `/market/{name}` | Delete an item | 200, 404 |

### Example Requests

```bash
# Get all items
curl http://localhost:8080/market

# Create an item
curl -X POST http://localhost:8080/market \
  -H "Content-Type: application/json" \
  -d '{"name":"Apple","price":1.99,"calories":52,"sugar":10.4}'

# Get a specific item
curl http://localhost:8080/market/Apple

# Update an item
curl -X PUT http://localhost:8080/market/Apple \
  -H "Content-Type: application/json" \
  -d '{"name":"Apple","price":2.49,"calories":52,"sugar":10.4}'

# Delete an item
curl -X DELETE http://localhost:8080/market/Apple
```

---

## CLI Usage

The same operations are available as CLI commands via Cobra:

```bash
# List all items
go run cmd/webserver/main.go list

# Add an item
go run cmd/webserver/main.go add --name Apple --price 1.99 --calories 52 --sugar 10.4

# Get a specific item
go run cmd/webserver/main.go get --name Apple

# Update an item
go run cmd/webserver/main.go update --name Apple --price 2.49

# Delete an item
go run cmd/webserver/main.go delete --name Apple
```

---

## gRPC

The service definition lives in `protobuf/market.proto`. To regenerate Go code after editing the proto file:

```bash
make proto
```

This runs `protoc` with both `--go_out` and `--go-grpc_out` flags to produce the typed Go client and server stubs.

---

## Design Decisions

**Interface-driven storage** — The server depends on a `MarketStore` interface, not a concrete type. The PostgreSQL implementation satisfies that interface. Swapping the storage backend requires zero changes to the HTTP or gRPC layers.

**`database/sql` over an ORM** — Using `lib/pq` and raw parameterised queries keeps the SQL explicit and readable. There is no hidden query generation. This also avoids the performance surprises that ORMs are known for at scale.

**Name as the public identifier** — Items are looked up by `name` in URLs rather than an internal integer ID. The database `id` is internal and never exposed through the API. This is a cleaner public contract — clients should not need to know or store database primary keys.

**Multi-stage Docker build** — Stage 1 compiles the binary inside a full Go environment. Stage 2 copies only the compiled binary into a minimal Alpine image. The result is a ~15MB final image instead of ~900MB, with no build tools or source code included.

**Health-checked Compose setup** — The `api` service uses `depends_on: condition: service_healthy` against the `db` service. Docker polls `pg_isready` before allowing the API to start, preventing connection failures on boot.

**SQL migrations** — Schema changes live in versioned `.sql` files under `/db/migrations`. The Postgres Docker image runs these automatically on first startup via the `/docker-entrypoint-initdb.d/` convention, making the setup fully reproducible.

---

## Docker Commands Reference

```bash
# Start everything
docker-compose up --build

# Run in background
docker-compose up -d --build

# View API logs
docker-compose logs -f api

# Stop everything (data preserved)
docker-compose down

# Full reset including database
docker-compose down -v
```
