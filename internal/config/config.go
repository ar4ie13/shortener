package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

type ShortURL string

func (u *ShortURL) String() string {
	return string(*u)
}

// Set validates and sets the flag value
func (u *ShortURL) Set(value string) error {
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

	*u = ShortURL(value)
	return nil
}

// Config struct used for program flag variables
type Config struct {
	LocalServerAddr  string
	ShortURLTemplate ShortURL
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) InitConfig() {
	flag.StringVar(&c.LocalServerAddr, "a", "localhost:8080", "local server address")

	defaultURL := "http://localhost:8080"
	if err := c.ShortURLTemplate.Set(defaultURL); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set default URL: %v\n", err)
		os.Exit(1)
	}

	flag.Var(&c.ShortURLTemplate, "b", "short url template")

}

func (c *Config) GetLocalServerAddr() string {

	return c.LocalServerAddr
}

func (c *Config) GetShortURLTemplate() string {
	return string(c.ShortURLTemplate)
}
