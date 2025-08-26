package main

import (
	"flag"
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
	cfg.InitConfig()
	flag.Parse()

	repo := repository.NewRepository()
	srv := service.NewService(repo)
	hdlr := handler.NewHandler(srv, cfg)

	if err := hdlr.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
