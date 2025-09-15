package main

import (
	"context"
	_ "eduBase/docs"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eduBase/internal/config"
	"eduBase/internal/logger"
	"eduBase/internal/server"
)

// @title           EduBase API
// @version         1.0
// @description     Внутренний API школьной базы (ученики, документы, медицина, согласия, контакты).
// @contact.name    EduBase Team
// @contact.email   admin@example.com
// @BasePath        /
// @schemes         http
// @host            localhost:8080
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}
	logger.Init(cfg.AppEnv)
	defer logger.Log.Sync()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Log.Fatal("db pool", logger.Err(err))
	}
	defer pool.Close()

	r := server.NewRouter(cfg, pool)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		logger.Log.Info("HTTP listening", logger.Str("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("server", logger.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
