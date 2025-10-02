package config

import (
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	errEmptyURL        = errors.New("URL template cannot be empty")
	errWrongHTTPScheme = errors.New("URL template must use http or https scheme")
	errMustIncludeHost = errors.New("URL template must include a host")
)

// ShortURLTemplate type for short URL template flag
type ShortURLTemplate string

// Config struct used for program flag variables
type Config struct {
	LocalServerAddr  string
	ShortURLTemplate ShortURLTemplate
	LogLevel         LogLevel
}

// NewConfig constructor for Config
func NewConfig() *Config {
	c := &Config{}
	c.InitConfig()
	return c
}

// String return short URL in string format
func (u *ShortURLTemplate) String() string {
	return string(*u)
}

// Set validates and sets the flag value
func (u *ShortURLTemplate) Set(value string) error {
	// Check if the value is empty
	if value == "" {
		return errEmptyURL
	}

	// Validate the URL format
	parsedURL, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// Ensure the scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errWrongHTTPScheme
	}

	// Ensure the host is not empty
	if parsedURL.Host == "" {
		return errMustIncludeHost
	}

	*u = ShortURLTemplate(value)
	return nil
}

// LogLevel type for custom log level flag
type LogLevel struct {
	Level zerolog.Level
}

// String returns log level as string
func (l *LogLevel) String() string {
	return l.Level.String()
}

// Set validates and sets the log level from string
func (l *LogLevel) Set(value string) error {
	level, err := zerolog.ParseLevel(strings.ToLower(value))
	if err != nil {
		return fmt.Errorf("invalid log level: %v", err)
	}
	l.Level = level
	return nil
}

// InitConfig initialize configuration
func (c *Config) InitConfig() {

	defaultServerAddr := "localhost:8080"
	defaultURL := "http://localhost:8080"
	defaultLogLevel := LogLevel{Level: zerolog.InfoLevel}

	flag.StringVar(&c.LocalServerAddr, "a", defaultServerAddr, "local server address")
	flag.Var(&c.ShortURLTemplate, "b", "short url template")
	flag.Var(&c.LogLevel, "l", "log level (debug, info, warn, error, fatal, panic)")

	if err := c.ShortURLTemplate.Set(defaultURL); err != nil {
		log.Fatal().Err(err).Msg("Failed to set default URL")
	}

	if err := c.LogLevel.Set(defaultLogLevel.String()); err != nil {
		log.Fatal().Err(err).Msg("Failed to set default log level")
	}

	flag.Parse()

	if serverAddr := os.Getenv("SERVER_ADDRESS"); serverAddr != "" {
		if _, err := strconv.Unquote("\"" + serverAddr + "\""); err != nil {
			parts := strings.SplitN(serverAddr, ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				log.Fatal().Err(err).Msg("Failed to set short URL template from BASE_URL")
			}
		}
		c.LocalServerAddr = serverAddr
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		err := c.ShortURLTemplate.Set(baseURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set log level from LOG_LEVEL")
		}
	}

	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		err := c.LogLevel.Set(logLevelStr)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set log level from LOG_LEVEL")
		}
	}
}

// GetLocalServerAddr returns localserver address string
func (c *Config) GetLocalServerAddr() string {
	return c.LocalServerAddr
}

// GetShortURLTemplate returns Short URL template string
func (c *Config) GetShortURLTemplate() string {
	return string(c.ShortURLTemplate)
}

// GetLogLevel returns logging level. Used in logger.NewLogger constructor.
func (c *Config) GetLogLevel() zerolog.Level {
	return c.LogLevel.Level
}
