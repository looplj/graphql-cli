package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Endpoint struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	URL         string            `yaml:"url,omitempty"`
	SchemaFile  string            `yaml:"schema_file,omitempty"`
	Headers     map[string]string `yaml:"headers,omitempty"`
}

type Config struct {
	Endpoints []Endpoint `yaml:"endpoints"`
}

func DefaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "graphql-cli")
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

func DefaultAuditLogPath() string {
	return filepath.Join(DefaultConfigDir(), "audit.log")
}

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}

		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	// expand env vars in headers
	for i := range cfg.Endpoints {
		for k, v := range cfg.Endpoints[i].Headers {
			cfg.Endpoints[i].Headers[k] = os.ExpandEnv(v)
		}
	}

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	if path == "" {
		path = DefaultConfigPath()
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

func (c *Config) GetEndpoint(name string) (*Endpoint, error) {
	if name == "" {
		return nil, fmt.Errorf("no endpoint specified, use -e to specify one (see 'graphql-cli endpoint list' for available endpoints)")
	}

	for i := range c.Endpoints {
		if c.Endpoints[i].Name == name {
			return &c.Endpoints[i], nil
		}
	}

	return nil, fmt.Errorf("endpoint %q not found", name)
}

func (c *Config) UpdateEndpoint(name string, url *string, description *string, headers map[string]string) error {
	for i := range c.Endpoints {
		if c.Endpoints[i].Name == name {
			if url != nil {
				c.Endpoints[i].URL = *url
			}

			if description != nil {
				c.Endpoints[i].Description = *description
			}

			for k, v := range headers {
				if c.Endpoints[i].Headers == nil {
					c.Endpoints[i].Headers = make(map[string]string)
				}

				c.Endpoints[i].Headers[k] = v
			}

			return nil
		}
	}

	return fmt.Errorf("endpoint %q not found", name)
}

func (c *Config) AddEndpoint(ep Endpoint) error {
	for _, e := range c.Endpoints {
		if e.Name == ep.Name {
			return fmt.Errorf("endpoint %q already exists", ep.Name)
		}
	}

	c.Endpoints = append(c.Endpoints, ep)

	return nil
}
