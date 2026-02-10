package postgresql

import (
	"context"
	"errors"
	"testing"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestDB(t *testing.T) (*DB, pgxmock.PgxPoolIface) {
	t.Helper()
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	db := &DB{
		pool: mock,
		zlog: zerolog.Nop(),
	}
	return db, mock
}

func TestDB_Close(t *testing.T) {
	db, mock := newTestDB(t)
	mock.ExpectClose()

	err := db.Close()

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDB_GetShortURL(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		originalURL := "https://example.com"
		expectedShortURL := "abc123"

		mock.ExpectQuery(`SELECT short_url FROM urls WHERE original_url = \$1`).
			WithArgs(originalURL).
			WillReturnRows(pgxmock.NewRows([]string{"short_url"}).AddRow(expectedShortURL))

		shortURL, err := db.GetShortURL(ctx, originalURL)

		assert.NoError(t, err)
		assert.Equal(t, expectedShortURL, shortURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		originalURL := "https://notfound.com"

		mock.ExpectQuery(`SELECT short_url FROM urls WHERE original_url = \$1`).
			WithArgs(originalURL).
			WillReturnError(pgx.ErrNoRows)

		_, err := db.GetShortURL(ctx, originalURL)

		assert.ErrorIs(t, err, myerrors.ErrNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		originalURL := "https://error.com"

		mock.ExpectQuery(`SELECT short_url FROM urls WHERE original_url = \$1`).
			WithArgs(originalURL).
			WillReturnError(errors.New("database error"))

		_, err := db.GetShortURL(ctx, originalURL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan a response row")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_GetURL(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "abc123"
		expectedOriginalURL := "https://example.com"

		mock.ExpectQuery(`SELECT original_url, is_deleted FROM urls WHERE short_url = \$1`).
			WithArgs(shortURL).
			WillReturnRows(pgxmock.NewRows([]string{"original_url", "is_deleted"}).
				AddRow(expectedOriginalURL, false))

		originalURL, err := db.GetURL(ctx, shortURL)

		assert.NoError(t, err)
		assert.Equal(t, expectedOriginalURL, originalURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NotFound", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "notfound"

		mock.ExpectQuery(`SELECT original_url, is_deleted FROM urls WHERE short_url = \$1`).
			WithArgs(shortURL).
			WillReturnError(pgx.ErrNoRows)

		_, err := db.GetURL(ctx, shortURL)

		assert.ErrorIs(t, err, myerrors.ErrNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Deleted", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "deleted123"
		originalURL := "https://deleted.com"

		mock.ExpectQuery(`SELECT original_url, is_deleted FROM urls WHERE short_url = \$1`).
			WithArgs(shortURL).
			WillReturnRows(pgxmock.NewRows([]string{"original_url", "is_deleted"}).
				AddRow(originalURL, true))

		_, err := db.GetURL(ctx, shortURL)

		assert.ErrorIs(t, err, myerrors.ErrShortURLIsDeleted)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "error"

		mock.ExpectQuery(`SELECT original_url, is_deleted FROM urls WHERE short_url = \$1`).
			WithArgs(shortURL).
			WillReturnError(errors.New("database error"))

		_, err := db.GetURL(ctx, shortURL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan a response row")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_Save(t *testing.T) {
	ctx := context.Background()
	userUUID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "abc123"
		originalURL := "https://example.com"

		mock.ExpectExec(`INSERT INTO urls\(short_url, original_url, user_uuid\) VALUES \(\$1, \$2, \$3\)`).
			WithArgs(shortURL, originalURL, userUUID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := db.Save(ctx, userUUID, shortURL, originalURL)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptyShortURL", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		err := db.Save(ctx, userUUID, "", "https://example.com")

		assert.ErrorIs(t, err, myerrors.ErrEmptyShortURLorURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptyOriginalURL", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		err := db.Save(ctx, userUUID, "abc123", "")

		assert.ErrorIs(t, err, myerrors.ErrEmptyShortURLorURL)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DuplicateOriginalURL", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "abc123"
		originalURL := "https://duplicate.com"

		pgErr := &pgconn.PgError{
			Code:    pgerrcode.UniqueViolation,
			Message: "duplicate key value violates unique constraint \"urls_original_url_key\"",
		}

		mock.ExpectExec(`INSERT INTO urls\(short_url, original_url, user_uuid\) VALUES \(\$1, \$2, \$3\)`).
			WithArgs(shortURL, originalURL, userUUID).
			WillReturnError(pgErr)

		err := db.Save(ctx, userUUID, shortURL, originalURL)

		assert.ErrorIs(t, err, myerrors.ErrURLExist)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DuplicateShortURL", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "duplicate"
		originalURL := "https://example.com"

		pgErr := &pgconn.PgError{
			Code:    pgerrcode.UniqueViolation,
			Message: "duplicate key value violates unique constraint \"urls_short_url_key\"",
		}

		mock.ExpectExec(`INSERT INTO urls\(short_url, original_url, user_uuid\) VALUES \(\$1, \$2, \$3\)`).
			WithArgs(shortURL, originalURL, userUUID).
			WillReturnError(pgErr)

		err := db.Save(ctx, userUUID, shortURL, originalURL)

		assert.ErrorIs(t, err, myerrors.ErrShortURLExist)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DatabaseError", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		shortURL := "abc123"
		originalURL := "https://error.com"

		mock.ExpectExec(`INSERT INTO urls\(short_url, original_url, user_uuid\) VALUES \(\$1, \$2, \$3\)`).
			WithArgs(shortURL, originalURL, userUUID).
			WillReturnError(errors.New("connection refused"))

		err := db.Save(ctx, userUUID, shortURL, originalURL)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save URL")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_SaveBatch(t *testing.T) {
	ctx := context.Background()
	userUUID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		batch := []model.URL{
			{UUID: uuid.New(), ShortURL: "s1", OriginalURL: "https://one.com"},
			{UUID: uuid.New(), ShortURL: "s2", OriginalURL: "https://two.com"},
		}

		eb := mock.ExpectBatch()
		eb.ExpectExec(`INSERT INTO urls`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))
		eb.ExpectExec(`INSERT INTO urls`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err := db.SaveBatch(ctx, userUUID, batch)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Empty batch", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		mock.ExpectBatch()

		err := db.SaveBatch(ctx, userUUID, []model.URL{})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("BatchError", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		batch := []model.URL{
			{UUID: uuid.New(), ShortURL: "s1", OriginalURL: "https://one.com"},
		}

		eb := mock.ExpectBatch()
		eb.ExpectExec(`INSERT INTO urls`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errors.New("insert failed"))

		err := db.SaveBatch(ctx, userUUID, batch)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to insert row")
	})

	t.Run("UniqueViolation", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		batch := []model.URL{
			{UUID: uuid.New(), ShortURL: "s1", OriginalURL: "https://duplicate.com"},
		}

		pgErr := &pgconn.PgError{
			Code:    pgerrcode.UniqueViolation,
			Message: "duplicate",
		}

		eb := mock.ExpectBatch()
		eb.ExpectExec(`INSERT INTO urls`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(pgErr)

		err := db.SaveBatch(ctx, userUUID, batch)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "https://duplicate.com")
	})
}

func TestDB_GetUserShortURLs(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		userUUID := uuid.New()

		mock.ExpectQuery(`SELECT short_url, original_url FROM urls WHERE user_uuid = \$1 and is_deleted = false`).
			WithArgs(userUUID).
			WillReturnRows(pgxmock.NewRows([]string{"short_url", "original_url"}).
				AddRow("s1", "https://one.com").
				AddRow("s2", "https://two.com"))

		result, err := db.GetUserShortURLs(ctx, userUUID)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "https://one.com", result["s1"])
		assert.Equal(t, "https://two.com", result["s2"])
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("NoURLs", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		userUUID := uuid.New()

		mock.ExpectQuery(`SELECT short_url, original_url FROM urls WHERE user_uuid = \$1 and is_deleted = false`).
			WithArgs(userUUID).
			WillReturnRows(pgxmock.NewRows([]string{"short_url", "original_url"}))

		_, err := db.GetUserShortURLs(ctx, userUUID)

		assert.ErrorIs(t, err, myerrors.ErrNotFound)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("QueryError", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		userUUID := uuid.New()

		mock.ExpectQuery(`SELECT short_url, original_url FROM urls WHERE user_uuid = \$1 and is_deleted = false`).
			WithArgs(userUUID).
			WillReturnError(errors.New("query failed"))

		_, err := db.GetUserShortURLs(ctx, userUUID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ScanError", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		userUUID := uuid.New()

		mock.ExpectQuery(`SELECT short_url, original_url FROM urls WHERE user_uuid = \$1 and is_deleted = false`).
			WithArgs(userUUID).
			WillReturnRows(pgxmock.NewRows([]string{"short_url", "original_url"}).
				AddRow("s1", "https://one.com").
				RowError(0, errors.New("row error")))

		_, err := db.GetUserShortURLs(ctx, userUUID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDB_DeleteUserShortURLs(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		userUUID := uuid.New()
		toDelete := map[uuid.UUID][]string{
			userUUID: {"s1", "s2"},
		}

		eb := mock.ExpectBatch()
		eb.ExpectExec(`UPDATE urls SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		eb.ExpectExec(`UPDATE urls SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := db.DeleteUserShortURLs(ctx, toDelete)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("EmptyMap", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		err := db.DeleteUserShortURLs(ctx, map[uuid.UUID][]string{})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("MultipleUsers", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		user1 := uuid.New()
		user2 := uuid.New()
		toDelete := map[uuid.UUID][]string{
			user1: {"s1"},
			user2: {"s2", "s3"},
		}

		eb := mock.ExpectBatch()
		eb.ExpectExec(`UPDATE urls SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		eb.ExpectExec(`UPDATE urls SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		eb.ExpectExec(`UPDATE urls SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err := db.DeleteUserShortURLs(ctx, toDelete)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateError", func(t *testing.T) {
		db, mock := newTestDB(t)
		defer mock.Close()

		userUUID := uuid.New()
		toDelete := map[uuid.UUID][]string{
			userUUID: {"s1"},
		}

		eb := mock.ExpectBatch()
		eb.ExpectExec(`UPDATE urls SET is_deleted = true`).
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(errors.New("update failed"))

		err := db.DeleteUserShortURLs(ctx, toDelete)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to update row")
	})
}
