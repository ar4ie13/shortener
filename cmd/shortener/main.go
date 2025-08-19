package main

import (
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
	repo := repository.NewRepository()
	srv := service.NewService(repo)
	hdlr := handler.NewHandler(srv)

	if err := hdlr.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
