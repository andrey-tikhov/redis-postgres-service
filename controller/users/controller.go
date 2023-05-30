package users

import (
	"context"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"redis-postgres-service/entity"
	"redis-postgres-service/repository/postgres"
)

type Controller interface {
	Add(ctx context.Context, req *entity.AddUserRequest) (*entity.AddUserResponse, error)
}

// compile time check that controller implements Controller interface
var _ Controller = (*controller)(nil)

// Params is an fx container for all Controller dependencies
type Params struct {
	fx.In

	Repository postgres.Repository
}

// New is a constructor provided to the fx for creating a Controller
func New(p Params) (Controller, error) {
	return &controller{
		repository: p.Repository,
	}, nil
}

type controller struct {
	repository postgres.Repository
}

// Add adds user to the `users` table as the incremental row
func (c *controller) Add(ctx context.Context, req *entity.AddUserRequest) (*entity.AddUserResponse, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return c.repository.AddUser(ctx, req)
}
