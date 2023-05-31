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

// HandlerConfig is a container for the handler configuration
type HandlerConfig struct {
	RequestBodyLimit int64 `yaml:"request_body_limit"`
}

// PgfxConfig is a container for the Postgres interface configuration (implemented by pgxpool.Pool)
type PgfxConfig struct {
	URL            string `yaml:"url"`
	Database       string `yaml:"database"`
	MaxConnections int    `yaml:"max_connections"`
}

// PgfxSecrets is a container for the Postgres interface secrets
type PgfxSecrets struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// PostgresRepoConfig is a container for the postgres repository configuration
type PostgresRepoConfig struct {
	Schema string `yaml:"schema"`
}

// RedisConfig is a container for redis repository configuration
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database int    `yaml:"database"`
}

// RedisSecrets is a container for redis repository secrets
type RedisSecrets struct {
	Password string `yaml:"password"`
}
