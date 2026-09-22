# Go URL Shortener – System Design

A production-shaped URL shortener built in Go. This project focuses on **system design maturity** — caching, persistence, failure modes, and honest trade-offs — rather than just a working demo.

**Live:** https://go-url-shortener-system-design.onrender.com

## Goal

Build a URL shortener that demonstrates real backend engineering judgment:

- Fast redirects under read-heavy load
- Durable storage
- Clear separation of concerns
- Explicit handling of consistency and failure
- A simple, usable web interface — no JavaScript

## Requirements

### Functional

- Shorten a long URL into a short, shareable link
- Redirect from short link to original URL
- Click count tracking per link
- Simple web UI for creating and browsing short links

### Non-Functional

- Low latency on the redirect path
- Data survives process restarts
- Read-heavy workload support
- Schema changes tracked and reversible via migrations

### Out of Scope (v1)

- Custom aliases
- User accounts / authentication
- Advanced analytics beyond click counts
- Rate limiting (planned as a separate project)

## High-Level Architecture

```
Browser
   │
   ├──► Web pages (Go html/template, server-rendered, no JS)
   │
   ▼
Go App Server
   │
   ├──► Redis (cache-aside)
   │
   └──► PostgreSQL (source of truth)
```

**Core principles:**

- The app server is stateless
- PostgreSQL is the single source of truth
- Redis sits in front of reads for hot keys, populated lazily
- Writes always go straight to Postgres
- The web frontend is a thin BFF layer over the same store — no separate API round-trip

## Design Evolution

The system was grown in deliberate stages instead of jumping straight to a distributed design.

### Stage 1 – MVP

- Single Go process
- In-memory map protected by `sync.RWMutex`
- Basic shorten + redirect
- No persistence

### Stage 2 – Single Instance Production

- Environment-based configuration, fail-fast on missing config
- Graceful shutdown (`SIGTERM`/`SIGINT`, in-flight requests drained)
- URL validation
- Collision handling on key generation
- Structured logging (`log/slog`)
- Dockerized, CI/CD via GitHub Actions

### Stage 3 – Durable Storage (PostgreSQL)

PostgreSQL is the source of truth. Schema is managed through versioned migrations (`golang-migrate`), not a static init script — every schema change is a small, reversible, ordered file under `db/migrations/`.

```sql
CREATE TABLE IF NOT EXISTS urls (
    code         CHAR(8) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    clicks       BIGINT NOT NULL DEFAULT 0
);
```

Short-code collisions are handled by attempting the insert and retrying only on a genuine primary-key violation (Postgres error `23505`) — no separate existence check before every write.

**Why PostgreSQL?** Strong consistency and unique constraints, excellent Go support (`pgx`), and a natural fit for the read-heavy, write-light access pattern this project has.

### Stage 4 – Caching Layer (Redis)

Most traffic is reads. Redis sits in front of PostgreSQL using a cache-aside strategy:

- Check Redis first on redirect
- On miss → load from Postgres → populate cache → return
- Writes go straight to Postgres; the cache is never written to on write, only lazily on the next read

**TTL:** 24 hours. **Trade-off:** a short window of possible staleness, acceptable for this use case — links are immutable, so staleness only ever affects freshly-created links that haven't been read yet.

A Redis failure never breaks a redirect: cache errors are logged and the request falls through to Postgres.

### Stage 5 – Read Scaling (Primary + Replica) — Planned

Designed, not implemented. All writes would go to the primary; redirects would prefer a read replica, with the application aware of replication lag. Deliberately deferred: real Postgres streaming replication is a meaningfully different scope of work from anything else in this project, and the learning goal (application-level read/write routing) doesn't require it to be built by hand to be understood.

### Stage 6 – Future Scaling (Design Only)

- Sharding by short-code prefix
- Redis Cluster
- Multi-region deployment

## Key Design Decisions

### Short Code Generation

- 8-character Base62 (`0-9a-zA-Z`)
- ~218 trillion possible combinations
- Collision handling via retry on insert, not a pre-check

### Click Tracking

- Incremented atomically in Postgres (`UPDATE urls SET clicks = clicks + 1`), never read-modify-write in application code
- Fired off asynchronously after a successful redirect, so click tracking never adds latency to the redirect path itself

### Consistency Model

- Strong consistency on the write path (Postgres)
- Cache may serve slightly stale data within the TTL window

### Failure Handling

| Failure     | Behavior                                      |
| ----------- | ---------------------------------------------- |
| Redis down  | Falls back to Postgres; logged, not fatal       |
| Postgres down | Requests fail; no silent fallback              |

## API Design

```
POST /api/shorten
Content-Type: application/json

{ "url": "https://example.com/very/long/path" }
```

**Response:**

```json
{
  "shortCode": "aB3xY9kL",
  "url": "https://go-url-shortener-system-design.onrender.com/aB3xY9kL"
}
```

```
GET /{code}
→ 302 redirect to the original URL
```

## Frontend

Server-rendered with Go's `html/template`, no JavaScript. The form on `/` posts to `POST /shorten`, which acts as a small BFF — it calls the same store the JSON API uses, then re-renders the page with the result, rather than round-tripping through the API itself.

- `GET /` — shorten form
- `GET /links` — table of every shortened link, with click counts
- `GET /about` — project overview

## Implementation Status

| Stage                    | Status         |
| ------------------------- | -------------- |
| MVP (in-memory)           | Implemented    |
| Production hygiene        | Implemented    |
| PostgreSQL + migrations   | Implemented    |
| Redis cache                | Implemented    |
| Click analytics            | Implemented    |
| Frontend (server-rendered) | Implemented    |
| Primary + Read Replica     | Planned        |
| Sharding / Multi-region    | Designed only  |
| Auth / Rate limiting       | Future projects |

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL (Neon in production)
- **Cache:** Redis (Upstash in production)
- **Migrations:** golang-migrate
- **Frontend:** Go `html/template`, no JavaScript
- **Containerization:** Docker + Docker Compose
- **CI/CD:** GitHub Actions (format, vet, test against real Postgres + Redis service containers, build)
- **Hosting:** Render

## Getting Started

```bash
git clone https://github.com/AbePlays/go-url-shortener-system-design.git
cd go-url-shortener-system-design
make compose-up      # starts app + Postgres + Redis
make migrate-up       # applies the schema
```

The app is available at `http://localhost:8080`.

See the `Makefile` for the full set of available commands (tests, formatting, linting, Docker, migrations).

## License

MIT
