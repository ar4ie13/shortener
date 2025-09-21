package main

import (
	"fmt"
	"github.com/ar4ie13/shortener/internal/config"
	"github.com/ar4ie13/shortener/internal/handler"
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
	repo := repository.NewRepository()
	srv := service.NewService(repo)
	hdlr := handler.NewHandler(srv, cfg)

	if err := hdlr.ListenAndServe(); err != nil {
		return fmt.Errorf("shortener service error: %w", err)
	}

	return nil
}
