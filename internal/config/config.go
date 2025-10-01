package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
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

// InitConfig initialize configuration
func (c *Config) InitConfig() {
	flag.StringVar(&c.LocalServerAddr, "a", "localhost:8080", "local server address")

	defaultURL := "http://localhost:8080"
	if err := c.ShortURLTemplate.Set(defaultURL); err != nil {
		log.Fatalf("Failed to set default URL: %v\n", err)
	}

	flag.Var(&c.ShortURLTemplate, "b", "short url template")

	flag.Parse()

	if serverAddr := os.Getenv("SERVER_ADDRESS"); serverAddr != "" {
		c.LocalServerAddr = serverAddr
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		err := c.ShortURLTemplate.Set(baseURL)
		if err != nil {
			log.Fatalf("Failed to set short URL template: %v\n", err)
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
