package memory

import (
	"errors"
	"fmt"
	"github.com/ar4ie13/shortener/internal/service"
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name              string
		expectedMapLength int
	}{
		{
			name:              "NewMemStorage",
			expectedMapLength: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMemStorage(); len(got.urlMemStore) != tt.expectedMapLength || len(got.slugMemStore) != tt.expectedMapLength {
				t.Errorf("NewMemStorage() urlLib lenth = %v, want %v", got, tt.expectedMapLength)
			}
			if got := NewMemStorage(); got.slugMemStore == nil || got.urlMemStore == nil {
				t.Errorf("NewMemStorage() slugMemStore or urlMemStore is nil")
			}
			if got := NewMemStorage(); got == nil {
				t.Errorf("NewMemStorage() struct is nil")
			}
		})
	}
}

func TestMemory_Get(t *testing.T) {
	type fields struct {
		slugMemStore map[string]string
		urlMemStore  map[string]string
	}
	type args struct {
		slug string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      string
		wantErr   bool
		wantError error
	}{
		{
			name: "Valid slug",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				slug: "abc123",
			},
			want:      "https://example.com",
			wantErr:   false,
			wantError: nil,
		},
		{
			name: "Non-existent slug",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc12": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc12",
				},
			},
			args: args{
				slug: "abc123",
			},
			want:      "",
			wantErr:   true,
			wantError: service.ErrNotFound,
		},
		{
			name: "Empty input parameter",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc12": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc12",
				},
			},
			args: args{
				slug: "",
			},
			want:      "",
			wantErr:   true,
			wantError: service.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MemStorage{
				slugMemStore: tt.fields.slugMemStore,
				urlMemStore:  tt.fields.urlMemStore,
			}
			got, err := repo.Get(tt.args.slug)
			if got != tt.want {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestMemory_Save(t *testing.T) {

	type fields struct {
		slugMemStore map[string]string
		urlMemStore  map[string]string
	}
	type args struct {
		slug string
		url  string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantErrName error
	}{
		{
			name: "Valid slug and URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				slug: "abc12",
				url:  "https://examplenew.com",
			},
			wantErr:     false,
			wantErrName: nil,
		},
		{
			name: "Valid slug and existent URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				slug: "abc12",
				url:  "https://example.com",
			},
			wantErr:     true,
			wantErrName: service.ErrURLExist,
		},
		{
			name: "Empty slug and existent URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				slug: "",
				url:  "https://example.com",
			},
			wantErr:     true,
			wantErrName: service.ErrEmptyIDorURL,
		},
		{
			name: "Valid slug and empty URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				slug: "abc",
				url:  "",
			},
			wantErr:     true,
			wantErrName: service.ErrEmptyIDorURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MemStorage{
				slugMemStore: tt.fields.slugMemStore,
				urlMemStore:  tt.fields.urlMemStore,
			}

			if err := repo.Save(tt.args.slug, tt.args.url); (err != nil) != tt.wantErr || !errors.Is(err, tt.wantErrName) {
				fmt.Println(err, tt.wantErrName)
				t.Errorf("Save() error = %s, wantErr %s", err, tt.wantErrName)
			}
		})
	}
}

func TestMemory_existsURL(t *testing.T) {
	type fields struct {
		slugMemStore map[string]string
		urlMemStore  map[string]string
	}
	type args struct {
		url string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Exists URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				url: "https://example.com",
			},
			want: true,
		},
		{
			name: "Not existsURL URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				url: "https://examplenew.com",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MemStorage{
				slugMemStore: tt.fields.slugMemStore,
				urlMemStore:  tt.fields.urlMemStore,
			}
			if got := repo.existsURL(tt.args.url); got != tt.want {
				t.Errorf("existsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemory_existsShortURL(t *testing.T) {
	type fields struct {
		slugMemStore map[string]string
		urlMemStore  map[string]string
	}
	type args struct {
		slug string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Exists URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				slug: "abc123",
			},
			want: true,
		},
		{
			name: "Not existsURL URL",
			fields: fields{
				slugMemStore: slugMemStore{
					"abc123": "https://example.com",
				},
				urlMemStore: urlMemStore{
					"https://example.com": "abc123",
				},
			},
			args: args{
				slug: "abc",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MemStorage{
				slugMemStore: tt.fields.slugMemStore,
				urlMemStore:  tt.fields.urlMemStore,
			}
			if got := repo.existsShortURL(tt.args.slug); got != tt.want {
				t.Errorf("existsShortURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
