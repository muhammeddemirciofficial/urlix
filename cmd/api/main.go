package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/muhammeddemirciofficial/urlix/internal/handler"
	"github.com/muhammeddemirciofficial/urlix/internal/middleware"
	"github.com/muhammeddemirciofficial/urlix/internal/repository"
	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

func main() {
	db, err := sql.Open(
		"pgx",
		"postgres://urlix:urlix@localhost:5432/urlix?sslmode=disable")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	rateLimiter := middleware.NewRaterLimiter(60, time.Minute)

	urlRepository := repository.NewPostgresURLRepository(db)
	urlService := service.NewURLService(urlRepository)
	urlHandler := handler.NewURLHandler(urlService)

	http.HandleFunc("/health", handler.Health)
	http.HandleFunc("/hello", handler.Hello)
	http.Handle("POST /api/urls", rateLimiter.Middleware(http.HandlerFunc(urlHandler.CreateURL)))
	http.HandleFunc("GET /r/{code}", urlHandler.Redirect)
	http.HandleFunc("GET /api/urls/{code}", urlHandler.GetStats)

	fmt.Println("URLIX is running on http://localhost:8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
