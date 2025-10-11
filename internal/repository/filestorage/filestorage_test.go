package filestorage

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

// MockFile is a mock for os.File
type MockFile struct {
	mock.Mock
}

func (m *MockFile) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockFile) WriteString(s string) (n int, err error) {
	args := m.Called(s)
	return args.Int(0), args.Error(1)
}

func (m *MockFile) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockOS is a mock for os functions
type MockOS struct {
	mock.Mock
}

func (m *MockOS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	args := m.Called(name, flag, perm)
	return args.Get(0).(*os.File), args.Error(1)
}

func TestNewFileStorage(t *testing.T) {
	type args struct {
		filepath string
		zlog     zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want *FileStorage
	}{
		{
			name: "TestNewFileStorage",
			args: args{
				filepath: "./storage.jsonl",
				zlog: zerolog.New(zerolog.ConsoleWriter{
					Out:        os.Stdout,
					TimeFormat: time.RFC3339,
				}).With().Timestamp().Logger().Level(zerolog.DebugLevel),
			},
			want: &FileStorage{
				urlMapping: urlMapping{},
				lastUUID:   lastUUID{},
				filePath:   "./storage.jsonl",
				zlog: zerolog.New(zerolog.ConsoleWriter{
					Out:        os.Stdout,
					TimeFormat: time.RFC3339,
				}).With().Timestamp().Logger().Level(zerolog.DebugLevel),
				mu: sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFileStorage(tt.args.filepath, tt.args.zlog); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFileStorage() = \n%v, \nwant %v", got, tt.want)
			}
		})
	}
}

// TestFileStorage_Store tests the Store function
func TestFileStorage_Store(t *testing.T) {
	tests := []struct {
		name          string
		shortURL      string
		url           string
		setupMocks    func(*MockOS, *MockFile)
		expectedError bool
	}{
		{
			name:     "successful store",
			shortURL: "abc123",
			url:      "https://example.com",
			setupMocks: func(mockOS *MockOS, mockFile *MockFile) {
				// Mock OpenFile to return our mock file
				mockOS.On("OpenFile", "test.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(0666)).
					Return(mockFile, nil)

				// Mock file operations
				expectedJSON := `{"uuid":2,"short_url":"abc123","url":"https://example.com"}`
				mockFile.On("Write", []byte(expectedJSON+"\n")).Return(len(expectedJSON+"\n"), nil)
				mockFile.On("WriteString", "\n").Return(1, nil)
				mockFile.On("Close").Return(nil)
			},
			expectedError: false,
		},
		{
			name:     "open file fails",
			shortURL: "abc123",
			url:      "https://example.com",
			setupMocks: func(mockOS *MockOS, mockFile *MockFile) {
				mockOS.On("OpenFile", "test.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(0666)).
					Return((*os.File)(nil), errors.New("file open error"))
			},
			expectedError: true,
		},
		{
			name:     "file write fails",
			shortURL: "abc123",
			url:      "https://example.com",
			setupMocks: func(mockOS *MockOS, mockFile *MockFile) {
				mockOS.On("OpenFile", "test.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(0666)).
					Return(mockFile, nil)

				expectedJSON := `{"uuid":2,"short_url":"abc123","url":"https://example.com"}`
				mockFile.On("Write", []byte(expectedJSON+"\n")).Return(0, errors.New("write error"))
				mockFile.On("Close").Return(nil)
			},
			expectedError: true,
		},
		{
			name:     "write newline fails",
			shortURL: "abc123",
			url:      "https://example.com",
			setupMocks: func(mockOS *MockOS, mockFile *MockFile) {
				mockOS.On("OpenFile", "test.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.FileMode(0666)).
					Return(mockFile, nil)

				expectedJSON := `{"uuid":2,"short_url":"abc123","url":"https://example.com"}`
				mockFile.On("Write", []byte(expectedJSON+"\n")).Return(len(expectedJSON+"\n"), nil)
				mockFile.On("WriteString", "\n").Return(0, errors.New("newline write error"))
				mockFile.On("Close").Return(nil)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockOS := new(MockOS)
			mockFile := new(MockFile)

			// Setup mocks
			tt.setupMocks(mockOS, mockFile)

			// Create FileStorage with test data
			storage := &FileStorage{
				filePath: "test.json",
				urlMapping: urlMapping{
					UUID:     0,
					ShortURL: "",
					URL:      "",
				},
				lastUUID: lastUUID{UUID: 1},
				mu:       sync.Mutex{},
			}

			// Replace os.OpenFile with our mock
			osOpenFile := mockOS.OpenFile
			originalOpenFile := osOpenFile

			defer func() { osOpenFile = originalOpenFile }()

			// Execute test
			if tt.expectedError {

				// Should not panic
				assert.NotPanics(t, func() {
					err := storage.Store(tt.shortURL, tt.url)
					assert.NoError(t, err)
				})

				// Verify UUID was incremented
				assert.Equal(t, 2, storage.lastUUID.UUID)
				assert.Equal(t, tt.shortURL, storage.urlMapping.ShortURL)
				assert.Equal(t, tt.url, storage.urlMapping.URL)
			}

		})
	}
}
