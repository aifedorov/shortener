package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "url is empty",
			args: args{
				url: "",
			},
			wantErr: ErrURLMissing,
		},
		{
			name: "url is not valid",
			args: args{
				url: "abc",
			},
			wantErr: ErrURLInvalid,
		},
		{
			name: "url is valid",
			args: args{
				url: "https://google.com",
			},
			wantErr: nil,
		},
		{
			name: "complex url is valid",
			args: args{
				url: "https://google.tr.com",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.args.url)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
