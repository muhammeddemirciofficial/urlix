# URLIX

A lightweight, production-minded URL shortener built with Go and PostgreSQL.

URLIX provides a small REST API for creating short URLs, custom aliases, redirects, click tracking, URL statistics, detailed click analytics, rate limiting, and optional URL expiration.

## Features

- Cryptographically secure random 6-character short codes
- Custom URL aliases
- Duplicate alias protection
- URL validation
- Optional URL expiration
- Expired URLs return `410 Gone`
- HTTP redirect with click counting
- Click analytics with user-agent and referer data
- PostgreSQL persistence
- In-memory repository for tests
- IP-based rate limiting
- Database health checks
- Dockerized application and PostgreSQL
- Automated database migrations
- OpenAPI 3 specification
- Swagger UI
- Unit and integration tests
- Race detector
- `go vet`
- GitHub Actions CI
- Graceful HTTP server shutdown
- Layered HTTP / service / repository architecture

## Architecture

```text
Client
  |
  v
HTTP Handler
  |
  v
Service
  |
  v
Repository Interface
  |
  +-------------------+
  |                   |
  v                   v
PostgreSQL          Memory
```

Dependency flow:

```text
HTTP -> Handler -> Service -> Repository
```

The service layer depends on the `URLRepository` interface, keeping business logic independent from the PostgreSQL implementation and making the application straightforward to test.

## API

### Health

```http
GET /health
```

Returns:

```json
{
  "status": "ok"
}
```

### Database Health

```http
GET /health/db
```

Returns `200 OK` when PostgreSQL is reachable and `503 Service Unavailable` otherwise.

### Create a Short URL

```http
POST /api/urls
Content-Type: application/json
```

Request:

```json
{
  "url": "https://github.com"
}
```

Response:

```json
{
  "code": "hGwSNp",
  "url": "https://github.com"
}
```

Returns `201 Created`.

### Create a Custom Alias

Request:

```json
{
  "url": "https://github.com",
  "alias": "github"
}
```

The alias becomes the public URL code, so the short URL is:

```text
http://localhost:8080/r/github
```

Aliases must be between 3 and 16 characters and cannot contain URL/path separator characters. A duplicate alias returns `409 Conflict`.

### Redirect

```http
GET /r/{code}
```

Example:

```bash
curl -i http://localhost:8080/r/hGwSNp
```

Returns `302 Found` and redirects to the original URL.

Each successful redirect increments the click count and records an analytics event.

### URL Statistics

```http
GET /api/urls/{code}
```

Response:

```json
{
  "code": "hGwSNp",
  "url": "https://github.com",
  "click_count": 3
}
```

### URL Analytics

```http
GET /api/urls/{code}/analytics
```

Returns the URL, total click count, and recorded click events including timestamp, user-agent, and referer.

### API Documentation

Swagger UI:

```text
GET /docs
```

OpenAPI specification:

```text
GET /openapi.yaml
```

When running locally, open:

```text
http://localhost:8080/docs
```

## Rate Limiting

URL creation is protected by an in-memory IP-based rate limiter.

Default limit:

```text
60 requests / minute / IP
```

Requests exceeding the limit receive `429 Too Many Requests`.

## URL Expiration

URL expiration is supported at the service and persistence layers. Expired URLs are rejected with:

```text
410 Gone
```

The current HTTP create endpoint intentionally exposes a small request contract containing `url` and optional `alias` only.

## Project Structure

```text
urlix/
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── migrate/
│       └── main.go
├── internal/
│   ├── handler/
│   │   ├── docs.go
│   │   ├── health.go
│   │   ├── hello.go
│   │   └── url.go
│   ├── middleware/
│   │   ├── rate_limit.go
│   │   └── rate_limit_test.go
│   ├── repository/
│   │   ├── memory.go
│   │   ├── memory_test.go
│   │   ├── postgres.go
│   │   └── postgres_test.go
│   └── service/
│       ├── url.go
│       └── url_test.go
├── migrations/
│   ├── 001_create_urls.sql
│   ├── 002_add_click_count.sql
│   ├── 003_add_expires_at.sql
│   └── 004_create_url_clicks.sql
├── .github/
│   └── workflows/
│       └── ci.yml
├── compose.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── LICENSE
├── openapi.yaml
└── README.md
```

## Requirements

- Go 1.27+
- Docker
- Docker Compose
- PostgreSQL 18 (used by the Docker setup)

## Configuration

Create `.env` from the example:

```bash
cp .env.example .env
```

Example:

```env
POSTGRES_USER=urlix
POSTGRES_PASSWORD=change-me
POSTGRES_DB=urlix
```

The application and migration container use `DATABASE_URL` internally through Docker Compose.

## Run with Docker

Build and start the complete stack:

```bash
docker compose up -d --build
```

The Compose setup starts:

1. PostgreSQL
2. Database migrations
3. URLIX API

Check the API:

```bash
curl http://localhost:8080/health
```

Open Swagger UI:

```text
http://localhost:8080/docs
```

Stop the stack:

```bash
docker compose down
```

## Run Locally

Start PostgreSQL with Docker:

```bash
docker compose up -d postgres
```

Set `DATABASE_URL` in your shell, then run migrations and the API as needed:

```bash
go run ./cmd/migrate
go run ./cmd/api
```

The API will be available at:

```text
http://localhost:8080
```

## Testing

Run the complete test suite:

```bash
go test ./...
```

Run the race detector:

```bash
go test -race ./...
```

Run static analysis:

```bash
go vet ./...
```

Build all packages:

```bash
go build ./...
```

## Example

Create a short URL:

```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com"}'
```

Create a custom alias:

```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com","alias":"github"}'
```

Redirect:

```bash
curl -i http://localhost:8080/r/github
```

Check statistics:

```bash
curl http://localhost:8080/api/urls/github
```

Check analytics:

```bash
curl http://localhost:8080/api/urls/github/analytics
```

## CI

GitHub Actions runs the following checks on pushes and pull requests targeting `main`:

- PostgreSQL-backed tests
- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- `go build ./...`
- Docker image build

## Tech Stack

- Go 1.27
- `net/http`
- PostgreSQL 18
- pgx/v5
- `database/sql`
- Docker / Docker Compose
- OpenAPI 3
- Swagger UI
- GitHub Actions

## Roadmap

URLIX v1.0.0 scope is complete.

Potential future work is intentionally outside the current release:

- Authentication and API keys
- Admin API
- Persistent distributed rate limiting
- More granular analytics aggregation
- Metrics / observability
- Production deployment manifests

## License

MIT
