package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"redis-postgres-service/entity"
	"redis-postgres-service/repository/postgres/pgfx"
)

const (
	_createUsersTableQuery = `CREATE TABLE IF NOT EXISTS public.users 
					(
					    id SERIAL PRIMARY KEY,
    			   		name TEXT,  
    			   		age INT
					);`
	_insertUserQuery = `INSERT INTO public.users(name, age) VALUES ($1, $2) RETURNING id`
)

type Repository interface {
	AddUser(ctx context.Context, request *entity.AddUserRequest) (*entity.AddUserResponse, error)
}

// compile time check that repository implements Repository interface
var _ Repository = (*repository)(nil)

// Params is an fx container for all Repository dependencies
type Params struct {
	fx.In

	Postgres pgfx.Postgres
	Logger   *zap.Logger
}

// New is a constructor provided to the fx for creating a Repository
func New(p Params) (Repository, error) {

	tag, err := p.Postgres.Exec(context.Background(), _createUsersTableQuery)
	if err != nil {
		return nil, errors.Errorf("failed to create a user table: %s", err)
	}
	p.Logger.With(zap.String("result", tag.String())).Info("dbpool result")

	return &repository{
		logger:         p.Logger,
		postgresClient: p.Postgres,
	}, nil
}

type repository struct {
	logger         *zap.Logger
	postgresClient pgfx.Postgres
}

// AddUser writes a row to the 'users' table and returns the number of row where data landed.
func (r *repository) AddUser(ctx context.Context, request *entity.AddUserRequest) (*entity.AddUserResponse, error) {
	logger := r.logger.With(zap.String("scope", "repository.adduser"))
	tx, err := r.postgresClient.BeginTx(ctx, pgx.TxOptions{})

	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			logger.
				With(zap.String("request", fmt.Sprintf("%+v", request))).
				Error("transaction rollback")
			tx.Rollback(ctx)
			return
		}
		tx.Commit(ctx)
	}()
	var id int64
	if err = tx.QueryRow(
		ctx,
		_insertUserQuery,
		request.Name,
		request.Age,
	).
		Scan(&id); err != nil {
		return nil, err
	}

	return &entity.AddUserResponse{
		Id: id,
	}, err
}
