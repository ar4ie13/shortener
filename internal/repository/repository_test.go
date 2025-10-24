package repository

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	pgconf "github.com/ar4ie13/shortener/internal/repository/db/postgresql/config"
	fileconf "github.com/ar4ie13/shortener/internal/repository/filestorage/config"
	"github.com/ar4ie13/shortener/internal/repository/memory"
	"github.com/ar4ie13/shortener/internal/service"
	"github.com/rs/zerolog"
)

func TestNewRepository(t *testing.T) {
	type args struct {
		ctx      context.Context
		fileconf fileconf.Config
		pgcfg    pgconf.Config
		zlog     zerolog.Logger
	}
	tests := []struct {
		name    string
		args    args
		want    service.Repository
		wantErr bool
	}{
		{
			name: "success memory",
			args: args{
				ctx: context.Background(),
				fileconf: fileconf.Config{
					FilePath: "",
				},
				pgcfg: pgconf.Config{
					DatabaseDSN: "",
				},
				zlog: zerolog.New(zerolog.ConsoleWriter{
					Out:        os.Stdout,
					TimeFormat: time.RFC3339,
				}).With().Timestamp().Logger().Level(zerolog.DebugLevel),
			},
			want:    memory.NewMemStorage(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRepository(context.Background(), tt.args.fileconf, tt.args.pgcfg, tt.args.zlog)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRepository() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRepository() got = %v, want %v", got, tt.want)
			}
			if got == nil {
				t.Errorf("NewRepository() memory is nil")
			}
			if got == nil {
				t.Errorf("NewRepository() Repository struct is nil")
			}
		})
	}

}
