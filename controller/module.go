package controller

import (
	"go.uber.org/fx"
	"redis-postgres-service/controller/incremental"
	"redis-postgres-service/controller/sign"
	"redis-postgres-service/controller/users"
)

var Module = fx.Options(
	fx.Provide(incremental.New),
	fx.Provide(users.New),
	fx.Provide(sign.New),
)
