package logger

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

// Logger wraps a zerolog.Logger for type safety and extensibility
type Logger struct {
	zerolog.Logger
}

// NewLogger creates a new Logger with the given zerolog level
func NewLogger(level zerolog.Level) *Logger {
	return &Logger{
		Logger: zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger().Level(level),
	}
}
