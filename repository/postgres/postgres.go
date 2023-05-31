package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	internalconfig "redis-postgres-service/config"
	"redis-postgres-service/entity"
	"redis-postgres-service/repository/postgres/pgfx"
)

const _configKey = "postgres_repo_config"

const (
	_createUsersTableQuery = `CREATE TABLE IF NOT EXISTS %s.users 
					(
					    id SERIAL PRIMARY KEY,
    			   		name TEXT,  
    			   		age INT
					);`
	_insertUserQuery = `INSERT INTO %s.users(name, age) VALUES ($1, $2) RETURNING id`
)

type Repository interface {
	AddUser(ctx context.Context, request *entity.AddUserRequest) (*entity.AddUserResponse, error)
}

// compile time check that repository implements Repository interface
var _ Repository = (*repository)(nil)

// Params is an fx container for all Repository dependencies
type Params struct {
	fx.In

	Postgres       pgfx.Postgres
	Logger         *zap.Logger
	ConfigProvider config.Provider
}

// New is a constructor provided to the fx for creating a Repository
func New(p Params) (Repository, error) {
	var cfg internalconfig.PostgresRepoConfig
	err := p.ConfigProvider.Get(_configKey).Populate(&cfg)
	if err != nil {
		return nil, errors.Errorf("failed to populate config: %s", err) // unreachable in tests, cause provider is populating from valid yaml.
	}

	query := fmt.Sprintf(_createUsersTableQuery, cfg.Schema)
	tag, err := p.Postgres.Exec(context.Background(), query)
	if err != nil {
		return nil, errors.Errorf("failed to create a user table: %s", err)
	}
	p.Logger.With(zap.String("result", tag.String())).Info("dbpool result")

	return &repository{
		logger:         p.Logger,
		postgresClient: p.Postgres,
		config:         &cfg,
	}, nil
}

type repository struct {
	logger         *zap.Logger
	postgresClient pgfx.Postgres
	config         *internalconfig.PostgresRepoConfig
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
	query := fmt.Sprintf(_insertUserQuery, r.config.Schema)
	var id int64
	if err = tx.QueryRow(
		ctx,
		query,
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
