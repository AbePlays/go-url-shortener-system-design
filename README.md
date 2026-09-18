# Go URL Shortener – System Design

A production-shaped URL shortener built in Go.
This project focuses on **system design maturity** — caching, replication, failure modes, and trade-offs — rather than just a working demo.

## Goal

Build a URL shortener that demonstrates real backend engineering judgment:

- Fast redirects under read-heavy load
- Durable storage
- Clear separation of concerns
- Explicit handling of consistency and failure
- A simple but usable web interface

## Requirements

### Functional
- Shorten a long URL into a short, shareable link
- Redirect from short link to original URL
- Basic click counting
- Simple web UI for creating short links

### Non-Functional
- Low latency on the redirect path
- Data survives process restarts
- Read-heavy workload support
- Design that can scale horizontally later

### Out of Scope (v1)
- Custom aliases
- User accounts / authentication
- Advanced analytics
- Rate limiting (planned as a separate project)

## High-Level Architecture

```
Browser
   │
   ▼
Load Balancer
   │
   ▼
Go App Servers (stateless)
   │
   ├──► Redis (cache)
   │
   └──► PostgreSQL
          ├── Primary   (writes)
          └── Replica   (reads)
```

**Core principles:**
- Application servers are stateless
- PostgreSQL is the source of truth
- Redis sits in front of reads for hot keys
- Writes always go to the primary
- Reads prefer the replica + cache

## Design Evolution

The system is grown in deliberate stages instead of jumping straight to a distributed design.

### Stage 1 – MVP
- Single Go process
- In-memory map protected by `sync.RWMutex`
- Basic shorten + redirect
- No persistence

**Purpose:** Validate the core flow quickly.
**Limitation:** Data is lost on restart; cannot scale beyond one process.

### Stage 2 – Single Instance Production
- Environment-based configuration
- Graceful shutdown
- URL validation
- Collision handling on key generation
- Structured logging

Still single-process, but production hygiene is in place.

### Stage 3 – Durable Storage (PostgreSQL)

PostgreSQL becomes the source of truth.

```sql
CREATE TABLE urls (
    code         CHAR(8) PRIMARY KEY,
    original_url TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    clicks       BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_urls_original ON urls (original_url);
```

**Why PostgreSQL?**
- Strong consistency and unique constraints
- Excellent Go support (`pgx`)
- Natural path to read replicas
- Easy to reason about

At this stage every redirect still hits the database.

### Stage 4 – Caching Layer (Redis)

Most traffic is reads. Redis is introduced as a cache in front of PostgreSQL.

**Strategy:** Cache-aside
- Check Redis first on redirect
- On miss → load from Postgres → populate cache
- Writes go to Postgres (cache is optionally warmed)

**TTL:** 24 hours (configurable)

**Trade-off:** Possible short window of inconsistency, which is acceptable for this use case.

### Stage 5 – Read Scaling (Primary + Replica)

- All writes go to the **Primary**
- Redirects prefer the **Replica**
- Application is aware of replication lag

A newly created short link may not be immediately visible on the replica. The cache helps hide most of this lag for popular links.

### Stage 6 – Future Scaling (Design Only)

These are designed but not fully implemented:

- Sharding by short-code prefix
- Redis Cluster
- Multi-region deployment
- Separate analytics pipeline (clicks → queue → aggregator)

## Key Design Decisions

### Short Code Generation
- 8-character Base62 (`0-9a-zA-Z`)
- ~218 trillion possible combinations
- Collision handling via retry
- Length chosen for a good balance between brevity and collision resistance

### Consistency Model
- Strong consistency on the write path (Postgres primary)
- Read-your-writes is **not** strictly guaranteed when reading from the replica
- Cache may serve slightly stale data within the TTL window

### Failure Handling
| Failure | Behavior |
|---------|----------|
| Redis down | Fall back to Postgres |
| Replica lag / down | Fall back to primary |
| Primary down | Writes fail; reads may still succeed from cache/replica |

## API Design

```http
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/path"
}
```

**Response:**
```json
{
  "code": "aB3xY9kL",
  "short_url": "https://short.example/aB3xY9kL"
}
```

```http
GET /{code}
→ 301/302 redirect to original URL
```

```http
GET /stats/{code}   (optional)
→ click count + metadata
```

The web UI uses the same backend via HTMX.

## Frontend

- Server-rendered or static HTML + CSS
- Vanilla JavaScript (Fetch API) for interacting with the backend
- Minimal and dependency-free

## Implementation Status

| Stage                        | Status          |
|-----------------------------|-----------------|
| MVP (in-memory)             | Implemented     |
| Production hygiene          | Implemented     |
| PostgreSQL                  | Implemented     |
| Redis cache                 | Implemented     |
| Primary + Read Replica      | Implemented     |
| Sharding / Multi-region     | Designed only   |
| Auth / Rate limiting        | Future projects |

## Trade-offs Summary

| Decision              | Benefit                          | Cost                              |
|-----------------------|----------------------------------|-----------------------------------|
| Postgres as source of truth | Durability + strong constraints | Higher latency than pure Redis   |
| Cache-aside Redis     | Very fast redirects              | Possible brief inconsistency      |
| Async replica         | Scales reads cleanly             | Replication lag                   |
| 8-char Base62         | Huge keyspace, still short       | Slightly longer than 6–7 chars    |
| HTMX frontend         | Simple and fast to build         | Less “app-like” than a full SPA   |

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Cache:** Redis
- **Frontend:** Go templates + HTMX
- **Containerization:** Docker + Docker Compose

## Getting Started

```bash
git clone https://github.com/<your-username>/go-url-shortener-system-design.git
cd go-url-shortener-system-design
docker compose up --build
```

The service will be available at `http://localhost:8080`.

## License

MIT
