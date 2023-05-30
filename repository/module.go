package repository

import (
	"go.uber.org/fx"
	"redis-postgres-service/repository/postgres"
	"redis-postgres-service/repository/postgres/pgfx"

	"redis-postgres-service/repository/redis"
)

var Module = fx.Options(
	pgfx.Module,
	fx.Provide(postgres.New),
	fx.Provide(redis.New),
)
