package repository

import (
	"context"

	"github.com/ar4ie13/shortener/internal/repository/db/postgresql"
	pgconf "github.com/ar4ie13/shortener/internal/repository/db/postgresql/config"
	"github.com/ar4ie13/shortener/internal/repository/filestorage"
	fileconf "github.com/ar4ie13/shortener/internal/repository/filestorage/config"
	"github.com/ar4ie13/shortener/internal/repository/memory"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/rs/zerolog"
)

// Repository is a main repository object
type Repository struct {
	m  *memory.MemStorage
	f  *filestorage.FileStorage
	db *postgresql.DB
}

// NewRepository return the correct interface for service depending on used store method
func NewRepository(
	ctx context.Context,
	fileconf fileconf.Config,
	pgcfg pgconf.Config,
	zlog zerolog.Logger,
) (service.Repository, error) {
	switch {
	case pgcfg.DatabaseDSN != "":
		db, err := postgresql.NewDB(ctx, pgcfg, zlog)
		if err != nil {
			return nil, err
		}
		err = postgresql.ApplyMigrations(pgcfg, zlog)
		if err != nil {
			return nil, err
		}
		return db, nil
	case fileconf.FilePath != "":
		filestore := filestorage.NewFileStorage(fileconf, zlog)
		err := filestore.Load()
		if err != nil {
			return nil, err
		}
		return filestore, nil
	default:
		return memory.NewMemStorage(), nil
	}
}
