package incremental

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"redis-postgres-service/entity"
	mock_redis "redis-postgres-service/mocks/repository/redis"
	"testing"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mock_redis.NewMockRepository(ctrl)
	c, err := New(Params{
		Repository: repo,
	})
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func Test_controller_Inc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockRepository struct {
		res int64
		err error
	}
	type args struct {
		req *entity.IncrementRequest
	}
	tests := []struct {
		name           string
		args           args
		mockRepository *mockRepository
		want           *entity.IncrementResponse
		assertion      assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			args: args{
				req: &entity.IncrementRequest{
					Key:   "Key",
					Value: 123,
				},
			},
			mockRepository: &mockRepository{
				res: 124,
				err: nil,
			},
			want: &entity.IncrementResponse{
				Value: 124,
			},
			assertion: assert.NoError,
		},
		{
			name: "repo fails",
			args: args{
				req: &entity.IncrementRequest{
					Key:   "Key",
					Value: 123,
				},
			},
			mockRepository: &mockRepository{
				res: 0,
				err: errors.New("some error"),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "nil request",
			args: args{
				req: nil,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := mock_redis.NewMockRepository(ctrl)
			if tt.mockRepository != nil {
				repo.EXPECT().
					AddIntValueForKey(
						ctx,
						gomock.Any(),
						gomock.Any(),
					).
					Return(
						tt.mockRepository.res,
						tt.mockRepository.err,
					)
			}
			c := &controller{
				repository: repo,
			}
			got, err := c.Inc(ctx, tt.args.req)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
