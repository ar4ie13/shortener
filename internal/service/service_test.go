package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/myerrors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) GetURL(ctx context.Context, shortURL string) (string, error) {
	args := m.Called(ctx, shortURL)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	args := m.Called(ctx, originalURL)
	return args.String(0), args.Error(1)
}

func (m *MockRepository) Save(ctx context.Context, userUUID uuid.UUID, shortURL string, url string) error {
	args := m.Called(ctx, userUUID, shortURL, url)
	return args.Error(0)
}

func (m *MockRepository) SaveBatch(ctx context.Context, userUUID uuid.UUID, batch []model.URL) error {
	args := m.Called(ctx, userUUID, batch)
	return args.Error(0)
}

func (m *MockRepository) GetUserShortURLs(ctx context.Context, userUUID uuid.UUID) (map[string]string, error) {
	args := m.Called(ctx, userUUID)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockRepository) DeleteUserShortURLs(ctx context.Context, shortURLsToDelete map[uuid.UUID][]string) error {
	args := m.Called(ctx, shortURLsToDelete)
	return args.Error(0)
}

func TestNewService(t *testing.T) {
	logger := zerolog.Nop()
	mockRepo := &MockRepository{}

	service := NewService(mockRepo, logger)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
	assert.NotNil(t, service.toDeleteChan)
	assert.Len(t, service.toDeleteChan, 0)
}

func TestService_GetURL(t *testing.T) {
	tests := []struct {
		name           string
		shortURL       string
		mockReturn     string
		mockError      error
		expectedURL    string
		expectErrorIs  error  // Use errors.Is for comparison
		expectedErrMsg string // Expected error message substring
	}{
		{
			name:          "empty short URL",
			shortURL:      "",
			expectErrorIs: myerrors.ErrEmptyID,
		},
		{
			name:        "successful retrieval",
			shortURL:    "abc123",
			mockReturn:  "https://example.com",
			mockError:   nil,
			expectedURL: "https://example.com",
		},
		{
			name:           "repository returns error",
			shortURL:       "abc123",
			mockReturn:     "",
			mockError:      errors.New("repo error"),
			expectedErrMsg: "failed to get URL: repo error",
		},
		{
			name:           "repository returns empty URL",
			shortURL:       "abc123",
			mockReturn:     "",
			mockError:      nil,
			expectedErrMsg: "failed to get URL:", // Match the prefix since exact error depends on implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			logger := zerolog.Nop()
			service := &Service{repo: mockRepo, zlog: logger}

			if tt.shortURL != "" {
				mockRepo.On("GetURL", mock.Anything, tt.shortURL).Return(tt.mockReturn, tt.mockError)
			}

			ctx := context.Background()
			userUUID := uuid.New()

			result, err := service.GetURL(ctx, userUUID, tt.shortURL)

			if tt.expectErrorIs != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectErrorIs))
			} else if tt.expectedErrMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, result)
			}

			if tt.shortURL != "" {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestService_GetUserShortURLs(t *testing.T) {
	tests := []struct {
		name           string
		mockReturn     map[string]string
		mockError      error
		expectedResult map[string]string
		expectedError  error
	}{
		{
			name:           "successful retrieval",
			mockReturn:     map[string]string{"abc123": "https://example.com"},
			mockError:      nil,
			expectedResult: map[string]string{"abc123": "https://example.com"},
			expectedError:  nil,
		},
		{
			name:           "repository returns error",
			mockReturn:     nil,
			mockError:      errors.New("repo error"),
			expectedResult: nil,
			expectedError:  errors.New("failed to get short urls: repo error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			logger := zerolog.Nop()
			service := &Service{repo: mockRepo, zlog: logger}

			userUUID := uuid.New()
			mockRepo.On("GetUserShortURLs", mock.Anything, userUUID).Return(tt.mockReturn, tt.mockError)

			ctx := context.Background()
			result, err := service.GetUserShortURLs(ctx, userUUID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_SaveURL(t *testing.T) {
	tests := []struct {
		name                  string
		urlLink               string
		mockSaveError         error
		mockGetShortURLError  error
		mockGetShortURLResult string
		expectedSlug          string
		expectedErrorIs       error
		shouldMockSave        bool
	}{
		{
			name:            "empty URL",
			urlLink:         "",
			expectedErrorIs: myerrors.ErrEmptyURL,
			shouldMockSave:  false,
		},
		{
			name:            "missing scheme",
			urlLink:         "example.com",
			expectedErrorIs: myerrors.ErrWrongHTTPScheme,
			shouldMockSave:  false,
		},
		{
			name:            "wrong scheme - ftp",
			urlLink:         "ftp://example.com",
			expectedErrorIs: myerrors.ErrWrongHTTPScheme,
			shouldMockSave:  false,
		},
		{
			name:            "wrong scheme - mailto",
			urlLink:         "mailto:test@example.com",
			expectedErrorIs: myerrors.ErrWrongHTTPScheme,
			shouldMockSave:  false,
		},
		{
			name:            "wrong scheme - file",
			urlLink:         "file:///path/to/file",
			expectedErrorIs: myerrors.ErrWrongHTTPScheme,
			shouldMockSave:  false,
		},
		{
			name:            "missing host",
			urlLink:         "http://",
			expectedErrorIs: myerrors.ErrMustIncludeHost,
			shouldMockSave:  false,
		},
		{
			name:            "valid URL successful save",
			urlLink:         "http://example.com/",
			mockSaveError:   nil,
			expectedErrorIs: nil,
			shouldMockSave:  true,
		},
		{
			name:                  "URL already exists",
			urlLink:               "http://example.com",
			mockSaveError:         myerrors.ErrURLExist,
			mockGetShortURLResult: "existing-slug",
			mockGetShortURLError:  nil,
			expectedSlug:          "existing-slug",
			expectedErrorIs:       myerrors.ErrURLExist,
			shouldMockSave:        true,
		},
		{
			name:                 "URL exists but can't get short URL",
			urlLink:              "http://example.com",
			mockSaveError:        myerrors.ErrURLExist,
			mockGetShortURLError: errors.New("not found"),
			expectedErrorIs:      myerrors.ErrNotFound,
			shouldMockSave:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			logger := zerolog.Nop()
			service := &Service{repo: mockRepo, zlog: logger}

			userUUID := uuid.New()
			ctx := context.Background()

			if tt.shouldMockSave {
				if tt.mockSaveError == nil {
					// For successful save, we expect Save to be called once
					mockRepo.On("Save", ctx, userUUID, mock.AnythingOfType("string"),
						mock.MatchedBy(func(s string) bool {
							return s == tt.urlLink || s == "http://example.com" // trimmed version
						})).Return(tt.mockSaveError)
				} else if errors.Is(tt.mockSaveError, myerrors.ErrURLExist) {
					// When URL exists, Save fails and GetShortURL is called
					mockRepo.On("Save", ctx, userUUID, mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).Return(tt.mockSaveError)
					mockRepo.On("GetShortURL", ctx, tt.urlLink).Return(
						tt.mockGetShortURLResult, tt.mockGetShortURLError)
				} else {
					// For other errors, just expect Save to be called
					mockRepo.On("Save", ctx, userUUID, mock.AnythingOfType("string"),
						mock.AnythingOfType("string")).Return(tt.mockSaveError)
				}
			}

			result, err := service.SaveURL(ctx, userUUID, tt.urlLink)

			if tt.expectedErrorIs != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErrorIs),
					"expected error %v, got %v (actual error: %v)", tt.expectedErrorIs, err, err)
			} else {
				assert.NoError(t, err)
				if tt.expectedSlug != "" {
					assert.NotEmpty(t, result)
				}
			}

			if tt.shouldMockSave {
				mockRepo.AssertExpectations(t)
			}
		})
	}

}

func TestService_SaveBatch(t *testing.T) {
	tests := []struct {
		name               string
		batch              []model.URL
		mockSaveBatchError error
		expectedResult     []model.URL
		expectedError      error
		shouldMockCall     bool
	}{
		{
			name:               "empty batch",
			batch:              []model.URL{},
			mockSaveBatchError: nil,
			expectedResult:     []model.URL{},
			expectedError:      nil,
			shouldMockCall:     true,
		},
		{
			name:           "batch with empty URL",
			batch:          []model.URL{{UUID: uuid.New(), OriginalURL: ""}},
			expectedError:  myerrors.ErrEmptyURL,
			shouldMockCall: false,
		},
		{
			name:           "batch with invalid URL format",
			batch:          []model.URL{{UUID: uuid.New(), OriginalURL: "not\\a\\valid\\url"}}, // Will likely fail scheme check
			expectedError:  myerrors.ErrWrongHTTPScheme,
			shouldMockCall: false,
		},
		{
			name:           "batch with missing scheme",
			batch:          []model.URL{{UUID: uuid.New(), OriginalURL: "example.com"}},
			expectedError:  myerrors.ErrWrongHTTPScheme,
			shouldMockCall: false,
		},
		{
			name:           "batch with missing host",
			batch:          []model.URL{{UUID: uuid.New(), OriginalURL: "http://"}},
			expectedError:  myerrors.ErrMustIncludeHost,
			shouldMockCall: false,
		},
		{
			name: "valid batch successful save",
			batch: []model.URL{
				{UUID: uuid.New(), OriginalURL: "http://example.com/"},
				{UUID: uuid.New(), OriginalURL: "https://google.com/path/"},
			},
			mockSaveBatchError: nil,
			expectedError:      nil,
			shouldMockCall:     true,
		},
		{
			name: "save batch repository error",
			batch: []model.URL{
				{UUID: uuid.New(), OriginalURL: "http://example.com/"},
			},
			mockSaveBatchError: errors.New("repo error"),
			expectedError:      errors.New("failed to save batch: repo error"),
			shouldMockCall:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			logger := zerolog.Nop()
			service := &Service{repo: mockRepo, zlog: logger}

			userUUID := uuid.New()
			ctx := context.Background()

			if tt.shouldMockCall {
				mockRepo.On("SaveBatch", ctx, userUUID, mock.AnythingOfType("[]model.URL")).Return(tt.mockSaveBatchError)
			}

			result, err := service.SaveBatch(ctx, userUUID, tt.batch)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				if len(tt.batch) > 0 {
					assert.Len(t, result, len(tt.batch))
					// Verify that each result has a short URL generated
					for i := range result {
						assert.NotEmpty(t, result[i].ShortURL)
						assert.Equal(t, tt.batch[i].UUID, result[i].UUID)
						// Verify URLs are trimmed
						assert.Equal(t, "http://example.com", result[0].OriginalURL) // first URL
						if len(result) > 1 {
							assert.Equal(t, "https://google.com/path", result[1].OriginalURL) // second URL
						}
					}
				} else {
					assert.Len(t, result, 0)
				}
			}

			if tt.shouldMockCall {
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

func TestService_SendShortURLForDelete(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	service := &Service{
		repo:         &MockRepository{},
		zlog:         logger,
		toDeleteChan: []chan map[uuid.UUID][]string{},
	}

	userUUID := uuid.New()
	shortURLs := []string{"abc123", "def456"}

	originalLength := len(service.toDeleteChan)
	service.SendShortURLForDelete(context.Background(), userUUID, shortURLs)
	newLength := len(service.toDeleteChan)

	assert.Equal(t, originalLength+1, newLength)
	assert.Len(t, service.toDeleteChan, 1)
}

func TestGenerateShortURL(t *testing.T) {
	tests := []struct {
		name          string
		length        int
		expectedError error
	}{
		{
			name:          "positive length",
			length:        8,
			expectedError: nil,
		},
		{
			name:          "zero length",
			length:        0,
			expectedError: myerrors.ErrShortURLLength,
		},
		{
			name:          "negative length",
			length:        -1,
			expectedError: myerrors.ErrShortURLLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateShortURL(tt.length)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.length)

				// Verify that the generated string only contains allowed characters
				for _, char := range result {
					assert.Contains(t, string(randGenerateSymbols), string(char))
				}
			}
		})
	}
}

func TestService_generateShortURLLengthValidation(t *testing.T) {
	// Additional test for edge cases in generateShortURL
	testCases := []int{-100, -1, 0}
	for _, length := range testCases {
		_, err := generateShortURL(length)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, myerrors.ErrShortURLLength))
	}
}

func TestService_generateShortURLPositiveCases(t *testing.T) {
	testCases := []int{1, 5, 8, 10, 100}
	for _, length := range testCases {
		result, err := generateShortURL(length)
		assert.NoError(t, err)
		assert.Len(t, result, length)

		// Verify all characters are from the allowed set
		for _, char := range result {
			assert.Contains(t, string(randGenerateSymbols), string(char))
		}
	}
}

func TestService_SaveURLWithTrailingSlash(t *testing.T) {
	mockRepo := &MockRepository{}
	logger := zerolog.Nop()
	service := &Service{repo: mockRepo, zlog: logger}

	userUUID := uuid.New()
	ctx := context.Background()

	originalURL := "http://example.com/"
	expectedTrimmedURL := "http://example.com"

	mockRepo.On("Save", ctx, userUUID, mock.AnythingOfType("string"), expectedTrimmedURL).Return(nil)

	result, err := service.SaveURL(ctx, userUUID, originalURL)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	mockRepo.AssertExpectations(t)
}

func TestService_SaveBatchWithTrailingSlashes(t *testing.T) {
	mockRepo := &MockRepository{}
	logger := zerolog.Nop()
	service := &Service{repo: mockRepo, zlog: logger}

	userUUID := uuid.New()
	ctx := context.Background()

	batch := []model.URL{
		{UUID: uuid.New(), OriginalURL: "http://example.com/"},
		{UUID: uuid.New(), OriginalURL: "https://google.com/path/"},
	}

	mockRepo.On("SaveBatch", ctx, userUUID, mock.AnythingOfType("[]model.URL")).Run(func(args mock.Arguments) {
		savedBatch := args.Get(2).([]model.URL)
		assert.Len(t, savedBatch, 2)
		assert.Equal(t, "http://example.com", savedBatch[0].OriginalURL)
		assert.Equal(t, "https://google.com/path", savedBatch[1].OriginalURL)
		// Also verify that short URLs were generated
		for _, item := range savedBatch {
			assert.NotEmpty(t, item.ShortURL)
		}
	}).Return(nil)

	result, err := service.SaveBatch(ctx, userUUID, batch)

	assert.NoError(t, err)
	assert.Len(t, result, 2)
	// Verify that the result URLs are trimmed and have generated short URLs
	for i, item := range result {
		assert.Equal(t, batch[i].UUID, item.UUID)
		assert.Equal(t, strings.TrimRight(batch[i].OriginalURL, "/"), item.OriginalURL)
		assert.NotEmpty(t, item.ShortURL)
	}
	mockRepo.AssertExpectations(t)
}

// Integration-style test for the deletion functionality
func TestService_DeleteFunctionality(t *testing.T) {
	t.Parallel()

	mockRepo := &MockRepository{}
	logger := zerolog.Nop()
	service := &Service{
		repo:         mockRepo,
		zlog:         logger,
		toDeleteChan: []chan map[uuid.UUID][]string{},
	}

	userUUID := uuid.New()
	shortURLs := []string{"abc123", "def456"}

	// This test verifies that the deletion channel mechanism works
	// Note: The actual deletion happens in a background goroutine
	service.SendShortURLForDelete(context.Background(), userUUID, shortURLs)

	assert.Len(t, service.toDeleteChan, 1)

	// Verify that a channel was added to the slice
	channel := service.toDeleteChan[0]
	assert.NotNil(t, channel)
}

func Benchmark_Service_generateShortURL(b *testing.B) {
	for b.Loop() {
		generateShortURL(8)
	}
}

func Benchmark_Service_SaveURL(b *testing.B) {
	for b.Loop() {

		srv := Service{repo: &HandyMockRepository{urls: map[string]string{}}}
		ctx := context.Background()
		userUUID, _ := uuid.NewUUID()
		url := func() string {
			body, _ := generateShortURL(8)
			return "http://" + body + ".com"
		}()
		srv.SaveURL(ctx, userUUID, url)
	}
}
