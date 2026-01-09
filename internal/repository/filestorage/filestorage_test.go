package filestorage

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	fileconf "github.com/ar4ie13/shortener/internal/repository/filestorage/config"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_LoadFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_load.jsonl")

	// Prepare test data
	data := []model.URL{
		{
			UUID:        uuid.New(),
			UserUUID:    uuid.New(),
			ShortURL:    "abc123",
			OriginalURL: "https://example.com",
			IsDeleted:   false,
		},
		{
			UUID:        uuid.New(),
			UserUUID:    uuid.New(),
			ShortURL:    "def456",
			OriginalURL: "https://example.org",
			IsDeleted:   true,
		},
	}

	// Write JSONL file
	file, err := os.Create(filePath)
	require.NoError(t, err)
	defer file.Close()

	for _, item := range data {
		line, _ := json.Marshal(item)
		file.Write(line)
		file.WriteString("\n")
	}

	// Load into FileStorage
	fs := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())
	err = fs.LoadFile()
	require.NoError(t, err)

	// Validate memory state
	assert.Equal(t, "https://example.com", fs.m.SlugMemStore["abc123"])
	assert.Equal(t, "https://example.org", fs.m.SlugMemStore["def456"])
	assert.Equal(t, "abc123", fs.m.URLMemStore["https://example.com"])
	assert.False(t, fs.m.IsSlugDeletedMemStore["abc123"])
	assert.True(t, fs.m.IsSlugDeletedMemStore["def456"])

	// Check user maps
	user1 := data[0].UserUUID
	assert.Equal(t, "https://example.com", fs.m.UserUUIDSlugMemStore[user1]["abc123"])
	assert.Equal(t, "abc123", fs.m.UserUUIDURLMemStore[user1]["https://example.com"])
}

func TestFileStorage_LoadFile_EmptyOrMissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Non-existent file → should not error
	nonExistent := filepath.Join(tmpDir, "missing.jsonl")
	fs := NewFileStorage(fileconf.Config{FilePath: nonExistent}, zerolog.Nop())
	err := fs.LoadFile()
	require.NoError(t, err)

	// Empty file
	emptyFile := filepath.Join(tmpDir, "empty.jsonl")
	os.WriteFile(emptyFile, []byte{}, 0644)
	fs2 := NewFileStorage(fileconf.Config{FilePath: emptyFile}, zerolog.Nop())
	err = fs2.LoadFile()
	require.NoError(t, err)
	assert.Empty(t, fs2.m.SlugMemStore)
}

func TestFileStorage_Save(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "save_test.jsonl")

	userUUID := uuid.New()
	fs := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())

	ctx := context.Background()
	err := fs.Save(ctx, userUUID, "xyz789", "https://save.test")
	require.NoError(t, err)

	// Verify in memory
	url, err := fs.GetURL(ctx, "xyz789")
	require.NoError(t, err)
	assert.Equal(t, "https://save.test", url)

	slug, err := fs.GetShortURL(ctx, "https://save.test")
	require.NoError(t, err)
	assert.Equal(t, "xyz789", slug)

	// Verify file content — use correct JSON field names (snake_case)
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	lines := string(content)
	assert.Contains(t, lines, `"short_url":"xyz789"`)
	assert.Contains(t, lines, `"original_url":"https://save.test"`)
	assert.Contains(t, lines, `"is_deleted":false`)
}

func TestFileStorage_SaveBatch(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "batch_test.jsonl")

	userUUID := uuid.New()
	batch := []model.URL{
		{UUID: uuid.New(), ShortURL: "b1", OriginalURL: "https://b1.com"},
		{UUID: uuid.New(), ShortURL: "b2", OriginalURL: "https://b2.com"},
	}

	fs := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())
	ctx := context.Background()

	err := fs.SaveBatch(ctx, userUUID, batch)
	require.NoError(t, err)

	// Check memory
	for _, item := range batch {
		slug, _ := fs.GetShortURL(ctx, item.OriginalURL)
		assert.Equal(t, item.ShortURL, slug)
		url, _ := fs.GetURL(ctx, item.ShortURL)
		assert.Equal(t, item.OriginalURL, url)
	}

	// Check file has 2 lines
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, 2, len(bytes.Split(bytes.TrimSpace(content), []byte("\n"))))
}

func TestFileStorage_GetUserShortURLs(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "user_urls.jsonl")

	user1 := uuid.New()
	user2 := uuid.New()

	fs := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())
	ctx := context.Background()

	fs.Save(ctx, user1, "u1-1", "https://u1a.com")
	fs.Save(ctx, user1, "u1-2", "https://u1b.com")
	fs.Save(ctx, user2, "u2-1", "https://u2a.com")

	result, err := fs.GetUserShortURLs(ctx, user1)
	require.NoError(t, err)
	expected := map[string]string{
		"u1-1": "https://u1a.com",
		"u1-2": "https://u1b.com",
	}
	assert.Equal(t, expected, result)

	// Non-existent user
	_, err = fs.GetUserShortURLs(ctx, uuid.New())
	assert.ErrorIs(t, err, myerrors.ErrNotFound)
}

func TestFileStorage_DeleteUserShortURLs(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "delete_test.jsonl")

	userUUID := uuid.New()
	fs := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())
	ctx := context.Background()

	// Save two URLs
	fs.Save(ctx, userUUID, "to-delete", "https://del.me")
	fs.Save(ctx, userUUID, "keep", "https://keep.me")

	// Delete one
	toDelete := map[uuid.UUID][]string{
		userUUID: {"to-delete"},
	}
	err := fs.DeleteUserShortURLs(ctx, toDelete)
	require.NoError(t, err)

	// Check in memory
	_, err = fs.GetURL(ctx, "to-delete")
	assert.ErrorIs(t, err, myerrors.ErrShortURLIsDeleted)

	url, err := fs.GetURL(ctx, "keep")
	require.NoError(t, err)
	assert.Equal(t, "https://keep.me", url)

	// Reload from file to ensure persistence
	fs2 := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())
	err = fs2.LoadFile()
	require.NoError(t, err)

	_, err = fs2.GetURL(ctx, "to-delete")
	assert.ErrorIs(t, err, myerrors.ErrShortURLIsDeleted)

	url, err = fs2.GetURL(ctx, "keep")
	require.NoError(t, err)
	assert.Equal(t, "https://keep.me", url)
}

func TestFileStorage_InvalidJSONInFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "invalid.jsonl")

	// Write invalid JSON
	os.WriteFile(filePath, []byte(`{"ShortURL": "bad", "OriginalURL": "https://bad.com"}\nNOT_JSON`), 0644)

	fs := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())
	err := fs.LoadFile()
	// Should fail on invalid JSON
	assert.Error(t, err)
	assert.NotContains(t, fs.m.SlugMemStore, "bad") // partial load should not happen
}

func TestFileStorage_Save_ErrorOnEmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "empty_save.jsonl")
	fs := NewFileStorage(fileconf.Config{FilePath: filePath}, zerolog.Nop())
	ctx := context.Background()
	userUUID := uuid.New()

	err := fs.Save(ctx, userUUID, "", "https://valid.com")
	assert.ErrorIs(t, err, myerrors.ErrEmptyShortURLorURL)

	err = fs.Save(ctx, userUUID, "valid", "")
	assert.ErrorIs(t, err, myerrors.ErrEmptyShortURLorURL)
}
