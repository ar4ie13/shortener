package main

import (
	"fmt"
	"github.com/ar4ie13/shortener/internal/config"
	"github.com/ar4ie13/shortener/internal/handler"
	"github.com/ar4ie13/shortener/internal/logger"
	"github.com/ar4ie13/shortener/internal/repository"
	"github.com/ar4ie13/shortener/internal/service"
	"log"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.NewConfig()
	zlog := logger.NewLogger(cfg.GetLogLevel())
	repo, err := repository.NewRepository(cfg.GetFileStorage(), zlog.Logger)
	if err != nil {
		return err
	}
	srv := service.NewService(repo)
	hdlr := handler.NewHandler(srv, cfg, zlog.Logger)

	if err = hdlr.ListenAndServe(); err != nil {
		return fmt.Errorf("shortener service error: %w", err)
	}

	return nil
}
