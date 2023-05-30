package gateway

import (
	"go.uber.org/fx"
	"redis-postgres-service/gateway/sha512"
)

var Module = fx.Options(
	fx.Provide(sha512.New),
)
