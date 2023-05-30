package sha512

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_gateway_SignHMACSHA512512(t *testing.T) {
	type args struct {
		text string
		key  string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		assertion assert.ErrorAssertionFunc
	}{
		{
			name: "Happy path",
			args: args{
				text: "123",
				key:  "key",
			},
			want:      "2ea823c645b1baf845ef76096a6d7fa9e568304ba9f7910bd52f01c03eec39cdfeec54e50b86b62ef5bfb9e6ce5c0be747ec13b3a199f9d235e99a36de369a84",
			assertion: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gateway{}
			got, err := g.SignHMACSHA512(context.Background(), tt.args.text, tt.args.key)
			tt.assertion(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
