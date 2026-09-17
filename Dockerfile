# Build stage
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /urlix ./cmd/api

# Runtime stage
FROM alpine:3.22

RUN addgroup -S urlix && adduser -S -G urlix urlix

WORKDIR /app

COPY --from=builder /urlix /app/urlix

USER urlix

EXPOSE 8080

ENTRYPOINT ["/app/urlix"]
