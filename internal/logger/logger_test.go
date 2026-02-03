package logger

import (
	"bytes"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("CreateDebugLogger", func(t *testing.T) {
		logger := NewLogger(zerolog.DebugLevel)

		require.NotNil(t, logger)
		assert.Equal(t, zerolog.DebugLevel, logger.GetLevel())
	})

	t.Run("CreateInfoLogger", func(t *testing.T) {
		logger := NewLogger(zerolog.InfoLevel)

		require.NotNil(t, logger)
		assert.Equal(t, zerolog.InfoLevel, logger.GetLevel())
	})

	t.Run("CreateWarnLogger", func(t *testing.T) {
		logger := NewLogger(zerolog.WarnLevel)

		require.NotNil(t, logger)
		assert.Equal(t, zerolog.WarnLevel, logger.GetLevel())
	})

	t.Run("CreateErrorLogger", func(t *testing.T) {
		logger := NewLogger(zerolog.ErrorLevel)

		require.NotNil(t, logger)
		assert.Equal(t, zerolog.ErrorLevel, logger.GetLevel())
	})

	t.Run("CreateDisabledLogger", func(t *testing.T) {
		logger := NewLogger(zerolog.Disabled)

		require.NotNil(t, logger)
		assert.Equal(t, zerolog.Disabled, logger.GetLevel())
	})
}

func TestLogger_Logging(t *testing.T) {
	t.Run("DebugLogsAtDebugLevel", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: zerolog.New(&buf).Level(zerolog.DebugLevel),
		}

		logger.Debug().Msg("debug message")

		assert.Contains(t, buf.String(), "debug message")
		assert.Contains(t, buf.String(), "debug")
	})

	t.Run("InfoLogsAtInfoLevel", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: zerolog.New(&buf).Level(zerolog.InfoLevel),
		}

		logger.Info().Msg("info message")

		assert.Contains(t, buf.String(), "info message")
		assert.Contains(t, buf.String(), "info")
	})

	t.Run("WarnLogsAtWarnLevel", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: zerolog.New(&buf).Level(zerolog.WarnLevel),
		}

		logger.Warn().Msg("warn message")

		assert.Contains(t, buf.String(), "warn message")
		assert.Contains(t, buf.String(), "warn")
	})

	t.Run("ErrorLogsAtErrorLevel", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: zerolog.New(&buf).Level(zerolog.ErrorLevel),
		}

		logger.Error().Msg("error message")

		assert.Contains(t, buf.String(), "error message")
		assert.Contains(t, buf.String(), "error")
	})

	t.Run("DebugNotLoggedAtInfoLevel", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: zerolog.New(&buf).Level(zerolog.InfoLevel),
		}

		logger.Debug().Msg("debug message")

		assert.Empty(t, buf.String())
	})

	t.Run("InfoNotLoggedAtWarnLevel", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: zerolog.New(&buf).Level(zerolog.WarnLevel),
		}

		logger.Info().Msg("info message")

		assert.Empty(t, buf.String())
	})

	t.Run("WarnNotLoggedAtErrorLevel", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &Logger{
			Logger: zerolog.New(&buf).Level(zerolog.ErrorLevel),
		}

		logger.Warn().Msg("warn message")

		assert.Empty(t, buf.String())
	})
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		Logger: zerolog.New(&buf).Level(zerolog.InfoLevel),
	}

	logger.Info().
		Str("key", "value").
		Int("count", 42).
		Msg("message with fields")

	output := buf.String()
	assert.Contains(t, output, "message with fields")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
	assert.Contains(t, output, "count")
	assert.Contains(t, output, "42")
}

func TestLogger_Msgf(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{
		Logger: zerolog.New(&buf).Level(zerolog.InfoLevel),
	}

	logger.Info().Msgf("formatted %s with %d", "message", 123)

	output := buf.String()
	assert.Contains(t, output, "formatted message with 123")
}
