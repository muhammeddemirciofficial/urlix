# urlix

A lightweight URL shortener built with Go and PostgreSQL.

urlix provides a simple REST API for creating short URLs, redirecting users, and tracking click counts.

## Features

- Create short URLs
- Random 6-character short codes
- PostgreSQL persistence
- In-memory repository for testing
- Click tracking
- URL statistics endpoint
- Input validation
- Dockerized PostgreSQL
- Unit tests
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
Repository
  |
  +----------+----------+
  |                     |
  v                     v
PostgreSQL            Memory
```

The project follows a simple dependency flow:

```text
HTTP -> Handler -> Service -> Repository
```

The service layer depends on the `URLRepository` interface rather than a concrete database implementation.

## API

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

### Redirect

```http
GET /r/{code}
```

Example:

```text
GET /r/hGwSNp
```

Returns:

```text
302 Found
Location: https://github.com
```

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
│   ├── repository/
│   │   ├── memory.go
│   │   └── postgres.go
│   └── service/
│       └── url.go
├── migrations/
│   ├── 001_create_urls.sql
│   └── 002_add_click_count.sql
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

Then open the generated short URL:

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
- Docker
- database/sql
- Go standard library (`net/http`)

## Roadmap

- [x] URL creation
- [x] URL redirection
- [x] PostgreSQL persistence
- [x] Click tracking
- [x] URL statistics
- [x] In-memory repository
- [x] Unit tests
- [ ] Integration tests
- [ ] GitHub Actions CI
- [ ] Rate limiting
- [ ] URL expiration
- [ ] Custom aliases
- [ ] API documentation

## License

MIT
