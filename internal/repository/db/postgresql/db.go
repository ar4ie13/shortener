package postgresql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ar4ie13/shortener/internal/repository/db/postgresql/config"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

type DB struct {
	pool *pgxpool.Pool
	zlog zerolog.Logger
}

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

func initPool(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}
	//poolCfg.ConnConfig.Tracer = &queryTracer{}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

func (db *DB) Close() error {
	db.pool.Close()
	return nil
}

func (db *DB) Get(ctx context.Context, shortURL string) (originalURL string, err error) {

	const queryStmt = `SELECT original_url FROM urls WHERE short_url = $1`

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		db.zlog.Debug().Msgf("request execution duration: %s", elapsed)
	}()

	row := db.pool.QueryRow(ctx, queryStmt, shortURL)

	err = row.Scan(&originalURL)
	if err != nil {
		return "", fmt.Errorf("failed to scan a response row: %w", err)
	}

	return originalURL, nil
}

func (db *DB) Save(ctx context.Context, shortURL string, originalURL string) error {

	const (
		queryStmtInsert        = `INSERT INTO urls(short_url, original_url) VALUES ($1, $2)`
		queryStmtCheckShortURL = `SELECT short_url FROM urls WHERE short_url = $1 LIMIT 1`
		queryStmtCheckURL      = `SELECT original_url FROM urls WHERE original_url = $1 LIMIT 1`
	)

	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		db.zlog.Debug().Msgf("request execution duration: %s", elapsed)
	}()

	var check string
	row := db.pool.QueryRow(ctx, queryStmtCheckURL, originalURL)
	if err := row.Scan(&check); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check the URL: %w", err)
	}
	if check == originalURL {
		return service.ErrURLExist
	}

	row = db.pool.QueryRow(ctx, queryStmtCheckShortURL, shortURL)
	if err := row.Scan(&check); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to check the ShortURL: %w", err)
	}
	if check == shortURL {
		return service.ErrShortURLExist
	}

	_, err := db.pool.Exec(ctx, queryStmtInsert, shortURL, originalURL)
	if err != nil {
		return fmt.Errorf("failed to save URL: %w", err)
	}

	db.zlog.Debug().Msgf("saved URL: %s", shortURL)

	return nil
}

// "host=localhost port=5432 user=shortener password=shortener dbname=shortener sslmode=disable"

/*
func CheckConn() error {
	ps := fmt.Sprint("host=localhost port=5432 user=videos password=userpassword dbname=videos sslmode=disable")

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return err
	}
	defer db.Close()
	// ...

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		panic(err)
	}
	return nil
}

*/
