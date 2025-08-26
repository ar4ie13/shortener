package config

import (
	"flag"
)

// Config struct used for program flag variables
type Config struct {
	LocalServerAddr  string
	ShortURLTemplate string
}

func NewConfig() *Config {
	return &Config{}
}

func (c *Config) InitConfig() {
	flag.StringVar(&c.LocalServerAddr, "a", "localhost:8080", "local server address")
	flag.StringVar(&c.ShortURLTemplate, "b", "http://localhost:8080", "short url template")
}

func (c *Config) GetLocalServerAddr() string {
	return c.LocalServerAddr
}

func (c *Config) GetShortURLTemplate() string {
	return c.ShortURLTemplate
}
