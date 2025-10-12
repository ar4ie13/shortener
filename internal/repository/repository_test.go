package repository

import (
	"github.com/ar4ie13/shortener/internal/repository/filestorage"
	"github.com/ar4ie13/shortener/internal/repository/memory"
	"github.com/rs/zerolog"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestNewRepository(t *testing.T) {
	type args struct {
		filepath string
		zlog     zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    *Repository
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				filepath: "./storage.jsonl",
				zlog: zerolog.New(zerolog.ConsoleWriter{
					Out:        os.Stdout,
					TimeFormat: time.RFC3339,
				}).With().Timestamp().Logger().Level(zerolog.DebugLevel),
			},
			want: &Repository{
				m: memory.NewMemStorage(),
				f: filestorage.NewFileStorage("./storage.jsonl", zerolog.New(zerolog.ConsoleWriter{
					Out:        os.Stdout,
					TimeFormat: time.RFC3339,
				}).With().Timestamp().Logger().Level(zerolog.DebugLevel)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRepository(tt.args.filepath, tt.args.zlog)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRepository() got = %v, want %v", got, tt.want)
			}
			if got.f == nil || got.m == nil {
				t.Errorf("NewRepository() filestorage or memory is nil")
			}
			if got == nil {
				t.Errorf("NewRepository() Repository struct is nil")
			}
		})
	}

}

func TestRepository_Load(t *testing.T) {
	type fields struct {
		m *memory.MemStorage
		f *filestorage.FileStorage
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				m: memory.NewMemStorage(),
				f: filestorage.NewFileStorage("./storage.jsonl", zerolog.New(zerolog.ConsoleWriter{
					Out:        os.Stdout,
					TimeFormat: time.RFC3339,
				}).With().Timestamp().Logger().Level(zerolog.DebugLevel)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &Repository{
				m: tt.fields.m,
				f: tt.fields.f,
			}
			if err := repo.Load(); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
