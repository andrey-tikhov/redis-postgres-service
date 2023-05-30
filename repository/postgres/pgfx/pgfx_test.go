package pgfx

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"strings"
	"testing"
)

// Execution of this test requires running postgres database with the following params
// database = postgrestest
// user = "user" password = "qwerty" user has rights to login to database
// url:port: "localhost:5432"
func TestNew(t *testing.T) {
	srcGood := config.Source(
		strings.NewReader(
			`{"postgres_config":{
						"url": "localhost:5432",
						"max_connections": 10,
						"database": "postgrestest"
						},
					"postgres_secrets":{"password":"qwerty","user":"user"}}`,
		),
	)
	providerGood, _ := config.NewYAML(srcGood)
	srcPoolCreationFails := config.Source(
		strings.NewReader(
			`{"postgres_config":{
						"url": "_badurl%%%%/",
						"max_connections": -1000,
						"database": "postgrestest"
						},
					"postgres_secrets":{"password":"qwerty","user":"user"}}`,
		),
	)
	providerPoolCreationFails, _ := config.NewYAML(srcPoolCreationFails)
	type args struct {
		p config.Provider
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			args: args{
				p: providerGood,
			},
			assertion: assert.NoError,
		},
		{
			name: "Connection fails. Unable to create a pool",
			args: args{
				p: providerPoolCreationFails,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testlc := fxtest.NewLifecycle(t)
			logger := zap.NewNop()
			r, err := New(Params{
				LC:             testlc,
				ConfigProvider: tt.args.p,
				Logger:         logger,
			})
			tt.assertion(t, err)
			if err == nil {
				assert.NotNil(t, r)
			}
			testlc.RequireStart().RequireStop()
		})
	}
}
