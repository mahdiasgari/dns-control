package config

import (
	"fmt"
	"os"

	"github.com/wraplink/dns-control/internal/domain"
	"gopkg.in/yaml.v3"
)

type Config struct {
	DNS   DNSConfig   `yaml:"dns"`
	API   APIConfig   `yaml:"api"`
	Agent AgentConfig `yaml:"agent"`

	Domains []domain.Rule `yaml:"domains"`
}

type DNSConfig struct {
	Listen   string   `yaml:"listen"`
	Upstream []string `yaml:"upstream"`
	TTL      uint32   `yaml:"ttl"`
}

type APIConfig struct {
	Listen string `yaml:"listen"`
}

type AgentConfig struct {
	Token string `yaml:"token"`
}

func Load(filename string) (Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf(
			"read config %q: %w",
			filename,
			err,
		)
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf(
			"parse config %q: %w",
			filename,
			err,
		)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if c.DNS.Listen == "" {
		return fmt.Errorf("dns.listen is required")
	}

	for i, upstream := range c.DNS.Upstream {
		if upstream == "" {
			return fmt.Errorf(
				"dns.upstream[%d] cannot be empty",
				i,
			)
		}
	}

	if c.DNS.TTL == 0 {
		return fmt.Errorf("dns.ttl must be greater than zero")
	}

	if c.API.Listen == "" {
		return fmt.Errorf("api.listen is required")
	}

	return nil
}
