package memory

import (
	"context"
	"testing"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_GetURL(t *testing.T) {
	repo := NewMemStorage()
	ctx := context.Background()

	shortURL := "abc123"
	originalURL := "https://example.com"

	// Save first
	userUUID := uuid.New()
	err := repo.Save(ctx, userUUID, shortURL, originalURL)
	require.NoError(t, err)

	// Retrieve
	got, err := repo.GetURL(ctx, shortURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, got)

	// Non-existent slug
	_, err = repo.GetURL(ctx, "nonexistent")
	assert.ErrorIs(t, err, myerrors.ErrNotFound)

	// Deleted slug
	repo.IsSlugDeletedMemStore[shortURL] = true
	_, err = repo.GetURL(ctx, shortURL)
	assert.ErrorIs(t, err, myerrors.ErrShortURLIsDeleted)
}

func TestMemStorage_GetShortURL(t *testing.T) {
	repo := NewMemStorage()
	ctx := context.Background()

	shortURL := "xyz789"
	originalURL := "https://example.org"

	userUUID := uuid.New()
	err := repo.Save(ctx, userUUID, shortURL, originalURL)
	require.NoError(t, err)

	// Retrieve
	got, err := repo.GetShortURL(ctx, originalURL)
	assert.NoError(t, err)
	assert.Equal(t, shortURL, got)

	// Non-existent URL
	_, err = repo.GetShortURL(ctx, "https://notfound.com")
	assert.ErrorIs(t, err, myerrors.ErrNotFound)

	// Deleted slug
	repo.IsSlugDeletedMemStore[shortURL] = true
	_, err = repo.GetShortURL(ctx, originalURL)
	assert.ErrorIs(t, err, myerrors.ErrNotFound)
}

func TestMemStorage_Save(t *testing.T) {
	repo := NewMemStorage()
	ctx := context.Background()
	userUUID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		err := repo.Save(ctx, userUUID, "test1", "https://valid.com")
		assert.NoError(t, err)
	})

	t.Run("Empty shortURL or URL", func(t *testing.T) {
		err := repo.Save(ctx, userUUID, "", "https://valid.com")
		assert.ErrorIs(t, err, myerrors.ErrEmptyShortURLorURL)

		err = repo.Save(ctx, userUUID, "test2", "")
		assert.ErrorIs(t, err, myerrors.ErrEmptyShortURLorURL)
	})

	t.Run("Duplicate URL", func(t *testing.T) {
		err := repo.Save(ctx, userUUID, "test3", "https://duplicate.com")
		require.NoError(t, err)

		err = repo.Save(ctx, userUUID, "test4", "https://duplicate.com")
		assert.ErrorIs(t, err, myerrors.ErrURLExist)
	})

	t.Run("Duplicate ShortURL", func(t *testing.T) {
		err := repo.Save(ctx, userUUID, "conflict", "https://first.com")
		require.NoError(t, err)

		err = repo.Save(ctx, userUUID, "conflict", "https://second.com")
		assert.ErrorIs(t, err, myerrors.ErrShortURLExist)
	})
}

func TestMemStorage_SaveBatch(t *testing.T) {
	repo := NewMemStorage()
	ctx := context.Background()
	userUUID := uuid.New()

	batch := []model.URL{
		{UUID: uuid.New(), ShortURL: "b1", OriginalURL: "https://b1.com"},
		{UUID: uuid.New(), ShortURL: "b2", OriginalURL: "https://b2.com"},
	}

	t.Run("Success", func(t *testing.T) {
		err := repo.SaveBatch(ctx, userUUID, batch)
		assert.NoError(t, err)

		// Verify mappings
		for _, item := range batch {
			url, ok := repo.SlugMemStore[item.ShortURL]
			assert.True(t, ok)
			assert.Equal(t, item.OriginalURL, url)

			slug, ok := repo.URLMemStore[item.OriginalURL]
			assert.True(t, ok)
			assert.Equal(t, item.ShortURL, slug)

			assert.Equal(t, item.OriginalURL, repo.UserUUIDSlugMemStore[userUUID][item.ShortURL])
			assert.Equal(t, item.ShortURL, repo.UserUUIDURLMemStore[userUUID][item.OriginalURL])
			assert.Equal(t, item.ShortURL, repo.UUIDMemStore[item.UUID])
			assert.False(t, repo.IsSlugDeletedMemStore[item.ShortURL])
		}
	})

	t.Run("Empty fields", func(t *testing.T) {
		badBatch := []model.URL{{ShortURL: "", OriginalURL: "https://bad.com"}}
		err := repo.SaveBatch(ctx, userUUID, badBatch)
		assert.ErrorIs(t, err, myerrors.ErrEmptyShortURLorURL)
	})

	t.Run("Conflict with existing data", func(t *testing.T) {
		// First save one URL
		err := repo.Save(ctx, userUUID, "existing", "https://exists.com")
		require.NoError(t, err)

		// Try to batch-save same URL
		conflictBatch := []model.URL{
			{UUID: uuid.New(), ShortURL: "new", OriginalURL: "https://exists.com"},
		}
		err = repo.SaveBatch(ctx, userUUID, conflictBatch)
		assert.ErrorIs(t, err, myerrors.ErrURLExist)
	})
}

func TestMemStorage_GetUserShortURLs(t *testing.T) {
	repo := NewMemStorage()
	ctx := context.Background()
	user1 := uuid.New()
	user2 := uuid.New()

	// Add URLs for user1
	repo.Save(ctx, user1, "u1s1", "https://u1-1.com")
	repo.Save(ctx, user1, "u1s2", "https://u1-2.com")

	// Add for user2
	repo.Save(ctx, user2, "u2s1", "https://u2-1.com")

	t.Run("Valid user", func(t *testing.T) {
		result, err := repo.GetUserShortURLs(ctx, user1)
		assert.NoError(t, err)
		expected := map[string]string{
			"u1s1": "https://u1-1.com",
			"u1s2": "https://u1-2.com",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Non-existent user", func(t *testing.T) {
		_, err := repo.GetUserShortURLs(ctx, uuid.New())
		assert.ErrorIs(t, err, myerrors.ErrNotFound)
	})

	t.Run("Excludes deleted slugs", func(t *testing.T) {
		repo.IsSlugDeletedMemStore["u1s1"] = true
		result, err := repo.GetUserShortURLs(ctx, user1)
		assert.NoError(t, err)
		expected := map[string]string{
			"u1s2": "https://u1-2.com",
		}
		assert.Equal(t, expected, result)
	})
}

func TestMemStorage_DeleteUserShortURLs(t *testing.T) {
	repo := NewMemStorage()
	ctx := context.Background()
	userUUID := uuid.New()

	// Save some URLs
	repo.Save(ctx, userUUID, "to-delete", "https://delete.me")
	repo.Save(ctx, userUUID, "keep", "https://keep.me")

	t.Run("Success deletion", func(t *testing.T) {
		toDelete := map[uuid.UUID][]string{
			userUUID: {"to-delete"},
		}
		err := repo.DeleteUserShortURLs(ctx, toDelete)
		assert.NoError(t, err)
		assert.True(t, repo.IsSlugDeletedMemStore["to-delete"])
		assert.False(t, repo.IsSlugDeletedMemStore["keep"])
	})

	t.Run("Invalid user UUID", func(t *testing.T) {
		toDelete := map[uuid.UUID][]string{
			uuid.New(): {"any"},
		}
		err := repo.DeleteUserShortURLs(ctx, toDelete)
		assert.ErrorIs(t, err, myerrors.ErrInvalidUserUUID)
	})

	t.Run("Non-existent slug in user's list", func(t *testing.T) {
		// Should not panic or error; just skip
		toDelete := map[uuid.UUID][]string{
			userUUID: {"nonexistent-slug"},
		}
		err := repo.DeleteUserShortURLs(ctx, toDelete)
		assert.NoError(t, err)
		// Ensure no side effect
		assert.False(t, repo.IsSlugDeletedMemStore["nonexistent-slug"]) // remains false or unset
	})
}
