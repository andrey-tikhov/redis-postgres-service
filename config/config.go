package config

import (
	"go.uber.org/config"
)

func New() (config.Provider, error) {
	configFile := config.File("config/base.yaml")
	secretsFile := config.File("config/secrets.yaml")
	provider, err := config.NewYAML(configFile, secretsFile)
	if err != nil {
		return nil, err
	}
	return provider, nil
}

type HandlerConfig struct {
	RequestBodyLimit int64 `yaml:"request_body_limit"`
}

type PostgresConfig struct {
	URL            string `yaml:"url"`
	MaxConnections int    `yaml:"max_connections"`
	Database       string `yaml:"database"`
}

type PostgresSecrets struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database int    `yaml:"database"`
}

type RedisSecrets struct {
	Password string `yaml:"password"`
}
