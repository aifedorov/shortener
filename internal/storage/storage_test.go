package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemoryStorage_GetURL(t *testing.T) {
	type fields struct {
		pathToURL map[string]string
	}
	type args struct {
		shortURL string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr error
	}{
		{
			name: "Get URL with existing value",
			fields: fields{
				pathToURL: map[string]string{
					"1": "https://google.com",
				},
			},
			args: args{
				shortURL: "1",
			},
			want:    "https://google.com",
			wantErr: nil,
		},
		{
			name: "Get URL with not existing value",
			fields: fields{
				pathToURL: map[string]string{
					"1": "https://google.com",
				},
			},
			args: args{
				shortURL: "2",
			},
			want:    "",
			wantErr: ErrURLNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemoryStorage{
				PathToURL: tt.fields.pathToURL,
			}
			got, err := ms.GetURL(tt.args.shortURL)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemoryStorage_SaveURL(t *testing.T) {
	type fields struct {
		pathToURL map[string]string
	}
	type args struct {
		baseURL   string
		targetURL string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantPrefix string
		want       string
		wantErr    error
	}{
		{
			name: "Save new URL",
			fields: fields{
				pathToURL: make(map[string]string),
			},
			args: args{
				baseURL:   "https://google.com",
				targetURL: "",
			},
			want:    "",
			wantErr: nil,
		},
		{
			name: "Save new URL",
			fields: fields{
				pathToURL: make(map[string]string),
			},
			args: args{
				baseURL:   "https://localhost:80",
				targetURL: "https://google.com",
			},
			wantPrefix: "https://localhost:80/",
			wantErr:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemoryStorage{
				PathToURL: tt.fields.pathToURL,
			}
			got, err := ms.SaveURL(tt.args.baseURL, tt.args.targetURL)
			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantPrefix != "" {
				assert.Contains(t, got, tt.wantPrefix)
			}

			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
