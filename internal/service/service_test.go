package service

import (
	"errors"
	"github.com/ar4ie13/shortener/internal/service/internal/mockery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

// HandyMockRepository implements Repository interface for testing
type HandyMockRepository struct {
	urls map[string]string
	err  error
}

func (m *HandyMockRepository) Get(id string) (string, error) {
	url, exists := m.urls[id]
	if !exists {
		return "", ErrNotFound
	}
	return url, nil
}

func (m *HandyMockRepository) Save(id string, url string) error {
	if id == "" || url == "" {
		return ErrEmptyIDorURL
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

func TestNewService(t *testing.T) {
	type args struct {
		r Repository
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{
		{
			name: "TestNewService",
			args: args{},
			want: &Service{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewService(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewService() = %v, want %v", got, tt.want)
			}
		})
	}
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
				&r,
			}
			_, err := s.GenerateShortURL(tt.args.url)
			if ((err != nil) != tt.wantErr) || (tt.wantErr && !errors.Is(err, tt.wantErrMsg)) {
				t.Errorf("%v", !errors.Is(err, tt.wantErrMsg))
				t.Errorf("GenerateShortURL() error = %v, wantErr %v", err, tt.wantErrMsg)
				return
			}

		})
	}
}

/*
func TestService_GetURL(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name       string
		fields     HandyMockRepository
		args       args
		want       string
		wantErr    bool
		wantErrMsg error
	}{
		{
			name: "Existent id",
			fields: HandyMockRepository{
				urls: map[string]string{
					"abc": "https://google.com",
					"xyz": "https://facebook.com",
				},
				err: nil,
			},
			args: args{
				id: "abc",
			},
			want:       "https://google.com",
			wantErr:    false,
			wantErrMsg: nil,
		},
		{
			name: "Non-existent id",
			fields: HandyMockRepository{
				urls: map[string]string{
					"abc": "https://google.com",
					"xyz": "https://facebook.com",
				},
				err: nil,
			},
			args: args{
				id: "ab",
			},
			want:       "",
			wantErr:    true,
			wantErrMsg: ErrNotFound,
		}, {
			name: "Empty id",
			fields: HandyMockRepository{
				urls: map[string]string{
					"abc": "https://google.com",
					"xyz": "https://facebook.com",
				},
				err: nil,
			},
			args: args{
				id: "",
			},
			want:       "",
			wantErr:    true,
			wantErrMsg: errEmptyID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := HandyMockRepository{
				tt.fields.urls,
				tt.fields.err,
			}
			s := Service{
				&r,
			}

			got, err := s.GetURL(tt.args.id)
			if ((err != nil) != tt.wantErr) || (tt.wantErr && !errors.Is(err, tt.wantErrMsg)) {
				t.Errorf("GetURL() error = %v, wantErr %v", err, tt.wantErrMsg)
				return
			}
			if got != tt.want {
				t.Errorf("GetURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
*/

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
			service := Service{r: mockRepo}

			if tt.shouldCallRepo {
				mockRepo.On("Get", tt.shortURL).Return(tt.mockReturnURL, tt.mockReturnErr)
			}

			result, err := service.GetURL(tt.shortURL)

			assert.Equal(t, tt.expectedURL, result)
			if tt.expectedErr != nil {
				require.Error(t, err)
				if errors.Is(tt.expectedErr, ErrNotFound) || errors.Is(tt.expectedErr, ErrEmptyIDorURL) {
					assert.Contains(t, err.Error(), "failed to get URL")
				}
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			if tt.shouldCallRepo {
				mockRepo.AssertCalled(t, "Get", tt.shortURL)
			} else {
				mockRepo.AssertNotCalled(t, "Get", mock.Anything)
			}
		})
	}
}
