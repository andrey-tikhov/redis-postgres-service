package incremental

import (
	"context"
	"errors"
	"go.uber.org/fx"
	"redis-postgres-service/entity"
	"redis-postgres-service/repository/redis"
)

type Controller interface {
	Inc(ctx context.Context, req *entity.IncrementRequest) (*entity.IncrementResponse, error)
}

// compile time check that controller implements Controller interface
var _ Controller = (*controller)(nil)

// Params is an fx container for all Controller dependencies
type Params struct {
	fx.In

	Repository redis.Repository
}

// New is a constructor provided to the fx for creating a Controller
func New(p Params) (Controller, error) {
	return &controller{
		repository: p.Repository,
	}, nil
}

type controller struct {
	repository redis.Repository
}

// Inc adds value provided in the request to the value stored in the downstream repository under the respective key
// resulting value is returned
func (c *controller) Inc(ctx context.Context, req *entity.IncrementRequest) (*entity.IncrementResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	res, err := c.repository.AddIntValueForKey(ctx, req.Key, req.Value)
	if err != nil {
		return nil, err
	}
	return &entity.IncrementResponse{
		Value: res,
	}, nil
}
