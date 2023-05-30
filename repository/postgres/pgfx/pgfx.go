package pgfx

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	internalconfig "redis-postgres-service/config"
)

const (
	_configKey  = "postgres_config"
	_secretsKey = "postgres_secrets"
)

const (
	_postgresURLprefix   = "postgres"
	_maxConnectionsParam = "pool_max_conns"
)

type Params struct {
	fx.In

	LC             fx.Lifecycle
	ConfigProvider config.Provider
	Logger         *zap.Logger
}

type Postgres interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

func New(p Params) (Postgres, error) {
	var cfg internalconfig.PostgresConfig
	err := p.ConfigProvider.Get(_configKey).Populate(&cfg)
	if err != nil {
		return nil, errors.Errorf("failed to populate config: %s", err) // unreachable in tests, cause provider is populating from valid yaml.
	}
	var secrets internalconfig.PostgresSecrets
	err = p.ConfigProvider.Get(_secretsKey).Populate(&secrets)
	if err != nil {
		return nil, errors.Errorf("failed to populate secrets: %s", err) // unreachable in tests, cause provider is populating from valid yaml.
	}

	url := fmt.Sprintf(
		"%s://%s:%s@%s/%s?%s=%d",
		_postgresURLprefix,
		secrets.User,
		secrets.Password,
		cfg.URL,
		cfg.Database,
		_maxConnectionsParam,
		cfg.MaxConnections,
	)
	dbpool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, errors.Errorf("failed to create a db: %s", err)
	}
	p.LC.Append(
		fx.Hook{
			OnStop: func(ctx context.Context) error {
				p.Logger.With(zap.String("scope", "pgfx.go")).Info("Onstop start for postgres executed")
				dbpool.Close()
				return nil
			},
		})
	return dbpool, nil
}
