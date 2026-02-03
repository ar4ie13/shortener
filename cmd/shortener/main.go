// This is a starting point for shortener service. It initializes configuration, objects and starts web server.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ar4ie13/shortener/internal/auditor"
	"github.com/ar4ie13/shortener/internal/auth"
	"github.com/ar4ie13/shortener/internal/config"
	"github.com/ar4ie13/shortener/internal/handlers"
	"github.com/ar4ie13/shortener/internal/logger"
	"github.com/ar4ie13/shortener/internal/repository"
	"github.com/ar4ie13/shortener/internal/service"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// main starts run functions which contains all objects initialization
func main() {
	fmt.Println("Build version: " + buildVersion)
	fmt.Println("Build date: " + buildDate)
	fmt.Println("Build commit: " + buildCommit)
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run function is used to init configuration, create all objects and start web server
func run() error {
	cfg := config.NewConfig()
	zlog := logger.NewLogger(cfg.GetLogLevel())
	authorize := auth.NewAuth(cfg.AuthConf)
	repo, err := repository.NewRepository(context.Background(), cfg.FilePath, cfg.PostgresDSN, zlog.Logger)
	if err != nil {
		return fmt.Errorf("cannot initialize repository: %w", err)
	}
	srv := service.NewService(repo, zlog.Logger)
	audit := auditor.NewAuditor(cfg.AuditConf, zlog.Logger)
	hdlr := handlers.NewHandler(srv, cfg, authorize, audit, zlog.Logger)

	if err = hdlr.ListenAndServe(); err != nil {
		return fmt.Errorf("HTTP server ListenAndServe: %v", err)
	}

	return nil
}
