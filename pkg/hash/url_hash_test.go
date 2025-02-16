package hash

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrlTokenHash_Hash(t *testing.T) {
	type args struct {
		n int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "given a valid token for mercadolibre.com.ar, expect the correct hash",
			args: args{
				n: 1890951313831759872,
			},
			wantErr: false,
		},
		{
			name: "given a negative token, expect an error",
			args: args{
				n: -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			h := UrlTokenHash{
				logger: logger,
			}
			_, err := h.Hash(tt.args.n)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
