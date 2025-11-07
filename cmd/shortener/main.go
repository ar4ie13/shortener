package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ar4ie13/shortener/internal/auth"
	"github.com/ar4ie13/shortener/internal/config"
	"github.com/ar4ie13/shortener/internal/handlers"
	"github.com/ar4ie13/shortener/internal/logger"
	"github.com/ar4ie13/shortener/internal/repository"
	"github.com/ar4ie13/shortener/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.NewConfig()
	zlog := logger.NewLogger(cfg.GetLogLevel())
	authorize := auth.NewAuth()
	repo, err := repository.NewRepository(context.Background(), cfg.FilePath, cfg.PostgresDSN, zlog.Logger)
	if err != nil {
		return fmt.Errorf("cannot initialize repository: %w", err)
	}
	srv := service.NewService(repo, zlog.Logger)
	hdlr := handlers.NewHandler(srv, cfg, authorize, zlog.Logger)

	if err = hdlr.ListenAndServe(); err != nil {
		return fmt.Errorf("shortener service error: %w", err)
	}

	return nil
}
