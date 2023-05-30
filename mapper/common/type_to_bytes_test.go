package common

import (
	"github.com/stretchr/testify/assert"
	"redis-postgres-service/entity"
	"testing"
)

func TestTypeToBytesIncrementRequest(t *testing.T) {
	type args[T any] struct {
		t *T
	}
	type testCase[T any] struct {
		name      string
		args      args[T]
		want      []byte
		assertion assert.ErrorAssertionFunc
	}
	tests := []testCase[entity.IncrementRequest]{
		{
			name: "Happy path",
			args: args[entity.IncrementRequest]{
				t: &entity.IncrementRequest{
					Key:   "Alex",
					Value: 24,
				},
			},
			want:      []byte(`{"key":"Alex","value":24}`),
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TypeToBytes(tt.args.t)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeToBytesBadType(t *testing.T) {
	ch := make(chan int)
	type args[T any] struct {
		t *T
	}
	type testCase[T any] struct {
		name      string
		args      args[T]
		want      []byte
		assertion assert.ErrorAssertionFunc
	}
	tests := []testCase[map[string]chan int]{
		{
			name: "Bad type",
			args: args[map[string]chan int]{
				t: &map[string]chan int{
					"ch": ch,
				},
			},
			want:      nil,
			assertion: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TypeToBytes(tt.args.t)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
