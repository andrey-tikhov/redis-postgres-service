package users

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"redis-postgres-service/entity"
	mock_postgres "redis-postgres-service/mocks/repository/postgres"
	"testing"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mock_postgres.NewMockRepository(ctrl)
	c, err := New(Params{
		Repository: repo,
	})
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func Test_controller_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockRepository struct {
		res *entity.AddUserResponse
		err error
	}
	type args struct {
		req *entity.AddUserRequest
	}
	tests := []struct {
		name           string
		args           args
		mockRepository *mockRepository
		want           *entity.AddUserResponse
		assertion      assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			args: args{
				req: &entity.AddUserRequest{
					Name: "Alex",
					Age:  22,
				},
			},
			mockRepository: &mockRepository{
				res: &entity.AddUserResponse{
					Id: 12,
				},
				err: nil,
			},
			want: &entity.AddUserResponse{
				Id: 12,
			},
			assertion: assert.NoError,
		},
		{
			name: "Repo fails",
			args: args{
				req: &entity.AddUserRequest{
					Name: "Alex",
					Age:  22,
				},
			},
			mockRepository: &mockRepository{
				res: nil,
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
			repositoryMock := mock_postgres.NewMockRepository(ctrl)
			if tt.mockRepository != nil {
				repositoryMock.EXPECT().
					AddUser(ctx, gomock.Any()).
					Return(
						tt.mockRepository.res,
						tt.mockRepository.err,
					)
			}
			c := &controller{
				repository: repositoryMock,
			}
			got, err := c.Add(ctx, tt.args.req)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
