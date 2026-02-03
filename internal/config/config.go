// Package config reads flags and environment variables to initialize service config
package config

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	auditconf "github.com/ar4ie13/shortener/internal/auditor/config"
	authconf "github.com/ar4ie13/shortener/internal/auth/config"
	pgconf "github.com/ar4ie13/shortener/internal/repository/db/postgresql/config"
	fileconf "github.com/ar4ie13/shortener/internal/repository/filestorage/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/jackc/pgx/v5/stdlib"
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
	FilePath         fileconf.Config
	PostgresDSN      pgconf.Config
	AuthConf         authconf.Config
	AuditConf        auditconf.AuditConf
	HTTPS            bool
	TLSCertPath      string
	TLSKeyPath       string
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
	var err error
	defaultServerAddr := "localhost:8080"
	defaultURL := "http://localhost:8080"
	defaultLogLevel := LogLevel{Level: zerolog.InfoLevel}
	defaultFileStorage := ""
	defaultDatabaseDSN := ""
	defaultSecretKey := "nHhjHgahbioHBGbBHJ"
	defaultTokenExpiration := time.Hour * 24
	defaultFileAuditPath := ""
	defaultRemoteAuditHost := ""
	defaultTLSCert := "cert.pem"
	defaultTLSKey := "key.pem"

	flag.StringVar(&c.LocalServerAddr, "a", defaultServerAddr, "local server address")
	flag.Var(&c.ShortURLTemplate, "b", "short url template")
	flag.Var(&c.LogLevel, "l", "log level (debug, info, warn, error, fatal, panic)")
	flag.StringVar(&c.FilePath.FilePath, "f", defaultFileStorage, "file storage path")
	flag.StringVar(&c.PostgresDSN.DatabaseDSN, "d", defaultDatabaseDSN, "database DSN")
	flag.StringVar(&c.AuthConf.SecretKey, "k", defaultSecretKey, "secret key")
	flag.DurationVar(&c.AuthConf.TokenExpiration, "e", defaultTokenExpiration, "token expiration")
	flag.StringVar(&c.AuditConf.FileConf.AuditFilePath, "audit-file", defaultFileAuditPath, "audit file path")
	flag.StringVar(&c.AuditConf.RemoteConf.RemoteServerURL, "audit-url", defaultRemoteAuditHost, "audit host url")
	flag.BoolVar(&c.AuditConf.Enabled, "audit-enabled", true, "enable/disable audit")
	flag.BoolVar(&c.HTTPS, "s", false, "enable https")
	flag.StringVar(&c.TLSCertPath, "tls-cert", defaultTLSCert, "TLS certificate")
	flag.StringVar(&c.TLSKeyPath, "tls-key", defaultTLSKey, "TLS key")

	if err = c.ShortURLTemplate.Set(defaultURL); err != nil {
		log.Fatal().Err(err).Msg("Failed to set default URL")
	}

	if err = c.LogLevel.Set(defaultLogLevel.String()); err != nil {
		log.Fatal().Err(err).Msg("Failed to set default log level")
	}

	flag.Parse()

	if serverAddr := os.Getenv("SERVER_ADDRESS"); serverAddr != "" {
		if _, err = strconv.Unquote("\"" + serverAddr + "\""); err != nil {
			parts := strings.SplitN(serverAddr, ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				log.Fatal().Err(err).Msg("Failed to set server address from SERVER_ADDRESS")
			}
		}
		c.LocalServerAddr = serverAddr
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		err = c.ShortURLTemplate.Set(baseURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set URL template from BASE_URL")
		}
	}

	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		err = c.LogLevel.Set(logLevelStr)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to set log level from LOG_LEVEL")
		}
	}

	if fileStorage := os.Getenv("FILE_STORAGE_PATH"); fileStorage != "" {
		c.FilePath.FilePath = fileStorage
	}

	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		c.PostgresDSN.DatabaseDSN = databaseDSN
	}

	if secretKey := os.Getenv("SECRET_KEY"); secretKey != "" {
		c.AuthConf.SecretKey = secretKey
	}

	if tokenExpirationStr := os.Getenv("TOKEN_EXPIRATION"); tokenExpirationStr != "" {
		c.AuthConf.TokenExpiration, err = time.ParseDuration(tokenExpirationStr)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot parse token expiration environment variable")
		}

	}

	if fileAuditPath := os.Getenv("AUDIT_FILE"); fileAuditPath != "" {
		c.AuditConf.FileConf.AuditFilePath = fileAuditPath
	}

	if hostAuditURL := os.Getenv("AUDIT_URL"); hostAuditURL != "" {
		c.AuditConf.RemoteConf.RemoteServerURL = hostAuditURL
	}

	if auditEnabled := os.Getenv("AUDIT_ENABLED"); auditEnabled != "" {
		c.AuditConf.Enabled, err = strconv.ParseBool(auditEnabled)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot parse audit enabled environment variable")
		}
	}

	if httpsEnabled := os.Getenv("ENABLE_HTTPS"); httpsEnabled != "" {
		c.HTTPS, err = strconv.ParseBool(httpsEnabled)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot parse https enabled environment variable")
		}
	}

	if tlsCertPath := os.Getenv("TLS_CERT_PATH"); tlsCertPath != "" {
		c.TLSCertPath = tlsCertPath
	}
	if tlsKeyPath := os.Getenv("TLS_KEY_PATH"); tlsKeyPath != "" {
		c.TLSKeyPath = tlsKeyPath
	}
}

// CheckPostgresConnection validates the connection to PostgreSQL database
func (c *Config) CheckPostgresConnection(ctx context.Context) error {
	db, err := sql.Open("pgx", c.PostgresDSN.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	ctxPg, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctxPg); err != nil {
		return err
	}
	return nil
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

func (c *Config) GetHTTPS() bool {
	return c.HTTPS
}

func (c *Config) GetTLSCertPath() string {
	return c.TLSCertPath
}

func (c *Config) GetTLSKeyPath() string {
	return c.TLSKeyPath
}
