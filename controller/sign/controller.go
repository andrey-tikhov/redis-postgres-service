package sign

import (
	"context"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"redis-postgres-service/entity"
	"redis-postgres-service/gateway/sha512"
)

type Controller interface {
	Sign(ctx context.Context, req *entity.SignRequest) (*entity.SignResponse, error)
}

// Params is an fx container for all Controller dependencies
type Params struct {
	fx.In

	Gateway sha512.Gateway
	Logger  *zap.Logger
}

// New is a constructor provided to the fx for creating a Controller
func New(p Params) (Controller, error) {
	return &controller{
		logger:  p.Logger,
		gateway: p.Gateway,
	}, nil
}

// compile time check that controller implements Controller interface
var _ Controller = (*controller)(nil)

type controller struct {
	logger  *zap.Logger
	gateway sha512.Gateway
}

// Sign returns SHA512 signature of text in the request using the key in the request
func (c *controller) Sign(ctx context.Context, req *entity.SignRequest) (*entity.SignResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	s, err := c.gateway.SignHMACSHA512(ctx, req.Key, req.Text)
	if err != nil {
		return nil, errors.Errorf("failed to sha512: %s", err)
	}
	return &entity.SignResponse{
		Hex: s,
	}, nil
}
