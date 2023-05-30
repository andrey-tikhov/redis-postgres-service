package redis

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/config"
	"go.uber.org/fx"
	internalconfig "redis-postgres-service/config"
)

const (
	_configKey  = "redis_config"
	_secretsKey = "redis_secrets"
)

type Repository interface {
	AddIntValueForKey(ctx context.Context, key string, value int64) (int64, error)
}

// compile time check that repository implements Repository interface
var _ Repository = (*repository)(nil)

// Params is an fx container for all Repository dependencies
type Params struct {
	fx.In

	LC             fx.Lifecycle
	ConfigProvider config.Provider
}

// New is a constructor provided to the fx for creating a Repository
func New(p Params) (Repository, error) {
	var cfg internalconfig.RedisConfig
	err := p.ConfigProvider.Get(_configKey).Populate(&cfg)
	if err != nil {
		return nil, errors.Errorf("failed to populate config: %s", err) // unreachable in tests, cause provider is populating from valid yaml.
	}
	var secrets internalconfig.RedisSecrets
	err = p.ConfigProvider.Get(_secretsKey).Populate(&secrets)
	if err != nil {
		return nil, errors.Errorf("failed to populate secrets: %s", err) // unreachable in tests, cause provider is populating from valid yaml.
	}
	fmt.Println(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: secrets.Password,
		DB:       cfg.Database,
	})
	if _, err = client.Ping(context.Background()).Result(); err != nil {
		return nil, errors.Errorf("failed to create a db, is redis running? Err: %s", err)
	}

	p.LC.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return &repository{
		client: client,
	}, nil
}

type repository struct {
	client *redis.Client
}

// AddIntValueForKey adds integer value for the key provided. Amounts stack
func (r *repository) AddIntValueForKey(ctx context.Context, key string, value int64) (int64, error) {
	res, err := r.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, errors.Errorf("redis increment failed: %s", err)
	}
	return res, nil
}
