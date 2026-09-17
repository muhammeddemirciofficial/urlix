package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/muhammeddemirciofficial/urlix/internal/handler"
	"github.com/muhammeddemirciofficial/urlix/internal/middleware"
	"github.com/muhammeddemirciofficial/urlix/internal/repository"
	"github.com/muhammeddemirciofficial/urlix/internal/service"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		panic("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	rateLimiter := middleware.NewRateLimiter(60, time.Minute)

	urlRepository := repository.NewPostgresURLRepository(db)
	urlService := service.NewURLService(urlRepository)
	urlHandler := handler.NewURLHandler(urlService)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.Health)
	mux.Handle("GET /health/db", handler.DatabaseHealth(db))
	mux.HandleFunc("/hello", handler.Hello)
	mux.Handle("POST /api/urls", rateLimiter.Middleware(http.HandlerFunc(urlHandler.CreateURL)))
	mux.HandleFunc("GET /r/{code}", urlHandler.Redirect)
	mux.HandleFunc("GET /api/urls/{code}", urlHandler.GetStats)
	mux.HandleFunc("GET /api/urls/{code}/analytics", urlHandler.GetAnalytics)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		fmt.Println("URLIX is running on http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic(fmt.Errorf("server shutdown: %w", err))
	}

	fmt.Println("server stopped")
}
