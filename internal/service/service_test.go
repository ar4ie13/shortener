package service

import (
	"context"
	"errors"
	"testing"

	"github.com/ar4ie13/shortener/internal/model"
	"github.com/ar4ie13/shortener/internal/service/internal/mockery"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// HandyMockRepository implements Repository interface for testing
type HandyMockRepository struct {
	urls map[string]string
	err  error
}

func (m *HandyMockRepository) GetURL(_ context.Context, id string) (string, error) {
	url, exists := m.urls[id]
	if !exists {
		return "", ErrNotFound
	}
	return url, nil
}

func (m *HandyMockRepository) GetShortURL(_ context.Context, urllink string) (string, error) {
	for k, v := range m.urls {
		if urllink == v {
			return k, nil
		}
	}
	return "", ErrNotFound
}

func (m *HandyMockRepository) Save(_ context.Context, userUUID uuid.UUID, id string, url string) error {
	if id == "" || url == "" {
		return ErrEmptyShortURLorURL
	}
	if m.err != nil {
		return m.err
	}
	for _, v := range m.urls {
		if v == url {
			return ErrURLExist
		}
	}
	m.urls[id] = url
	return nil
}
func (m *HandyMockRepository) SaveBatch(_ context.Context, userUUID uuid.UUID, batch []model.URL) error {
	for i := range batch {
		if batch[i].ShortURL == "" || batch[i].OriginalURL == "" {
			return ErrEmptyShortURLorURL
		}
		if m.err != nil {
			return m.err
		}

	}
	return nil
}

func (m *HandyMockRepository) GetUserShortURLs(ctx context.Context, userUUID uuid.UUID) (map[string]string, error) {
	return m.urls, m.err
}
func (m *HandyMockRepository) DeleteUserShortURLs(ctx context.Context, shortURLsToDelete map[uuid.UUID][]string) error {
	return m.err
}

func TestService_GenerateShortURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name       string
		fields     HandyMockRepository
		args       args
		wantErr    bool
		wantErrMsg error
	}{
		{
			name: "Non existent URL",
			fields: HandyMockRepository{
				urls: map[string]string{
					"abc": "https://google.com",
					"xyz": "https://facebook.com",
				},
				err: nil,
			},
			args: args{
				url: "http://abc.com",
			},
			wantErr:    false,
			wantErrMsg: nil,
		},
		{
			name: "Existent URL",
			fields: HandyMockRepository{
				urls: map[string]string{
					"abc": "https://google.com",
					"xyz": "https://facebook.com",
				},
				err: nil,
			},

			args: args{
				url: "https://google.com",
			},

			wantErr:    true,
			wantErrMsg: ErrURLExist,
		},
		{
			name: "Empty test",
			fields: HandyMockRepository{
				urls: map[string]string{
					"abc": "https://google.com",
					"xyz": "https://facebook.com",
				},
				err: nil,
			},

			args: args{
				url: "",
			},

			wantErr:    true,
			wantErrMsg: ErrEmptyURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := HandyMockRepository{
				tt.fields.urls,
				tt.fields.err,
			}
			s := Service{
				repo:         &r,
				toDeleteChan: []chan map[uuid.UUID][]string{},
				zlog:         zerolog.Logger{},
			}
			_, err := s.SaveURL(context.Background(), uuid.New(), tt.args.url)
			if ((err != nil) != tt.wantErr) || (tt.wantErr && !errors.Is(err, tt.wantErrMsg)) {
				t.Errorf("%v", !errors.Is(err, tt.wantErrMsg))
				t.Errorf("SaveURL() error = %v, wantErr %v", err, tt.wantErrMsg)
				return
			}

		})
	}
}

func Test_generateShortURL(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "Length 8",
			args: args{
				length: 8,
			},
			want:    8,
			wantErr: false,
		},
		{
			name: "Length 34",
			args: args{
				length: 34,
			},
			want:    34,
			wantErr: false,
		},
		{
			name: "Length 0",
			args: args{
				length: 0,
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateShortURL(tt.args.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateShortURL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.want {
				t.Errorf("generateShortURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestService_GetURL_Mockery by using mockery
func TestService_GetURL_Mockery(t *testing.T) {
	tests := []struct {
		name           string
		shortURL       string
		mockReturnURL  string
		mockReturnErr  error
		expectedURL    string
		expectedErr    error
		shouldCallRepo bool
	}{
		{
			name:           "success",
			shortURL:       "abc123",
			mockReturnURL:  "https://example.com",
			mockReturnErr:  nil,
			expectedURL:    "https://example.com",
			expectedErr:    nil,
			shouldCallRepo: true,
		},
		{
			name:           "empty short URL",
			shortURL:       "",
			mockReturnURL:  "",
			mockReturnErr:  nil,
			expectedURL:    "",
			expectedErr:    errEmptyID,
			shouldCallRepo: false,
		},
		{
			name:           "not found",
			shortURL:       "abc",
			mockReturnURL:  "",
			mockReturnErr:  ErrNotFound,
			expectedURL:    "",
			expectedErr:    ErrNotFound, // The function wraps this error
			shouldCallRepo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mockery.NewMockRepository(t)
			service := Service{repo: mockRepo}

			if tt.shouldCallRepo {
				mockRepo.On("GetURL", context.Background(), tt.shortURL).Return(tt.mockReturnURL, tt.mockReturnErr)
			}

			result, err := service.GetURL(context.Background(), uuid.New(), tt.shortURL)

			assert.Equal(t, tt.expectedURL, result)
			if tt.expectedErr != nil {
				require.Error(t, err)
				if errors.Is(tt.expectedErr, ErrNotFound) || errors.Is(tt.expectedErr, ErrEmptyShortURLorURL) {
					assert.Contains(t, err.Error(), "failed to get URL")
				}
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertCalled(t, "GetURL", context.Background(), tt.shortURL)
			} else {
				mockRepo.AssertNotCalled(t, "GetURL", mock.Anything)
			}
		})
	}
}
