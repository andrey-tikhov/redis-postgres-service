package sign

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"redis-postgres-service/entity"
	mock_sha512 "redis-postgres-service/mocks/gateway/sha512"
	"testing"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	gw := mock_sha512.NewMockGateway(ctrl)
	c, err := New(Params{
		Logger:  zap.NewNop(),
		Gateway: gw,
	})
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func Test_controller_Sign(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockGateway struct {
		res string
		err error
	}
	type args struct {
		req *entity.SignRequest
	}
	tests := []struct {
		name        string
		args        args
		mockGateway *mockGateway
		want        *entity.SignResponse
		assertion   assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			args: args{
				&entity.SignRequest{
					Text: "123",
					Key:  "key",
				},
			},
			mockGateway: &mockGateway{
				res: "2ea823c645b1baf845ef76096a6d7fa9e568304ba9f7910bd52f01c03eec39cdfeec54e50b86b62ef5bfb9e6ce5c0be747ec13b3a199f9d235e99a36de369a84",
				err: nil,
			},
			want: &entity.SignResponse{
				Hex: "2ea823c645b1baf845ef76096a6d7fa9e568304ba9f7910bd52f01c03eec39cdfeec54e50b86b62ef5bfb9e6ce5c0be747ec13b3a199f9d235e99a36de369a84",
			},
			assertion: assert.NoError,
		},
		{
			name: "Gateway fails",
			args: args{
				&entity.SignRequest{
					Text: "123",
					Key:  "key",
				},
			},
			mockGateway: &mockGateway{
				res: "",
				err: errors.New("some error"),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "Nil request",
			args: args{
				nil,
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			gatewayMock := mock_sha512.NewMockGateway(ctrl)
			if tt.mockGateway != nil {
				gatewayMock.EXPECT().
					SignHMACSHA512(
						ctx,
						gomock.Any(),
						gomock.Any(),
					).
					Return(tt.mockGateway.res, tt.mockGateway.err)
			}
			c := &controller{
				logger:  zap.NewNop(),
				gateway: gatewayMock,
			}
			got, err := c.Sign(ctx, tt.args.req)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
