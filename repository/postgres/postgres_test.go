package postgres

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"redis-postgres-service/entity"
	mock_pgfx "redis-postgres-service/mocks/repository/postgres/pgfx"
	"testing"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockPostgres struct {
		res pgconn.CommandTag
		err error
	}

	tests := []struct {
		name         string
		mockPostgres *mockPostgres
		assertion    assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			mockPostgres: &mockPostgres{
				res: pgconn.NewCommandTag("some tag"),
				err: nil,
			},
			assertion: assert.NoError,
		},
		{
			name: "Table creation fails",
			mockPostgres: &mockPostgres{
				res: pgconn.NewCommandTag("some tag"),
				err: errors.New("some error"),
			},
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			postgres := mock_pgfx.NewMockPostgres(ctrl)

			postgres.EXPECT().
				Exec(
					ctx,
					gomock.Any(),
					gomock.Any(),
				).
				Return(
					tt.mockPostgres.res,
					tt.mockPostgres.err,
				)
			got, err := New(Params{
				Logger:   zap.NewNop(),
				Postgres: postgres,
			})
			if err == nil {
				assert.NotNil(t, got)
			}
			tt.assertion(t, err)
		})
	}
}

func Test_repository_AddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	type mockPostgresBeginTx struct {
		err error
	}
	type mockTxRollback struct {
		err error
	}
	type mockTxCommit struct {
		err error
	}
	type mockRowScan struct {
		res int64
		err error
	}
	type args struct {
		request *entity.AddUserRequest
	}
	tests := []struct {
		name                string
		args                args
		mockPostgresBeginTx *mockPostgresBeginTx
		mockTxRollback      *mockTxRollback
		mockTxCommit        *mockTxCommit
		mockRowScan         *mockRowScan
		want                *entity.AddUserResponse
		assertion           assert.ErrorAssertionFunc
	}{
		{
			name: "Happy Path",
			args: args{
				request: &entity.AddUserRequest{
					Name: "Name",
					Age:  23,
				},
			},
			mockPostgresBeginTx: &mockPostgresBeginTx{
				err: nil,
			},
			mockTxCommit: &mockTxCommit{
				err: nil,
			},
			mockRowScan: &mockRowScan{
				res: 1,
				err: nil,
			},
			want: &entity.AddUserResponse{
				Id: 1,
			},
			assertion: assert.NoError,
		},
		{
			name: "Begin Tx failed",
			args: args{
				request: &entity.AddUserRequest{
					Name: "Name",
					Age:  23,
				},
			},
			mockPostgresBeginTx: &mockPostgresBeginTx{
				err: errors.New("some error"),
			},
			want:      nil,
			assertion: assert.Error,
		},
		{
			name: "Row scan fails",
			args: args{
				request: &entity.AddUserRequest{
					Name: "Name",
					Age:  23,
				},
			},
			mockPostgresBeginTx: &mockPostgresBeginTx{
				err: nil,
			},
			mockTxRollback: &mockTxRollback{
				err: nil,
			},
			mockRowScan: &mockRowScan{
				res: 0,
				err: errors.New("some error"),
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockPostgres := mock_pgfx.NewMockPostgres(ctrl)
			mockTx := mock_pgfx.NewMockTx(ctrl)
			mockRow := mock_pgfx.NewMockRow(ctrl)

			mockPostgres.EXPECT().
				BeginTx(
					ctx,
					gomock.Any(),
				).
				Return(
					mockTx,
					tt.mockPostgresBeginTx.err,
				)
			if tt.mockPostgresBeginTx.err == nil {
				mockTx.EXPECT().
					QueryRow(
						ctx,
						gomock.Any(),
						gomock.Any(),
					).
					Return(
						mockRow,
					)
			}
			if tt.mockRowScan != nil {
				mockRow.EXPECT().
					Scan(
						gomock.Any(),
					).
					DoAndReturn(
						func(dest interface{}) error {
							if tt.mockRowScan.err != nil {
								return tt.mockRowScan.err
							}
							typed := dest.(*int64)
							*typed = tt.mockRowScan.res
							return nil
						})
			}
			if tt.mockTxRollback != nil {
				mockTx.EXPECT().
					Rollback(ctx).
					Return(tt.mockTxRollback.err)
			}
			if tt.mockTxCommit != nil {
				mockTx.EXPECT().
					Commit(ctx).
					Return(tt.mockTxCommit.err)
			}

			r := &repository{
				logger:         zap.NewNop(),
				postgresClient: mockPostgres,
			}
			got, err := r.AddUser(ctx, tt.args.request)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
