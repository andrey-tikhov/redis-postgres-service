package app

import (
	"context"
	"go.uber.org/fx"
	"net/http"
	"redis-postgres-service/config"
	"redis-postgres-service/controller"
	"redis-postgres-service/gateway"
	"redis-postgres-service/handler"
	"redis-postgres-service/handler/validation"
	"redis-postgres-service/logging"
	"redis-postgres-service/repository"
)

var Module = fx.Options(
	logging.Module,
	handler.Module,
	config.Module,
	controller.Module,
	repository.Module,
	gateway.Module,
	fx.Invoke(StartAndListen),
)

// StartAndListen is a core service function that
// 1. adds validation to the handler endpoints
// 2. creates the server
// 3. adds OnStart fx.Hook that launches server listening
// 4. adds OnStop fx.Hook that executes server shutdown when app is stopped
func StartAndListen(h handler.Handler, lc fx.Lifecycle) {
	mux := http.NewServeMux()
	mux.Handle(
		"/redis/incr",
		validation.HttpPostCheck(
			validation.NotNilRequest(
				http.HandlerFunc(h.Incremental),
			),
		),
	)
	mux.Handle(
		"/sign/hmacsha512",
		validation.HttpPostCheck(
			validation.NotNilRequest(
				http.HandlerFunc(h.Signature),
			),
		),
	)
	mux.Handle(
		"/postgres/users",
		validation.HttpPostCheck(
			validation.NotNilRequest(
				http.HandlerFunc(h.AddUser),
			),
		),
	)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go srv.ListenAndServe()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return srv.Shutdown(ctx)
			},
		})
}
