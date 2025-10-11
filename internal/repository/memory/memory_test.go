package memory

/*
import (
	"errors"
	"fmt"
	"github.com/ar4ie13/shortener/internal/repository/filestorage"
	"testing"
)

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name            string
		wantSliceLength int
		filepath        string
	}{
		{
			name:            "NewMemStorage",
			wantSliceLength: 0,
			filepath:        "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := NewMemStorage(tt.filepath); len(got.urlLib) != tt.wantSliceLength {
				t.Errorf("NewMemStorage() urlLib lenth = %v, want %v", got, tt.wantSliceLength)
			}
			if got, _ := NewMemStorage(tt.filepath); got.urlLib == nil {
				t.Errorf("NewMemStorage() urlLib is nil")
			}
			if got, _ := NewMemStorage(tt.filepath); got == nil {
				t.Errorf("NewMemStorage() struct is nil")
			}

		})
	}
}

func TestRepository_Get(t *testing.T) {
	type fields struct {
		urlLib urlLib
	}
	type args struct {
		id string
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
			name: "Valid ID",
			fields: fields{
				urlLib: []filestorage.URLMapping{
					{
						1,
						"abc123",
						"https://example.com",
					},
				},
			},
			args: args{
				id: "abc123",
			},
			want:      "https://example.com",
			wantErr:   false,
			wantError: nil,
		},
		{
			name: "Non-existent ID",
			fields: fields{
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc12",
						URL:      "https://example.com",
					},
				},
			},
			args: args{
				id: "abc123",
			},
			want:      "",
			wantErr:   true,
			wantError: ErrNotFound,
		},
		{
			name: "Empty input parameter",
			fields: fields{
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc12",
						URL:      "https://example.com",
					},
				},
			},
			args: args{
				id: "",
			},
			want:      "",
			wantErr:   true,
			wantError: ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MemStorage{
				urlLib: tt.fields.urlLib,
			}
			got, err := repo.Get(tt.args.id)
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

func TestRepository_Save(t *testing.T) {

	type fields struct {
		urlLib urlLib
	}
	type args struct {
		id  string
		url string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantErrName error
	}{
		{
			name: "Valid ID and URL without storage file",
			fields: fields{
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc123",
						URL:      "https://example.com",
					},
				},
			},
			args: args{
				id:  "abc12",
				url: "https://examplenew.com",
			},
			wantErr:     true,
			wantErrName: ErrFileStorage,
		},
		{
			name: "Valid ID and existent URL",
			fields: fields{
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc123",
						URL:      "https://example.com",
					},
				},
			},
			args: args{
				id:  "abc12",
				url: "https://example.com",
			},
			wantErr:     true,
			wantErrName: ErrURLExist,
		},
		{
			name: "Empty ID and existent URL",
			fields: fields{
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc123",
						URL:      "https://example.com",
					},
				},
			},
			args: args{
				id:  "",
				url: "https://example.com",
			},
			wantErr:     true,
			wantErrName: ErrEmptyIDorURL,
		},
		{
			name: "Valid ID and empty URL",
			fields: fields{
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc123",
						URL:      "https://example.com",
					},
				},
			},
			args: args{
				id:  "abc",
				url: "",
			},
			wantErr:     true,
			wantErrName: ErrEmptyIDorURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &MemStorage{
				urlLib: tt.fields.urlLib,
			}

			if err := repo.Save(tt.args.id, tt.args.url); (err != nil) != tt.wantErr || !errors.Is(err, tt.wantErrName) {
				fmt.Println(err, tt.wantErrName)
				t.Errorf("Save() error = %s, wantErr %s", err, tt.wantErrName)
			}
		})
	}
}

func TestRepository_exists(t *testing.T) {
	type fields struct {
		urlLib urlLib
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
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc123",
						URL:      "https://example.com",
					},
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
				urlLib: []filestorage.URLMapping{
					{
						UUID:     1,
						ShortURL: "abc123",
						URL:      "https://example.com",
					},
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
				urlLib: tt.fields.urlLib,
			}
			if got := repo.existsURL(tt.args.url); got != tt.want {
				t.Errorf("existsURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

*/
