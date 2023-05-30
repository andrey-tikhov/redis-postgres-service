package common

import (
	"github.com/stretchr/testify/assert"
	"redis-postgres-service/entity"
	"testing"
)

func TestBytesToTypeIncrementRequest(t *testing.T) {
	type args struct {
		b []byte
	}
	type testCase[T any] struct {
		name      string
		args      args
		want      *entity.IncrementRequest
		assertion assert.ErrorAssertionFunc
	}
	tests := []testCase[entity.IncrementRequest]{
		{
			name: "Happy path",
			args: args{
				b: []byte(`{"key":"Alex","value":23}`),
			},
			want: &entity.IncrementRequest{
				Key:   "Alex",
				Value: 23,
			},
			assertion: assert.NoError,
		},
		{
			name: "empty body",
			args: args{
				b: nil,
			},
			want:      nil,
			assertion: assert.NoError,
		},
		{
			name: "Happy path",
			args: args{
				b: []byte(`{"key":_egv`),
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BytesToType[entity.IncrementRequest](tt.args.b)
			assert.Equal(t, tt.want, got)
			tt.assertion(t, err)
		})
	}
}

func TestBytesToTypeAddUserRequest(t *testing.T) {
	type args struct {
		b []byte
	}
	type testCase[T any] struct {
		name      string
		args      args
		want      *entity.AddUserRequest
		assertion assert.ErrorAssertionFunc
	}
	tests := []testCase[entity.AddUserRequest]{
		{
			name: "Happy path",
			args: args{
				b: []byte(`{"name":"Alex","age":23}`),
			},
			want: &entity.AddUserRequest{
				Name: "Alex",
				Age:  23,
			},
			assertion: assert.NoError,
		},
		{
			name: "Happy path",
			args: args{
				b: []byte(`{"key":_egv`),
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BytesToType[entity.AddUserRequest](tt.args.b)
			assert.Equal(t, tt.want, got)
			tt.assertion(t, err)
		})
	}
}
