package config

import (
	"flag"
	"os"
)

// Config holds all configuration for the service.
type Config struct {
	Addr   string
	Secret string
}

// Default returns a config with sensible defaults.
func Default() *Config {
	return &Config{
		Addr:   ":7700",
		Secret: "",
	}
}

// Load returns config layered: defaults < env < flags.
func Load() *Config {
	c := Default()

	// Env
	if v := os.Getenv("IPKIT_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("IPKIT_SECRET"); v != "" {
		c.Secret = v
	}

	// Flags
	flag.StringVar(&c.Addr, "addr", c.Addr, "listen address")
	flag.StringVar(&c.Secret, "secret", c.Secret, "auth token secret (auto-generated if empty)")
	flag.Parse()

	return c
}
