package postgresql

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

type queryTracer struct {
	zlog *zerolog.Logger
}

func (t *queryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	t.zlog.Debug().Msgf("Running query %s (%v)", data.SQL, data.Args)
	return ctx
}

func (t *queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	t.zlog.Debug().Msgf("%v", data.CommandTag)
}
