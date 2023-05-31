package integration_tests

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	uberconfig "go.uber.org/config"
	"go.uber.org/fx"
	"io"
	"net"
	"net/http"
	"redis-postgres-service/controller"
	"redis-postgres-service/gateway/sha512"
	"redis-postgres-service/handler"
	"redis-postgres-service/handler/validation"
	"redis-postgres-service/logging"
	mocksha512 "redis-postgres-service/mocks/gateway/sha512"
	mockpgfx "redis-postgres-service/mocks/repository/postgres/pgfx"
	mockredis "redis-postgres-service/mocks/repository/redis"
	"redis-postgres-service/repository/postgres"
	"redis-postgres-service/repository/postgres/pgfx"
	"redis-postgres-service/repository/redis"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAddUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	expectedResponse1 := []byte(`{"id":1}`)
	expectedResponse2 := []byte(`{"id":2}`)
	t.Run("test for 2 concurrent successful calls to Postgres", func(t *testing.T) {
		NewPostrgesRepo := func() pgfx.Postgres {
			pgfx_mock := mockpgfx.NewMockPostgres(ctrl)
			mockTx := mockpgfx.NewMockTx(ctrl)
			mockRow1 := mockpgfx.NewMockRow(ctrl)
			mockRow2 := mockpgfx.NewMockRow(ctrl)
			pgfx_mock.
				EXPECT().
				Exec(gomock.Any(), gomock.Any(), gomock.Any()).
				MaxTimes(1).
				MinTimes(1).
				Return(pgconn.NewCommandTag("CREATE TABLE"), nil)
			pgfx_mock.
				EXPECT().
				BeginTx(gomock.Any(), gomock.Any()).
				MaxTimes(2).
				MinTimes(2).
				Return(mockTx, nil)
			mockTx.
				EXPECT().
				QueryRow(
					gomock.Any(),
					gomock.Any(),
					"user1",
					26,
				).
				MaxTimes(1).
				MinTimes(1).
				Return(
					mockRow1,
				)
			mockTx.
				EXPECT().
				QueryRow(
					gomock.Any(),
					gomock.Any(),
					"user2",
					26,
				).
				MaxTimes(1).
				MinTimes(1).
				Return(
					mockRow2,
				)
			mockTx.
				EXPECT().
				Commit(gomock.Any()).
				MaxTimes(2).
				MinTimes(2).
				Return(nil)
			mockRow1.
				EXPECT().
				Scan(
					gomock.Any(),
				).
				DoAndReturn(
					func(dest interface{}) error {
						typed := dest.(*int64)
						*typed = 1
						return nil
					})
			mockRow2.
				EXPECT().
				Scan(
					gomock.Any(),
				).
				DoAndReturn(
					func(dest interface{}) error {
						typed := dest.(*int64)
						*typed = 2
						return nil
					})
			return pgfx_mock
		}
		NewSignGateway := func() sha512.Gateway {
			return mocksha512.NewMockGateway(ctrl)
		}
		NewRedisRepo := func() redis.Repository {
			return mockredis.NewMockRepository(ctrl)
		}
		NewMux := func(lc fx.Lifecycle) *http.ServeMux {
			mux := http.NewServeMux()
			server := &http.Server{
				Addr:    "127.0.0.1:8000",
				Handler: mux,
			}
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					ln, err := net.Listen("tcp", server.Addr)
					if err != nil {
						return err
					}
					go server.Serve(ln)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return server.Shutdown(ctx)
				},
			})
			return mux
		}
		NewConfig := func() uberconfig.Provider {
			configOption := uberconfig.Source(strings.NewReader(`{"postgres_repo_config":{"schema":"public"},"handler":{"request_body_limit":1048576}}`))
			provider, _ := uberconfig.NewYAML(configOption)
			return provider
		}
		Register := func(mux *http.ServeMux, h handler.Handler) {
			mux.Handle(
				"/postgres/users",
				validation.HttpPostCheck(
					validation.NotNilRequest(
						http.HandlerFunc(h.AddUser),
					),
				),
			)
		}
		app := fx.New(
			fx.Provide(NewPostrgesRepo),
			fx.Provide(NewSignGateway),
			fx.Provide(NewRedisRepo),
			fx.Provide(NewMux),
			fx.Provide(NewConfig),
			handler.Module,
			controller.Module,
			logging.Module,
			fx.Provide(postgres.New),
			fx.Invoke(Register),
		)
		startCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err := app.Start(startCtx)
		assert.NoError(t, err)
		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(
				"POST",
				"http://localhost:8000/postgres/users",
				strings.NewReader(`{"name":"user1","age":26}`),
			)
			client := http.Client{
				Timeout: time.Second * 10,
			}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			result, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, expectedResponse1, result)
		}()
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(
				"POST",
				"http://localhost:8000/postgres/users",
				strings.NewReader(`{"name":"user2","age":26}`),
			)
			client := http.Client{
				Timeout: time.Second * 10,
			}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			result, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, expectedResponse2, result)
		}()
		wg.Wait()
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = app.Stop(stopCtx)
		assert.NoError(t, err)
	})
}
