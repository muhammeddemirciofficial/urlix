urlix
A lightweight URL shortener built with GO and PostgreSQL.

urlix provides a simple REST API for creating short URLs, redirecting users, and tracking click counts.

Features

    * Create short URLs
    * Random 6-character short codes
    * PostgreSQL persistence
    * In-memory repository for testing
    * Click tracking
    * URL statistics endpoint
    * Input validation
    * Dockerized PostgreSQL
    * Unit tests
    * Clean separation between HTTP, service and repository layers

Architecture

                    ┌──────────────┐
                    │    Client    │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │    Handler   │
                    │     HTTP     │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │    Service   │
                    │  Business    │
                    │    Logic     │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │ Repository   │
                    └──────┬───────┘
                           │
                    ┌──────┴───────┐
                    ▼              ▼
              ┌───────────┐  ┌───────────┐
              │ PostgreSQL│  │   Memory  │
              └───────────┘  └───────────┘

The project follows a simple dependency flow:

    HTTP → Handler → Service → Repository

The service layer depends on the URLRepository interface rather than a concrete database implementation.

API
    Create a short URL

    POST /api/urls
    Content-Type: application/json

    Request:
        {
            "url": "https://github.com"
        }

    Response:
        {
            "code": "hGwSNp",
            "url": "https://github.com"
        }

    Redirect
        GET /r/{code}

    Example:
        GET /r/hGwSNp

    Returns:
        302 Found
        Location: https://github.com

    Each succesful redirect increments the URL's click count.

    URL statictics

        GET /api/urls/{code}

    Response:

        {
            "code": "hGwSNp",
            "url": "https://github.com",
            "click_count": 3
        }

    Health check
       GET /health
    Hello
        GET /hello

Project Structure

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
    └── README.md

Requirements
    * GO 1.27+
    * Docker
    * Docker Compose

Run Locally

Start PostgreSQL:

    docker compose up -d

Run the application:

    go run ./cmd/api

The API will be available at:

    http://localhost:8080

Test
Run all tests:

    go test ./...

Build the application:

    go build ./...

Example
Create a short URL:

    curl -X POST http://localhost:8080/api/urls \
    -H "Content-Type: application/json" \
    -d '{"url":"https://github.com"}'

The open the generated short URL:

    curl -i http://localhost:8080/r/{code}

Check statistics:

    curl http://localhost:8080/api/urls/{code}

Tech Stack
    * Go
    * PostgreSQL
    * pgx
    * Docker
    * database/sql
    * Go net/http

Roadmap
    * Url creation
    * Url redirection
    * PostgreSQL persistence
    * Click tracking
    * URL statistics
    * In-memory repository
    * Unit tests
    * Integration tests
    * Github Actions CI
    * Rate limiting
    * URL expiration
    * Custom aliases
    * API documentation

License
MIT




