package main

import (
	"go.uber.org/fx"
	"redis-postgres-service/app"
)

func opts() fx.Option {
	return fx.Options(
		app.Module,
	)
}

func main() {
	fx.New(opts()).Run()
}
