package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/repository/db/postgresql/config"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

// DB is a main postgres repository object
type DB struct {
	pool *pgxpool.Pool
	zlog zerolog.Logger
}

// NewDB construct postgres DB object
func NewDB(ctx context.Context, cfg config.Config, zlog zerolog.Logger) (*DB, error) {
	pool, err := initPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	return &DB{
		pool: pool,
		zlog: zlog,
	}, nil
}

// initPool initializes pgx connection pool
func initPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

// Close closes pgx pool
func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

// GetShortURL gets short_url from db by provided URL
func (db *DB) GetShortURL(ctx context.Context, originalURL string) (shortURL string, err error) {

	const queryStmt = `SELECT short_url FROM urls WHERE original_url = $1`

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		db.zlog.Debug().Msgf("request execution duration: %s", elapsed)
	}()

	row := db.pool.QueryRow(ctx, queryStmt, originalURL)

	err = row.Scan(&shortURL)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return "", service.ErrNotFound
		default:
			return "", fmt.Errorf("failed to scan a response row: %w", err)
		}
	}

	return shortURL, nil
}

// GetURL gets URL by provided shortURL
func (db *DB) GetURL(ctx context.Context, shortURL string) (originalURL string, err error) {
	var isDeleted bool
	const queryStmt = `SELECT original_url, is_deleted FROM urls WHERE short_url = $1`

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		db.zlog.Debug().Msgf("request execution duration: %s", elapsed)
	}()

	row := db.pool.QueryRow(ctx, queryStmt, shortURL)

	err = row.Scan(&originalURL, &isDeleted)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return "", service.ErrNotFound
		default:
			return "", fmt.Errorf("failed to scan a response row: %w", err)
		}
	}
	if isDeleted {
		return "", service.ErrShortURLIsDeleted
	}

	return originalURL, nil
}

// Save saves tuple with shortURL, URL and UUID
func (db *DB) Save(ctx context.Context, userUUID uuid.UUID, shortURL string, originalURL string) error {

	if shortURL == "" || originalURL == "" {
		return service.ErrEmptyShortURLorURL
	}

	const (
		queryStmtInsert = `INSERT INTO urls(short_url, original_url, user_uuid) VALUES ($1, $2, $3)`
	)

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		db.zlog.Debug().Msgf("request execution duration: %s", elapsed)
	}()

	_, err := db.pool.Exec(ctx, queryStmtInsert, shortURL, originalURL, userUUID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			switch {
			case strings.Contains(err.Error(), "urls_original_url_key"):
				return fmt.Errorf("error while saving URL %s: %w", originalURL, service.ErrURLExist)
			case strings.Contains(err.Error(), "urls_short_url_key"):
				return fmt.Errorf("error while saving URL %s: %w", shortURL, service.ErrShortURLExist)
			}
		}
		return fmt.Errorf("failed to save URL: %w", err)
	}

	db.zlog.Debug().Msgf("saved URL: %s", shortURL)

	return nil
}

// SaveBatch performs bulk insert to postgres database
func (db *DB) SaveBatch(ctx context.Context, userUUID uuid.UUID, batch []model.URL) error {
	query := `INSERT INTO urls (uuid, short_url, original_url, user_uuid) VALUES (@uuid, @shortURL, @originalURL, @userUUID)`

	insertBatch := &pgx.Batch{}
	for _, v := range batch {
		args := pgx.NamedArgs{
			"uuid":        v.UUID,
			"shortURL":    v.ShortURL,
			"originalURL": v.OriginalURL,
			"userUUID":    userUUID,
		}
		insertBatch.Queue(query, args)
	}

	results := db.pool.SendBatch(ctx, insertBatch)
	defer results.Close()

	for _, v := range batch {
		_, err := results.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return fmt.Errorf("error while saving URL %s: %w", v.OriginalURL, err)
			}
			return fmt.Errorf("unable to insert row: %w", err)
		}
	}

	return results.Close()
}

func (db *DB) GetUserShortURLs(ctx context.Context, userUUID uuid.UUID) (map[string]string, error) {
	const queryStmt = `SELECT short_url, original_url FROM urls WHERE user_uuid = $1 and is_deleted = false`

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		db.zlog.Debug().Msgf("request execution duration: %s", elapsed)
	}()

	rows, err := db.pool.Query(ctx, queryStmt, userUUID)
	if err != nil {
		return nil, err
	}

	//if !rows.Next() {
	//	return nil, service.ErrNotFound
	//}

	userShortURLs := make(map[string]string)
	for rows.Next() {
		var shortURL string
		var originalURL string

		err = rows.Scan(&shortURL, &originalURL)
		if err != nil {
			return nil, err
		}
		userShortURLs[shortURL] = originalURL
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(userShortURLs) == 0 {
		return nil, service.ErrNotFound
	}

	return userShortURLs, nil
}
func (db *DB) DeleteUserShortURLs(ctx context.Context, shortURLsToDelete map[uuid.UUID][]string) error {

	query := `UPDATE urls SET is_deleted = true WHERE short_url = @shortURL AND user_uuid = @userUUID`
	insertBatch := &pgx.Batch{}
	for k, v := range shortURLsToDelete {
		for i := range v {
			args := pgx.NamedArgs{
				"userUUID": k,
				"shortURL": v[i],
			}
			insertBatch.Queue(query, args)
		}
	}

	results := db.pool.SendBatch(ctx, insertBatch)
	defer results.Close()

	for range shortURLsToDelete {
		_, err := results.Exec()
		if err != nil {
			//var pgErr *pgconn.PgError
			//if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			//	return fmt.Errorf("error while saving URL %s: %w", v.OriginalURL, err)
			//}
			return fmt.Errorf("unable to update row: %w", err)
		}
	}

	return results.Close()
}
