package redis

import (
	"context"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redismock/v9"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
	"go.uber.org/fx/fxtest"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	s := miniredis.RunT(t)
	srcGood := config.Source(
		strings.NewReader(fmt.Sprintf(
			`{"redis_config":{
						"port": "%s",
						"host": "%s",
						"database": 0},
					"redis_secrets":{"password":"qwerty"}}`,
			s.Port(),
			s.Host(),
		)),
	)
	providerGood, _ := config.NewYAML(srcGood)
	srcBadConfig := config.Source(
		strings.NewReader(
			`{"redis_config":{
						"port": "some port",
						"host": "some host",
						"database": 0},
					"redis_secrets":{"password":"qwerty"}}`,
		),
	)
	providerBadConfig, _ := config.NewYAML(srcBadConfig)

	type args struct {
		provider config.Provider
	}
	tests := []struct {
		name      string
		args      args
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			args: args{
				provider: providerGood,
			},
			assertion: assert.NoError,
		},
		{
			name: "Redis unreachable",
			args: args{
				provider: providerBadConfig,
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testlc := fxtest.NewLifecycle(t)
			got, err := New(Params{
				ConfigProvider: tt.args.provider,
				LC:             testlc,
			})
			tt.assertion(t, err)
			if err == nil {
				assert.NotNil(t, got)
			}
			testlc.RequireStart().RequireStop()
		})
	}
}

func Test_repository_AddIntValueForKey(t *testing.T) {
	type mockRedis struct {
		res int64
		err error
	}
	type args struct {
		key   string
		value int64
	}
	tests := []struct {
		name      string
		args      args
		mockRedis *mockRedis
		want      int64
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			args: args{
				value: 64,
				key:   "some_key",
			},
			mockRedis: &mockRedis{
				res: 64,
				err: nil,
			},
			want:      64,
			assertion: assert.NoError,
		},
		{
			name: "Redis fails",
			args: args{
				value: 64,
				key:   "some_key",
			},
			mockRedis: &mockRedis{
				res: 0,
				err: errors.New("some error"),
			},
			want:      0,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, mock := redismock.NewClientMock()
			if tt.mockRedis != nil {
				expectedCall := mock.ExpectIncrBy(tt.args.key, tt.args.value)
				expectedCall.SetVal(tt.mockRedis.res)
				expectedCall.SetErr(tt.mockRedis.err)
			}
			r := &repository{
				client: client,
			}
			got, err := r.AddIntValueForKey(ctx, tt.args.key, tt.args.value)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
