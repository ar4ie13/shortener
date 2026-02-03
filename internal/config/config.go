// Package config reads flags and environment variables to initialize service config
package config

import (
	"context"
	"database/sql"
	"encoding/json"
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
	LocalServerAddr  string              `json:"local_server_addr,omitempty"`
	ShortURLTemplate ShortURLTemplate    `json:"short_url_template,omitempty"`
	LogLevel         LogLevel            `json:"log_level,omitempty"`
	FilePath         fileconf.Config     `json:"file_config,omitempty"`
	PostgresDSN      pgconf.Config       `json:"postgres_config,omitempty"`
	AuthConf         authconf.Config     `json:"auth_config,omitempty"`
	AuditConf        auditconf.AuditConf `json:"audit_config,omitempty"`
	HTTPS            bool                `json:"https_enabled,omitempty"`
	TLSCertPath      string              `json:"tls_cert_path,omitempty"`
	TLSKeyPath       string              `json:"tls_key_path,omitempty"`
	ConfigPath       string              `json:"config_path,omitempty"`
}

type ConfigTest struct {
	LocalServerAddr  string              `json:"local_server_addr,omitempty"`
	ShortURLTemplate ShortURLTemplate    `json:"short_url_template,omitempty"`
	LogLevel         LogLevel            `json:"log_level,omitempty"`
	FilePath         fileconf.Config     `json:"file_config,omitempty"`
	PostgresDSN      pgconf.Config       `json:"postgres_config,omitempty"`
	AuthConf         authconf.Config     `json:"auth_config,omitempty"`
	AuditConf        auditconf.AuditConf `json:"audit_config,omitempty"`
	HTTPS            bool                `json:"https_enabled,omitempty"`
	TLSCertPath      string              `json:"tls_cert_path,omitempty"`
	TLSKeyPath       string              `json:"tls_key_path,omitempty"`
	ConfigPath       string              `json:"config_path,omitempty"`
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
	Level zerolog.Level `json:"level"`
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
	defaultTLSCert := ""
	defaultTLSKey := ""
	defaultConfigPath := ""

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
	flag.StringVar(&c.ConfigPath, "c", defaultConfigPath, "config file path")

	if err = c.ShortURLTemplate.Set(defaultURL); err != nil {
		log.Fatal().Err(err).Msg("Failed to set default URL")
	}

	if err = c.LogLevel.Set(defaultLogLevel.String()); err != nil {
		log.Fatal().Err(err).Msg("Failed to set default log level")
	}

	// Determine config file path (check env first, then pre-scan args for -c flag)
	configPath := defaultConfigPath
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		configPath = envConfig
	}
	for i, arg := range os.Args[1:] {
		if arg == "-c" && i+1 < len(os.Args)-1 {
			configPath = os.Args[i+2]
			break
		}
		if strings.HasPrefix(arg, "-c=") {
			configPath = strings.TrimPrefix(arg, "-c=")
			break
		}
	}

	// Load JSON config file (lowest priority)
	if err = c.loadConfigFile(configPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config file")
	}

	// Parse flags (medium priority - overrides JSON)
	flag.Parse()

	// Environment variables (highest priority - overrides flags and JSON)
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

	if configPath := os.Getenv("CONFIG"); configPath != "" {
		c.ConfigPath = configPath
	}
}

// loadConfigFile loads configuration from JSON file into the Config struct.
// Returns error only for JSON parsing issues, not for missing files.
func (c *Config) loadConfigFile(path string) error {
	if path == "" {
		return nil
	}

	file, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var fileConfig Config
	if err = json.Unmarshal(file, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply non-zero values from file
	if fileConfig.LocalServerAddr != "" {
		c.LocalServerAddr = fileConfig.LocalServerAddr
	}
	if fileConfig.ShortURLTemplate != "" {
		c.ShortURLTemplate = fileConfig.ShortURLTemplate
	}
	if fileConfig.LogLevel.Level != zerolog.NoLevel {
		c.LogLevel = fileConfig.LogLevel
	}
	if fileConfig.FilePath.FilePath != "" {
		c.FilePath = fileConfig.FilePath
	}
	if fileConfig.PostgresDSN.DatabaseDSN != "" {
		c.PostgresDSN = fileConfig.PostgresDSN
	}
	if fileConfig.AuthConf.SecretKey != "" {
		c.AuthConf.SecretKey = fileConfig.AuthConf.SecretKey
	}
	if fileConfig.AuthConf.TokenExpiration != 0 {
		c.AuthConf.TokenExpiration = fileConfig.AuthConf.TokenExpiration
	}
	if fileConfig.AuditConf.FileConf.AuditFilePath != "" {
		c.AuditConf.FileConf.AuditFilePath = fileConfig.AuditConf.FileConf.AuditFilePath
	}
	if fileConfig.AuditConf.RemoteConf.RemoteServerURL != "" {
		c.AuditConf.RemoteConf.RemoteServerURL = fileConfig.AuditConf.RemoteConf.RemoteServerURL
	}
	if fileConfig.HTTPS {
		c.HTTPS = fileConfig.HTTPS
	}
	if fileConfig.TLSCertPath != "" {
		c.TLSCertPath = fileConfig.TLSCertPath
	}
	if fileConfig.TLSKeyPath != "" {
		c.TLSKeyPath = fileConfig.TLSKeyPath
	}

	return nil
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
