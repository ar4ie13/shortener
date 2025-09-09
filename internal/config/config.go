package config

import (
	"flag"
	"fmt"
	"log"
	"net/url"
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
	return &Config{}
}

// String return short URL in string format
func (u *ShortURLTemplate) String() string {
	return string(*u)
}

// Set validates and sets the flag value
func (u *ShortURLTemplate) Set(value string) error {
	// Check if the value is empty
	if value == "" {
		return fmt.Errorf("URL template cannot be empty")
	}

	// Validate the URL format
	parsedURL, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// Ensure the scheme is http or https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("URL template must use http or https scheme")
	}

	// Ensure the host is not empty
	if parsedURL.Host == "" {
		return fmt.Errorf("URL template must include a host")
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
}

// GetLocalServerAddr returns localserver address string
func (c *Config) GetLocalServerAddr() string {

	return c.LocalServerAddr
}

// GetShortURLTemplate returns Short URL template string
func (c *Config) GetShortURLTemplate() string {
	return string(c.ShortURLTemplate)
}
