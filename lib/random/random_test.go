package random

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenRandomString(t *testing.T) {
	type args struct {
		s    string
		size int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "size is zero value",
			args: args{
				s:    "abc",
				size: 0,
			},
			wantErr: true,
		},
		{
			name: "size is negative value",
			args: args{
				s:    "abc",
				size: -1,
			},
			wantErr: true,
		},
		{
			name: "size is valid",
			args: args{
				s:    "abc",
				size: 2,
			},
			wantErr: false,
		},
		{
			name: "size is too big value",
			args: args{
				s:    "abc",
				size: 33,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GenRandomString(tt.args.s, tt.args.size)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}
