# urlix

A lightweight URL shortener built with Go and PostgreSQL.

urlix provides a simple REST API for creating short URLs, custom aliases, redirects, click tracking, URL statistics, rate limiting, and optional URL expiration.

## Features

- Create short URLs
- Cryptographically secure random 6-character codes
- Custom URL aliases
- Duplicate code collision retry
- URL validation
- Optional URL expiration
- Expired URLs return `410 Gone`
- Click tracking
- URL statistics
- PostgreSQL persistence
- In-memory repository for testing
- IP-based rate limiting
- Dockerized PostgreSQL
- Unit and integration tests
- Race detector tests
- GitHub Actions CI
- Clean separation between HTTP, service, and repository layers

## Architecture

```text
Client
  |
  v
Handler (HTTP)
  |
  v
Service (Business Logic)
  |
  v
Repository Interface
  |
  +-------------------+
  |                   |
  v                   v
PostgreSQL          Memory
```

The project follows a simple dependency flow:

```text
HTTP -> Handler -> Service -> Repository
```

The service layer depends on the `URLRepository` interface rather than a concrete database implementation.

## API

### Create a Short URL

`POST /api/urls`

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

```http
POST /api/urls
```

Request:

```json
{
  "url": "https://github.com",
  "alias": "github"
}
```

Response:

```json
{
  "code": "github",
  "url": "https://github.com"
}
```

Aliases must be between 3 and 16 characters and cannot contain URL/path separator characters.

A duplicate alias returns:

```text
409 Conflict
```

### Create an Expiring URL

```http
POST /api/urls
```

Request:

```json
{
  "url": "https://github.com",
  "expires_at": "2026-12-31T23:59:59Z"
}
```

The URL remains active until the specified expiration time.

Expired URLs return:

```text
410 Gone
```

Aliases can also be combined with expiration:

```json
{
  "url": "https://github.com",
  "alias": "github",
  "expires_at": "2026-12-31T23:59:59Z"
}
```

### Redirect

```http
GET /r/{code}
```

Example:

```bash
curl -i http://localhost:8080/r/hGwSNp
```

Returns an HTTP `302 Found` response and redirects the client to the original URL.

Each successful redirect increments the URL's click count.

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

### Health Check

```http
GET /health
```

### Hello

```http
GET /hello
```

## Project Structure

```text
urlix/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── handler/
│   │   ├── health.go
│   │   ├── hello.go
│   │   └── url.go
│   ├── middleware/
│   │   ├── rate_limit.go
│   │   └── rate_limit_test.go
│   ├── repository/
│   │   ├── memory.go
│   │   ├── postgres.go
│   │   └── postgres_test.go
│   └── service/
│       ├── url.go
│       └── url_test.go
├── migrations/
│   ├── 001_create_urls.sql
│   ├── 002_add_click_count.sql
│   └── 003_add_expires_at.sql
├── .github/
│   └── workflows/
│       └── ci.yml
├── compose.yaml
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

## Requirements

- Go 1.27+
- Docker
- Docker Compose
- PostgreSQL

## Run Locally

Start PostgreSQL:

```bash
docker compose up -d
```

Run the application:

```bash
go run ./cmd/api
```

The API will be available at:

```text
http://localhost:8080
```

## Test

Run all tests:

```bash
go test ./...
```

Run the race detector:

```bash
go test -race ./...
```

Build the application:

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

Create an expiring URL:

```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com","expires_at":"2026-12-31T23:59:59Z"}'
```

Redirect using the generated code:

```bash
curl -i http://localhost:8080/r/{code}
```

Check URL statistics:

```bash
curl http://localhost:8080/api/urls/{code}
```

## Tech Stack

- Go
- PostgreSQL
- pgx
- `database/sql`
- Docker
- `net/http`
- Go standard library
- GitHub Actions

## Roadmap

- [x] URL creation
- [x] URL redirection
- [x] PostgreSQL persistence
- [x] Click tracking
- [x] URL statistics
- [x] In-memory repository
- [x] Unit tests
- [x] Integration tests
- [x] GitHub Actions CI
- [x] Race detector tests
- [x] Rate limiting
- [x] URL validation
- [x] Custom aliases
- [x] URL expiration
- [ ] API documentation
- [ ] OpenAPI specification
- [ ] Detailed click analytics
- [ ] Authentication
- [ ] Admin API

## License

MIT
